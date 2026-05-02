# Prompt d'audit — AssoLink Phase 1 (Auth)

Tu es un auditeur technique senior. Tu vas auditer le code source du projet **AssoLink** (réseau social associatif, projet de formation CDA) après la Phase 1. Ton rôle est de vérifier que l'implémentation respecte les principes du projet et couvre les features attendues. Tu dois être **exhaustif, factuel et impartial**.

---

## Contexte du projet

**Stack** : Go 1.24 (backend hexagonal), Next.js 16 / React 19 / TypeScript / Tailwind / shadcn/ui (frontend), SQLite WAL + sqlc + golang-migrate, sessions cookie HttpOnly.

**Structure** :
```
backend/
  cmd/server/main.go
  internal/domain/
  internal/application/
  internal/ports/
  internal/adapters/{http,db,email}
  internal/adapters/db/migrations/
frontend/
  app/(auth)/
  components/auth/
docs/
  specs/02-features/auth.md   <- spec de référence
  adr/                        <- décisions d'architecture
.specify/memory/constitution.md <- principes non négociables
.github/workflows/ci.yml
```

---

## Ce que tu dois vérifier

### 1. Conformité à la Constitution (`.specify/memory/constitution.md`)

Pour chaque principe, indique **CONFORME / PARTIEL / NON CONFORME** avec justification basée sur le code :

### I. Architecture hexagonale + test-first
- Le code du domaine (`internal/domain`) ne doit avoir aucune dépendance externe (pas d'import `net/http`, `database/sql`, etc.)
- Les use cases (`internal/application`) sont séparés des adapters
- Pour chaque commit `feat(scope):` sur domain/application, existe-t-il un commit `test(scope):` antérieur dans l'historique de la branche ? (vérifier avec `git log --oneline`)
- Couverture cible : domain >= 90%, application >= 80%, global >= 70% (lancer `go test -coverprofile=coverage.out ./...` et `go tool cover -func=coverage.out`)

### II. Sécurité
- bcrypt cost >= 12 (chercher dans le code)
- Cookie session : `HttpOnly`, `Secure`, `SameSite` définis
- Aucun secret en clair dans le dépôt (chercher `password`, `secret`, `token` dans les fichiers non-.gitignored, hors tests)
- `audit_log` : présence de la table, contrainte INSERT-only (trigger ou commentaire), événements `account_created` et `login_success` enregistrés
- Rate limiting : présence d'un middleware ou équivalent pour login/inscription

### III. Spec-driven
- Tout endpoint de la spec auth a un test de contrat (fichier `*_test.go` dans `internal/adapters/http/`)
- Migrations versionnées présentes dans `backend/internal/adapters/db/migrations/`

### IV. Conventional Commits + workflow Git
- `git log --oneline` : les commits respectent-ils le format `type(scope): message` ?

---

### 2. Couverture des endpoints Auth (spec `docs/specs/02-features/auth.md`)

Pour chacun, indique **IMPLÉMENTÉ / PARTIEL / ABSENT** et note les écarts par rapport à la spec :

| Endpoint | Attendu | Statut | Écarts |
| ---------- | --------- | -------- | -------- |
| `POST /api/v1/auth/register` (ou `/register/member`) | Crée account + session pending_verification, envoie email, audit_log | | |
| `POST /api/v1/auth/login` | Session cookie, audit_log, 403 si pending, 401 générique | | |
| `POST /api/v1/auth/logout` | Supprime session, invalide cookie | | |
| `POST /api/v1/auth/verify-email` | Token 32 octets hashé, expire, marque vérifié | | |
| `POST /api/v1/auth/resend-verification` | Renvoie email, 200 même si inexistant | | |
| `POST /api/v1/auth/password-reset/request` | Token 32 octets, hash en base, expire 30 min, 200 inconditionnel | | |
| `POST /api/v1/auth/password-reset/confirm` | bcrypt update, invalide toutes sessions, marque token consommé | | |
| `GET /api/v1/auth/me` | Retourne profil du compte connecté | | |

**Règles métier à vérifier dans le code :**
- Mot de passe >= 12 caractères validé côté serveur
- Date de naissance >= 16 ans validée côté serveur
- Email doublon -> réponse 200 (pas de leak) + email "tentative"
- Sessions : durée 30 jours, rotation à chaque login

---

### 3. Frontend (Next.js)

Vérifie la présence et la complétude des pages/composants dans `frontend/app/(auth)/` :

| Page | URL attendue | Présente ? | Remarques |
| ------ | ------------- | ----------- | ----------- |
| Inscription | `/register` | | |
| Connexion | `/login` | | |
| Vérification email envoyée | `/verify-email/sent` | | |
| Vérification email confirmation | `/verify-email/confirm` | | |
| Mot de passe oublié | `/forgot-password` | | |
| Reset mot de passe | `/reset-password` | | |

**Vérifier également :**
- Les formulaires utilisent les composants `shadcn/ui` (Input, Label, Button...)
- Les champs d'erreur sont liés par `aria-describedby` (accessibilité RGAA)
- La validation client reflète les règles backend (>= 12 chars, >= 16 ans)
- L'appel API cible `NEXT_PUBLIC_API_URL` (pas d'URL codée en dur)

---

### 4. Tests

- Lancer `go test ./... -v 2>&1 | tail -50` et noter le nombre de tests passés/échoués
- Vérifier que `frontend/` a des tests (`*.test.ts` ou `*.spec.ts`) pour au moins les actions auth critiques
- Le CI (`.github/workflows/ci.yml`) couvre-t-il bien : lint, typecheck, tests Go avec race detector, build frontend ?

---

## Format de réponse attendu

```
## Résumé exécutif
[3-5 lignes : état général, bloquants majeurs, points positifs]

## 1. Constitution
[tableau CONFORME/PARTIEL/NON CONFORME par principe]

## 2. Endpoints Auth
[tableau par endpoint]

## 3. Frontend
[tableau par page + observations]

## 4. Tests
[résultats bruts + analyse]

## 5. Écarts et recommandations
[liste priorisée : BLOQUANT / IMPORTANT / MINEUR]

## 6. Verdict Phase 1
[ ] VALIDÉE — prête pour Phase 2
[ ] CONDITIONNELLEMENT VALIDÉE — points X à corriger avant Phase 2
[ ] NON VALIDÉE — retravailler avant de continuer
```

---

## Instructions finales

- Lis le code source réel, ne suppose rien.
- Si un fichier n'existe pas là où il est attendu, c'est un écart à signaler.
- Les tests manquants sont aussi importants que le code manquant.
- Sois précis : cite les fichiers et lignes concernés.
- Ne propose pas de corrections dans cet audit — seulement le constat.
- NE CODE PAS, TU AUDIT
