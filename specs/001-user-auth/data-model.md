# Phase 1 - Data Model: Authentification utilisateur

Schéma SQLite pour la feature `auth`. Cinq tables : `accounts`, `sessions`, `email_verifications`, `password_resets`, `audit_log`. Toutes les tables utilisent des UUID v7 (TEXT 36 chars) comme clés primaires.

## Schéma SQL (référence — la version canonique vit dans `backend/internal/adapters/db/migrations/0001_init_auth.up.sql`)

```sql
-- =====================================
-- accounts
-- =====================================
CREATE TABLE accounts (
    id              TEXT    PRIMARY KEY,                    -- UUID v7
    email           TEXT    NOT NULL,                        -- normalisé NFKC + lowercase
    email_canonical TEXT    NOT NULL,                        -- pour unicité, NFKC + lowercase + strip dots gmail (?)
    password_hash   TEXT    NOT NULL,                        -- bcrypt cost 12
    status          TEXT    NOT NULL CHECK (status IN ('pending_verification','verified','disabled')),
    first_name      TEXT    NOT NULL,
    last_name       TEXT    NOT NULL,
    date_of_birth   TEXT    NOT NULL,                        -- ISO 8601 'YYYY-MM-DD'
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
    id              TEXT    PRIMARY KEY,                    -- UUID v7 (= valeur du cookie __Host-session)
    account_id      TEXT    NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    created_at      TEXT    NOT NULL DEFAULT (datetime('now','subsec')),
    expires_at      TEXT    NOT NULL,                        -- created_at + 30 jours
    last_seen_at    TEXT    NOT NULL DEFAULT (datetime('now','subsec')),
    ip              TEXT,                                    -- pour audit, pas de comparaison stricte
    user_agent      TEXT
);
CREATE INDEX idx_sessions_account_id ON sessions(account_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- =====================================
-- email_verifications
-- =====================================
CREATE TABLE email_verifications (
    id              TEXT    PRIMARY KEY,                    -- UUID v7
    account_id      TEXT    NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    token_hash      TEXT    NOT NULL,                        -- SHA-256 du token (pas le token brut)
    created_at      TEXT    NOT NULL DEFAULT (datetime('now','subsec')),
    expires_at      TEXT    NOT NULL,                        -- created_at + 24 h
    consumed_at     TEXT
);
CREATE UNIQUE INDEX idx_email_verifications_token_hash ON email_verifications(token_hash);
CREATE INDEX idx_email_verifications_account_id ON email_verifications(account_id);

-- =====================================
-- password_resets
-- =====================================
CREATE TABLE password_resets (
    id              TEXT    PRIMARY KEY,                    -- UUID v7
    account_id      TEXT    NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    token_hash      TEXT    NOT NULL,                        -- SHA-256 du token
    created_at      TEXT    NOT NULL DEFAULT (datetime('now','subsec')),
    expires_at      TEXT    NOT NULL,                        -- created_at + 30 min
    consumed_at     TEXT
);
CREATE UNIQUE INDEX idx_password_resets_token_hash ON password_resets(token_hash);
CREATE INDEX idx_password_resets_account_id ON password_resets(account_id);

-- =====================================
-- audit_log (APPEND-ONLY)
-- =====================================
CREATE TABLE audit_log (
    id              TEXT    PRIMARY KEY,                    -- UUID v7
    occurred_at     TEXT    NOT NULL DEFAULT (datetime('now','subsec')),
    event_type      TEXT    NOT NULL,                        -- ex: 'account.created', 'auth.login_success'
    account_id      TEXT,                                    -- nullable (événement avant compte: tentative de login)
    email_hash      TEXT,                                    -- SHA-256 de l'email (jamais en clair)
    ip              TEXT,
    user_agent      TEXT,
    metadata        TEXT                                     -- JSON optionnel pour détails (raison d'échec, etc.)
);
CREATE INDEX idx_audit_log_occurred_at ON audit_log(occurred_at);
CREATE INDEX idx_audit_log_account_id ON audit_log(account_id);
CREATE INDEX idx_audit_log_event_type ON audit_log(event_type);

-- Triggers: empêche UPDATE/DELETE sur audit_log
CREATE TRIGGER audit_log_no_update BEFORE UPDATE ON audit_log
BEGIN SELECT RAISE(ABORT, 'audit_log is append-only'); END;

CREATE TRIGGER audit_log_no_delete BEFORE DELETE ON audit_log
BEGIN SELECT RAISE(ABORT, 'audit_log is append-only'); END;
```

