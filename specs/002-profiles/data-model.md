# Data Model: Profils & Types de compte

**Phase**: 1 | **Date**: 2026-05-02 | **Source**: `docs/specs/04-data/mld.md`

## Entités

### `accounts` (modifiée)

Suppression des colonnes PII (migrées vers `members`). Ajout de `kind`. Modification du CHECK `status`.

```sql
CREATE TABLE accounts (
  id           TEXT PRIMARY KEY,
  email        TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  status       TEXT NOT NULL CHECK(status IN ('pending_verification','active','suspended','deleted')),
  kind         TEXT NOT NULL DEFAULT 'member' CHECK(kind IN ('member','association')),
  created_at   INTEGER NOT NULL,
  updated_at   INTEGER NOT NULL,
  deleted_at   INTEGER
);
-- Colonnes supprimées vs Phase 1 :
--   first_name, last_name, date_of_birth, email_canonical, last_login_at
```

**Invariants**:
- `status` ne peut pas régresser de `active` → `pending_verification`.
- `deleted_at` non null ⇒ `status = 'deleted'` (enforced par trigger).
- `kind` immutable après création.

### `members` (nouvelle)

```sql
CREATE TABLE members (
  account_id   TEXT PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
  first_name   TEXT NOT NULL,
  last_name    TEXT NOT NULL,
  birth_date   TEXT NOT NULL,   -- ISO 8601 YYYY-MM-DD
  nickname     TEXT,
  about_me     TEXT CHECK(about_me IS NULL OR length(about_me) <= 500),
  avatar_path  TEXT,
  visibility   TEXT NOT NULL DEFAULT 'public'
               CHECK(visibility IN ('public','members_only','private'))
);
```

**Invariants**:
- `birth_date` → âge ≥ 18 ans à la date d'inscription (validé dans le domaine Go, pas en SQL).
- Un seul `members` par `account_id` (PK).

### `associations` (nouvelle)

```sql
CREATE TABLE associations (
  account_id   TEXT PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
  legal_name   TEXT NOT NULL,
  siren        TEXT CHECK(siren IS NULL OR siren REGEXP '^\d{9}$'),
  rna          TEXT CHECK(rna IS NULL OR rna REGEXP '^W\d{9}$'),
  postal_code  TEXT NOT NULL,
  about        TEXT CHECK(about IS NULL OR length(about) <= 2000),
  logo_path    TEXT,
  visibility   TEXT NOT NULL DEFAULT 'public'
               CHECK(visibility IN ('public','members_only','private'))
);

CREATE INDEX idx_associations_postal_code ON associations(postal_code);
```

**Note**: SQLite ne supporte pas REGEXP nativement — la validation SIREN/RNA est faite dans le domaine Go, pas en SQL. Le CHECK SQL est documentaire.

**Invariants**:
- Au moins un des deux (`siren`, `rna`) doit être renseigné OU les deux peuvent être NULL (validation applicative, pas contrainte SQL en V1).

### `memberships` (nouvelle)

```sql
CREATE TABLE memberships (
  id                     TEXT PRIMARY KEY,
  member_account_id      TEXT NOT NULL REFERENCES accounts(id),
  association_account_id TEXT NOT NULL REFERENCES accounts(id),
  role                   TEXT NOT NULL CHECK(role IN ('owner','admin','member')),
  status                 TEXT NOT NULL CHECK(status IN ('pending','active','left')),
  joined_at              INTEGER NOT NULL,
  UNIQUE(member_account_id, association_account_id)
);

CREATE INDEX idx_memberships_asso ON memberships(association_account_id);
```

**Invariants**:
- Une association a toujours exactement un `owner` actif (enforced applicativement).
- `member_account_id` doit référencer un compte `kind='member'` (enforced applicativement).

## Migration 0002

**Fichier**: `backend/internal/adapters/db/migrations/0002_profiles.up.sql`

Ordre d'exécution dans une transaction :

1. Renommer les valeurs de status existantes.
2. Créer `members`, `associations`, `memberships`.
3. Migrer les données PII de `accounts` vers `members`.
4. Recréer `accounts` sans les colonnes PII et avec les nouvelles contraintes.
5. Rétablir les index.

```sql
BEGIN;

-- Étape 1 : normaliser les status avant la recréation de la table
UPDATE accounts SET status = 'active'    WHERE status = 'verified';
UPDATE accounts SET status = 'suspended' WHERE status = 'disabled';

-- Étape 2 : nouvelles tables
CREATE TABLE members ( ... );
CREATE TABLE associations ( ... );
CREATE TABLE memberships ( ... );

-- Étape 3 : migration PII
INSERT INTO members (account_id, first_name, last_name, birth_date, visibility)
SELECT id, first_name, last_name, date_of_birth, 'public'
FROM accounts;

-- Étape 4 : recréation de accounts
CREATE TABLE accounts_new ( ... ); -- nouvelle structure
INSERT INTO accounts_new SELECT id, email, password_hash, status, 'member', created_at, updated_at, NULL FROM accounts;
DROP TABLE accounts;
ALTER TABLE accounts_new RENAME TO accounts;

-- Étape 5 : index
CREATE UNIQUE INDEX idx_accounts_email ON accounts(email);
CREATE INDEX idx_associations_postal_code ON associations(postal_code);

COMMIT;
```

## Nouveaux domaines Go

### `internal/domain/member/`

```
member.go          — entité Member + NewMember(accountID, firstName, lastName, birthDate) + invariants
member_test.go     — tests purs : âge minimum, longueur about_me
```

### `internal/domain/association/`

```
association.go     — entité Association + NewAssociation(...) + ValidateSIREN/RNA
association_test.go
```

### `internal/application/auth/` (mise à jour)

```
register_member.go     — use case RegisterMember (scission de register.go)
register_member_test.go
register_association.go
register_association_test.go
```

### `internal/application/profile/`

```
get_profile.go         — use case GetProfile (me + public)
get_profile_test.go
update_profile.go      — use case UpdateProfile
update_profile_test.go
upload_avatar.go       — use case UploadAvatar
upload_avatar_test.go
```

## Transitions de state `status`

```
pending_verification ──(verify email)──→ active
active               ──(admin action)──→ suspended
suspended            ──(admin action)──→ active
active/suspended     ──(RGPD request)──→ deleted
```

`deleted` est terminal en V1 (purge physique après 30j par batch, ADR futur).
