# Research: Profils & Types de compte

**Phase**: 0 | **Date**: 2026-05-02 | **Plan**: plan.md

## R-001 — Migration SQLite : renommage/suppression de colonnes sur `accounts`

**Decision**: Recréation de la table via `CREATE TABLE accounts_new … / INSERT INTO … SELECT … / DROP TABLE accounts / ALTER TABLE accounts_new RENAME TO accounts`.

**Rationale**: SQLite ne supporte pas `ALTER TABLE DROP COLUMN` de manière fiable avant 3.35.0 et la contrainte `CHECK` ne peut pas être modifiée sans recréer la table. La recréation est la stratégie standard golang-migrate pour SQLite.

**Alternatives considered**:
- `ALTER TABLE ADD COLUMN` uniquement (insuffisant — ne supprime pas first_name/last_name/date_of_birth ni ne modifie le CHECK status).
- Migration applicative en Go (risque d'état intermédiaire incohérent en cas d'interruption).

**Impact**: La migration 0002 doit être écrite en une transaction atomique. Le fichier `.down.sql` recrée l'ancienne structure.

---

## R-002 — Renommage des valeurs de `status` (verified → active, disabled → suspended)

**Decision**: `UPDATE accounts SET status = 'active' WHERE status = 'verified'` + `UPDATE accounts SET status = 'suspended' WHERE status = 'disabled'` dans la même transaction que la recréation de table.

**Rationale**: Les données Phase 1 utilisent `verified`/`disabled`. Le CHECK de la nouvelle table accepte `active`/`suspended`/`deleted`. L'UPDATE doit précéder la recréation de la table avec le nouveau CHECK.

**Alternatives considered**:
- Renommage au niveau applicatif (patchwork — laisse le schéma incohérent).

---

## R-003 — Transfert des colonnes PII de `accounts` vers `members`

**Decision**: Dans la migration 0002, après la création de `members`, insérer les données avec :
```sql
INSERT INTO members (account_id, first_name, last_name, birth_date, visibility)
SELECT id, first_name, last_name, date_of_birth, 'public'
FROM accounts;
```
Puis recréer `accounts` sans ces colonnes.

**Rationale**: La migration doit être atomique. Ordre obligatoire : (1) créer `members`, (2) migrer les données, (3) recréer `accounts` sans les colonnes PII.

**Alternatives considered**:
- Garder les colonnes PII sur `accounts` en redondance (rejeté — viole le MLD et crée des incohérences futures).

---

## R-004 — sqlc : requêtes pour `members`, `associations`, `memberships`

**Decision**: Ajouter les fichiers `.sql` dans `backend/internal/adapters/db/queries/` et regénérer avec `sqlc generate`.

**Rationale**: Pattern établi en Phase 1. Pas d'ORM, sqlc seul (Constitution ADR-0003).

**New query files needed**:
- `members.sql` : CreateMember, GetMemberByAccountID, UpdateMember.
- `associations.sql` : CreateAssociation, GetAssociationByAccountID, UpdateAssociation.
- `memberships.sql` : CreateMembership, GetMembershipByIDs.
- `accounts.sql` update : modifier GetByEmailCanonical pour inclure `kind`, ajouter GetByID.

---

## R-005 — Stockage avatar

**Decision**: Stocker l'original dans `data/uploads/avatars/{account_id}/{uuid}.{ext}` (ADR-0011). Pas de processing AVIF en V1.

**Rationale**: Constitution principe II — AVIF déféré V2. Volume V1 négligeable (< 500 comptes).

**Alternatives considered**:
- Vercel Blob / S3 (rejeté — souveraineté numérique, Constitution principe I).
- Processing AVIF serveur (rejeté — hors scope V1 par Constitution).

---

## R-006 — Validation SIREN / RNA

**Decision**: Validation par regex dans le domaine Go.
- SIREN : `^[0-9]{9}$`
- RNA : `^W[0-9]{9}$`

**Rationale**: Règles purement syntaxiques, pas besoin d'appel externe (INSEE API déféré V2).

**Alternatives considered**:
- Vérification existence via API INSEE (V2 — hors scope V1).

---

## R-007 — Endpoint GET /api/v1/auth/me : mise à jour

**Decision**: Ajouter `kind` au `meResponse` existant. Pas de rupture — ajout de champ JSON.

**Rationale**: Le frontend Phase 2 a besoin du `kind` pour router vers le bon formulaire profil.

---

## Résolution des NEEDS CLARIFICATION

Tous les points sont résolus. Aucun NEEDS CLARIFICATION restant.
