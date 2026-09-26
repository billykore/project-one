-- Sessions already belong to users by user_id. The username snapshot is
-- mutable profile data and must not be used as a session relationship.
ALTER TABLE user_tokens DROP COLUMN username;
