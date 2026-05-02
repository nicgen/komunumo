-- name: CreateMembership :exec
INSERT INTO memberships (id, member_account_id, association_account_id, role, status, joined_at)
VALUES (
    sqlc.arg(id),
    sqlc.arg(member_account_id),
    sqlc.arg(association_account_id),
    sqlc.arg(role),
    sqlc.arg(status),
    sqlc.arg(joined_at)
);

-- name: GetMembershipByAccountIDs :one
SELECT * FROM memberships
WHERE member_account_id      = sqlc.arg(member_account_id)
  AND association_account_id = sqlc.arg(association_account_id);
