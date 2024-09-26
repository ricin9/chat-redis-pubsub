-- +migrate Up
ALTER TABLE
    room_users
ADD
    COLUMN last_read TIMESTAMP;

UPDATE
    room_users
SET
    last_read = CURRENT_TIMESTAMP;

-- +migrate Down
ALTER TABLE
    room_users DROP COLUMN last_read;