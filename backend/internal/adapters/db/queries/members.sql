-- name: CreateMember :exec
INSERT INTO members (account_id, first_name, last_name, birth_date, nickname, about_me, avatar_path, visibility)
VALUES (
    sqlc.arg(account_id),
    sqlc.arg(first_name),
    sqlc.arg(last_name),
    sqlc.arg(birth_date),
    sqlc.arg(nickname),
    sqlc.arg(about_me),
    sqlc.arg(avatar_path),
    sqlc.arg(visibility)
);

-- name: GetMemberByAccountID :one
SELECT * FROM members WHERE account_id = sqlc.arg(account_id);

-- name: UpdateMember :exec
UPDATE members
SET nickname    = sqlc.arg(nickname),
    about_me    = sqlc.arg(about_me),
    avatar_path = sqlc.arg(avatar_path),
    visibility  = sqlc.arg(visibility)
WHERE account_id = sqlc.arg(account_id);
