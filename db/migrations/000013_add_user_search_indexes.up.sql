-- Enable trigram extension for fuzzy text search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Prefix index for LIKE 'query%' patterns (optimizes leading-wildcard-free prefix matching)
CREATE INDEX idx_users_username_prefix ON users USING btree (username varchar_pattern_ops);

-- Trigram GIN index for similarity-based fuzzy matching
CREATE INDEX idx_users_username_trgm ON users USING gin (username gin_trgm_ops);
