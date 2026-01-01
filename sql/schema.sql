CREATE TABLE users (
    username TEXT PRIMARY KEY,
    identity_public_key BLOB NOT NULL,
    exchange_public_key BLOB NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE files (
    id TEXT PRIMARY KEY,
    sender TEXT NOT NULL,
    recipient TEXT NOT NULL,
    file_name TEXT NOT NULL,
    encrypted_key BLOB NOT NULL,
    auto_delete BOOLEAN NOT NULL DEFAULT 0,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(sender) REFERENCES users(username),
    FOREIGN KEY(recipient) REFERENCES users(username)
);

CREATE TABLE sessions (
    token TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(username) REFERENCES users(username)
);

CREATE TABLE challenges (
    username TEXT PRIMARY KEY,
    nonce TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(username) REFERENCES users(username)
);
