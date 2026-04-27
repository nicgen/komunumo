-- Email Verification Queries

-- name: CreateEmailVerification :exec
INSERT INTO email_verifications (
    id,
    account_id,
    token_hash,
    created_at,
    expires_at
) VALUES (
    sqlc.arg(id),
    sqlc.arg(account_id),
    sqlc.arg(token_hash),
    sqlc.arg(created_at),
    sqlc.arg(expires_at)
);

-- name: GetActiveEmailVerificationByHash :one
SELECT * FROM email_verifications
WHERE token_hash = sqlc.arg(token_hash)
  AND consumed_at IS NULL
  AND expires_at > sqlc.arg(now_time);

-- name: ConsumeEmailVerification :exec
UPDATE email_verifications SET consumed_at = sqlc.arg(consumed_at)
WHERE id = sqlc.arg(id) AND consumed_at IS NULL;

-- name: RevokeActiveEmailVerificationsForAccount :exec
UPDATE email_verifications SET consumed_at = sqlc.arg(consumed_at)
WHERE account_id = sqlc.arg(account_id) AND consumed_at IS NULL;

-- Password Reset Queries

-- name: CreatePasswordReset :exec
INSERT INTO password_resets (
    id,
    account_id,
    token_hash,
    created_at,
    expires_at
) VALUES (
    sqlc.arg(id),
    sqlc.arg(account_id),
    sqlc.arg(token_hash),
    sqlc.arg(created_at),
    sqlc.arg(expires_at)
);

-- name: GetActivePasswordResetByHash :one
SELECT * FROM password_resets
WHERE token_hash = sqlc.arg(token_hash)
  AND consumed_at IS NULL
  AND expires_at > sqlc.arg(now_time);

-- name: ConsumePasswordReset :exec
UPDATE password_resets SET consumed_at = sqlc.arg(consumed_at)
WHERE id = sqlc.arg(id) AND consumed_at IS NULL;

-- name: RevokeActivePasswordResetsForAccount :exec
UPDATE password_resets SET consumed_at = sqlc.arg(consumed_at)
WHERE account_id = sqlc.arg(account_id) AND consumed_at IS NULL;
