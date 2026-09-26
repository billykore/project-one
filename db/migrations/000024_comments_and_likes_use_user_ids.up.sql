-- Comments and likes are ownership relationships. Store their immutable user
-- identifiers and keep usernames only as display data resolved at read time.
ALTER TABLE comments ADD COLUMN user_id INTEGER;
UPDATE comments AS comments
SET user_id = users.id
FROM users
WHERE users.username = comments.username;
ALTER TABLE comments ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE comments
    ADD CONSTRAINT comments_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
DROP INDEX IF EXISTS idx_comments_username;
CREATE INDEX idx_comments_user_id ON comments(user_id);
ALTER TABLE comments DROP CONSTRAINT IF EXISTS comments_username_fkey;
ALTER TABLE comments DROP COLUMN username;

ALTER TABLE post_likes ADD COLUMN user_id INTEGER;
UPDATE post_likes AS likes
SET user_id = users.id
FROM users
WHERE users.username = likes.username;
ALTER TABLE post_likes ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE post_likes
    ADD CONSTRAINT post_likes_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE post_likes DROP CONSTRAINT IF EXISTS post_likes_pkey;
DROP INDEX IF EXISTS idx_post_likes_username;
ALTER TABLE post_likes DROP CONSTRAINT IF EXISTS post_likes_username_fkey;
ALTER TABLE post_likes DROP COLUMN username;
ALTER TABLE post_likes ADD PRIMARY KEY (post_id, user_id);
CREATE INDEX idx_post_likes_user_id ON post_likes(user_id);
