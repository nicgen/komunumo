-- name: AppendAuditEvent :exec
INSERT INTO audit_log (
    id,
    occurred_at,
    event_type,
    account_id,
    email_hash,
    ip,
    user_agent,
    metadata
) VALUES (
    sqlc.arg(id),
    sqlc.arg(occurred_at),
    sqlc.arg(event_type),
    sqlc.arg(account_id),
    sqlc.arg(email_hash),
    sqlc.arg(ip),
    sqlc.arg(user_agent),
    sqlc.arg(metadata)
);
