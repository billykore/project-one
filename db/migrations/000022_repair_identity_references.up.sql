-- Usernames are mutable profile data. Existing child rows still reference the
-- legacy username key, so make those constraints safe while the application
-- progressively moves all cross-context relations to user_id.
ALTER TABLE comments DROP CONSTRAINT IF EXISTS comments_username_fkey;
ALTER TABLE comments
    ADD CONSTRAINT comments_username_fkey
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE post_likes DROP CONSTRAINT IF EXISTS post_likes_username_fkey;
ALTER TABLE post_likes
    ADD CONSTRAINT post_likes_username_fkey
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE ON UPDATE CASCADE;

-- Backfill the historical session rows whose user_id was not populated by the
-- token repository, then enforce the ownership relation.
UPDATE user_tokens AS tokens
SET user_id = users.id
FROM users
WHERE tokens.username = users.username;

DELETE FROM user_tokens AS tokens
WHERE NOT EXISTS (SELECT 1 FROM users WHERE users.id = tokens.user_id);

ALTER TABLE user_tokens DROP CONSTRAINT IF EXISTS user_tokens_user_id_fkey;
ALTER TABLE user_tokens
    ADD CONSTRAINT user_tokens_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
