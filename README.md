# Library Management System — Microservices

**A library management system built on microservices with Go, Keycloak, Kafka, gRPC, and Apache APISIX.**

*A personal portfolio and learning project — designed, written, and run end to end by myself.*

---

## Abstract

This project is a library management system decomposed into six independent microservices — *auth*,
*user management*, *book management*, *borrowing*, *notification*, and *activity logbook* — built as a
**portfolio piece and a deliberate learning exercise** rather than a commercial product. The goal was never
to accumulate features, but to confront the problems that actually define distributed systems, framed around
a domain simple enough to reason about and real enough to hurt: how to stop a single copy of a book from
being borrowed by two people at once, how to guarantee that an event delivered twice is not processed twice,
and how to estimate how long a borrowing queue will take to clear. Every technical decision therefore has a
reason behind it and can be defended: synchronous service-to-service calls go over **gRPC** while
asynchronous events flow through **Kafka**; stock consistency is protected in two layers, by a **Redis
distributed lock** and by a database-level unique constraint; identity is not reinvented but delegated to
**Keycloak** with JWKS validation; rate limiting and capacity protection live at the edge in **Apache
APISIX** as a single gateway; persistence is polyglot (MySQL for transactional data, PostgreSQL for audit
trails, Elasticsearch for search, Redis for caching and locking); and waiting-list wait times are estimated
with an **M/M/c queueing model and the Erlang C formula**. In short, the value of this project lies in its
architectural reasoning, not in its endpoint count.

---

## Functional Requirements Analysis

### Language & Runtime

| Concern | Choice | Notes |
| --- | --- | --- |
| Language | **Go 1.25** | One module per service, no shared monorepo module |
| HTTP framework | **Gin 1.12** | Layered delivery → usecase → repository structure |
| ORM / data access | **GORM 1.31** (+ `go-sql-driver/mysql`, `pgx/v5`) | Raw SQL where transactional control matters |
| Validation | **go-playground/validator v10** | Request-level DTO validation |
| Schema migration | **golang-migrate** | 24 versioned `.up/.down.sql` files; no `AutoMigrate` |
| Configuration | **godotenv** | `.env.shared` for common values, `.env.docker` per service |
| Scheduling | **robfig/cron v3** | Expired waiting-list sweeps |

### Databases & Storage

Persistence is intentionally **polyglot** — each store was picked for the access pattern it is good at,
and every service owns its own schema (no shared database between services).

| Store | Version | Used by | Why this store |
| --- | --- | --- | --- |
| **MySQL** | 9.5 | Book Management, Borrowing | Transactional writes; unique constraints are used as a hard concurrency guard |
| **PostgreSQL** | 15 | User Management, Notification | Richer typing and constraint support for profile and delivery-log data |
| **TimescaleDB** | 2.28 (on PG 15) | Activity Logbook | Audit records are append-only time-series data, stored as a hypertable |
| **Redis** | 8.6 | All services | Cache-aside layer with standardized key prefixes, **and** the distributed lock backend |
| **Elasticsearch** | 8.12 | Book Management | Full-text and fuzzy book search, which a relational `LIKE` cannot serve well |

### Inter-Service Communication

Two transports, each chosen for a different reason — not one protocol forced onto every problem.

- **gRPC 1.83 + Protocol Buffers** — synchronous calls where the caller genuinely cannot proceed without an
  answer: borrowing queries *user status* (`:50051`) and *book stock* (`:50052`) before accepting a request.
- **Apache Kafka 3.7 (KRaft mode) + `segmentio/kafka-go`** — asynchronous, fire-and-forget domain events
  (registration, borrowing, email, audit). Consumers are **idempotent by `event_id`** and retry with an
  **exponential backoff** strategy, so redelivery is safe and a transient failure does not drop a message.

### Internet Technologies

| Technology | Version | Role in the system |
| --- | --- | --- |
| **Keycloak** | 24.0.2 | OIDC identity provider (`library-realm`). Authentication is delegated, not reimplemented; services verify JWTs locally against the realm's **JWKS** endpoint via `MicahParks/keyfunc`, plus a role-based middleware for authorization |
| **Apache APISIX** | 3.17 | The single ingress (`:9080` / `:9443`) in standalone YAML mode. No application service publishes a host port — **all external traffic must pass the gateway**. Handles global CORS, layered `limit-count` rate limiting, a uniform JSON 429 body via `response-rewrite`, a system-wide capacity guard returning 503, and Prometheus metrics |
| **docker-mailserver** | 16.0.1 | Local SMTP endpoint for email verification during development, driven by Go's `net/smtp` with `PlainAuth` |
| **Open Library API** | — | External book metadata enrichment: a book created by ISBN is auto-populated from `openlibrary.org` |
| **Kafka UI** | latest | Browser-based topic, consumer-group, and message inspection |

### Algorithmic Methods

1. **Queueing theory — M/M/c model with the Erlang C formula.**
   The core algorithmic contribution. Copies of a book are modeled as `c` parallel servers, borrow requests
   as a Poisson arrival process `λ`, and returns as an exponential service rate `μ`. From these,
   [`erlangc.go`](services/borrowing-management-services/internal/helper/erlangc.go) derives utilization `ρ`,
   the probability of waiting `P(W > 0)`, mean queue length `Lq`, expected wait `Wq` in minutes, a saturation
   flag for `ρ ≥ 1`, and a recommended optimal copy count `c*` that keeps utilization below a healthy 0.8.
   This is what lets the waiting list answer *"roughly how long until my turn?"* instead of only *"you are
   number 7"*.

2. **Distributed mutual exclusion.** A Redis `SETNX` lock with a unique owner token and a TTL guards the
   borrow path, so only one request may decrement a given book's stock at a time — and the lock can only be
   released by the holder that acquired it.

3. **Two-layer concurrency guard (defense in depth).** The Redis lock is treated as an optimization, never as
   the source of truth. A unique constraint on the `active_slot` column is the final authority, with MySQL
   duplicate-entry errors detected and translated into a clean domain conflict. If Redis is unavailable or a
   lock expires early, correctness still holds.

4. **Cache-aside caching** with standardized key prefixes, and token-derived TTLs so cached authorization
   state can never outlive the token it was derived from.

---

## Services

Six services, each with its own database and its own lifecycle. Ports below are **internal** to the Docker
network — clients reach everything through the APISIX gateway on `:9080`.

