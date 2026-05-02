# Implementation Plan: Profils & Types de compte

**Branch**: `feat/002-profiles` | **Date**: 2026-05-02 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-profiles/spec.md`

## Summary

Étendre l'infrastructure d'auth Phase 1 pour introduire deux types de comptes (member / association), migrer les données PII de `accounts` vers `members`, exposer les endpoints de profil, et livrer les deux parcours d'inscription distincts. Cette phase est fondatrice pour toutes les phases suivantes (follows, posts, events, memberships).

## Technical Context

**Language/Version**: Go 1.24 (backend), TypeScript 5.6+ avec Next.js 16 / React 19 (frontend)
**Primary Dependencies**:
- Backend : `golang-migrate`, `sqlc`, `modernc.org/sqlite`, `net/http`, `golang.org/x/crypto/bcrypt`, Brevo email (existant).
- Frontend : Next.js 16 App Router, Tailwind v4, shadcn/ui, `zod`, `react-hook-form`.
**Storage**: SQLite WAL. Migration `0002_profiles` : recréation de `accounts` + création `members`/`associations`/`memberships` + migration données PII. Voir `data-model.md`.
**Testing**: Go `testing` + `testify/require` (domain pur + intégration httptest). Vitest + @testing-library/react (frontend).
**Target Platform**: Backend Linux container Scaleway Paris. Frontend Vercel `cdg1`.
**Project Type**: web (backend hexagonal + frontend Next.js séparés).
**Performance Goals**: `GET /api/v1/me/profile` p95 < 100 ms. `PATCH /api/v1/me/profile` p95 < 200 ms. Avatar upload p95 < 2 s.
**Constraints**:
- Migration atomique sans perte de données (tous les comptes Phase 1 migrés).
- RGAA AAA sur les nouveaux formulaires (register/member, register/association, /profile).
- PII (birth_date) jamais exposé aux visiteurs non autorisés.
- Avatar : stockage original uniquement, pas d'AVIF en V1 (Constitution principe II).
**Scale/Scope**: ~500 comptes attendus à 6 mois. Volume `members` et `associations` négligeable pour SQLite.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principe | Conformité | Justification |
|----------|-----------|---------------|
| **I. Souveraineté numérique** | PASS | Backend Scaleway Paris, avatars stockés localement (`data/uploads/`), aucune dépendance hors UE ajoutée. |
| **II. Sobriété + Accessibilité RGAA** | PASS | Nouveaux formulaires shadcn/ui + Radix UI. `aria-describedby` sur tous les champs d'erreur (pattern établi Phase 1). Avatar sans processing AVIF en V1. |
| **III. Hexagonale + test-first** | PASS | Nouveaux domaines `member`, `association` écrits après leurs tests. Chaque commit `feat(profiles):` précédé d'un `test(profiles):` vérifiable dans `git log`. |
| **IV. Spec-driven** | PASS | spec.md + plan.md + data-model.md + contracts/ présents. ADRs 0001, 0003, 0006, 0011 référencés. |
| **V. Sécurité** | PASS | PII protégé par la couche visibilité. Audit log sur account_created et profile.updated. Validation SIREN/RNA dans le domaine. Avatar MIME-check côté serveur. |
| **VI. Workflow Git** | PASS | Branche `feat/002-profiles`, Conventional Commits, PR CI verte avant merge. |

**Verdict** : aucune violation. Entrée en Phase 0 validée.

## Project Structure

### Documentation (this feature)

```text
specs/002-profiles/
├── plan.md              # Ce fichier
├── spec.md              # Spec de travail Speckit
├── research.md          # Phase 0 — décisions et recherches
├── data-model.md        # Phase 1 — schéma SQLite + invariants + migration
├── quickstart.md        # Phase 1 — démarrage local
├── contracts/
│   ├── openapi.profiles.yaml   # Contrats API nouveaux endpoints
│   └── ports.md                # Interfaces Go des nouveaux ports
└── tasks.md             # Phase 2 — généré par /speckit-tasks
```

### Source Code

```text
backend/
├── internal/
│   ├── domain/
│   │   ├── member/
│   │   │   ├── member.go          # Entité Member + NewMember + invariants (âge, about_me)
│   │   │   └── member_test.go
│   │   ├── association/
│   │   │   ├── association.go     # Entité Association + ValidateSIREN/RNA
│   │   │   └── association_test.go
│   │   └── account/
│   │       └── account.go        # Mise à jour : Kind, Status (active/suspended/deleted)
│   ├── application/
│   │   ├── auth/
│   │   │   ├── register_member.go         # Use case RegisterMember (scission register.go)
│   │   │   ├── register_member_test.go
│   │   │   ├── register_association.go
│   │   │   └── register_association_test.go
│   │   └── profile/
│   │       ├── get_profile.go             # GetMyProfile + GetPublicProfile
│   │       ├── get_profile_test.go
│   │       ├── update_profile.go
│   │       ├── update_profile_test.go
│   │       ├── upload_avatar.go
│   │       └── upload_avatar_test.go
│   ├── ports/
│   │   ├── member_repository.go
│   │   ├── association_repository.go
│   │   ├── membership_repository.go
│   │   └── file_store.go
│   └── adapters/
│       ├── http/
│       │   ├── register_handler.go        # POST /register/member + /register/association
│       │   ├── register_handler_test.go
│       │   ├── profile_handler.go         # GET+PATCH /me/profile, GET /accounts/{id}/profile
│       │   ├── profile_handler_test.go
│       │   └── avatar_handler.go          # POST /me/avatar
│       ├── db/
│       │   ├── migrations/
│       │   │   ├── 0002_profiles.up.sql
│       │   │   └── 0002_profiles.down.sql
│       │   ├── queries/
│       │   │   ├── members.sql
│       │   │   ├── associations.sql
│       │   │   └── memberships.sql
│       │   ├── member_repository.go
│       │   ├── association_repository.go
│       │   └── membership_repository.go
│       └── storage/
│           └── local_file_store.go        # Implémente ports.FileStore
frontend/
├── app/
│   └── (auth)/
│       ├── register/
│       │   └── page.tsx                   # Sélection type de compte
│       ├── register/member/
│       │   └── page.tsx
│       └── register/association/
│           └── page.tsx
└── components/
    ├── auth/
    │   ├── register-member-form.tsx
    │   └── register-association-form.tsx
    └── profile/
        ├── member-profile-form.tsx
        └── association-profile-form.tsx
```

**Structure Decision**: même séparation backend hexagonal / frontend Next.js que Phase 1. Le `register.go` Phase 1 est scindé en `register_member.go` et `register_association.go` — pas de modification destructive de l'existant avant que les nouveaux use cases soient testés.

## Scope & Deferrals

### [REPORTÉ V2] Accord parental pour mineurs
Âge minimum fixé à 18 ans en V1. La gestion de l'accord parental (consentement RGPD pour 13-17 ans) est déférée V2.

### [REPORTÉ V2] Avatar AVIF
Stockage de l'original uniquement. Le processing AVIF (génération côté serveur) est déféré V2 per Constitution principe II.

### [REPORTÉ V2] Vérification SIREN via API INSEE
Validation syntaxique uniquement en V1 (regex 9 chiffres). La vérification d'existence via l'API INSEE est déférée V2.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|--------------------------------------|
| Recréation table `accounts` | SQLite ne supporte pas ALTER TABLE DROP COLUMN ni modification de CHECK | Impossible autrement sur SQLite sans recréer la table |
| Scission `register.go` | `register/member` et `register/association` ont des invariants distincts | Un seul use case avec branche `if kind ==` violerait la séparation des responsabilités |
