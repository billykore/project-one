CREATE TABLE feature_flags (
    id           BIGSERIAL PRIMARY KEY,
    key          VARCHAR(128) NOT NULL,
    name         VARCHAR(255) NOT NULL,
    purpose      TEXT NOT NULL,
    owner        VARCHAR(255) NOT NULL,
    lifecycle    VARCHAR(16) NOT NULL DEFAULT 'active',
    safe_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX ux_feature_flags_key ON feature_flags (lower(trim(key)));