| # | Service | Port | Data Store | Responsibility |
| --- | --- | --- | --- | --- |
| 1 | **Auth** | `9091` | — (Keycloak is the source of truth) | Register, login, logout, token refresh, email verification and resend, current-user lookup |
| 2 | **User Management** | `9190` HTTP · `50051` gRPC | PostgreSQL · Redis | User profiles, password changes, role-based admin operations; serves **user status** over gRPC |
| 3 | **Book Management** | `9290` HTTP · `50052` gRPC | MySQL · Redis · Elasticsearch | Book CRUD, full-text search, ISBN enrichment from Open Library; serves **book stock** over gRPC |
| 4 | **Borrowing** | `9390` | MySQL · Redis | Borrow, return, cancel, status transitions, waiting list with Erlang C wait estimation, cron-driven expiry handling. The orchestrator of the system |
| 5 | **Notification** | `9490` | PostgreSQL · Redis | Consumes email events and dispatches mail via SMTP, with a persistent delivery log |
| 6 | **Activity Logbook** | `9590` | TimescaleDB · Redis | Consumes audit events into an append-only trail, queryable by `trace_id` |

### A shared layout

Every service follows the same layered structure, so moving between them requires no relearning.
Dependencies point inward — `delivery` knows `usecase`, `usecase` knows `repository`, and never the reverse:

```
cmd/app/            Entry point — main.go only
internal/
  bootstrap/        Boots the app: config → infrastructure → registry → router
  config/           Environment variables parsed into typed structs
  delivery/         Inbound edge: controller, middleware, router (+ grpc where applicable)
  dto/              Request and response contracts
  helper/           Cross-cutting utilities: errors, responses, trace IDs, audit redaction
  infrastructure/   Outbound edge: database, redis, kafka, and external systems
  models/           Database entities
  registry/         Dependency wiring (manual DI, no framework)
  repository/       Data access
  usecase/          Business logic
```

What differs between services is *which* `infrastructure/` adapters exist — and that difference is
exactly the architecture.

---

### 1. Auth Service

> Delegates identity to Keycloak and owns the full account lifecycle: registration, login, logout,
> token refresh, and email verification.

**Why it has no database:** Keycloak is the single source of truth for user credentials. This service
holds no user table at all — only Redis, used as a token blacklist so a logout takes effect immediately
instead of waiting for the token to expire on its own.

| Internet technology | How this service uses it |
| --- | --- |
| **Keycloak** | Creates users through the Admin API, exchanges credentials via the password grant, and refreshes tokens |
| **Apache APISIX** | Sits behind the strictest rate limits in the system: 5 logins, 3 registrations, and 2 resend-email requests per window |
| **Kafka** (producer) | Publishes `USER_CREATED` and `RESEND_VERIFICATION_EMAIL` so email delivery never blocks the HTTP response |

```
auth-services/
├── Dockerfile
├── cmd/app/main.go
└── internal/
    ├── bootstrap/index.go
    ├── config/config.go
    ├── delivery/
    │   ├── controller/auth_controller.go
    │   ├── middleware/
    │   │   ├── auth_middleware.go              JWT validation against Keycloak JWKS
    │   │   └── audit_middleware.go             Emits audit events to the logbook
    │   └── router/
    │       ├── auth_router.go
    │       └── initrouter/initial_router.go
    ├── dto/auth_dto.go
    ├── helper/
    │   ├── token.go                            Role extraction + token TTL calculation
    │   ├── verification_link.go                Signs the HS512 email-verification link
    │   ├── redact.go                           Strips passwords before anything is logged
    │   ├── audit_body.go, trace.go             Trace ID propagation
    │   └── error.go, response_global.go, depedencies.go
    ├── infrastructure/
    │   ├── keycloak/keycloak.go                Admin API + token endpoint client
    │   ├── kafka/{producer,event}/
    │   └── redis/redis.go                      Token blacklist
    ├── registry/
    ├── repository/auth_repository.go
    └── usecase/auth_usecase.go
```

---

### 2. User Management Service

> Owns the application-side user profile that Keycloak does not store, and answers *"is this user
> eligible to borrow?"* for other services over gRPC.

| Internet technology | How this service uses it |
| --- | --- |
| **Keycloak** | Reads and updates user attributes and password changes through the Admin API |
| **gRPC** (server, `:50051`) | Exposes user status so Borrowing can check eligibility synchronously |
| **Kafka** (consumer + producer) | Consumes `USER_CREATED` to build the local profile; publishes `USER_STATUS_UPDATED` |
| **Apache APISIX** | Separates sensitive routes (password, profile: 5/min) from admin writes (15/min) and reads (60/min) |

```
user-management-services/
├── Dockerfile
├── cmd/app/main.go
└── internal/
    ├── bootstrap/index.go
    ├── config/config.go
    ├── database/
    │   ├── postgres.go
    │   └── migration/                          ⚠ no infrastructure/ prefix here
    ├── delivery/
    │   ├── controller/{user,admin}_controller.go
    │   ├── grpc/
    │   │   ├── grpc_server.go
    │   │   └── grpc_status_handler.go          Serves user status to Borrowing
    │   ├── middleware/{auth,role,audit}_middleware.go
    │   └── router/{user,admin}_router.go       /user/my/* vs /user/* (SUPER_ADMIN)
    ├── dto/user_dto.go
    ├── helper/
    │   ├── validate_phone_number.go
    │   ├── generate_card.go                    Library card number generation
    │   ├── role.go, date_formatter.go, redact.go, trace.go
    │   └── error.go, response_global.go, audit_body.go
    ├── infrastructure/
    │   ├── keycloak/user_keycloak.go
    │   └── kafka/{consumer,producer,event}/
    ├── models/user.go
    ├── proto/user/                             Generated gRPC stubs
    ├── registry/
    ├── repository/user_repository.go
    └── usecase/{user,admin}_usecase.go
```

---

### 3. Book Management Service

> Manages the catalogue — CRUD, stock, and search — and enriches a book automatically from its ISBN
> instead of asking the librarian to type everything in.

| Internet technology | How this service uses it |
| --- | --- |
| **Open Library API** | `FetchByISBN` pulls title, authors, publisher, page count, and cover from `openlibrary.org`; only ISBN and stock are entered by hand |
| **Elasticsearch** | Mirrors the catalogue into an index so search is full-text and typo-tolerant, not a relational `LIKE` |
| **gRPC** (server, `:50052`) | Exposes live stock so Borrowing can validate availability before it commits |
| **Kafka** (consumer + producer) | Consumes `BORROWING_CREATED` to adjust available stock asynchronously |
| **Apache APISIX** | Allows the most generous read limit in the system (120/min) — browsing is cheap, writing is not (15/min) |

