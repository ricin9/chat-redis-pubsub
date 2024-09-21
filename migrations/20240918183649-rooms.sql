-- +migrate Up
CREATE TABLE rooms (
    room_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO
    rooms (name)
VALUES
    ('General');

CREATE TABLE room_users (
    room_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    FOREIGN KEY (room_id) REFERENCES rooms(room_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    PRIMARY KEY (room_id, user_id)
);

INSERT INTO
    room_users (room_id, user_id)
SELECT
    1,
    user_id
from
    users;

CREATE TABLE messages (
    message_id INTEGER PRIMARY KEY AUTOINCREMENT,
    room_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    reply_to INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (room_id) REFERENCES rooms(room_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (reply_to) REFERENCES messages(message_id)
);

-- +migrate Down
DROP TABLE messages;

DROP TABLE room_users;

DROP TABLE rooms;

DROP TRIGGER insert_new_users_to_general_room;