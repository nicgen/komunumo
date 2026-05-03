# Prompt d'initialisation — Phase 3

> Copier-coller ce prompt au début d'une nouvelle session Claude pour démarrer la Phase 3.

---

Nous reprenons le projet **komunumo / assolink** (réseau social associatif, Go 1.24 + Next.js 16).

## État du projet

- **Phase 1** (Auth) : complète et auditée.
- **Phase 2** (Profils — members, associations, memberships) : complète et auditée le 2026-05-03.
  - Migration `0002_profiles` en place (SQLite WAL).
  - Domaines : `account`, `member`, `association`. PII sur `members`, pas sur `accounts`.
  - Endpoints : register/member, register/association, GET/PATCH /me/profile, GET /accounts/{id}/profile.
  - `account.New` prend 3 arguments (email, passwordHash, kind). Pas de PII sur `accounts`.
  - Cookie session : `Secure` flag via `APP_ENV=production` (pas `r.TLS`).
- **CI** : complète et verte sur `main`.
  - Lefthook : gitleaks (pre-commit), commitlint (commit-msg), tests+lint (pre-push).
  - GitHub Actions : backend, frontend, docs, CodeQL, SonarCloud (main only), Trivy (main only).
  - Branche `main` protégée : PR obligatoire, CI verte requise.

## Workflow multi-agents

Lire `docs/process/multi-agent-workflow.md` pour le détail complet.

En résumé :
1. **Speckit** génère les specs et contrats.
2. **Claude** planifie, écrit les prompts Gemini, audite le code produit.
3. **Gemini** implémente les tâches.
4. **Claude** audite et corrige la CI.
5. PR `dev` → `main` quand CI verte.

## Règles commitlint — à inclure dans chaque prompt Gemini

Scopes autorisés (liste fermée dans `commitlint.config.mjs`) :
`auth`, `backend`, `frontend`, `posts`, `chat`, `notif`, `profile`, `profiles`, `follows`, `groups`, `events`, `search`, `audit`, `rgpd`, `db`, `api`, `web`, `ws`, `ops`, `adr`, `specs`, `docs`, `ci`, `deps`, `scaffold`, `release`, `learnings`

- Header max **120 caractères**
- Ne jamais inventer un nouveau scope — signaler le manque à Claude
- Ne jamais modifier `.gitignore` sans justification explicite
- Ne jamais modifier `.github/workflows/` sans vérifier les versions d'actions via `gh api`

## Objectif Phase 3

Implémenter **Follows + Posts** — le cœur social du réseau.

Démarrer par la planification avec Speckit, puis task list, puis prompt Gemini.

Specs existantes à consulter :
- `docs/specs/02-features/follows.md`
- `docs/specs/02-features/posts.md`
- `docs/specs/03-api/openapi.yaml`
- `docs/specs/04-data/mcd.mmd`
