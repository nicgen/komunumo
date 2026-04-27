---
description: "Tasks - feature 001-user-auth"
---

# Tasks: Authentification utilisateur

**Input**: Design documents from `/specs/001-user-auth/`
**Prerequisites**: plan.md, spec.md (4 user stories), research.md (D-001..D-012), data-model.md (5 tables), contracts/ (openapi.auth.yaml + ports.md), quickstart.md.

**Tests**: REQUIS — la Constitution v1.0.0 (principe III) impose que tout commit `feat(auth): …` touchant `internal/domain` ou `internal/application` soit précédé d'un commit `test(auth): …` correspondant. Les phases incluent donc des tâches de test FIRST, validées rouges avant implémentation.

**Organization**: tâches groupées par user story pour permettre la livraison incrémentale (MVP = US1 + US2).

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: parallélisable (fichier différent, pas de dépendance bloquante).
- **[Story]**: US1 / US2 / US3 / US4 — label requis pour les tâches de phase user story.
- Tous les chemins sont relatifs à la racine du repo (ex: `backend/internal/...`).

## Path Conventions (rappel)

```text
backend/
├── cmd/server/main.go              # Composition root
├── internal/
│   ├── domain/{account,session,token,audit}/
│   ├── application/{auth,audit}/
│   ├── ports/
│   └── adapters/{http,db,email,password,clock}/
├── sqlc.yaml
└── Makefile

frontend/
└── app/(auth)/{register,login,verify-email,reset-password}/
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: bootstrap des deux sous-projets backend/frontend pour pouvoir lancer test/lint/migrate.

- [x] T001 Initialize Go module at `backend/go.mod` with module path `komunumo/backend` and Go 1.24
- [x] T002 [P] Create `backend/Makefile` with targets `test`, `test-domain`, `test-application`, `test-adapters`, `test-http`, `migrate-up`, `migrate-down`, `sqlc`, `run`
- [x] T003 [P] Add backend deps via `go get`: `modernc.org/sqlite`, `golang.org/x/crypto/bcrypt`, `github.com/google/uuid`, `github.com/go-chi/chi/v5`, `github.com/golang-migrate/migrate/v4`, `github.com/stretchr/testify`
- [x] T004 [P] Configure `backend/sqlc.yaml` pointing to `internal/adapters/db/queries/` (input) and `internal/adapters/db/sqlc/` (output) with engine sqlite
- [x] T005 [P] Configure `backend/.golangci.yml` with linters: errcheck, govet, ineffassign, staticcheck, gosec, misspell
- [x] T006 [P] Initialize frontend Next.js 16 project at `frontend/` (App Router, TS, Tailwind v4) with `pnpm create next-app`
- [x] T007 [P] Install frontend deps in `frontend/`: `zod`, `react-hook-form`, shadcn/ui CLI bootstrap, axe-core (`@axe-core/react`), `@playwright/test`
- [x] T008 [P] Create `frontend/next.config.ts` with rewrites `/api/:path*` → `process.env.KOMUNUMO_API_INTERNAL_URL`
- [x] T009 [P] Create `infra/local/docker-compose.yml` with Traefik + mkcert volume for `*.local.hello-there.net` HTTPS

**Checkpoint**: `make test` (backend, vide) et `pnpm test` (frontend, vide) passent ; `make migrate-up` est appelable.

> **Setup note**: `make migrate-up` requires the `migrate` CLI binary (`github.com/golang-migrate/migrate/v4/cmd/migrate`). It is **not** installed by `go get` — must be installed separately: `go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest`. Future onboarding scripts / CI must ensure this binary is present.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: schéma BDD, ports vides, security headers, environnement — tout ce que les 4 user stories partagent.

**CRITIQUE** : aucune US ne peut commencer avant la fin de cette phase.

- [x] T010 Create migration `backend/internal/adapters/db/migrations/0001_init_auth.up.sql` with the 5 tables + indexes + audit triggers (cf. data-model.md)
- [x] T011 [P] Create migration `backend/internal/adapters/db/migrations/0001_init_auth.down.sql` reversing T010
- [x] T012 Create `backend/internal/adapters/db/openconn.go` opening SQLite via `modernc.org/sqlite` with `PRAGMA journal_mode=WAL` + `PRAGMA foreign_keys=ON` + `PRAGMA busy_timeout=5000`
- [x] T013 [P] Create empty domain types `backend/internal/domain/account/account.go` (struct Account, Status enum, errors)
- [x] T014 [P] Create empty domain types `backend/internal/domain/session/session.go` (struct Session + errors)
- [x] T015 [P] Create empty domain types `backend/internal/domain/token/token.go` (struct Token, Kind enum + errors)
- [x] T016 [P] Create empty domain types `backend/internal/domain/audit/event.go` (struct Event + EventType constants from data-model nomenclature)
- [x] T017 [P] Declare port interfaces in `backend/internal/ports/account_repository.go`, `session_repository.go`, `token_repository.go`, `audit_repository.go`, `email_sender.go`, `password_hasher.go`, `clock.go`, `token_generator.go`, `rate_limiter.go`, `unit_of_work.go` (signatures from contracts/ports.md)
- [x] T018 [P] Implement `backend/internal/adapters/clock/system_clock.go` (returns `time.Now().UTC()`)
- [x] T019 [P] Implement `backend/internal/adapters/password/bcrypt_hasher.go` (cost 12, no test yet)
- [x] T020 [P] Implement `backend/internal/adapters/tokengen/uuid_token_gen.go` (UUID v7 + crypto/rand 32 bytes + SHA-256)
- [x] T021 [P] Implement `backend/internal/adapters/ratelimit/in_memory.go` (token bucket, key-based map)
- [x] T022 [P] Implement `backend/internal/adapters/http/middleware/security_headers.go` (CSP strict, HSTS 31536000, X-Frame-Options DENY, Referrer-Policy strict-origin-when-cross-origin)
- [x] T023 [P] Implement `backend/internal/adapters/http/middleware/csrf.go` (double-submit cookie `__Host-csrf` + header `X-CSRF-Token`)
- [x] T024 [P] Implement `backend/internal/adapters/http/middleware/rate_limit.go` consuming `ports.RateLimiter`
- [x] T025 [P] Implement structured logging in `backend/internal/adapters/log/slog.go` (slog JSON, redacts `password`, `password_hash`, `token`, `email` fields)
- [x] T026 Wire composition root in `backend/cmd/server/main.go`: load env, open DB, instantiate adapters, mount router (handlers vides → 501)
  - **Note**: env loading implemented as `envOr(key, fallback)` (reads process env with hardcoded fallbacks). File-based loading (`.env`) and 1Password CLI integration are **not yet implemented** — deferred to Phase 7 or a dedicated infra task.
- [x] T027 [P] Create frontend layout `frontend/app/(auth)/layout.tsx` (centered, accessible main landmark, skip-link)
- [x] T028 [P] Create frontend `frontend/lib/api.ts` (typed fetch wrapper, includes credentials, attaches CSRF header from cookie)
- [x] T029 [P] Create frontend `frontend/lib/auth.ts` (server-side helper reading `__Host-session` cookie + GET /api/v1/auth/me)

**Checkpoint**: `make migrate-up` produit les 5 tables ; `make run` démarre, écoute `:8080`, sert un `404` propre ; `pnpm dev` démarre, page `/login` route 200 mais vide.

---

## Phase 3: User Story 1 - Inscription + vérification email (Priority: P1) MVP

**Goal**: un visiteur crée un compte, reçoit un email de vérification, clique le lien, son compte passe `verified`.

**Independent Test** (cf. spec.md US1): formulaire `/register` rempli → email reçu → lien cliqué → page de confirmation → ligne `accounts.status = 'verified'`. Vérifiable au navigateur + Mailpit/journal Brevo.

### Tests for US1 (TEST-FIRST — Constitution principe III)

- [x] T030 [P] [US1] Domain test `backend/internal/domain/account/account_test.go`: invariants email canonical (NFKC + lowercase), age >= 16, status transitions
- [x] T031 [P] [US1] Domain test `backend/internal/domain/account/password_test.go`: password policy (>= 12 chars, classes), erreurs typées
- [x] T032 [P] [US1] Domain test `backend/internal/domain/token/token_test.go`: TTL 24h, single-consume, revoke active
- [x] T033 [P] [US1] Application test `backend/internal/application/auth/register_test.go`: cas nominal, email déjà utilisé, age < 16, password trop faible, échec Brevo (transactionnel — pas de compte créé)
- [x] T034 [P] [US1] Application test `backend/internal/application/auth/verify_email_test.go`: token valide, expiré, déjà consommé, inconnu
- [x] T035 [P] [US1] Application test `backend/internal/application/auth/resend_verification_test.go`: rate limit 1/email/15min, révocation des tokens précédents
- [x] T036 [US1] Adapter integration test `backend/internal/adapters/db/account_repository_test.go`: Create + FindByEmailCanonical + UpdateStatus sur SQLite fichier temp
- [x] T037 [P] [US1] Adapter integration test `backend/internal/adapters/db/token_repository_test.go`: Create + FindActiveByHash + Consume + RevokeActiveForAccount
- [x] T038 [P] [US1] Adapter integration test `backend/internal/adapters/db/audit_repository_test.go`: Append OK ; UPDATE/DELETE doivent échouer (vérifie les triggers)
- [x] T039 [P] [US1] Adapter test `backend/internal/adapters/email/brevo_test.go`: mock HTTP server, vérifie payload + headers Brevo, retours 200/4xx/5xx mappés en erreurs typées
- [x] T040 [P] [US1] HTTP contract test `backend/internal/adapters/http/auth_handler_register_test.go`: 201 JSON, 303 form-encoded, 400 validation, 429 rate limit, anti-énumération sur email existant
- [x] T041 [P] [US1] HTTP contract test `backend/internal/adapters/http/auth_handler_verify_email_test.go`: 200/303 OK, 400 token invalide, 410 expiré
- [x] T042 [P] [US1] Frontend a11y test `frontend/tests/unit/register.test.tsx`: axe-core 0 violations sur la page `/register`

### Implementation for US1

- [x] T043 [US1] Implement `backend/internal/domain/account/account.go` + `password.go` to make T030–T031 green
- [x] T044 [US1] Implement `backend/internal/domain/token/token.go` to make T032 green
- [x] T045 [P] [US1] Write SQL queries `backend/internal/adapters/db/queries/accounts.sql` (CreateAccount, GetAccountByEmailCanonical, UpdateAccountStatus, etc.) and run `make sqlc`
- [x] T046 [P] [US1] Write SQL queries `backend/internal/adapters/db/queries/tokens.sql` (CreateToken, GetActiveTokenByHash, ConsumeToken, RevokeActiveTokensForAccount) and run `make sqlc`
- [x] T047 [P] [US1] Write SQL queries `backend/internal/adapters/db/queries/audit_log.sql` (AppendAuditEvent) and run `make sqlc`
- [x] T048 [US1] Implement `backend/internal/adapters/db/account_repository.go` to make T036 green
- [x] T049 [P] [US1] Implement `backend/internal/adapters/db/token_repository.go` to make T037 green
- [x] T050 [P] [US1] Implement `backend/internal/adapters/db/audit_repository.go` to make T038 green
- [x] T051 [P] [US1] Implement `backend/internal/adapters/db/unit_of_work.go` (Tx via context key)
- [x] T052 [US1] Implement `backend/internal/adapters/email/brevo.go` (HTTP client API Brevo v3, fail-fast on non-2xx) to make T039 green
- [x] T053 [P] [US1] Create email templates `backend/internal/adapters/email/templates/verify_email.html` (RGAA AAA, alt text, contraste)
- [x] T054 [US1] Implement `backend/internal/application/auth/register.go` (UnitOfWork: insert account + insert token + insert audit; send email; if email fails → rollback) to make T033 green
- [x] T055 [US1] Implement `backend/internal/application/auth/verify_email.go` (UnitOfWork: load token, consume, update account status, audit) to make T034 green
- [x] T056 [US1] Implement `backend/internal/application/auth/resend_verification.go` to make T035 green
- [x] T057 [US1] Implement HTTP handlers `backend/internal/adapters/http/auth_handler.go` for `POST /api/v1/auth/register`, `POST /api/v1/auth/verify-email`, `POST /api/v1/auth/resend-verification` (dual content-type) to make T040–T041 green
- [x] T058 [US1] Mount US1 routes in `backend/internal/adapters/http/router.go` with rate limit middleware
- [x] T059 [P] [US1] Create page `frontend/app/(auth)/register/page.tsx` (RSC + form action, server validation echo, `<form action method=post>`)
- [x] T060 [P] [US1] Create page `frontend/app/(auth)/verify-email/sent/page.tsx` (informationnel, lien resend)
- [x] T061 [P] [US1] Create page `frontend/app/(auth)/verify-email/confirm/page.tsx` (server component, lit `?token=`, POST /api/v1/auth/verify-email, redirige vers /login?verified=1)
- [x] T062 [P] [US1] Add seed data `backend/scripts/seed.sql` with 1 verified test account for the rest of the stories

**Checkpoint US1**: parcours inscription → email → vérification fonctionnel, audit log contient `account.created` + `account.email_verified`, RGAA AAA validée par axe-core.

---

## Phase 4: User Story 2 - Connexion avec session persistante (Priority: P1) MVP

**Goal**: un compte vérifié se connecte, reçoit un cookie `__Host-session` 30 j, accède à `/home`.

**Independent Test** (cf. spec.md US2): seed compte → POST /api/v1/auth/login → cookie présent (HttpOnly+Secure+SameSite=Strict) → GET /api/v1/auth/me retourne profil.

### Tests for US2

- [ ] T063 [P] [US2] Domain test `backend/internal/domain/session/session_test.go`: TTL 30j, expired check, last_seen lazy update
- [ ] T064 [P] [US2] Application test `backend/internal/application/auth/login_test.go`: succès, mauvais mdp (réponse uniforme), compte inconnu (réponse uniforme), pending_verification, rate limit 5/IP/15min
- [ ] T065 [P] [US2] Application test `backend/internal/application/auth/me_test.go`: session valide, expirée, inconnue
- [ ] T066 [P] [US2] Adapter integration test `backend/internal/adapters/db/session_repository_test.go`: Create + FindByID (ignore expired) + TouchLastSeen + DeleteAllForAccount
- [ ] T067 [P] [US2] Adapter test `backend/internal/adapters/password/bcrypt_hasher_test.go`: cost 12 vérifié + timing entre 200ms et 500ms (skippé si CI dimensionnée différemment, marque en T-shirt)
- [ ] T068 [P] [US2] HTTP contract test `backend/internal/adapters/http/auth_handler_login_test.go`: 200 cookies posés, 401 uniforme, 403 pending, 429 rate limit, valide attributs cookies
- [ ] T069 [P] [US2] HTTP contract test `backend/internal/adapters/http/auth_handler_me_test.go`: 200 avec session, 401 sans cookie
- [ ] T070 [P] [US2] HTTP middleware test `backend/internal/adapters/http/middleware/require_auth_test.go`: charge la session, refuse expirée, met à jour last_seen
- [ ] T071 [P] [US2] Frontend a11y test `frontend/tests/unit/login.test.tsx`: axe-core 0 violations

### Implementation for US2

- [ ] T072 [US2] Implement `backend/internal/domain/session/session.go` to make T063 green
- [ ] T073 [P] [US2] Write SQL queries `backend/internal/adapters/db/queries/sessions.sql` (CreateSession, GetActiveSession, TouchLastSeen, DeleteSession, DeleteAllSessionsForAccount, DeleteExpiredSessions) and run `make sqlc`
- [ ] T074 [US2] Implement `backend/internal/adapters/db/session_repository.go` to make T066 green
- [ ] T075 [US2] Implement `backend/internal/application/auth/login.go` (verify password constant-time even si compte inexistant; create session; audit) to make T064 green
- [ ] T076 [US2] Implement `backend/internal/application/auth/me.go` (return current account by session) to make T065 green
- [ ] T077 [US2] Implement HTTP handler `POST /api/v1/auth/login` + `GET /api/v1/auth/me` in `backend/internal/adapters/http/auth_handler.go` (set cookies `__Host-session` + `__Host-csrf`) to make T068–T069 green
- [ ] T078 [US2] Implement middleware `backend/internal/adapters/http/middleware/require_auth.go` to make T070 green
- [ ] T079 [US2] Mount US2 routes + middleware in `backend/internal/adapters/http/router.go`
- [ ] T080 [P] [US2] Create page `frontend/app/(auth)/login/page.tsx` (RSC + form action, accepte ?next=)
- [ ] T081 [P] [US2] Create page `frontend/app/home/page.tsx` (RSC, lit la session via `lib/auth.ts`, redirige vers /login si non auth)

**Checkpoint US2** (= MVP démontrable au jury): `inscription → vérif email → login → /home → /me` end-to-end, RGAA AAA validée, audit log contient les événements attendus.

---

## Phase 5: User Story 3 - Réinitialisation de mot de passe (Priority: P2)

**Goal**: oubli de mot de passe → email avec lien → choix nouveau mot de passe → toutes les sessions révoquées → reconnexion.

**Independent Test** (cf. spec.md US3): seed → /reset-password → mail → /reset-password/confirm → connexion ancien mdp KO, nouveau OK.

### Tests for US3

- [ ] T082 [P] [US3] Application test `backend/internal/application/auth/request_password_reset_test.go`: anti-énumération (200 même si compte inconnu), rate limit 3/email/h, génère token TTL 30min
- [ ] T083 [P] [US3] Application test `backend/internal/application/auth/confirm_password_reset_test.go`: token valide → password changé + sessions révoquées + audit + email confirmation; token expiré/consommé/inconnu → erreurs typées
- [ ] T084 [P] [US3] HTTP contract test `backend/internal/adapters/http/auth_handler_password_reset_test.go`: 200 uniforme, 410 expiré, 400 token invalide
- [ ] T085 [P] [US3] Frontend a11y test `frontend/tests/unit/reset-password.test.tsx`: axe-core 0 violations sur les 2 pages

### Implementation for US3

- [ ] T086 [US3] Implement `backend/internal/application/auth/request_password_reset.go` to make T082 green
- [ ] T087 [US3] Implement `backend/internal/application/auth/confirm_password_reset.go` (UnitOfWork: consume token + update password_hash + delete all sessions + audit + send password_changed email) to make T083 green
- [ ] T088 [US3] Implement HTTP handlers `POST /api/v1/auth/password-reset/request` + `/confirm` in `backend/internal/adapters/http/auth_handler.go` to make T084 green
- [ ] T089 [US3] Mount US3 routes in `backend/internal/adapters/http/router.go`
- [ ] T090 [P] [US3] Create email template `backend/internal/adapters/email/templates/reset_password.html` (RGAA AAA)
- [ ] T091 [P] [US3] Create email template `backend/internal/adapters/email/templates/password_changed.html` (RGAA AAA, sans CTA)
- [ ] T092 [P] [US3] Create page `frontend/app/(auth)/reset-password/page.tsx` (demande reset)
- [ ] T093 [P] [US3] Create page `frontend/app/(auth)/reset-password/confirm/page.tsx` (consommation token, formulaire nouveau mot de passe)

**Checkpoint US3**: parcours reset complet, ancienne session du compte expirée immédiatement, audit log contient `auth.password_reset_requested` + `auth.password_changed`.

---

## Phase 6: User Story 4 - Déconnexion volontaire (Priority: P2)

**Goal**: bouton "Se déconnecter" → suppression session côté serveur + cookies effacés.

**Independent Test** (cf. spec.md US4): logout → /me retourne 401 → cookie absent → ré-injection ancien `session_id` → 401.

### Tests for US4

- [ ] T094 [P] [US4] Application test `backend/internal/application/auth/logout_test.go`: session supprimée en BDD, audit `auth.logout`
- [ ] T095 [P] [US4] HTTP contract test `backend/internal/adapters/http/auth_handler_logout_test.go`: 204/303 + Set-Cookie Max-Age=0; 401 sans session

### Implementation for US4

- [ ] T096 [US4] Implement `backend/internal/application/auth/logout.go` to make T094 green
- [ ] T097 [US4] Implement HTTP handler `POST /api/v1/auth/logout` in `backend/internal/adapters/http/auth_handler.go` to make T095 green
- [ ] T098 [US4] Mount US4 route + RequireAuth middleware in `backend/internal/adapters/http/router.go`
- [ ] T099 [P] [US4] Add logout button in `frontend/app/components/header.tsx` (form POST avec CSRF)

**Checkpoint US4**: parcours déconnexion fonctionnel, ancien `session_id` rejeté.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: hardening, observabilité, accessibilité approfondie, validation soutenance.

- [ ] T100 [P] Run `axe-core` + `lighthouse-ci` on `/register`, `/login`, `/verify-email/*`, `/reset-password/*` (target: a11y >= 90, perf >= 90, RGESN < 100KB critical path)
- [ ] T101 [P] Manual NVDA (Windows) or Orca (Linux) audit on the 4 parcours, capture vidéo dans `docs/specs/04-quality/rgaa-audit-auth.md`
- [ ] T102 [P] Add Playwright E2E `frontend/tests/e2e/auth.spec.ts` covering the 4 user stories
- [ ] T103 [P] Add CI grep test ensuring no password/email/token appears in `slog` outputs (1000 simulated requests in `backend/scripts/grep_logs_test.sh`)
- [ ] T104 [P] Document deviations from quickstart.md in `specs/001-user-auth/quickstart.md` "Dépannage"
- [ ] T105 Add cron job `backend/cmd/jobs/cleanup_sessions/main.go` calling `SessionRepository.DeleteExpired` (run daily, post-MVP)
- [ ] T106 [P] Add ADR-0015 in `docs/adr/0015-rate-limit-token-bucket.md` documenting D-008 implementation choices
- [ ] T107 Run `quickstart.md` smoke test end-to-end and tick its done-criteria checklist
- [ ] T108 Update `docs/specs/03-api/openapi.yaml` (consolidate `contracts/openapi.auth.yaml` content)
- [ ] T109 Update `docs/specs/02-uml/sequence-auth.puml` if implementation diverged from initial diagram
- [ ] T110 Update memory file `MEMORY.md` and `project_certif_state.md` with feature completion status

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: aucune dépendance — peut commencer immédiatement.
- **Foundational (Phase 2)**: dépend de Setup — bloque toutes les US.
- **US1 (Phase 3)** = MVP partie 1 — dépend de Foundational.
- **US2 (Phase 4)** = MVP partie 2 — dépend de Foundational ; peut s'exécuter en parallèle de US1, mais validation E2E dépend des deux.
- **US3 (Phase 5)**: dépend de Foundational + US1 (compte vérifié requis pour le test E2E) + US2 (login pour vérifier l'invalidation).
- **US4 (Phase 6)**: dépend de US2 (session existante à fermer).
- **Polish (Phase 7)**: dépend de toutes les US livrées.

### TDD Ordering (within each user story)

1. **Tests d'abord** (vérifier rouge `make test-domain` / `test-application` / etc.)
2. **Domain** (purs, sans IO)
3. **Adapters DB** (avec sqlc + migrations)
4. **Adapters externes** (email, password, clock — plupart faits en Foundational)
5. **Application use cases** (utilisent ports)
6. **HTTP handlers** (sérialisation, status codes)
7. **Frontend pages** (RSC + forms)

### Parallel Opportunities

- Setup: T002–T009 quasiment tous parallélisables (différents fichiers).
- Foundational: T013–T029 sont [P] modulo T010 (migration must run first).
- Tests US1: T030–T042 indépendants → tous [P] sur des fichiers distincts.
- Tests US2: T063–T071 idem.
- Frontend pages d'une même US: parallélisables (T059–T061, T080–T081, T092–T093).

---

## Parallel Example: Phase 2 Foundational

```bash
# Après T010 + T012 (DB infra), lancer en parallèle :
Task: "T013 Domain Account skeleton"
Task: "T014 Domain Session skeleton"
Task: "T015 Domain Token skeleton"
Task: "T016 Domain Audit skeleton"
Task: "T017 All port interfaces"
Task: "T018 SystemClock"
Task: "T019 BcryptHasher"
Task: "T020 UUIDTokenGen"
Task: "T021 RateLimiter in-memory"
Task: "T022 SecurityHeaders middleware"
Task: "T023 CSRF middleware"
Task: "T024 RateLimit middleware"
Task: "T025 slog with redaction"
```

## Parallel Example: Tests US1 (TDD red)

```bash
Task: "T030 account_test.go"
Task: "T031 password_test.go"
Task: "T032 token_test.go"
Task: "T033 register_test.go"
Task: "T034 verify_email_test.go"
Task: "T035 resend_verification_test.go"
Task: "T036 account_repository_test.go"
Task: "T037 token_repository_test.go"
Task: "T038 audit_repository_test.go (asserts triggers)"
Task: "T039 brevo_test.go"
Task: "T040 auth_handler_register_test.go"
Task: "T041 auth_handler_verify_email_test.go"
Task: "T042 register.test.tsx (axe)"
```

---

## Implementation Strategy

### MVP First (US1 + US2)

1. Compléter Phase 1 + 2 (Setup + Foundational) — environ 2 jours solo.
2. Compléter Phase 3 (US1 inscription) — TDD rigoureux, environ 3 jours.
3. Compléter Phase 4 (US2 login) — environ 2 jours.
4. **STOP & VALIDATE**: tester end-to-end, capture vidéo soutenance.
5. Demo possible: parcours inscription + login fonctionnel = première vertical complète, prouve l'architecture hexagonale.

### Incremental Delivery

1. Setup + Foundational → infra prête.
2. US1 → demo "création de compte" (mergé sur main via PR + Option A admin merge).
3. US2 → demo "connexion" (PR séparé ou inclus selon timing).
4. US3 → demo "reset password".
5. US4 → demo "logout" (souvent tout petit, fusionnable avec US3 dans un seul PR).
6. Polish → audit RGAA + Playwright E2E.

### Solo dev cadence

- 1 commit `test(auth): …` par tâche de test, suivi d'un commit `feat(auth): …` quand le test passe vert.
- 1 PR par US (ou 1 PR US1+US2 pour le MVP, 1 PR US3+US4 pour la complétion).
- Option A admin merge avec ≥ J+1 entre ouverture et merge (cf. constitution VI).

---

## Notes

- Test-first NON optionnel : la Constitution principe III impose un commit `test(scope):` antérieur au commit `feat(scope):` correspondant pour toute modification de `internal/domain` ou `internal/application`.
- Les tâches T036–T039 et T066–T067 sont des "tests d'intégration" (utilisent SQLite fichier temp / mock HTTP) et restent rapides (≤ 200 ms chacune).
- L'exception `[no-test-first: motif]` reste possible pour les pures wirings (ex: T026 composition root) mais doit apparaître dans le commit message.
- Aucun secret en clair dans les tests : `KOMUNUMO_BREVO_API_KEY=test-key-noop` partout.
- Pas de mock de la DB dans les tests `application/auth/*` — utiliser des fakes en mémoire des ports (cf. fakes pattern, à coder une seule fois en début de Phase 3 sous `backend/internal/ports/fakes/`).

**Total: 110 tâches** (T001–T110).
