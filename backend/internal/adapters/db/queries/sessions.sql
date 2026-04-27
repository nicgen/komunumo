-- Session Queries

-- name: CreateSession :exec
INSERT INTO sessions (
    id,
    account_id,
    created_at,
    expires_at,
    last_seen_at,
    ip,
    user_agent
) VALUES (
    ?1,
    ?2,
    ?3,
    ?4,
    ?5,
    ?6,
    ?7
);

-- name: GetSessionByID :one
SELECT id, account_id, created_at, expires_at, last_seen_at, ip, user_agent
FROM sessions
WHERE id = ?1;

-- name: TouchSessionLastSeen :exec
UPDATE sessions SET last_seen_at = ?1 WHERE id = ?2;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = ?1;

-- name: DeleteAllSessionsForAccount :exec
DELETE FROM sessions WHERE account_id = ?1;

-- name: DeleteExpiredSessions :execrows
DELETE FROM sessions WHERE expires_at <= ?1;
