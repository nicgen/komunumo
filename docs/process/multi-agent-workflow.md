# Workflow multi-agents : Claude + Gemini + Speckit

## Principe général

Ce projet utilise trois agents avec des rôles distincts pour éviter les limitations de chacun :

| Agent | Rôle | Forces |
| ----- | ---- | ------- |
| **Speckit** | Génération des specs et contrats | Cohérence spec ↔ code, ADR automatiques |
| **Claude** | Planification, prompts Gemini, audit, corrections CI | Raisonnement architectural, vision globale |
| **Gemini** | Implémentation des tâches | Volume de code, exécution de tâches longues |

Le flux par phase :

```
Speckit (specs + contrats)
    ↓
Claude (plan + task list + prompt Gemini)
    ↓
Gemini (implémentation)
    ↓
Claude (audit du code produit + corrections CI)
    ↓
PR dev → main (CI verte requise)
```

## Pourquoi ce découpage

- **Claude** est limité en tokens sur les longues implémentations — Gemini prend le relais.
- **Gemini** suit bien les directives précises mais dévie sans briefing explicite — Claude structure.
- **Claude** est meilleur pour les corrections chirurgicales, l'audit de sécurité et la CI.

---

## Règles critiques pour les prompts Gemini

Ces règles doivent figurer dans **chaque prompt transmis à Gemini** sans exception.

### 1. Commitlint — scopes stricts

Les scopes autorisés sont une liste fermée dans `commitlint.config.mjs`. Gemini **ne doit jamais inventer un nouveau scope**. Avant tout commit, vérifier que le scope est dans la liste.

Liste actuelle :
`auth`, `backend`, `frontend`, `posts`, `chat`, `notif`, `profile`, `profiles`, `follows`, `groups`, `events`, `search`, `audit`, `rgpd`, `db`, `api`, `web`, `ws`, `ops`, `adr`, `specs`, `docs`, `ci`, `deps`, `scaffold`, `release`, `learnings`

Si un scope manque → le signaler à Claude avant de commiter.

### 2. Header max 120 caractères

`fix(profiles): adjust test fixtures — X-Forwarded-For header, FindByEmailCanonical, tasks T001-T024 marked done`
→ 111 caractères, bloquant en CI. Garder les headers concis.

### 3. Ne jamais modifier `.gitignore` pour y ajouter `.github/`

Gemini a ajouté `.github/` au `.gitignore` lors d'une session précédente, ce qui aurait ignoré tous les workflows CI. Toute modification de `.gitignore` doit être explicite et justifiée.

### 4. Indentation YAML stricte dans les workflows

Un mauvais indent dans `.github/workflows/ci.yml` casse silencieusement un job entier. Toujours valider avec `yamllint` ou un diff attentif avant de commit.

### 5. Vérifier les versions des actions GitHub

Les versions d'actions peuvent ne pas exister (`aquasecurity/trivy-action@0.24.0` → inexistant).
Vérifier via `gh api repos/<owner>/<repo>/releases/latest` avant d'utiliser une version.

---

## Erreurs rencontrées en session (2026-05-03)

| Erreur | Cause | Fix |
| ------ | ----- | --- |
| ESLint exit 2 | Import de `eslint-plugin-jsx-a11y` alors que `eslint-config-next` l'enregistre déjà | Supprimer l'import et `plugins:`, garder seulement les `rules:` |
| `.github/` dans `.gitignore` | Gemini a ajouté la ligne sans justification | Suppression immédiate |
| Indent YAML commitlint job | `fetch-depth: 0` mal indenté | Correction manuelle |
| `sonar` job sur toutes les branches | SonarCloud free plan = main only | Restreindre à `github.ref == 'refs/heads/main'` |
| `trivy-action@0.24.0` inexistant | Version copiée/inventée | Mettre à jour à la dernière via `gh api` |
| Scopes `profiles`, `frontend`, `backend` manquants | Gemini a utilisé des scopes non déclarés | Ajouter les scopes dans `commitlint.config.mjs` avant de lancer Gemini |
| Push direct sur `main` bloqué | Branche protégée, PR obligatoire | Toujours passer par PR |
| Double run CI sur `dev` | PR ouverte → push = événement `push` + `pull_request` | Comportement normal, pas un bug |
| Gitleaks faux positif Makefile | `curl -u "$(TOKEN):"` matche `curl-auth-user` | Allowlist regex dans `.gitleaks.toml` |

---

## Checklist avant de lancer Gemini sur une phase

- [ ] Les specs sont finalisées dans `docs/specs/` et `specs/00N-feature/`
- [ ] Le plan et la task list sont dans `specs/00N-feature/tasks.md`
- [ ] Les scopes commitlint nécessaires sont déclarés dans `commitlint.config.mjs`
- [ ] Le prompt Gemini contient la liste des scopes et la règle header 120 chars
- [ ] Le prompt Gemini contient l'interdiction de modifier `.gitignore` sans justification
- [ ] La CI est verte sur `dev` avant de démarrer

## Checklist d'audit post-Gemini (Claude)

- [ ] `go test -race ./cmd/... ./internal/...` passe
- [ ] `pnpm lint && pnpm typecheck && pnpm test` passent
- [ ] Les checkboxes du tasks.md sont à jour
- [ ] Aucun fichier sensible dans le diff (`.env`, `cookies.txt`, clés)
- [ ] Les versions des actions GitHub sont réelles (vérifier via `gh api`)
- [ ] Pas de modification de `.gitignore` non justifiée
- [ ] La CI est verte sur la PR avant merge
