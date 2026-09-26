ALTER TABLE user_tokens DROP CONSTRAINT IF EXISTS user_tokens_user_id_fkey;

ALTER TABLE post_likes DROP CONSTRAINT IF EXISTS post_likes_username_fkey;
ALTER TABLE post_likes
    ADD CONSTRAINT post_likes_username_fkey
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE;

ALTER TABLE comments DROP CONSTRAINT IF EXISTS comments_username_fkey;
ALTER TABLE comments
    ADD CONSTRAINT comments_username_fkey
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE;
