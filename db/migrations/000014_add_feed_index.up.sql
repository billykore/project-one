CREATE INDEX idx_posts_username_created_at_id
ON posts (username, created_at DESC, id DESC)
WHERE deleted_at IS NULL;