```
book-management-services/
├── Dockerfile
├── cmd/app/main.go
└── internal/
    ├── bootstrap/index.go
    ├── client/open_library.go                  External Open Library HTTP client
    ├── config/conifg.go
    ├── delivery/
    │   ├── controller/book_controller.go
    │   ├── grpc/
    │   │   ├── grpc_server.go
    │   │   └── book_stock_handler.go           Serves stock to Borrowing
    │   ├── middleware/{auth,role,audit}_middleware.go
    │   └── router/book_router.go
    ├── dto/book_dto.go
    ├── helper/
    │   ├── category_resolver.go                Maps Open Library subjects to local categories
    │   └── date_formatter.go, error.go, response.go, trace.go, audit_body.go
    ├── infrastructure/
    │   ├── database/mysql.go + migration/
    │   ├── elasticsearch/elasticsearch.go
    │   ├── kafka/{consumer,producer,event}/
    │   └── redis/redis.go
    ├── models/books.go
    ├── proto/book/                             Generated gRPC stubs
    ├── registry/
    ├── repository/
    │   ├── book_repository.go                  MySQL — source of truth
    │   ├── elasticsearch_repository.go         Search index
    │   └── cache_repository.go                 Redis cache-aside
    └── usecase/book_usecase.go
```

---

### 4. Borrowing Service

> The orchestrator and the hardest service in the system: it decides whether a loan may happen, holds
> the concurrency guarantees, and runs the Erlang C waiting list.

**Why this one carries the most code:** it is the only service that must be correct under concurrency.
Everything interesting lives here — the Redis lock, the MySQL duplicate-key guard, the queueing model,
and the cron job that promotes waiting users when a copy is returned.

| Internet technology | How this service uses it |
| --- | --- |
| **gRPC** (client) | Calls User Management (`:50051`) for eligibility and Book Management (`:50052`) for stock before accepting a request |
| **Kafka** (producer + consumer) | Publishes `BORROWING_CREATED`; consumes user-status events to keep its local Redis mirror fresh |
| **Redis** | Two roles: cache-aside for reads, **and** the distributed lock that serializes stock mutation |
| **Apache APISIX** | Throttled hardest of all: 1 write per 2 seconds per identity, because this is the endpoint worth abusing |

```
borrowing-management-services/
├── Dockerfile
├── cmd/app/main.go
└── internal/
    ├── bootstrap/index.go
    ├── client/
    │   ├── user_grpc_client.go                 → User Management :50051
    │   └── book_grpc_client.go                 → Book Management :50052
    ├── config/config.go
    ├── delivery/
    │   ├── controller/{borrowing,waiting_list}_controller.go
    │   ├── middleware/{auth,role,audit}_middleware.go
    │   └── router/{borrowing,waiting_list}_router.go
    ├── dto/{borrowing,waiting_list,borrowing_consumer}_dto.go
    ├── helper/
    │   ├── erlangc.go                          ★ M/M/c + Erlang C implementation
    │   ├── detect_mysql_error.go               Turns duplicate-key errors into domain conflicts
    │   ├── borrowing_code.go                   Human-readable loan codes
    │   └── role.go, date_formatter.go, error.go, response.go, trace.go, audit_body.go
    ├── infrastructure/
    │   ├── database/mysql.go + migration/      Includes the active_slot unique constraint
    │   ├── kafka/{producer,consume,event}/
    │   └── redis/
    │       ├── lock.go                         ★ SETNX distributed lock with owner token
    │       └── redis_cache.go
    ├── models/{borrowings,waiting_list}.go
    ├── proto/{user,book}/                      Generated gRPC client stubs
    ├── registry/
    ├── repository/
    │   ├── {borrowing,waiting_list}_repository.go
    │   ├── cache/{borrowing,waiting_list}_cache.go
    │   └── user_redis.go                       Local mirror of user status
    └── usecase/
        ├── borrowing_usecase.go                Lock → gRPC checks → write → publish
        ├── waiting_list_usecase.go             Calls CalculateErlangC for the estimate
        └── cronjob/waiting_list_cron.go        Expires and promotes queue entries
```

---

### 5. Notification Service

> Listens for events and turns them into email, keeping a persistent log of every delivery attempt.

**Why it is event-driven only:** nothing in the system waits for an email to be sent. Registration
returns immediately; this service picks the event up afterwards. Its single HTTP route exists purely
to read back the delivery log.

| Internet technology | How this service uses it |
| --- | --- |
| **SMTP / docker-mailserver** | Sends verification mail with Go's `net/smtp` and `PlainAuth` |
| **Kafka** (consumer) | Two separate consumer groups — one for first-time verification, one for resends — so a resend backlog cannot stall new registrations |
| **Apache APISIX** | Exposes `/api/v1/notifications/log-email` for reads only (60/min) |

```
notification-services/
├── Dockerfile
├── cmd/app/main.go
└── internal/
    ├── bootstrap/index.go
    ├── config/config.go
    ├── delivery/
    │   ├── controller/email_controller.go
    │   ├── middleware/{auth,role,audit}_middleware.go
    │   └── router/email_router.go              Read-only delivery log
    ├── dto/email_dto.go
    ├── helper/
    ├── infrastructure/
    │   ├── database/postgres.go + migration/
    │   ├── kafka/{consumer,producer,event}/    Two consumer groups
    │   └── redis/redis.go
    ├── models/email_logs.go
    ├── registry/
    ├── repository/
    │   ├── email_repository.go
    │   └── cache/email_cache_repository.go
    └── usecase/email_usecase.go                SMTP dispatch + delivery logging
```

---

### 6. Activity Logbook Service

> An append-only audit trail: every mutating request across all six services is recorded and can be
> replayed by `trace_id`.

**Why TimescaleDB and why read-only:** audit records are time-series data — written once, never updated,
queried by time range — so they are stored in a hypertable. The service deliberately exposes **no write
endpoint**: the only way in is through Kafka, which means no caller can forge or amend an audit record
over HTTP. Reads are restricted to `SUPER_ADMIN`, with cursor pagination because the table only grows.

| Internet technology | How this service uses it |
| --- | --- |
| **Kafka** (consumer) | The sole ingress. Consumes `LOGBOOK_CREATED` with `event_id` idempotency and backoff retry, so redelivery cannot duplicate an audit row |
| **Apache APISIX** | Read-only route (60/min), `SUPER_ADMIN` enforced in the service middleware |

