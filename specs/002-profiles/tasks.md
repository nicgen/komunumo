# Tasks: Profils & Types de compte

**Input**: Design documents from `/specs/002-profiles/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓, quickstart.md ✓
**Branch**: `feat/002-profiles`

**Tests**: Inclus — Constitution Principe III impose test-first. Chaque commit `feat(profiles):` doit être précédé d'un commit `test(profiles):` vérifiable dans `git log`.

**Organization**: Tasks groupées par User Story. Foundation obligatoire avant toute US.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Parallélisable (fichiers distincts, sans dépendances non satisfaites)
- **[Story]**: US concernée (US1–US4)
- Chemins absolus depuis la racine du repo

---

## Phase 1: Setup (Migration + Requêtes sqlc)

**Purpose**: Poser les fichiers SQL (migration + queries) qui débloquent le reste. Aucun code Go ni TypeScript ici.

- [X] T001 Écrire `backend/internal/adapters/db/migrations/0002_profiles.up.sql` (recréation accounts, création members/associations/memberships, migration PII, index — cf. data-model.md §Migration 0002)
- [X] T002 Écrire `backend/internal/adapters/db/migrations/0002_profiles.down.sql` (rollback : DROP membres/associations/memberships, recréation accounts Phase 1 avec colonnes PII)
- [X] T003 Écrire `backend/internal/adapters/db/queries/members.sql` (-- name: CreateMember, GetMemberByAccountID, UpdateMember)
- [X] T004 [P] Écrire `backend/internal/adapters/db/queries/associations.sql` (-- name: CreateAssociation, GetAssociationByAccountID, UpdateAssociation)
- [X] T005 [P] Écrire `backend/internal/adapters/db/queries/memberships.sql` (-- name: CreateMembership, GetMembershipByAccountIDs)
- [X] T006 Régénérer le code sqlc : `cd backend && sqlc generate` (dépend de T003–T005)
- [X] T007 Appliquer la migration localement : `migrate -database "sqlite3://./backend/data/assolink.db" -path backend/internal/adapters/db/migrations up` — vérifier `.tables` + `SELECT count(*) FROM members`

**Checkpoint**: tables créées, sqlc généré → code Go compilable.

---

## Phase 2: Foundational (Domaine + Ports + Adapters DB)

**Purpose**: Entités domaine, interfaces de ports et adapters DB. Bloque toutes les US.

**⚠️ CRITIQUE**: Aucune US ne peut démarrer avant la fin de cette phase.

- [X] T008 Mettre à jour `backend/internal/domain/account/account.go` : ajouter types `Kind` (member/association) et `Status` (active/suspended/deleted + pending_verification) ; mise à jour de `Account` struct (supprimer FirstName/LastName/DateOfBirth, ajouter Kind)
- [X] T009 [P] Écrire les tests `backend/internal/domain/member/member_test.go` : NewMember (âge ≥ 18 ans OK, < 18 ans → erreur), about_me > 500 → erreur, FirstName vide → erreur
- [X] T010 [P] Créer `backend/internal/domain/member/member.go` : struct Member, func NewMember(accountID, firstName, lastName, birthDate string) (*Member, error) + invariants âge + about_me (dépend de T009 — tests must fail first)
- [X] T011 [P] Écrire les tests `backend/internal/domain/association/association_test.go` : ValidateSIREN (9 chiffres OK, 8 chiffres KO), ValidateRNA (W+9 OK, mauvais format KO), about > 2000 → erreur
- [X] T012 [P] Créer `backend/internal/domain/association/association.go` : struct Association, func NewAssociation(...) + ValidateSIREN/ValidateRNA (dépend de T011)
- [X] T013 Ajouter `backend/internal/ports/member_repository.go` : interface MemberRepository { Create, FindByAccountID, Update }
- [X] T014 [P] Ajouter `backend/internal/ports/association_repository.go` : interface AssociationRepository { Create, FindByAccountID, Update }
- [X] T015 [P] Ajouter `backend/internal/ports/membership_repository.go` : interface MembershipRepository { Create, FindByAccountIDs }
- [X] T016 [P] Ajouter `backend/internal/ports/file_store.go` : interface FileStore { StoreAvatar, AvatarURL } (cf. contracts/ports.md)
- [X] T017 Étendre `backend/internal/ports/account_repository.go` : ajouter FindByID et UpdateKindAndStatus à l'interface AccountRepository
- [X] T018 [P] Créer les fakes `backend/internal/ports/fakes/member_repository.go`, `association_repository.go`, `membership_repository.go`, `file_store.go` (fake en mémoire, un fichier par interface)
- [X] T019 Mettre à jour `backend/internal/ports/fakes/account_repository.go` : implémenter FindByID et UpdateKindAndStatus
- [X] T020 Créer `backend/internal/adapters/db/member_repository.go` : implémentation sqlc de MemberRepository (dépend de T006, T013)
- [X] T021 [P] Créer `backend/internal/adapters/db/association_repository.go` : implémentation sqlc de AssociationRepository (dépend de T006, T014)
- [X] T022 [P] Créer `backend/internal/adapters/db/membership_repository.go` : implémentation sqlc de MembershipRepository (dépend de T006, T015)
- [X] T023 Étendre `backend/internal/adapters/db/account_repository.go` : implémenter FindByID et UpdateKindAndStatus (dépend de T017)
- [X] T024 Créer `backend/internal/adapters/storage/local_file_store.go` : implémentation FileStore — StoreAvatar écrit dans `data/uploads/avatars/{accountID}/{uuid}.{ext}`, AvatarURL retourne `/uploads/avatars/…` (dépend de T016)

**Checkpoint**: `go build ./...` passe, fakes compilent, domaines testés.

---

## Phase 3: US1 — Inscription Personne (Priority: P1) 🎯 MVP

**Goal**: `POST /api/v1/auth/register/member` fonctionnel de bout en bout (backend + frontend).

**Independent Test**:
```bash
curl -s -X POST http://localhost:8080/api/v1/auth/register/member \
  -H "Content-Type: application/json" \
  -d '{"email":"lea@test.com","password":"Password1234!","first_name":"Léa","last_name":"Martin","birth_date":"2000-01-15"}' | jq .
