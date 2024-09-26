-- +migrate Up
ALTER TABLE users ADD COLUMN last_online TIMESTAMP;

UPDATE users SET last_online = created_at;

-- +migrate Down
DROP COLUMN last_online;