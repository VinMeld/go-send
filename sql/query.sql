-- name: CreateUser :exec
INSERT INTO users (username, identity_public_key, exchange_public_key)
VALUES (?, ?, ?);

-- name: GetUser :one
SELECT * FROM users
WHERE username = ? LIMIT 1;

-- name: CreateFile :exec
INSERT INTO files (id, sender, recipient, file_name, encrypted_key, auto_delete, timestamp)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetFile :one
SELECT * FROM files
WHERE id = ? LIMIT 1;

-- name: ListFiles :many
SELECT * FROM files
WHERE recipient = ?
ORDER BY timestamp DESC;

-- name: DeleteFile :exec
DELETE FROM files
WHERE id = ?;

-- name: CreateSession :exec
INSERT INTO sessions (token, username, expires_at)
VALUES (?, ?, ?);

-- name: GetSession :one
SELECT * FROM sessions
WHERE token = ? LIMIT 1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE token = ?;

-- name: CreateChallenge :exec
INSERT INTO challenges (username, nonce)
VALUES (?, ?)
ON CONFLICT(username) DO UPDATE SET nonce = excluded.nonce;

-- name: GetChallenge :one
SELECT nonce FROM challenges
WHERE username = ? LIMIT 1;

-- name: DeleteChallenge :exec
DELETE FROM challenges
WHERE username = ?;