# → 201 Created
curl -s -X POST http://localhost:8080/api/v1/auth/register/member \
  -d '{"email":"young@test.com","password":"Password1234!","first_name":"A","last_name":"B","birth_date":"2015-01-01"}' | jq .
# → 422 "vous devez avoir au moins 18 ans"
```

### Tests — US1

- [X] T025 [US1] Écrire `backend/internal/application/auth/register_member_test.go` : cas OK (account kind=member + member row créés, email envoyé, audit_log), âge < 18 → ErrTooYoung, email dupliqué → ErrEmailTaken, password faible → ErrWeakPassword

### Implémentation — US1

- [X] T026 [US1] Créer `backend/internal/application/auth/register_member.go` : struct RegisterMemberService + func RegisterMember(ctx, ip, RegisterMemberInput) error — scission de register.go Phase 1 (ne pas modifier register.go tant que les tests ne passent pas) (dépend de T025)
- [X] T027 [US1] Écrire `backend/internal/adapters/http/register_handler_test.go` (section member) : POST /auth/register/member 201, 400 JSON malformé, 422 âge, 429 rate-limit
- [X] T028 [US1] Créer `backend/internal/adapters/http/register_handler.go` : func HandleRegisterMember — JSON decode, appel RegisterMemberService, 201/400/422/429 (dépend de T027)
- [X] T029 [US1] Mettre à jour `backend/internal/adapters/http/router.go` : enregistrer POST /api/v1/auth/register/member
- [X] T030 [P] [US1] Créer `frontend/components/auth/register-member-form.tsx` : formulaire Zod+RHF (email, password, first_name, last_name, birth_date) avec aria-describedby sur chaque champ d'erreur (pattern Phase 1)
- [X] T031 [P] [US1] Créer `frontend/app/(auth)/register/member/page.tsx` : page d'inscription Personne (utilise register-member-form)

**Checkpoint**: US1 complète — inscription personne ≥ 18 ans fonctionne, < 18 → 422, email dupliqué → 409.

---

## Phase 4: US2 — Inscription Association (Priority: P2)

**Goal**: `POST /api/v1/auth/register/association` + page sélection du type de compte.

**Independent Test**:
```bash
curl -s -X POST http://localhost:8080/api/v1/auth/register/association \
  -H "Content-Type: application/json" \
  -d '{"email":"asso@test.com","password":"Password1234!","legal_name":"Les Amis du Code","postal_code":"75011","first_name":"Anne","last_name":"Dupont","birth_date":"1985-06-20"}' | jq .
