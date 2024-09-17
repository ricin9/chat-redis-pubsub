-- +migrate Up
CREATE TABLE sessions (
    session_id CHARACTER(20) PRIMARY KEY NOT NULL,
    user_id INTEGER NOT NULL,
    ip TEXT,
    user_agent TEXT,
    created_at INTEGER DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expires_at INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- +migrate Down
DROP TABLE sessions;