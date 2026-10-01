CREATE TABLE IF NOT EXISTS logbook_audits (

    id          UUID NOT NULL DEFAULT gen_random_uuid(),

    event_id    UUID NOT NULL,

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

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id, occurred_at),

    UNIQUE (event_id, occurred_at)
);

SELECT create_hypertable('logbook_audits', 'occurred_at', chunk_time_interval => INTERVAL '1 day');

CREATE INDEX IF NOT EXISTS idx_logbook_user_id   ON logbook_audits (user_id);
CREATE INDEX IF NOT EXISTS idx_logbook_endpoint  ON logbook_audits (endpoint);
CREATE INDEX IF NOT EXISTS idx_logbook_created   ON logbook_audits (occurred_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_logbook_trace     ON logbook_audits (trace_id);
CREATE INDEX IF NOT EXISTS idx_logbook_root ON logbook_audits (occurred_at DESC, id DESC) WHERE is_root;


ALTER TABLE logbook_audits SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'service_name',
    timescaledb.compress_orderby   = 'occurred_at DESC, id DESC'
);

SELECT add_compression_policy('logbook_audits', INTERVAL '7 days');
SELECT add_retention_policy('logbook_audits', INTERVAL '365 days');
