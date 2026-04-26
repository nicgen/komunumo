# Implementation Plan: Authentification utilisateur

**Branch**: `feat/001-user-auth` | **Date**: 2026-04-26 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-user-auth/spec.md`

## Summary

Implémenter le parcours d'authentification end-to-end (inscription, vérification email, connexion, reset mot de passe, déconnexion, middleware `RequireAuth`) en respectant la Constitution v1.0.0 d'AssoLink/Komunumo : architecture hexagonale Go avec test-first sur le domaine, sessions cookies HttpOnly/Secure/SameSite=Strict (cf. ADR-0004), bcrypt cost 12+, audit log append-only V1, RGAA AAA sur tous les écrans concernés (formulaires shadcn/ui sur Radix UI). Cette feature constitue la première verticale end-to-end et valide la chaîne complète domain → application → ports → adapters (http/db/email).

## Technical Context

**Language/Version**: Go 1.24 (backend), TypeScript 5.6+ avec Next.js 16 / React 19 (frontend)
**Primary Dependencies**:
- Backend : `golang-migrate` pour migrations, `sqlc` pour la génération de code SQL → Go (cf. ADR-0003), `modernc.org/sqlite` (driver pure Go), `golang.org/x/crypto/bcrypt`, `github.com/go-chi/chi/v5` ou `net/http` standard, client HTTP Brevo (API REST).
- Frontend : Next.js App Router + RSC (cf. ADR-0002), Tailwind v4 + shadcn/ui sur Radix UI (cf. ADR-0006), `zod` pour validation, `react-hook-form` si interactivité avancée nécessaire.
**Storage**: SQLite WAL + FTS5, migrations versionnées via `golang-migrate`, requêtes via `sqlc` (pas d'ORM). Tables touchées : `accounts`, `sessions`, `email_verifications`, `password_resets`, `audit_log`.
**Testing**:
- Backend : `testing` standard + `testify/require` ; tests domaine purs (sans DB) ; tests d'intégration sur SQLite fichier temporaire ; tests de contrat sur les handlers HTTP via `httptest`.
- Frontend : `vitest` + `@testing-library/react` ; tests E2E `playwright` sur les parcours critiques (post-MVP).
**Target Platform**:
- Backend : container Linux (Debian slim) sur Scaleway DEV1-S Paris (démo) / Traefik local (dev).
- Frontend : Vercel région `cdg1` (Edge runtime exclu pour l'auth — sessions = Node runtime).
**Project Type**: web (backend + frontend séparés, monorepo dépôt unique).
**Performance Goals**: `/login` p95 < 400 ms (incluant ~250 ms de bcrypt cost 12), `/register` p95 < 600 ms (incluant l'envoi d'email Brevo en synchrone V1).
**Constraints**:
- Pages `/register`, `/login`, `/reset-password` < 100 KB HTML+CSS+JS critique (RGESN, cf. Constitution principe II).
- 100 % des parcours d'auth atteignent RGAA AAA (axe-core + lighthouse-ci + audit manuel NVDA/Orca).
- Aucun mot de passe ni email en clair dans les logs (vérifié par grep en CI sur 1 000 requêtes simulées).
- bcrypt cost ≥ 12 (vérifié par test unitaire mesurant le temps de hashage).
**Scale/Scope**:
- MVP : 50 comptes seed + ~500 vrais comptes attendus à 6 mois (cf. vision.md).
- Volume `audit_log` V1 : ~10 000 entrées attendues à 6 mois (négligeable pour SQLite).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principe | Conformité | Justification |
|----------|-----------|---------------|
| **I. Souveraineté numérique** | PASS | Backend Scaleway Paris, email Brevo (FR), aucune dépendance hors UE pour le chemin critique. |
| **II. Sobriété (RGESN) + Accessibilité (RGAA)** | PASS | shadcn/ui sur Radix UI (accessibilité par défaut), formulaires < 100 KB, dégradation gracieuse sans JS, audit RGAA AAA dans Success Criteria (SC-003, SC-006). |
| **III. Hexagonale + test-first** | PASS | Code organisé en `domain/application/ports/adapters` ; chaque commit `feat(auth): …` sera précédé d'un commit `test(auth): …` correspondant, vérifiable dans `git log`. |
| **IV. Spec-driven** | PASS | Feature spec.md ✓ ; plan.md (ce fichier) ✓ ; ADRs référencés (0001, 0002, 0003, 0004, 0006, 0012, 0014) ; tasks.md à venir via `/speckit-tasks`. |
| **V. Sécurité par défaut** | PASS | Cookies `__Host-session` (HttpOnly/Secure/SameSite=Strict), bcrypt cost 12, audit_log append-only via trigger SQLite, anti-énumération, rate limiting. CSP, HSTS, X-Frame-Options posés au niveau adapter HTTP. |
| **VI. Workflow Git industrialisé** | PASS | Branche `feat/001-user-auth`, Conventional Commits, PR avec CI verte avant merge. Override admin uniquement post-revue solo (≥ J+1). |

**Verdict** : aucune violation, on peut entrer en Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/001-user-auth/
├── plan.md              # Ce fichier
├── research.md          # Phase 0 - décisions tech consolidées
├── data-model.md        # Phase 1 - schéma SQLite + invariants
├── quickstart.md        # Phase 1 - démarrage local pour cette feature
├── contracts/           # Phase 1 - contrats API + interfaces Go
│   ├── openapi.auth.yaml          # Endpoints HTTP (extrait de docs/specs/03-api/openapi.yaml)
│   └── ports.md                   # Interfaces Go des ports (côté domain)
├── checklists/
│   └── requirements.md  # Validation spec (déjà créé)
└── tasks.md             # Phase 2 - généré par /speckit-tasks
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   └── server/
│       └── main.go                    # Composition root (DI manuel)
├── internal/
│   ├── domain/
│   │   ├── account/
│   │   │   ├── account.go             # Entité Account + invariants
│   │   │   ├── account_test.go        # Tests purs (sans DB)
│   │   │   ├── password.go            # Value object + règles robustesse
│   │   │   └── password_test.go
│   │   ├── session/
│   │   │   ├── session.go
│   │   │   └── session_test.go
│   │   └── audit/
│   │       └── event.go               # Types d'événements typés
│   ├── application/
│   │   ├── auth/
│   │   │   ├── register.go            # Use case "inscrire un compte"
│   │   │   ├── register_test.go       # Test avec mocks de ports
│   │   │   ├── login.go
│   │   │   ├── login_test.go
│   │   │   ├── verify_email.go
│   │   │   ├── reset_password.go
│   │   │   └── logout.go
│   │   └── audit/
│   │       └── log.go                 # Use case "journaliser un événement"
│   ├── ports/
│   │   ├── account_repository.go      # Interface
│   │   ├── session_repository.go
│   │   ├── token_repository.go        # email_verifications + password_resets
│   │   ├── audit_repository.go
│   │   ├── email_sender.go            # SendEmail(ctx, to, template, data)
│   │   ├── password_hasher.go         # Hash/Verify
│   │   └── clock.go                   # Tests déterministes
│   └── adapters/
│       ├── http/
│       │   ├── auth_handler.go        # POST /api/v1/auth/{register,login,...}
│       │   ├── auth_handler_test.go   # Tests d'intégration via httptest
│       │   ├── middleware/
│       │   │   ├── require_auth.go    # Lit cookie, charge la session
│       │   │   ├── rate_limit.go      # Token bucket par IP + compte
│       │   │   ├── csrf.go            # Double-submit cookie
│       │   │   └── security_headers.go # CSP, HSTS, X-Frame-Options
│       │   └── router.go              # Composition chi/net.http
│       ├── db/
│       │   ├── sqlc/                  # Code généré par sqlc (commit oui)
│       │   ├── queries/               # Fichiers SQL source pour sqlc
│       │   │   ├── accounts.sql
│       │   │   ├── sessions.sql
│       │   │   ├── tokens.sql
│       │   │   └── audit_log.sql
│       │   ├── migrations/            # Fichiers golang-migrate
│       │   │   ├── 0001_init_auth.up.sql
│       │   │   └── 0001_init_auth.down.sql
│       │   ├── account_repository.go  # Implémente ports.AccountRepository
│       │   ├── session_repository.go
│       │   ├── token_repository.go
│       │   └── audit_repository.go
│       ├── email/
│       │   ├── brevo.go               # Implémente ports.EmailSender via API Brevo
│       │   ├── brevo_test.go          # Tests avec mock HTTP server
│       │   └── templates/             # Templates HTML accessibles
│       │       ├── verify_email.html
│       │       ├── reset_password.html
│       │       └── password_changed.html
│       ├── password/
│       │   └── bcrypt_hasher.go       # Implémente ports.PasswordHasher (cost 12)
│       └── clock/
│           └── system_clock.go
├── go.mod
├── go.sum
└── sqlc.yaml

frontend/
├── app/
│   ├── (auth)/
│   │   ├── register/
│   │   │   └── page.tsx               # RSC + form action
│   │   ├── login/
│   │   │   └── page.tsx
│   │   ├── verify-email/
│   │   │   ├── sent/
│   │   │   │   └── page.tsx
│   │   │   └── confirm/
│   │   │       └── page.tsx           # Lit token via searchParams, POST backend
│   │   ├── reset-password/
│   │   │   ├── page.tsx               # Demande de reset
│   │   │   └── confirm/
│   │   │       └── page.tsx           # Confirmation avec token
│   │   └── layout.tsx                 # Layout centré accessible
│   └── api/                           # Pas d'API Next.js
├── components/
│   └── ui/                            # shadcn/ui imports
├── lib/
│   ├── api.ts                         # Wrapper fetch typé vers backend Go
│   └── auth.ts                        # Helper côté serveur (lecture cookie)
└── tests/
    ├── unit/
    └── e2e/
```

**Structure Decision**: web application avec backend Go hexagonal et frontend Next.js séparés. Le frontend ne contient **aucune route API** : toute logique métier passe par le backend. Un proxy (Vercel rewrites en preview/prod, `next.config.ts` rewrites en dev) forward les requêtes `/api/*` vers `https://api.local.hello-there.net` (dev) ou la cible démo Scaleway. Cette séparation respecte la Constitution principe III (architecture hexagonale stricte) et permet d'envisager une consommation future par d'autres clients (mobile, ActivityPub) sans réécrire le domaine.

## Complexity Tracking

> Aucune violation de la Constitution Check, rien à tracker.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|--------------------------------------|
| _(aucune)_ | _ _ | _ _ |
