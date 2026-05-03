# Quickstart — Phase 2 Profils

**Date**: 2026-05-02 | **Branch**: `feat/002-profiles`

## Prérequis

- Go 1.24, `golang-migrate` CLI installé (`go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest`)
- `sqlc` installé (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)
- Node.js 22 + pnpm 9

## Backend

### 1. Appliquer la migration

```bash
cd backend
migrate -database "sqlite3://./data/assolink.db" -path internal/adapters/db/migrations up
```

Vérifier que les tables `members`, `associations`, `memberships` existent et que les données Phase 1 sont migrées :

```bash
sqlite3 data/assolink.db ".tables"
sqlite3 data/assolink.db "SELECT count(*) FROM members;"
sqlite3 data/assolink.db "SELECT id, status, kind FROM accounts LIMIT 5;"
```

### 2. Régénérer le code sqlc

```bash
cd backend
sqlc generate
```

### 3. Lancer les tests

```bash
cd backend
go test -race ./... 2>&1 | grep -v scratch
```

Couverture cible : domain ≥ 90 %, application ≥ 80 %, global ≥ 70 %.

### 4. Démarrer le serveur

```bash
cd backend
go run ./cmd/server
# API sur http://localhost:8080
```

### 5. Smoke test curl

```bash
# Inscription Personne
curl -s -X POST http://localhost:8080/api/v1/auth/register/member \
  -H "Content-Type: application/json" \
  -d '{"email":"lea@test.com","password":"Password1234!","first_name":"Léa","last_name":"Martin","birth_date":"2000-01-15"}' | jq .

# Inscription Association
curl -s -X POST http://localhost:8080/api/v1/auth/register/association \
  -H "Content-Type: application/json" \
  -d '{"email":"asso@test.com","password":"Password1234!","legal_name":"Les Amis du Code","postal_code":"75011","first_name":"Anne","last_name":"Dupont","birth_date":"1985-06-20"}' | jq .

# Login + profil
TOKEN=$(curl -sc /tmp/cookies http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"lea@test.com","password":"Password1234!"}' -o /dev/null)
curl -sb /tmp/cookies http://localhost:8080/api/v1/me/profile | jq .
```

## Frontend

```bash
cd frontend
pnpm install
pnpm dev
# App sur http://localhost:3000
```

Pages à vérifier :
- `/register` → sélection du type de compte
- `/register/member` → formulaire Personne
- `/register/association` → formulaire Association
- `/profile` (connecté) → profil éditable

## Rollback migration

```bash
cd backend
migrate -database "sqlite3://./data/assolink.db" -path internal/adapters/db/migrations down 1
```
