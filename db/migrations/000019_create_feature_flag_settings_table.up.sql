CREATE TABLE feature_flag_settings (
    id                 BIGSERIAL PRIMARY KEY,
    flag_id            BIGINT NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    environment        VARCHAR(16) NOT NULL,
    mode               VARCHAR(16) NOT NULL DEFAULT 'disabled_all',
    rollout_percentage INTEGER NOT NULL DEFAULT 0 CHECK (rollout_percentage BETWEEN 0 AND 100),
    revision           INTEGER NOT NULL DEFAULT 1,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (flag_id, environment)
);
