-- name: CreateAccount :exec
INSERT INTO accounts (
    id,
    email,
    email_canonical,
    password_hash,
    status,
    first_name,
    last_name,
    date_of_birth,
    created_at,
    updated_at
) VALUES (
    sqlc.arg(id),
    sqlc.arg(email),
    sqlc.arg(email_canonical),
    sqlc.arg(password_hash),
    sqlc.arg(status),
    sqlc.arg(first_name),
    sqlc.arg(last_name),
    sqlc.arg(date_of_birth),
    sqlc.arg(created_at),
    sqlc.arg(updated_at)
);

-- name: GetAccountByEmailCanonical :one
SELECT * FROM accounts WHERE email_canonical = sqlc.arg(email_canonical);

-- name: GetAccountByID :one
SELECT * FROM accounts WHERE id = sqlc.arg(id);

-- name: UpdateAccountStatus :exec
UPDATE accounts SET status = sqlc.arg(status), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);

-- name: UpdateAccountPasswordHash :exec
UPDATE accounts SET password_hash = sqlc.arg(password_hash), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);

-- name: TouchAccountLastLogin :exec
UPDATE accounts SET last_login_at = sqlc.arg(last_login_at)
WHERE id = sqlc.arg(id);