## Invariants métier (côté domain)

### `Account`
- `email` est unique sur sa version canonique (`email_canonical`).
- `password_hash` n'est jamais nul ; n'est jamais lu hors du domaine `account.password`.
- `status` ne peut transitioner que selon le diagramme suivant :
  ```
  (creation) → pending_verification
  pending_verification → verified  (sur consommation d'un email_verification non expiré)
  verified → disabled              (action admin V2, hors scope)
  disabled → verified              (action admin V2, hors scope)
  ```
- `date_of_birth` doit être tel que l'âge calculé à `created_at` ≥ 16 ans.
- `email` est normalisé en NFKC + lowercase **avant** stockage dans `email_canonical`.

### `Session`
- `expires_at` > `created_at`, durée canonique 30 jours.
- `id` est généré côté serveur uniquement (UUID v7), n'est **jamais** dérivé d'une input utilisateur.
- À chaque login réussi, **toutes** les sessions précédentes du compte sont conservées (multi-device toléré). Lors d'un changement de mot de passe, **toutes** les sessions sont supprimées.
- `last_seen_at` est mis à jour de manière paresseuse (au plus une fois par minute) pour ne pas surcharger les écritures.

### `EmailVerification`
- `token_hash` est unique globalement (deux comptes ne peuvent pas avoir le même token actif).
- `consumed_at` est posé une seule fois ; tout token consommé devient inutilisable.
- Un seul token actif (non consommé, non expiré) par compte à la fois ; redemander un email annule les précédents (`consumed_at = now()` sur les anciens).

### `PasswordReset`
- Mêmes invariants qu'`EmailVerification`, avec `expires_at = created_at + 30 min`.

### `AuditLogEntry`
- INSERT-only (garanti par triggers SQLite).
- `email_hash` est obligatoire si l'événement concerne une tentative d'identification (succès ou échec).
- `event_type` suit la nomenclature `<domain>.<action>` (ex: `account.created`, `auth.login_failed`, `auth.password_changed`).
- `metadata` est un JSON valide ou NULL (jamais une chaîne libre non structurée).

## Nomenclature des événements d'audit (V1)

| `event_type` | Quand | Donnée associée typique |
|--------------|-------|--------------------------|
| `account.created` | Création d'un compte (avant vérification email) | `email_hash`, `ip`, `user_agent` |
| `account.email_verified` | Vérification email réussie | `account_id`, `email_hash` |
| `auth.login_success` | Login réussi | `account_id`, `ip`, `user_agent` |
| `auth.login_failed` | Login échoué (mauvais mdp, compte inexistant, compte non vérifié) | `email_hash`, `ip`, `user_agent`, `metadata: {"reason": "wrong_password"\|"unknown_account"\|"pending_verification"}` |
| `auth.password_reset_requested` | Demande de reset | `email_hash`, `ip` |
| `auth.password_changed` | Changement effectif du mdp via reset | `account_id`, `ip` |
| `auth.logout` | Déconnexion volontaire | `account_id`, `ip` |
| `auth.session_expired` | Session expirée naturellement | `account_id` |

## Migrations attendues pour la feature

- `0001_init_auth.up.sql` / `0001_init_auth.down.sql` — toutes les tables ci-dessus.

Pas d'index FTS5 ni de modifications structurelles attendues pour cette feature ; les recherches plein texte arriveront avec la feature `search` (F3, hors scope).

## Volumétrie projetée (6 mois, ~500 comptes actifs)

| Table | Lignes attendues | Taille estimée |
|-------|------------------|----------------|
| `accounts` | 500 | < 100 KB |
| `sessions` | ~1500 (3 sessions/compte en moyenne) | < 300 KB |
| `email_verifications` | ~1000 (renvois inclus) | < 200 KB |
| `password_resets` | ~200 | < 40 KB |
| `audit_log` | ~10 000 (20 événements/compte/an) | < 2 MB |

**Total** : largement sous le seuil de bascule SQLite → PostgreSQL (cf. ADR-0003 mentionne ~1k–5k writes/sec comme limite réaliste, on est très en-deçà).
