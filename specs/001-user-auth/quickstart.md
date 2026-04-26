# Phase 1 - Quickstart : feature `auth` en local

Démarrage minimal pour développer/tester la feature `auth` (backend Go + frontend Next.js) sur la branche `feat/001-user-auth`.

## Prérequis

| Outil | Version cible | Notes |
|-------|--------------|-------|
| Go | 1.24+ | `go version` |
| Node.js | 22 LTS | gérer via `nvm` ; Next.js 16 minimum |
| pnpm | 9+ | `npm i -g pnpm` |
| `golang-migrate` | 4.17+ | `go install -tags 'sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@latest` |
| `sqlc` | 1.27+ | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |
| `mkcert` | 1.4+ | TLS local pour `*.local.hello-there.net` |
| Brevo API key | — | Cf. ADR-0012, créer compte gratuit, mettre dans `.env.local` |

## Variables d'environnement

`backend/.env` (jamais commité) :

```env
KOMUNUMO_DB_PATH=./var/komunumo.db
KOMUNUMO_BCRYPT_COST=12
KOMUNUMO_SESSION_TTL=720h           # 30 jours
KOMUNUMO_VERIFICATION_TTL=24h
KOMUNUMO_RESET_TTL=30m
KOMUNUMO_BREVO_API_KEY=xkeysib-…
KOMUNUMO_BREVO_FROM=hello@komunumo.fr
KOMUNUMO_BASE_URL=https://app.local.hello-there.net
KOMUNUMO_LISTEN_ADDR=:8080
KOMUNUMO_TRUST_PROXY=true            # derrière Traefik en dev
```

`frontend/.env.local` :

```env
NEXT_PUBLIC_APP_NAME=Komunumo
KOMUNUMO_API_INTERNAL_URL=http://localhost:8080
```

## Première installation

```bash
# 1. Cloner / pull
git checkout feat/001-user-auth

# 2. Backend
cd backend
go mod download
make migrate-up                      # applique les migrations 0001_init_auth
make sqlc                            # génère internal/adapters/db/sqlc/*
make test-domain                     # tests purs (≤ 1 s)

# 3. Frontend
cd ../frontend
pnpm install
pnpm dev
```

## Lancer la feature en local

Trois terminaux :

```bash
# Terminal 1 — backend Go
cd backend
make run                             # écoute sur :8080

# Terminal 2 — frontend Next.js
cd frontend
pnpm dev                             # écoute sur :3000, proxy /api → :8080

# Terminal 3 — Traefik (optionnel pour HTTPS local)
docker compose -f infra/local/docker-compose.yml up traefik
# expose https://app.local.hello-there.net + https://api.local.hello-there.net
```

Sans Traefik, `__Host-` cookies ne fonctionnent pas car ils exigent HTTPS. Pour les tests E2E manuels, **utiliser Traefik + mkcert**. Pour les tests unitaires/intégration backend, le mode HTTP est OK car les tests ne dépendent pas du préfixe `__Host-`.

## Parcours end-to-end (smoke test)

1. Ouvrir `https://app.local.hello-there.net/register`.
2. Soumettre un compte (`alice@asso-paris.fr`, mot de passe ≥ 12 chars, date de naissance ≥ 16 ans).
3. Vérifier en BDD :
   ```bash
   sqlite3 backend/var/komunumo.db \
     "SELECT email, status FROM accounts; SELECT event_type FROM audit_log ORDER BY occurred_at;"
   ```
   Attendu : 1 ligne `pending_verification`, 1 événement `account.created`.
4. Récupérer le lien de vérification dans la console Brevo (ou stub local en dev — cf. `EMAIL_DEV_STDOUT=1` qui fait afficher l'email sur stdout).
5. Cliquer sur le lien → page `/verify-email/confirm?token=…` → redirection vers `/login?verified=1`.
6. Se connecter — vérifier la présence des cookies `__Host-session` et `__Host-csrf` dans DevTools.
7. Appeler `GET /api/v1/auth/me` → retourne le profil.
8. Cliquer sur "Se déconnecter" → cookies effacés, redirection `/`.

## Tests automatisés

### Backend

```bash
cd backend

# Tests purs domain (zéro DB)
make test-domain

# Tests use cases avec mocks de ports
make test-application

# Tests d'intégration adapters/db (SQLite fichier temp)
make test-adapters

# Tests handlers HTTP (httptest)
make test-http

# Tout
make test
```

### Frontend

```bash
cd frontend
pnpm test               # vitest (composants, hooks)
pnpm lint               # eslint + prettier
pnpm test:axe           # accessibilité automatisée sur les pages auth
```

### E2E (post-MVP, smoke)

```bash
pnpm test:e2e -- auth
```

## Commandes utiles

| But | Commande |
|-----|----------|
| Réinitialiser la BDD locale | `rm backend/var/komunumo.db && make migrate-up` |
| Voir les logs slog en JSON | `make run \| jq` |
| Régénérer sqlc | `make sqlc` |
| Vérifier que `audit_log` est bien append-only | `sqlite3 var/komunumo.db "DELETE FROM audit_log;"` doit échouer |
| Mesurer le coût bcrypt actuel | `go test ./internal/adapters/password -run TestBcryptCost -v` |
| Lighthouse RGAA sur `/login` | `pnpm lhci autorun --collect.url=https://app.local.hello-there.net/login` |

## Dépannage

| Symptôme | Cause probable | Fix |
|----------|---------------|-----|
| Cookies non posés en dev | HTTPS absent ou domaine mismatch | Lancer Traefik + mkcert |
| `bcrypt: timing` test fail | Machine sous-dimensionnée | Adapter le seuil dans `password_test.go` ou réduire le cost en config locale (12 reste obligatoire en CI/prod) |
| 429 immédiat sur `/login` | Test d'avant n'a pas vidé le rate limiter en mémoire | Redémarrer le backend ou ajouter `RATE_LIMIT_DISABLED=true` en dev |
| `email not sent` | Quota Brevo épuisé ou clé absente | Activer `EMAIL_DEV_STDOUT=1` pour stub |
| Sessions expirées immédiatement | Horloge système décalée | Synchroniser via `chronyd` ou ajuster `KOMUNUMO_SESSION_TTL` |

## Critères de "done" pour cette feature

- [ ] 100 % des tâches `tasks.md` cochées.
- [ ] `make test` vert (backend) + `pnpm test` vert (frontend).
- [ ] `make test-axe` vert (RGAA AAA sur les 4 pages auth).
- [ ] `lighthouse-ci` ≥ 90 sur Performance, Accessibilité, Best-practices, SEO pour `/register`, `/login`, `/reset-password`.
- [ ] Audit manuel NVDA (Windows) ou Orca (Linux) sur les 4 parcours — capture vidéo dans le dossier de soutenance.
- [ ] PR `feat/001-user-auth` mergée sur `main` via Option A admin merge avec CI verte (≥ J+1 après ouverture).
- [ ] Mis à jour : `docs/specs/03-api/openapi.yaml` (consolidation), `docs/specs/02-uml/sequence-auth.puml` (si schéma initial à ajuster), `MEMORY.md` projet.
