# Feature Specification: Profils & Types de compte

**Feature Branch**: `feat/002-profiles`
**Created**: 2026-05-02
**Status**: Approved
**Formal spec**: `docs/specs/02-features/profile.md`

## Résumé

Phase 2 du projet AssoLink. Étend l'infrastructure d'auth (Phase 1) pour :
1. Scinder l'inscription en deux parcours : Personne (`register/member`) et Association (`register/association`).
2. Créer les tables de profil `members` et `associations` avec migration des données Phase 1.
3. Exposer les endpoints profil (`GET/PATCH /api/v1/me/profile`, `GET /api/v1/accounts/{id}/profile`).
4. Aligner le schéma `accounts` sur le MLD (status, kind, suppression colonnes PII migrées).

## User Stories

- **US1** — En tant que visiteur, je veux m'inscrire en tant que Personne (≥ 18 ans) afin d'accéder aux fonctionnalités sociales.
- **US2** — En tant que visiteur, je veux inscrire mon Association afin de publier des annonces et gérer des membres.
- **US3** — En tant qu'utilisateur connecté, je veux consulter et modifier mon profil public.
- **US4** — En tant que visiteur, je veux consulter le profil public d'un autre compte (selon visibilité).

## Acceptance Scenarios

### US1 — Inscription Personne

**Given** un visiteur, **When** il soumet `POST /api/v1/auth/register/member` avec email, password, first_name, last_name, birth_date (≥ 18 ans), **Then** accounts(kind=member, status=pending_verification) + members(visibility=public) sont créés, email de vérification envoyé, audit_log("account_created").

**Given** birth_date < 18 ans, **When** soumis, **Then** 422 "vous devez avoir au moins 18 ans".

### US2 — Inscription Association

**Given** un visiteur, **When** il soumet `POST /api/v1/auth/register/association` avec email, password, legal_name, postal_code, first_name/last_name/birth_date du créateur (≥ 18 ans), **Then** accounts(kind=association, status=pending_verification) + associations(visibility=public) + memberships(role=owner, status=active) créés. SIREN invalide → 422 "siren must be 9 digits".

### US3 — Profil connecté

**Given** connecté kind=member, **When** `GET /api/v1/me/profile`, **Then** 200 avec données member + kind. **When** `PATCH /api/v1/me/profile` {nickname, about_me}, **Then** 200 + audit_log("profile.updated").

### US4 — Profil public

**Given** profil visibility=public, **When** `GET /api/v1/accounts/{id}/profile` sans auth, **Then** 200 sans birth_date. **Given** visibility=private, **Then** 404.

## Règles métier clés

- Age ≥ 18 ans, validé côté serveur (birth_date sur members).
- SIREN : 9 chiffres exactement (regex `^\d{9}$`).
- RNA : W suivi de 9 chiffres (regex `^W\d{9}$`).
- about_me ≤ 500 caractères, about (asso) ≤ 2000 caractères.
- Avatar : ≤ 2 Mo, formats JPEG/PNG/WebP, stockage original (pas d'AVIF en V1).
- Visibility : `public` | `members_only` | `private`.
- Migration sans perte : tous les comptes Phase 1 migrés vers members avec first_name/last_name/birth_date.