# → 201 Created (accounts kind=association, associations row, memberships role=owner)
curl -s -X POST http://localhost:8080/api/v1/auth/register/association \
  -d '{"email":"bad@test.com","password":"Password1234!","legal_name":"Test","postal_code":"75011","siren":"12345","first_name":"A","last_name":"B","birth_date":"1990-01-01"}' | jq .
# → 422 "siren must be 9 digits"
```

### Tests — US2

- [x] T032 [US2] Écrire `backend/internal/application/auth/register_association_test.go` : cas OK (account kind=association + associations + memberships role=owner créés), SIREN invalide → ErrInvalidSIREN, RNA invalide → ErrInvalidRNA, âge < 18 → ErrTooYoung, postal_code manquant → ErrValidation

### Implémentation — US2

- [x] T033 [US2] Créer `backend/internal/application/auth/register_association.go` : struct RegisterAssociationService + func RegisterAssociation(ctx, ip, RegisterAssociationInput) error — crée account(kind=association) + association + membership(role=owner, status=active) en transaction (dépend de T032)
- [x] T034 [US2] Écrire tests dans `backend/internal/adapters/http/register_handler_test.go` (section association) : 201, 422 SIREN, 422 âge, 400 champs manquants
- [x] T035 [US2] Ajouter `HandleRegisterAssociation` dans `backend/internal/adapters/http/register_handler.go` (dépend de T034)
- [x] T036 [US2] Mettre à jour `backend/internal/adapters/http/router.go` : enregistrer POST /api/v1/auth/register/association
- [x] T037 [P] [US2] Créer `frontend/app/(auth)/register/page.tsx` : page de sélection — deux boutons "Je suis une Personne" et "Je suis une Association" (liens vers /register/member et /register/association)
- [x] T038 [P] [US2] Créer `frontend/components/auth/register-association-form.tsx` : formulaire Zod+RHF (email, password, legal_name, postal_code, siren?, rna?, first_name, last_name, birth_date) avec aria-describedby
- [x] T039 [P] [US2] Créer `frontend/app/(auth)/register/association/page.tsx` (utilise register-association-form)

**Checkpoint**: US1 + US2 complètes — les deux parcours d'inscription fonctionnent indépendamment.

---

## Phase 5: US3 — Profil connecté (Priority: P3)

**Goal**: `GET /api/v1/me/profile`, `PATCH /api/v1/me/profile` et `POST /api/v1/me/avatar` fonctionnels.

**Independent Test**:
```bash
TOKEN=$(curl -sc /tmp/cookies http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"lea@test.com","password":"Password1234!"}' -o /dev/null)
curl -sb /tmp/cookies http://localhost:8080/api/v1/me/profile | jq .
# → 200 { kind: "member", first_name: "Léa", ... }
curl -sb /tmp/cookies -X PATCH http://localhost:8080/api/v1/me/profile \
  -H "Content-Type: application/json" \
  -d '{"nickname":"lea42","about_me":"Passionnée de code"}' | jq .
