ALTER TABLE users ADD COLUMN username VARCHAR(255) UNIQUE NOT NULL DEFAULT substr(md5(random()::text), 1, 10);
