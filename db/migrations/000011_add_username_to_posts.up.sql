ALTER TABLE posts ADD COLUMN username VARCHAR(255) NOT NULL DEFAULT substr(md5(random()::text), 1, 10);
