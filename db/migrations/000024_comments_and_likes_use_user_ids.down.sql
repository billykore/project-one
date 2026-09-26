ALTER TABLE comments ADD COLUMN username VARCHAR(255);
UPDATE comments AS comments
SET username = users.username
FROM users
WHERE users.id = comments.user_id;
ALTER TABLE comments ALTER COLUMN username SET NOT NULL;
ALTER TABLE comments
    ADD CONSTRAINT comments_username_fkey
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE ON UPDATE CASCADE;
DROP INDEX IF EXISTS idx_comments_user_id;
CREATE INDEX idx_comments_username ON comments(username);
ALTER TABLE comments DROP CONSTRAINT IF EXISTS comments_user_id_fkey;
ALTER TABLE comments DROP COLUMN user_id;

ALTER TABLE post_likes ADD COLUMN username VARCHAR(255);
UPDATE post_likes AS likes
SET username = users.username
FROM users
WHERE users.id = likes.user_id;
ALTER TABLE post_likes ALTER COLUMN username SET NOT NULL;
ALTER TABLE post_likes
    ADD CONSTRAINT post_likes_username_fkey
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE post_likes DROP CONSTRAINT IF EXISTS post_likes_pkey;
DROP INDEX IF EXISTS idx_post_likes_user_id;
ALTER TABLE post_likes DROP CONSTRAINT IF EXISTS post_likes_user_id_fkey;
ALTER TABLE post_likes DROP COLUMN user_id;
ALTER TABLE post_likes ADD PRIMARY KEY (post_id, username);
CREATE INDEX idx_post_likes_username ON post_likes(username);
