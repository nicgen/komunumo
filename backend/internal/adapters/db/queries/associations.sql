-- name: CreateAssociation :exec
INSERT INTO associations (account_id, legal_name, siren, rna, postal_code, about, logo_path, visibility)
VALUES (
    sqlc.arg(account_id),
    sqlc.arg(legal_name),
    sqlc.arg(siren),
    sqlc.arg(rna),
    sqlc.arg(postal_code),
    sqlc.arg(about),
    sqlc.arg(logo_path),
    sqlc.arg(visibility)
);

-- name: GetAssociationByAccountID :one
SELECT * FROM associations WHERE account_id = sqlc.arg(account_id);

-- name: UpdateAssociation :exec
UPDATE associations
SET about       = sqlc.arg(about),
    logo_path   = sqlc.arg(logo_path),
    postal_code = sqlc.arg(postal_code),
    visibility  = sqlc.arg(visibility)
WHERE account_id = sqlc.arg(account_id);