```
activity-logbook-services/
├── Dockerfile
├── cmd/app/main.go
└── internal/
    ├── bootstrap/index.go
    ├── config/config.go
    ├── delivery/
    │   ├── controller/logbook_controller.go    Read-only: no create endpoint exists
    │   ├── middleware/{auth,role}_middleware.go
    │   └── router/logbook_router.go            SUPER_ADMIN only
    ├── dto/logbook_dto.go
    ├── helper/cursor.go                        Cursor pagination for an ever-growing table
    ├── infrastructure/
    │   ├── database/postgres.go + migration/   Hypertable + TimescaleDB extension
    │   ├── kafka/consumer/consumer.go          Idempotent by event_id, with backoff
    │   └── redis/redis.go
    ├── models/logbook.go
    ├── registry/
    ├── repository/{logbook,cache}_repository.go
    └── usecase/logbook_usecase.go
```

---

### How a request flows

A borrow request is the clearest illustration of the whole design:

```
Client → APISIX (rate limit, CORS) → Borrowing Service
           ├── JWT verified locally against Keycloak JWKS
           ├── gRPC → User Management  : is this user active and eligible?
           ├── gRPC → Book Management  : is stock available?
           ├── Redis lock acquired     : serialize access to this book
           ├── MySQL write             : unique constraint has the final say
           ├── Redis lock released
           └── Kafka event published ──┬── Notification Service → SMTP
                                       └── Activity Logbook     → audit trail
```

---

## Installation

### Prerequisites

| Requirement | Version | Note |
| --- | --- | --- |
| **Docker Engine + Compose v2** | 24+ | The only hard requirement — every component runs in a container |
| **Git** | any | |
| Free RAM | ~8 GB | Elasticsearch, Kafka, Keycloak, and 5 databases run side by side |
| Go | 1.25+ | *Optional* — only to build or run a service outside Docker |
| golang-migrate CLI | 4.x | *Optional* — `migrate` is already bundled inside every service image |

> **Linux only:** Elasticsearch needs a raised memory-map limit, otherwise its container exits on boot.
> ```bash
> sudo sysctl -w vm.max_map_count=262144
> ```

### Step 1 — Clone the repository

```bash
git clone https://github.com/<your-username>/library-app-microservices-keycloak.git
cd library-app-microservices-keycloak
```

### Step 2 — Create the environment files

