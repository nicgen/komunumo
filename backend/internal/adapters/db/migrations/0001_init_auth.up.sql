-- =====================================
-- accounts
-- =====================================
CREATE TABLE accounts (
    id              TEXT    PRIMARY KEY,
    email           TEXT    NOT NULL,
    email_canonical TEXT    NOT NULL,
    password_hash   TEXT    NOT NULL,
    status          TEXT    NOT NULL CHECK (status IN ('pending_verification','verified','disabled')),
    first_name      TEXT    NOT NULL,
    last_name       TEXT    NOT NULL,
    date_of_birth   TEXT    NOT NULL,
    created_at      TEXT    NOT NULL DEFAULT (datetime('now','subsec')),
    updated_at      TEXT    NOT NULL DEFAULT (datetime('now','subsec')),
    last_login_at   TEXT
);
CREATE UNIQUE INDEX idx_accounts_email_canonical ON accounts(email_canonical);
CREATE INDEX idx_accounts_status ON accounts(status);

-- =====================================
-- sessions
-- =====================================
CREATE TABLE sessions (
    id              TEXT    PRIMARY KEY,
    account_id      TEXT    NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    created_at      TEXT    NOT NULL DEFAULT (datetime('now','subsec')),
    expires_at      TEXT    NOT NULL,
    last_seen_at    TEXT    NOT NULL DEFAULT (datetime('now','subsec')),
    ip              TEXT,
    user_agent      TEXT
);
CREATE INDEX idx_sessions_account_id ON sessions(account_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- =====================================
-- email_verifications
-- =====================================
CREATE TABLE email_verifications (
    id              TEXT    PRIMARY KEY,
    account_id      TEXT    NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    token_hash      TEXT    NOT NULL,
    created_at      TEXT    NOT NULL DEFAULT (datetime('now','subsec')),
    expires_at      TEXT    NOT NULL,
    consumed_at     TEXT
);
CREATE UNIQUE INDEX idx_email_verifications_token_hash ON email_verifications(token_hash);
CREATE INDEX idx_email_verifications_account_id ON email_verifications(account_id);

-- =====================================
-- password_resets
-- =====================================
CREATE TABLE password_resets (
    id              TEXT    PRIMARY KEY,
    account_id      TEXT    NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    token_hash      TEXT    NOT NULL,
    created_at      TEXT    NOT NULL DEFAULT (datetime('now','subsec')),
    expires_at      TEXT    NOT NULL,
    consumed_at     TEXT
);
CREATE UNIQUE INDEX idx_password_resets_token_hash ON password_resets(token_hash);
CREATE INDEX idx_password_resets_account_id ON password_resets(account_id);

-- =====================================
-- audit_log (APPEND-ONLY)
-- =====================================
CREATE TABLE audit_log (
    id              TEXT    PRIMARY KEY,
    occurred_at     TEXT    NOT NULL DEFAULT (datetime('now','subsec')),
    event_type      TEXT    NOT NULL,
    account_id      TEXT,
    email_hash      TEXT,
    ip              TEXT,
    user_agent      TEXT,
    metadata        TEXT
);
CREATE INDEX idx_audit_log_occurred_at ON audit_log(occurred_at);
CREATE INDEX idx_audit_log_account_id ON audit_log(account_id);
CREATE INDEX idx_audit_log_event_type ON audit_log(event_type);

CREATE TRIGGER audit_log_no_update BEFORE UPDATE ON audit_log
BEGIN SELECT RAISE(ABORT, 'audit_log is append-only'); END;

CREATE TRIGGER audit_log_no_delete BEFORE DELETE ON audit_log
BEGIN SELECT RAISE(ABORT, 'audit_log is append-only'); END;
