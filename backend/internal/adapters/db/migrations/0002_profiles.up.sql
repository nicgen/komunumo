-- Step 1: normalise status values (verified→active, disabled→suspended)
UPDATE accounts SET status = 'active'    WHERE status = 'verified';
UPDATE accounts SET status = 'suspended' WHERE status = 'disabled';

-- Step 2: new tables
CREATE TABLE members (
    account_id  TEXT    PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    first_name  TEXT    NOT NULL,
    last_name   TEXT    NOT NULL,
    birth_date  TEXT    NOT NULL,
    nickname    TEXT,
    about_me    TEXT    CHECK(about_me IS NULL OR length(about_me) <= 500),
    avatar_path TEXT,
    visibility  TEXT    NOT NULL DEFAULT 'public'
                CHECK(visibility IN ('public','members_only','private'))
);

CREATE TABLE associations (
    account_id  TEXT    PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    legal_name  TEXT    NOT NULL,
    siren       TEXT,
    rna         TEXT,
    postal_code TEXT    NOT NULL,
    about       TEXT    CHECK(about IS NULL OR length(about) <= 2000),
    logo_path   TEXT,
    visibility  TEXT    NOT NULL DEFAULT 'public'
                CHECK(visibility IN ('public','members_only','private'))
);

CREATE TABLE memberships (
    id                     TEXT    PRIMARY KEY,
    member_account_id      TEXT    NOT NULL REFERENCES accounts(id),
    association_account_id TEXT    NOT NULL REFERENCES accounts(id),
    role                   TEXT    NOT NULL CHECK(role IN ('owner','admin','member')),
    status                 TEXT    NOT NULL CHECK(status IN ('pending','active','left')),
    joined_at              TEXT    NOT NULL,
    UNIQUE(member_account_id, association_account_id)
);

-- Step 3: migrate PII from accounts to members
INSERT INTO members (account_id, first_name, last_name, birth_date, visibility)
SELECT id, first_name, last_name, date_of_birth, 'public'
FROM accounts;

-- Step 4: recreate accounts without PII, with kind + deleted_at, with updated status CHECK.
-- email_canonical and last_login_at are retained for existing auth functionality.
CREATE TABLE accounts_new (
    id              TEXT    PRIMARY KEY,
    email           TEXT    NOT NULL,
    email_canonical TEXT    NOT NULL,
    password_hash   TEXT    NOT NULL,
    status          TEXT    NOT NULL CHECK(status IN ('pending_verification','active','suspended','deleted')),
    kind            TEXT    NOT NULL DEFAULT 'member' CHECK(kind IN ('member','association')),
    created_at      TEXT    NOT NULL,
    updated_at      TEXT    NOT NULL,
    last_login_at   TEXT,
    deleted_at      TEXT
);

INSERT INTO accounts_new
    (id, email, email_canonical, password_hash, status, kind, created_at, updated_at, last_login_at)
SELECT
    id, email, email_canonical, password_hash, status, 'member', created_at, updated_at, last_login_at
FROM accounts;

DROP TABLE accounts;
ALTER TABLE accounts_new RENAME TO accounts;

-- Step 5: recreate indexes
CREATE UNIQUE INDEX idx_accounts_email_canonical ON accounts(email_canonical);
CREATE INDEX idx_accounts_status ON accounts(status);
CREATE INDEX idx_associations_postal_code ON associations(postal_code);
CREATE INDEX idx_memberships_asso ON memberships(association_account_id);
