-- User management queries
-- name: CreateUser :one
INSERT INTO users (username, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users 
SET username = $2, email = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- Drive115 credentials management queries
-- name: CreateDrive115Credentials :one
INSERT INTO drive115_credentials (user_id, name, uid, cid, seid, kid, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetDrive115CredentialsByID :one
SELECT * FROM drive115_credentials WHERE id = $1;

-- name: GetDrive115CredentialsByUserID :many
SELECT * FROM drive115_credentials 
WHERE user_id = $1 
ORDER BY created_at DESC;

-- name: GetActiveDrive115CredentialsByUserID :many
SELECT * FROM drive115_credentials 
WHERE user_id = $1 AND is_active = true 
ORDER BY created_at DESC;

-- name: GetDrive115CredentialsByUserIDAndName :one
SELECT * FROM drive115_credentials 
WHERE user_id = $1 AND name = $2;

-- name: UpdateDrive115Credentials :one
UPDATE drive115_credentials 
SET name = $2, uid = $3, cid = $4, seid = $5, kid = $6, is_active = $7, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: SetDrive115CredentialsActive :exec
UPDATE drive115_credentials 
SET is_active = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeactivateAllUserDrive115Credentials :exec
UPDATE drive115_credentials 
SET is_active = false, updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1;

-- name: DeleteDrive115Credentials :exec
DELETE FROM drive115_credentials WHERE id = $1;

-- User session management queries
-- name: CreateUserSession :one
INSERT INTO user_sessions (user_id, session_token, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserSessionByToken :one
SELECT us.*, u.username, u.email 
FROM user_sessions us
JOIN users u ON us.user_id = u.id
WHERE us.session_token = $1 AND us.expires_at > CURRENT_TIMESTAMP;

-- name: GetUserSessionsByUserID :many
SELECT * FROM user_sessions 
WHERE user_id = $1 AND expires_at > CURRENT_TIMESTAMP
ORDER BY created_at DESC;

-- name: DeleteUserSession :exec
DELETE FROM user_sessions WHERE session_token = $1;

-- name: DeleteUserSessionsByUserID :exec
DELETE FROM user_sessions WHERE user_id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM user_sessions WHERE expires_at <= CURRENT_TIMESTAMP; 