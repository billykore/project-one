CREATE TABLE feature_flag_overrides (
    id          BIGSERIAL PRIMARY KEY,
    flag_id     BIGINT NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    environment VARCHAR(16) NOT NULL,
    username    VARCHAR(255) NOT NULL,
    type        VARCHAR(8) NOT NULL CHECK (type IN ('include', 'exclude')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (flag_id, environment, username)
);
