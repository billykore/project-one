ALTER TABLE follows ADD COLUMN follower_username VARCHAR(255);
ALTER TABLE follows ADD COLUMN followed_username VARCHAR(255);

UPDATE follows
SET follower_username = users.username
FROM users
WHERE users.id = follows.follower_id;

UPDATE follows
SET followed_username = users.username
FROM users
WHERE users.id = follows.followed_id;

ALTER TABLE follows ALTER COLUMN follower_username SET NOT NULL;
ALTER TABLE follows ALTER COLUMN followed_username SET NOT NULL;
