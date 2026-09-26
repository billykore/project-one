ALTER TABLE user_tokens ADD COLUMN username VARCHAR(255);
UPDATE user_tokens AS tokens
SET username = users.username
FROM users
WHERE users.id = tokens.user_id;
ALTER TABLE user_tokens ALTER COLUMN username SET NOT NULL;
