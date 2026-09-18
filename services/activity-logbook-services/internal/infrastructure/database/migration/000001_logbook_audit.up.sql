CREATE TABLE IF NOT EXISTS logbook_audits (

    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    trace_id    UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),

    user_id     UUID NOT NULL,

    service_name VARCHAR(100) NOT NULL,

    endpoint    VARCHAR(500) NOT NULL,

    method      VARCHAR(10) NOT NULL,

    http_status VARCHAR(200) NOT NULL,

    http_code   INT NOT NULL,

    ip_address  VARCHAR(45),

    request_body JSONB,

    response_body JSONB,

    execution_time_ms INT NOT NULL,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()

);

CREATE INDEX IF NOT EXISTS idx_logbook_user_id   ON logbook_audits (user_id);
CREATE INDEX IF NOT EXISTS idx_logbook_endpoint  ON logbook_audits (endpoint);
CREATE INDEX IF NOT EXISTS idx_logbook_created   ON logbook_audits (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_logbook_trace     ON logbook_audits (trace_id);