# → 200 + audit_log("profile.updated")
```

### Tests — US3

- [X] T040 [US3] Écrire `backend/internal/application/profile/get_profile_test.go` : GetMyProfile retourne MemberProfile ou AssociationProfile selon kind, session invalide → ErrUnauthorized
- [X] T041 [P] [US3] Écrire `backend/internal/application/profile/update_profile_test.go` : UpdateProfile met à jour nickname+about_me, about_me > 500 → erreur, audit_log("profile.updated") appelé
- [X] T042 [P] [US3] Écrire `backend/internal/application/profile/upload_avatar_test.go` : UploadAvatar stocke le fichier et retourne l'URL, taille > 2 Mo → erreur, MIME invalide → erreur

### Implémentation — US3

- [X] T043 [US3] Créer `backend/internal/application/profile/get_profile.go` : struct GetProfileService + func GetMyProfile(ctx, sessionID) (ProfileOutput, error) — discriminé par kind (dépend de T040)
- [X] T044 [US3] Créer `backend/internal/application/profile/update_profile.go` : func UpdateProfile(ctx, sessionID, UpdateProfileInput) error + audit_log (dépend de T041)
- [X] T045 [US3] Créer `backend/internal/application/profile/upload_avatar.go` : func UploadAvatar(ctx, sessionID, io.Reader, size, mimeType) (string, error) — délègue à FileStore (dépend de T042)
- [X] T046 [US3] Écrire `backend/internal/adapters/http/profile_handler_test.go` : GET /me/profile 200 (member+asso), PATCH /me/profile 200, 401 sans session, POST /me/avatar 200
- [X] T047 [US3] Créer `backend/internal/adapters/http/profile_handler.go` : HandleGetMyProfile + HandleUpdateMyProfile + HandleUploadAvatar (dépend de T046)
- [X] T048 [US3] Mettre à jour `backend/internal/adapters/http/router.go` : GET/PATCH /api/v1/me/profile, POST /api/v1/me/avatar
- [X] T049 [P] [US3] Créer `frontend/components/profile/member-profile-form.tsx` : formulaire édition profil Personne (nickname, about_me, visibility) — aria-describedby RGAA
- [X] T050 [P] [US3] Créer `frontend/components/profile/association-profile-form.tsx` : formulaire édition profil Association (about, visibility, postal_code) — aria-describedby RGAA
- [X] T051 [US3] Créer `frontend/app/profile/page.tsx` : page profil connecté — charge GET /me/profile, affiche le bon formulaire selon kind, soumet PATCH

**Checkpoint**: US3 complète — un utilisateur connecté peut voir et modifier son profil ; avatar uploadable.

---

## Phase 6: US4 — Profil public (Priority: P4)

**Goal**: `GET /api/v1/accounts/{id}/profile` respectant la visibilité (public → 200, private → 404).

**Independent Test**:
```bash
# Profil public
curl -s http://localhost:8080/api/v1/accounts/{ID_LEA}/profile | jq .
# → 200 sans birth_date
# Profil privé (après PATCH visibility=private)
curl -s http://localhost:8080/api/v1/accounts/{ID_PRIVE}/profile | jq .
# → 404
```

### Tests — US4

- [X] T052 [US4] Ajouter dans `backend/internal/application/profile/get_profile_test.go` : GetPublicProfile visibility=public → ProfileOutput sans birth_date, visibility=private → ErrNotFound, members_only sans session → ErrNotFound

### Implémentation — US4

- [X] T053 [US4] Ajouter `GetPublicProfile(ctx, accountID, viewerSessionID string) (ProfileOutput, error)` dans `backend/internal/application/profile/get_profile.go` (dépend de T052)
- [X] T054 [US4] Ajouter dans `backend/internal/adapters/http/profile_handler_test.go` : GET /accounts/{id}/profile 200 public, 404 private
- [X] T055 [US4] Ajouter `HandleGetPublicProfile` dans `backend/internal/adapters/http/profile_handler.go` (dépend de T054)
- [X] T056 [US4] Mettre à jour `backend/internal/adapters/http/router.go` : GET /api/v1/accounts/{accountId}/profile

**Checkpoint**: Toutes les US sont fonctionnelles. Les 4 flux inscription/profil sont testables de bout en bout.

---

## Phase 7: Polish & Cross-Cutting

**Purpose**: Alignement avec les specs transverses et validation finale.

- [X] T057 Ajouter le champ `kind` à la réponse de `GET /api/v1/auth/me` dans `backend/internal/application/auth/me.go` (R-007 — non-breaking, le champ est absent en Phase 1)
- [X] T058 [P] Vérifier RGAA AAA sur `register-member-form.tsx`, `register-association-form.tsx`, `member-profile-form.tsx`, `association-profile-form.tsx` : aria-label, aria-describedby, role, focus-visible présents sur tous les champs et boutons
- [X] T059 [P] Ajouter les tests d'intégration DB pour `member_repository.go`, `association_repository.go`, `membership_repository.go` dans `backend/internal/adapters/db/` (pattern existant avec testhelper_test.go)
- [X] T060 Exécuter le smoke test complet `quickstart.md` : migrate up + sqlc generate + go test -race + curl des 4 endpoints principaux
- [X] T061 Vérifier que `audit_log` contient bien `account_created` (US1+US2) et `profile.updated` (US3) via `sqlite3 data/assolink.db "SELECT * FROM audit_log ORDER BY created_at DESC LIMIT 10;"`

---

## Dependencies & Execution Order

### Dépendances de phases

- **Phase 1** (Setup): Aucune dépendance — démarre immédiatement.
- **Phase 2** (Foundational): Dépend de T006 (sqlc generate) — **bloque toutes les US**.
- **US1 (Phase 3)**: Dépend de Phase 2 complète. P1 → démarre en premier.
- **US2 (Phase 4)**: Dépend de Phase 2 complète. Peut démarrer en parallèle d'US1 côté frontend, mais l'application layer dépend des mêmes repos foundationnels.
- **US3 (Phase 5)**: Dépend de Phase 2 + US1/US2 (les profils sont créés lors de l'inscription).
- **US4 (Phase 6)**: Dépend de US3 (GetPublicProfile étend GetProfile).
- **Polish (Phase 7)**: Dépend de toutes les US.

### Dépendances intra-phase

- T006 dépend de T003, T004, T005
- T007 dépend de T006
- T010 dépend de T009 (tests must fail first)
- T012 dépend de T011 (tests must fail first)
- T020–T022 dépendent de T006
- T026 dépend de T025 (test-first)
- T028 dépend de T027 (test-first)
- T033 dépend de T032 (test-first)
- T035 dépend de T034 (test-first)
- T043–T045 dépendent de leurs tests respectifs (T040–T042)
- T047 dépend de T046 (test-first)
- T053 dépend de T052 (test-first)
- T055 dépend de T054 (test-first)

### Opportunités de parallélisation

```bash
# Phase 1 — en parallèle après T001/T002 :
T003 members.sql | T004 associations.sql | T005 memberships.sql

