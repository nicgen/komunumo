# Modèle Logique de Données (MLD) - AssoLink

Tables dérivées du MCD (`mcd.mmd`), niveau implémentation SQLite.

Conventions :
- Toutes les clés primaires sont des UUID v7 (sauf `audit_log.id` BIGINT auto-incrément, `session.id` token aléatoire 32 octets base64).
- Timestamps en `INTEGER` (Unix epoch microsecondes) pour SQLite.
- Soft-delete : colonne `deleted_at` nullable. La purge dure (RGPD article 17) est faite en batch après 30 jours.
- Index : voir section dédiée par table.

## Tables principales

### `accounts`

| Colonne | Type | Contraintes |
|---------|------|-------------|
| id | TEXT (UUID) | PK |
| email | TEXT | NOT NULL UNIQUE |
| password_hash | TEXT | NOT NULL |
| status | TEXT | NOT NULL CHECK IN ('pending_verification','active','suspended','deleted') — Phase 2 : renommer verified→active, disabled→suspended |
| kind | TEXT | NOT NULL CHECK IN ('member','association') — ajouté migration 0002 |
| created_at | INTEGER | NOT NULL |
| updated_at | INTEGER | NOT NULL |
| deleted_at | INTEGER | NULL |

Index : `idx_accounts_email_lower` sur `lower(email)`, `idx_accounts_status` sur `status`.

### `members`

| Colonne | Type | Contraintes |
|---------|------|-------------|
| account_id | TEXT | PK FK -> accounts(id) |
| first_name | TEXT | NOT NULL |
| last_name | TEXT | NOT NULL |
| birth_date | TEXT (ISO date) | NOT NULL |
| nickname | TEXT | NULL |
| about_me | TEXT | NULL |
| avatar_path | TEXT | NULL |
| visibility | TEXT | NOT NULL CHECK IN ('public','members_only','private') |

### `associations`

| Colonne | Type | Contraintes |
|---------|------|-------------|
| account_id | TEXT | PK FK -> accounts(id) |
| legal_name | TEXT | NOT NULL |
| siren | TEXT | NULL |
| rna | TEXT | NULL |
| postal_code | TEXT | NOT NULL |
| about | TEXT | NULL |
| logo_path | TEXT | NULL |
| visibility | TEXT | NOT NULL CHECK IN ('public','members_only','private') |

Index : `idx_associations_postal_code` sur `postal_code` (filtre carte V1).

### `memberships`

| Colonne | Type | Contraintes |
|---------|------|-------------|
| id | TEXT (UUID) | PK |
| member_account_id | TEXT | NOT NULL FK -> accounts(id) |
| association_account_id | TEXT | NOT NULL FK -> accounts(id) |
| role | TEXT | NOT NULL CHECK IN ('owner','admin','member') |
| status | TEXT | NOT NULL CHECK IN ('pending','active','left') |
| joined_at | INTEGER | NOT NULL |

Index : UNIQUE `(member_account_id, association_account_id)` ; `idx_memberships_asso` sur `association_account_id`.

### `follows`

| Colonne | Type | Contraintes |
|---------|------|-------------|
| id | TEXT (UUID) | PK |
| follower_account_id | TEXT | NOT NULL FK -> accounts(id) |
| target_account_id | TEXT | NOT NULL FK -> accounts(id) |
| status | TEXT | NOT NULL CHECK IN ('pending','accepted','declined') |
| created_at | INTEGER | NOT NULL |

Index : UNIQUE `(follower_account_id, target_account_id)` ; `idx_follows_target` sur `target_account_id`.

### `posts`

| Colonne | Type | Contraintes |
|---------|------|-------------|
| id | TEXT (UUID) | PK |
| author_account_id | TEXT | NOT NULL FK -> accounts(id) |
| content | TEXT | NOT NULL CHECK length <= 5000 |
| visibility | TEXT | NOT NULL CHECK IN ('public','followers','members','private') |
| media_path | TEXT | NULL |
| media_alt | TEXT | NULL |
| created_at | INTEGER | NOT NULL |
| updated_at | INTEGER | NOT NULL |
| deleted_at | INTEGER | NULL |

Index : `idx_posts_author_created` sur `(author_account_id, created_at DESC)` ; FTS5 virtual table `posts_fts(content)`.

### `post_private_viewers`

PK composite `(post_id, viewer_account_id)`.

### `comments`, `events`, `event_rsvp`, `conversations`, `conversation_participants`, `messages`

Structures équivalentes au MCD. Index :
- `comments` : `(post_id, created_at)`.
- `events` : `(association_account_id, starts_at)`, `(postal_code, starts_at)`.
- `messages` : `(conversation_id, created_at DESC)`.

### `notifications`

Index : `(recipient_account_id, created_at DESC)`, `idx_notifications_unread` partiel sur `read_at IS NULL`.

### `sessions`

| Colonne | Type |
|---------|------|
| id | TEXT (token base64) PK |
| account_id | TEXT FK |
| ip_hash | TEXT (sha256(ip+pepper)) |
| user_agent | TEXT |
| created_at | INTEGER |
| expires_at | INTEGER |
| revoked_at | INTEGER NULL |

Index : `idx_sessions_account` sur `account_id`, `idx_sessions_expires` partiel sur `revoked_at IS NULL`.

### `audit_log`

| Colonne | Type |
|---------|------|
| id | INTEGER PK AUTOINCREMENT |
| actor_account_id | TEXT FK |
| action | TEXT |
| target_type | TEXT |
| target_id | TEXT |
| payload_json | TEXT |
| prev_hash | BLOB(32) |
| hash | BLOB(32) |
| at | INTEGER |

Index : `idx_audit_actor` sur `actor_account_id`, `idx_audit_at` sur `at`.

Append-only enforcement : trigger `BEFORE UPDATE` qui RAISE.

### `email_verifications`, `password_resets`

Mêmes pattern : token_hash, expires_at, consumed_at.

### `user_preferences`

PK = `account_id`. Toutes colonnes ont une valeur par défaut SQL.

## Recherche full-text (FTS5)

```sql
CREATE VIRTUAL TABLE posts_fts USING fts5(
  content,
  author_id UNINDEXED,
  visibility UNINDEXED,
  tokenize='unicode61 remove_diacritics 2'
);

CREATE TRIGGER posts_fts_insert AFTER INSERT ON posts BEGIN
  INSERT INTO posts_fts(rowid, content, author_id, visibility)
    VALUES (new.rowid, new.content, new.author_account_id, new.visibility);
END;
-- + triggers update / delete
```

## Pragmas SQLite à activer au démarrage

```sql
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA foreign_keys = ON;
PRAGMA temp_store = MEMORY;
PRAGMA mmap_size = 268435456;
PRAGMA busy_timeout = 5000;
```
