CREATE TABLE IF NOT EXISTS logbook_audits (

    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    event_id    UUID,

    trace_id    VARCHAR(225) NOT NULL,

    user_id     UUID,

    service_name VARCHAR(100) NOT NULL,

    endpoint    VARCHAR(500) NOT NULL,

    method      VARCHAR(10) NOT NULL,

    http_status VARCHAR(200),

    http_code   INT,

    kind       VARCHAR(100),

    ip_address  VARCHAR(45),

    request_body JSONB,

    response_body JSONB,

    execution_time_ms INT NOT NULL,

    is_root     BOOLEAN NOT NULL DEFAULT FALSE,

    occurred_at  TIMESTAMPTZ NOT NULL,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()

);

CREATE INDEX IF NOT EXISTS idx_logbook_user_id   ON logbook_audits (user_id);
CREATE INDEX IF NOT EXISTS idx_logbook_endpoint  ON logbook_audits (endpoint);
CREATE INDEX IF NOT EXISTS idx_logbook_created   ON logbook_audits (occurred_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_logbook_trace     ON logbook_audits (trace_id);

