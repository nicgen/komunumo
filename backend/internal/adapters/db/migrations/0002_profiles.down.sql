-- Restore accounts with PII columns and old status CHECK.
-- PII is recovered from members where available.
CREATE TABLE accounts_old (
    id              TEXT    PRIMARY KEY,
    email           TEXT    NOT NULL,
    email_canonical TEXT    NOT NULL,
    password_hash   TEXT    NOT NULL,
    status          TEXT    NOT NULL CHECK(status IN ('pending_verification','verified','disabled')),
    first_name      TEXT    NOT NULL DEFAULT '',
    last_name       TEXT    NOT NULL DEFAULT '',
    date_of_birth   TEXT    NOT NULL DEFAULT '2000-01-01',
    created_at      TEXT    NOT NULL,
    updated_at      TEXT    NOT NULL,
    last_login_at   TEXT
);

INSERT INTO accounts_old
    (id, email, email_canonical, password_hash, status,
     first_name, last_name, date_of_birth, created_at, updated_at, last_login_at)
SELECT
    a.id, a.email, a.email_canonical, a.password_hash,
    CASE a.status
        WHEN 'active'    THEN 'verified'
        WHEN 'suspended' THEN 'disabled'
        ELSE a.status
    END,
    COALESCE(m.first_name, ''),
    COALESCE(m.last_name,  ''),
    COALESCE(m.birth_date, '2000-01-01'),
    a.created_at, a.updated_at, a.last_login_at
FROM accounts a
LEFT JOIN members m ON m.account_id = a.id;

DROP TABLE memberships;
DROP TABLE associations;
DROP TABLE members;

DROP TABLE accounts;
ALTER TABLE accounts_old RENAME TO accounts;

CREATE UNIQUE INDEX idx_accounts_email_canonical ON accounts(email_canonical);
CREATE INDEX idx_accounts_status ON accounts(status);
