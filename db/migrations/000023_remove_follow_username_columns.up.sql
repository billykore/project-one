-- Follows are a Social Graph relationship. Both ends are stable user IDs;
-- usernames are resolved from Identity only when rendering a read model.
ALTER TABLE follows DROP COLUMN follower_username;
ALTER TABLE follows DROP COLUMN followed_username;