# Phase 2 — en parallèle (fichiers distincts) :
T009+T010 (domain/member) | T011+T012 (domain/association)
T013 | T014 | T015 | T016 (nouveaux ports)
T020 | T021 | T022 (DB adapters, après sqlc)
T018 (fakes) — dès que ports définis

# US1 + US2 — parallèle backend/frontend :
T025→T026→T027→T028 (backend US1) || T030+T031 (frontend US1)
T032→T033→T034→T035 (backend US2) || T037+T038+T039 (frontend US2)

# US3 — tests en parallèle :
T040 | T041 | T042
```

---

## Parallel Example: Phase 2 Foundational

```bash
# Lancer en parallèle :
Agent: "Écrire member_test.go + member.go (T009–T010)"
Agent: "Écrire association_test.go + association.go (T011–T012)"
Agent: "Ajouter les 4 ports interfaces (T013–T016)"

# Puis, après sqlc generate :
Agent: "Créer member_repository.go (T020)"
Agent: "Créer association_repository.go (T021)"
Agent: "Créer membership_repository.go (T022)"
```

---

## Implementation Strategy

### MVP First (US1 uniquement)

1. Compléter Phase 1 (SQL + sqlc)
2. Compléter Phase 2 (domain + ports + adapters)
3. Compléter Phase 3 : US1
4. **STOP et VALIDER** : smoke test inscription personne
5. Merger sur dev si OK

### Incremental Delivery

1. Phase 1 + Phase 2 → Foundation prête
2. US1 → inscription personne testée → valider
3. US2 → inscription association testée → valider
4. US3 → profil connecté testé → valider
5. US4 → profil public testé → valider
6. Polish → PR finale vers dev

---

## Notes

- **Test-first obligatoire** : chaque `feat(profiles):` commit doit être précédé d'un `test(profiles):` commit dans `git log` (Constitution Principe III).
- **Migration atomique** : `0002_profiles.up.sql` doit rouler dans une seule transaction BEGIN/COMMIT.
- **PII protégé** : `birth_date` ne doit jamais apparaître dans les réponses publiques (US4 — `GetPublicProfile` exclut ce champ).
- **Pas d'AVIF** : `local_file_store.go` stocke l'original uniquement (JPEG/PNG/WebP), sans processing (V1).
- **Fake pattern** : reproduire le pattern existant `ports/fakes/*.go` pour les nouveaux fakes.
- **register.go Phase 1** : ne pas modifier ni supprimer tant que `register_member.go` n'est pas testé et opérationnel — scission après les tests verts.
