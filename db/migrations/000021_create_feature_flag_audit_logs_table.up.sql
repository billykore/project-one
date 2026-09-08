CREATE TABLE feature_flag_audit_logs (
    id             BIGSERIAL PRIMARY KEY,
    flag_id        BIGINT NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    environment    VARCHAR(16),
    field          VARCHAR(64) NOT NULL,
    previous_value TEXT,
    new_value      TEXT,
    actor          VARCHAR(255) NOT NULL,
    reason         TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_feature_flag_audit_flag_created
    ON feature_flag_audit_logs (flag_id, created_at DESC);