Environment files are **not committed** (they hold credentials). Create them before the first boot —
the full key list and templates are in [Configuration](#configuration) below.

```bash
# Root level: infrastructure credentials, exposed ports, Keycloak settings
cp .env.example .env
cp .env.shared.example .env.shared

# Per-service: service-specific settings
for s in auth user-management book-management borrowing-management notification activity-logbook; do
  cp "services/${s}-services/.env.docker.example" "services/${s}-services/.env.docker"
  cp "services/${s}-services/.env.shared.example"  "services/${s}-services/.env.shared"
done
```

Then open `.env` and replace every `CHANGE_ME` placeholder with your own values.

### Step 3 — Start the infrastructure first

Databases, Kafka, Redis, Elasticsearch, and Keycloak must be healthy *before* the Go services boot,
because each service validates its connections at startup.

```bash
docker compose up -d keycloak kafka redis-cache elasticsearch mailserver \
  user-management-db book-management-db borrowing-management-db notification-db logbook-db
```

Wait until every container reports `healthy` — this takes roughly 1–2 minutes on a first run:

```bash
docker compose ps
```

### Step 4 — Configure the Keycloak realm

The realm is **not** auto-imported, so this step is manual and only needed once.

1. Open the admin console at **http://localhost:8081** and sign in with `KEYCLOAK_ADMIN` /
   `KEYCLOAK_ADMIN_PASSWORD` from your `.env`.
2. Create a realm named **`library-realm`** (must match `KEYCLOAK_REALM`).
3. Create a confidential client:
   - *Client ID*: the value you set for `KEYCLOAK_CLIENT_ID`
   - *Client authentication*: **On**
   - *Direct access grants*: **Enabled** (the login endpoint uses the password grant)
   - *Service accounts roles*: **Enabled** (the auth service creates users via the Admin API)
4. Under **Service accounts roles**, assign the `realm-management` client roles `manage-users`,
   `view-users`, and `manage-realm`.
5. Create three **realm roles**, exactly as the middleware expects them:
   ```
   SUPER_ADMIN
   ADMIN
   USER_PUBLIC
   ```
6. Copy the client secret from the client's **Credentials** tab into `KEYCLOAK_CLIENT_SECRET`
   in both `.env` and `.env.shared`.

### Step 5 — Build and start the application services

```bash
docker compose up -d --build
```

### Step 6 — Run the database migrations

Each service image already ships the `migrate` binary, so nothing extra needs installing. Run these
once, in any order:

```bash
# PostgreSQL services
docker compose exec user-management-service \
  migrate -path ./migration -database "postgres://$DB_USER:$DB_PASS@user-management-db:5432/$DB_NAME?sslmode=disable" up

docker compose exec notification-service \
  migrate -path ./migration -database "postgres://$DB_USER:$DB_PASS@notification-db:5432/$DB_NAME?sslmode=disable" up

docker compose exec logbook-service \
  migrate -path ./migration -database "postgres://$DB_USER:$DB_PASS@logbook-db:5432/$DB_NAME?sslmode=disable" up

# MySQL services — multiStatements is required, migrations contain multiple DDL statements
docker compose exec book-management-service \
  migrate -path ./migration -database "mysql://$DB_USER:$DB_PASS@tcp(book-management-db:3306)/$DB_NAME?multiStatements=true" up

docker compose exec borrowing-management-service \
  migrate -path ./migration -database "mysql://$DB_USER:$DB_PASS@tcp(borrowing-management-db:3306)/$DB_NAME?multiStatements=true" up
```

Substitute `$DB_USER`, `$DB_PASS`, and `$DB_NAME` with that service's credentials from `.env`.
Passwords containing special characters must be URL-encoded (`@` → `%40`, `#` → `%23`, `/` → `%2F`).

> See **[MIGRATION.md](MIGRATION.md)** for the complete migration reference — rollbacks, dirty-state
> recovery, running `migrate` from the host instead of a container, and the PostgreSQL/MySQL differences.

### Step 7 — Verify the installation

```bash
# All containers up and healthy
docker compose ps

# The gateway is reachable and routing (401 = APISIX routed correctly, auth simply missing)
curl -i http://localhost:9080/api/v1/book
```

A `401 Unauthorized` here is the **expected** result: it proves APISIX matched the route and the
service rejected an unauthenticated request.

---

## Usage

All client traffic goes through the **APISIX gateway on `http://localhost:9080`**. No application
service publishes a host port, so calling a service directly is not possible by design.

### Service endpoints and dashboards

| Interface | URL | Purpose |
| --- | --- | --- |
| **API Gateway** | `http://localhost:9080` | The single entry point for every API call |
| Keycloak Admin | `http://localhost:8081` | Realm, client, role, and user administration |
| Kafka UI | `http://localhost:2004` | Topics, consumer groups, message inspection |
| Elasticsearch | `http://localhost:9911` | Direct index queries while debugging search |
| Prometheus metrics | `http://localhost:9080/apisix/prometheus/metrics` | Gateway traffic and rate-limit counters |

### 1. Register a user

```bash
curl -X POST http://localhost:9080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "nik": "3201234567890001",
    "username": "rendi",
    "email": "rendi@example.com",
    "firstName": "Rendi",
    "lastName": "Putra",
    "password": "YourStrongPassword123"
  }'
```

This creates the user in Keycloak and publishes a `user.created` event. The Notification service
consumes it and sends a verification email, which you can read in the local mail server.

### 2. Verify the email, then log in

```bash
# The verification link arrives by email; it can also be called directly
curl "http://localhost:9080/api/v1/auth/verify-email?token=<TOKEN_FROM_EMAIL>"

# Resend it if needed (rate-limited to 2 requests per 5 minutes at the gateway)
curl -X POST http://localhost:9080/api/v1/auth/resend-email \
  -H "Content-Type: application/json" \
  -d '{"email":"rendi@example.com"}'
```

```bash
curl -X POST http://localhost:9080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"rendi","password":"YourStrongPassword123"}'
```

Response:

```json
{
  "responseCode": 200,
  "responseMessage": "Login successful",
  "data": {
    "accessToken": "eyJhbGciOiJSUzI1NiIs...",
    "refreshToken": "eyJhbGciOiJIUzUxMiIs...",
    "accessTokenExpiresIn": 300,
    "refreshTokenExpiresIn": 1800,
    "tokenType": "Bearer"
  }
}
```

Keep the token handy for the calls below:

```bash
TOKEN="<paste accessToken here>"
```

### 3. Add a book (requires `ADMIN` or `SUPER_ADMIN`)

Only `isbn`, `totalStock`, and `availableStock` are required — the remaining metadata is fetched
automatically from the **Open Library API** when the ISBN is found.

```bash
curl -X POST http://localhost:9080/api/v1/book \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"isbn":"9780134190440","totalStock":3,"availableStock":3}'
```

### 4. Search the catalogue

```bash
# Full-text search, served by Elasticsearch
curl -G http://localhost:9080/api/v1/book \
  -H "Authorization: Bearer $TOKEN" \
  --data-urlencode "search=clean code" \
  --data-urlencode "page=1" \
  --data-urlencode "limit=10"
```

### 5. Borrow a book (requires `USER_PUBLIC`)

This is the path that exercises the whole system — gRPC eligibility and stock checks, the Redis lock,
the unique-constraint guard, and the Kafka fan-out:

```bash
curl -X POST http://localhost:9080/api/v1/borrowing \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"bookID":"<BOOK_UUID>","dueDate":"2026-11-01T00:00:00Z"}'
```

If the stock is exhausted, the request is rejected with a conflict rather than overselling — that
rejection is the concurrency guard doing its job.

### 6. Join the waiting list — the Erlang C queue estimate

When no copy is available, a user can queue. The response is where the **M/M/c queueing model**
becomes visible:

```bash
curl -X POST http://localhost:9080/api/v1/borrowing/waiting-list \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"bookID":"<BOOK_UUID>"}'
```

```json
{
  "data": {
    "queueNumber": 3,
    "status": "WAITING",
    "arrivalRate": 2.5,
    "serviceRate": 0.8,
    "numServers": 3,
    "utilization": 1.04,
    "probWait": 0.87,
    "avgQueueLen": 7.21,
    "avgWaitMin": 173.04,
    "optimalServers": 4
  }
}
```

Read that as: arrivals outpace returns (`utilization` > 1, the system is saturated), a new arrival
waits with 87% probability for roughly **173 minutes**, and the library would need **4 copies**
instead of 3 to bring utilization back to a healthy level. The queue does not just report a position —
it explains itself.

```bash
# Leave the queue
curl -X PATCH "http://localhost:9080/api/v1/borrowing/waiting-list/<ID>/cancel" \
  -H "Authorization: Bearer $TOKEN"
```

### 7. Return a book (requires `ADMIN` or `SUPER_ADMIN`)

Returning frees a slot, which lets the cron job promote the next person in the waiting list.

```bash
curl -X PATCH "http://localhost:9080/api/v1/borrowing/<BORROWING_ID>/status" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"RETURNED"}'
```

### 8. Inspect the audit trail (requires `SUPER_ADMIN`)

Every mutating request is recorded with a `trace_id`, so one action can be followed across all six
services:

```bash
curl "http://localhost:9080/api/v1/logbook/<TRACE_ID>" \
  -H "Authorization: Bearer $TOKEN"
```

### 9. Observe the rate limiter

Rate limiting is enforced at the edge, not in application code. Firing the login endpoint repeatedly
(limit: 5 per 61 seconds per IP) shows it:

```bash
for i in $(seq 1 8); do
  curl -s -o /dev/null -w "request %{http_code}\n" \
    -X POST http://localhost:9080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"rendi","password":"wrong"}'
done
```

The sixth request onward returns `429` with a uniform JSON body and a `Retry-After` header.

### Everyday operations

```bash
# Follow one service's logs
docker compose logs -f borrowing-management-service

# Rebuild a single service after a code change
docker compose up -d --build book-management-service

# Stop everything, keeping the data
docker compose down

# Stop everything and wipe all volumes (full reset — databases included)
docker compose down -v
```

---

## Configuration

Configuration is layered across three files, loaded with `godotenv`:

| File | Scope | Consumed by |
| --- | --- | --- |
| `.env` (root) | Infrastructure credentials, host-exposed ports, Keycloak admin settings | `docker-compose.yml` |
| `.env.shared` (root) | Values every service shares: Keycloak, Kafka brokers and topics, Redis, gRPC targets | all services |
| `services/<name>/.env.docker` | Settings belonging to one service only | that service |

> **None of these files are committed.** They are listed in `.gitignore` because they contain
> credentials. Every value shown below is a **placeholder** — replace each one before running.

### Root `.env`

```dotenv
# ─── Keycloak ──────────────────────────────────────────────────────────────
KEYCLOAK_PORT_EXPOSE=8081
KEYCLOAK_URL=http://keycloak:8080
KEYCLOAK_REALM=library-realm
KEYCLOAK_ADMIN=CHANGE_ME
KEYCLOAK_ADMIN_PASSWORD=CHANGE_ME
KEYCLOAK_CLIENT_ID=CHANGE_ME
KEYCLOAK_CLIENT_SECRET=CHANGE_ME

# ─── Shared database defaults ──────────────────────────────────────────────
DB_PORT_GLOBAL=5432
MYSQL_PORT_GLOBAL=3306
DB_SSLMODE=disable
DB_TIMEZONE=Asia/Jakarta

# ─── User Management (PostgreSQL) ──────────────────────────────────────────
userManagement_DB_HOST=user-management-db
userManagement_DB_NAME=library_user_management
userManagement_DB_USER=CHANGE_ME
userManagement_DB_PASSWORD=CHANGE_ME
userManagement_DB_PORT_EXPOSE=3333

# ─── Book Management (MySQL) ───────────────────────────────────────────────
bookManagement_MYSQL_HOST=book-management-db
bookManagement_MYSQL_DATABASE=library_book_management
bookManagement_MYSQL_USER=CHANGE_ME
bookManagement_MYSQL_PASSWORD=CHANGE_ME
bookManagement_MYSQL_ROOT_PASSWORD=CHANGE_ME
bookManagement_MYSQL_CHARSET=utf8mb4
bookManagement_MYSQL_COLLATION=utf8mb4_unicode_ci
bookManagement_MYSQL_PORT_EXPOSE=4444

# ─── Borrowing Management (MySQL) ──────────────────────────────────────────
borrowingManagement_MYSQL_HOST=borrowing-management-db
borrowingManagement_MYSQL_DATABASE=library_borrowing_management
borrowingManagement_MYSQL_USER=CHANGE_ME
borrowingManagement_MYSQL_PASSWORD=CHANGE_ME
borrowingManagement_MYSQL_ROOT_PASSWORD=CHANGE_ME
borrowingManagement_MYSQL_CHARSET=utf8mb4
borrowingManagement_MYSQL_COLLATION=utf8mb4_unicode_ci
borrowingManagement_MYSQL_PORT_EXPOSE=5555

# ─── Notification (PostgreSQL) ─────────────────────────────────────────────
notificationService_POSTGRES_HOST=notification-db
notificationService_POSTGRES_NAME=library_notification_service
notificationService_POSTGRES_USER=CHANGE_ME
notificationService_POSTGRES_PASSWORD=CHANGE_ME
notificationService_POSTGRES_PORT_EXPOSE=6666

# ─── Activity Logbook (TimescaleDB) ────────────────────────────────────────
logbookService_POSTGRES_HOST=logbook-db
logbookService_POSTGRES_NAME=library_logbook_service
logbookService_POSTGRES_USER=CHANGE_ME
logbookService_POSTGRES_PASSWORD=CHANGE_ME
logbookService_POSTGRES_PORT_EXPOSE=7777

# ─── Redis ─────────────────────────────────────────────────────────────────
REDIS_HOST=redis-cache
REDIS_PORT=6379
REDIS_PASSWORD=CHANGE_ME
REDIS_DB=0
REDIS_PORT_EXPOSE=6200
REDIS_PERSISTENCE_HOST=redis-cache
REDIS_PERSISTENCE_PASSWORD=CHANGE_ME
REDIS_DB_PERSISTENCE=1
REDIS_PERSISTENCE_PORT_EXPOSE=6201

# ─── Kafka & Elasticsearch ─────────────────────────────────────────────────
KAFKA_PORT_EXPOSE=2003
KAFKA_UI_PORT_EXPOSE=2004
KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS=kafka:9092
ELASTICSEARCH_PORT_EXPOSE=9911

# ─── Mail server ───────────────────────────────────────────────────────────
MAILSERVER_HOST=mailserver
MAILSERVER_DOMAIN=CHANGE_ME
STMP_PORT_EXPOSE=99
SUBMISSION_PORT_EXPOSE=100
RELAY_HOST=CHANGE_ME
RELAY_PORT=587
RELAY_USER=CHANGE_ME
RELAY_PASSWORD=CHANGE_ME

# ─── Service ports (internal) ──────────────────────────────────────────────
AUTH_APP_PORT=9091
USER_MANAGEMENT_PORT=9190
BOOK_MANAGEMENT_PORT=9290
BORROWING_MANAGEMENT_PORT=9390
NOTIFICATION_SERVICE_PORT=9490
LOGBOOK_SERVICE_PORT=9590
USER_MANAGEMENT_GRPC_HOST=user-management-service
USER_MANAGEMENT_GRPC_PORT=50051
```

### Root `.env.shared`

```dotenv
# ─── Keycloak (as the services see it) ─────────────────────────────────────
KEYCLOAK_URL=http://keycloak:8080
KEYCLOAK_REALM=library-realm
KEYCLOAK_ISSUER=http://localhost:8081/realms/library-realm
KEYCLOAK_CLIENT_ID=CHANGE_ME
KEYCLOAK_CLIENT_SECRET=CHANGE_ME

# ─── Kafka ─────────────────────────────────────────────────────────────────
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC_USER_CREATED=USER_CREATED
KAFKA_TOPIC_USER_UPDATED=USER_STATUS_UPDATED
KAFKA_TOPIC_USER_LOGIN=USER_AUTHENTICATED
KAFKA_TOPIC_BBORROWING_CREATED=BORROWING_CREATED
KAFKA_TOPIC_LOGBOOK_CREATED=LOGBOOK_CREATED
KAFKA_TOPIC_RESEND_VERIFICATION_EMAIL=RESEND_VERIFICATION_EMAIL

# ─── Redis ─────────────────────────────────────────────────────────────────
REDIS_HOST=redis-cache
REDIS_PORT=6379
REDIS_PASSWORD=CHANGE_ME
REDIS_DB=0
REDIS_CACHE_TTL=48h
REDIS_PERSISTENCE_HOST=redis-cache
REDIS_PERSISTENCE_PASSWORD=CHANGE_ME
REDIS_DB_PERSISTENCE=1

# ─── gRPC targets ──────────────────────────────────────────────────────────
USER_MANAGEMENT_GRPC_HOST=user-management-service
USER_MANAGEMENT_GRPC_PORT=50051
BOOK_MANAGEMENT_GRPC_HOST=book-management-service
BOOK_MANAGEMENT_GRPC_PORT=50052

# ─── Shared DB defaults ────────────────────────────────────────────────────
DB_PORT_GLOBAL=5432
MYSQL_PORT_GLOBAL=3306
DB_SSLMODE=disable
DB_TIMEZONE=Asia/Jakarta
```

### Per-service `.env.docker`

Each service only declares what belongs to it:

```dotenv
# services/auth-services/.env.docker
AUTH_APP_PORT=9091
BASE_URL=http://localhost:9091
JWT_SECRET=CHANGE_ME            # signs the email-verification token only
JWT_EXPIRATION=11            # verification-link lifetime, in minutes
REDIS_HOST=redis-cache
REDIS_PORT=6379
REDIS_PASSWORD=CHANGE_ME
REDIS_DB=0
```

```dotenv
# services/user-management-services/.env.docker
USER_MANAGEMENT_PORT=9190
KAFKA_CONSUMER_GROUP_USER_MANAGEMENT=USER_MANAGEMENT_GROUP
```

```dotenv
# services/book-management-services/.env.docker
BOOK_MANAGEMENT_PORT=9290
ELASTICSEARCH_URL=http://elasticsearch:9200
OPEN_LIBRARY_API_URL=https://openlibrary.org/api/books
KAFKA_CONSUMER_GROUP_BOOK_MANAGEMENT=BOOK_MANAGEMENT_GROUP
```

```dotenv
# services/borrowing-management-services/.env.docker
BORROWING_MANAGEMENT_PORT=9390
BORROWING_MANAGEMENT_GROUP_ID=borrowing-management-group
REDIS_PERSISTENCE_HOST=redis-cache
REDIS_PERSISTENCE_PORT_EXPOSE=6201
REDIS_PERSISTENCE_PASSWORD=CHANGE_ME
REDIS_DB_PERSISTENCE=1
TZ=Asia/Jakarta
```

```dotenv
# services/notification-services/.env.docker
NOTIFICATION_SERVICE_PORT=9490
SMTP_HOST=mailserver
SMTP_PORT=587
SMTP_USER=CHANGE_ME
SMTP_PASSWORD=CHANGE_ME
KAFKA_CONSUMER_GROUP_NOTIFICATION_SERVICE=NOTIFICATION_SERVICE_GROUP
KAFKA_CONSUMER_GROUP_NOTIFICATION_SERVICE_RESEND=NOTIFICATION_SERVICE_GROUP_RESEND
```

```dotenv
# services/activity-logbook-services/.env.docker
LOGBOOK_SERVICE_PORT=9590
KAFKA_CONSUMER_GROUP_LOGBOOK_SERVICE=ACTIVITY_LOGBOOK_GROUP
```

### Gateway configuration

APISIX runs in **standalone mode** and reads its configuration from YAML rather than etcd, so routes
are version-controlled instead of being set through an admin API:

| File | Contents |
| --- | --- |
| `apisix_conf/config.yaml` | APISIX runtime settings (standalone mode, enabled plugins) |
| `apisix_conf/apisix.yaml` | Global rules and all upstream routes, including every rate-limit tier |

To change a rate limit, edit the relevant `limit-count` block in `apisix_conf/apisix.yaml` and reload:

```bash
docker compose restart apache-apisix-gateway
```

### Configuration notes

- **Two Redis logical databases, one instance.** `REDIS_DB=0` holds the cache; `REDIS_DB_PERSISTENCE=1`
  holds state that must survive a cache flush. Both point at the same container, isolated by index.
- **`REDIS_CACHE_TTL` applies to domain data only** (books, user profiles). Authorization-related entries
  such as the token blacklist do not use it: their TTL is computed from the token's own expiry, so a cached
  entry can never outlive the token it was derived from.
- **`JWT_SECRET` is not used for API authentication.** API tokens are RS256-signed by Keycloak and
  verified through JWKS; this secret only signs the email-verification link.
- **Consumer group names must be unique per service.** Two services sharing a group ID would split a
  topic's partitions between them, and each would silently see only part of the messages.
- **`KEYCLOAK_URL` and `KEYCLOAK_ISSUER` are deliberately different.** Services reach Keycloak over the
  Docker network (`http://keycloak:8080`), while the `iss` claim inside a token reflects the browser-facing
  address (`http://localhost:8081`). Token validation fails if the issuer does not match exactly.

## Limitations

This is a learning project, and the boundaries below are real. Some are deliberate trade-offs that kept
the scope focused on distributed-systems problems; others are simply gaps. They are listed openly because
a reviewer would find them in ten minutes anyway, and because knowing what is missing is part of knowing
the system.

### Testing and automation

- **There are no automated tests.** Not a single `_test.go` file exists in the repository. This is the
  largest gap, and an awkward one: the two most valuable things to test are also the easiest —
  [`erlangc.go`](services/borrowing-management-services/internal/helper/erlangc.go) is a pure function with
  known mathematical expectations, and the Redis lock can be tested against a real Redis instance.
- **There is no CI pipeline.** No `.github/workflows`, no build or lint gate. Nothing stops a commit that
  does not compile.
- **There is no `Makefile` or bootstrap script.** Installation is a sequence of manual commands, which is
  why the [Installation](#installation) section has to be as long as it is.

### The queueing model

- **The Erlang C inputs are hardcoded, not measured.** In
  [`waiting_list_usecase.go`](services/borrowing-management-services/internal/usecase/waiting_list_usecase.go),
  the arrival rate (`λ = 2.0`), service rate (`μ = 1/168`), and server count (`c = 1`) are fixed constants
  rather than derived from borrowing history and the book's actual stock. The formula itself is implemented
  correctly, but every book currently receives the same estimate, and `c = 1` ignores the real number of
  copies. **The math is sound; the inputs are placeholders.** Feeding it real `λ` from historical borrow
  events and real `c` from book stock is the single change that would turn a correct implementation into a
  genuinely useful one.

### Observability

- **Logging is unstructured.** Every service uses the standard library `log` package — over a hundred
  `log.Printf` and `log.Println` calls — so logs cannot be queried by field, level, or request.
- **There is no distributed tracing.** A `trace_id` is generated and persisted in the audit trail, but it
  is not attached to application logs, and no OpenTelemetry exporter is wired up. (The `otel` packages in
  `go.mod` are indirect dependencies of the Elasticsearch client, not an integration.) Debugging a request
  that crosses four services therefore means reading four separate log streams by hand.
- **Metrics stop at the gateway.** APISIX exposes Prometheus metrics for traffic and rate limiting, but no
  service exports application-level metrics, and no Prometheus or Grafana instance is deployed to scrape them.
- **Services expose no health endpoint.** Docker Compose health checks cover the infrastructure containers
  only, and APISIX performs no active upstream health checking — a dead service is discovered when a request
  to it fails.
- **Startup failures are fatal.** Bootstrapping leans heavily on `log.Fatal`, so a dependency that is
  briefly unreachable at boot kills the process. The Compose restart policy papers over this, but a
  backoff-and-retry startup would be the correct fix.

### Security and transport

- **Internal traffic is unencrypted.** gRPC clients connect with `insecure.NewCredentials()`, and the
  gateway reaches services over plain HTTP. Everything relies on the Docker network being trusted, which is
  fine on one machine and not fine anywhere else. Production would need TLS, and ideally mTLS between services.
- **The distributed lock is single-node.** It is a correct single-instance lock — a Lua script guarantees
  only the owner can release it, and acquisition retries with jitter — but it is not a Redlock quorum, and it
  has no TTL-renewal watchdog. If an operation outlived its lock TTL, a second holder could acquire the same
  key. This is survivable **only** because the lock is treated as an optimization and the database unique
  constraint is the actual guarantee; the lock is not load-bearing for correctness.
- **Secrets live in `.env` files.** Appropriate for local development, unacceptable in production, where
  these belong in a secret manager.
- **Email is local-only.** `docker-mailserver` is a development mail sink; no real delivery provider,
  SPF/DKIM setup, or bounce handling is configured.

### Operational onboarding

- **A fresh clone cannot run yet.** `.env`, every `.env.docker`, and `apisix_conf/` are listed in
  `.gitignore`, and no `.env.example` files are committed, so a newcomer has no working configuration to
  start from. The templates in [Configuration](#configuration) exist to close that gap manually.
- **The Keycloak realm is configured by hand.** There is no realm-import JSON, so the realm, client, service
  account roles, and the three realm roles must be created through the admin console on every fresh
  environment — a step that is easy to get subtly wrong.
- **There is no API specification.** No OpenAPI or Swagger document exists; the `curl` examples in
  [Usage](#usage) are the closest thing to a contract.

### Scaling

- **Nothing has been run with more than one replica per service.** The design anticipates it — Kafka consumer
  groups partition work, and stock mutation is guarded independently of process count — but it is untested,
  so calling it horizontally scalable would be a claim, not a fact.
- **MySQL DDL is not transactional.** A migration that fails midway leaves the schema half-applied and the
  version marked dirty, requiring manual cleanup (see [MIGRATION.md](MIGRATION.md)). The practical mitigation
  is one DDL statement per migration file.

### What would come first

If this project continued, the order would be:

1. **Unit tests** for the Erlang C helper and the Redis lock, then a CI workflow to run them.
2. **Real inputs for the queueing model** — `λ` from borrow history, `c` from actual stock.
3. **Structured logging with the `trace_id` attached**, so the audit trail and the logs finally agree.
4. **Committed `.env.example` files and `apisix_conf/`**, plus a Keycloak realm export, so the project can be
   cloned and started by someone else.

## Conclusion
This microservices-based Library Management System constitutes a deliberate synthesis of domain-oriented service decomposition, polyglot persistence, and mathematically formalized resource scheduling. By decomposing core business operations into autonomous runtimes coordinated through gRPC for synchronous state validation and Apache Kafka for asynchronous event propagation, the architecture minimizes inter-service temporal coupling while preserving strict database-per-service isolation. Centralizing edge concerns within Apache APISIX establishes a unified perimeter for traffic governance — tiered rate limiting, a system-wide capacity guard, and uniform error semantics — thereby insulating internal services from volumetric abuse; authorization itself, by contrast, remains deliberately decentralized and is adjudicated within each service against the realm's JWKS. The theoretical contribution is most evident in the adoption of the Erlang C formulation: by modelling available copies as `c` parallel servers within an M/M/c regime, the system departs from conventional ordinal waitlists and elevates queue membership into a predictive utilization model capable of informing inventory capacity planning.

From an operational standpoint, the two-tiered concurrency framework exemplifies a defense-in-depth strategy for state mutation under contention. By subordinating a pessimistic Redis advisory lock — whose ownership semantics are enforced through an atomic Lua release — to a declarative unique constraint on `active_slot` in MySQL, the architecture secures the throughput benefits of application-level serialization without delegating correctness to it. The lock operates as a contention optimizer, whereas the relational invariant remains the authoritative arbiter; premature lock expiry, Redis unavailability, or abrupt process termination consequently degrades performance rather than integrity. This stratification is reinforced by idempotent Kafka consumption keyed on `event_id` with exponential backoff, and by an append-only audit trail whose exclusive ingress is the event stream — a design decision that structurally precludes the forgery or retroactive amendment of audit records through the HTTP surface.

A critical appraisal must nevertheless concede that the system's analytical ambitions presently exceed their empirical grounding. The principal structural weakness is the decoupling of the queueing model from runtime telemetry: the parameters λ, μ, and c are statically configured rather than estimated from observed borrowing behaviour and actual stock, with the consequence that a correctly implemented formulation yields a uniform and therefore non-discriminating estimate across the entire catalogue. This limitation is compounded by the absence of automated test coverage and continuous integration, by observability that terminates at the gateway, and by the circumstance that a `trace_id` is persisted in the audit trail yet never propagated into application logs, rendering cross-service diagnosis an exercise in manual reconstruction; intra-cluster transport likewise remains unencrypted. Strategically, subsequent development should therefore proceed in descending order of epistemic return: binding the queueing parameters to event-derived estimates, instituting unit coverage for the Erlang C computation and the locking primitive behind a continuous-integration gate, establishing structured logging with OpenTelemetry context propagation, and enforcing mutual TLS between services. Resolving these deficiencies would transition the system from a structurally sound architectural prototype into a defensible production-grade distributed platform.