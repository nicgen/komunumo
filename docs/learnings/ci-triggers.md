# CI — quand ça tourne et pourquoi deux runs

## Règle de base

Le workflow `.github/workflows/ci.yml` se déclenche sur deux événements :

```yaml
on:
  push:
    branches: [main, dev]
  pull_request:
    branches: [main, dev]
```

Conséquence : **un push sur `dev` avec une PR ouverte vers `main` déclenche deux runs distincts.**

- Run 1 → événement `push` sur `dev`
- Run 2 → événement `pull_request` (la PR pointe vers `main`, sa branche source `dev` vient d'être mise à jour)

C'est le comportement normal de GitHub Actions. Une fois la PR mergée et fermée, un push sur `dev` ne déclenche plus qu'un seul run.

---

## Ce qui tourne selon le déclencheur

| Job | push dev | push main | PR vers main/dev |
| --- | -------- | --------- | ---------------- |
| `changes` (paths-filter) | oui | oui | oui |
| `commitlint` | non | non | **oui** |
| `backend` | si backend modifié ou push | si backend modifié ou push | si backend modifié ou push |
| `frontend` | si frontend modifié ou push | si frontend modifié ou push | si frontend modifié ou push |
| `docs` | si docs modifié | si docs modifié | si docs modifié |
| `codeql` | **oui (toujours)** | **oui (toujours)** | **oui (toujours)** |
| `trivy` | non | **oui (main uniquement)** | non |
| `sonar` | non | **oui (main uniquement)** | non |
| `ci-success` | oui | oui | oui |

---

## Ce qui bloque le merge (required checks)

Le job `ci-success` agrège : `backend`, `frontend`, `docs`, `commitlint`.
Un résultat `failure` dans l'un de ces quatre bloque la PR.

`skipped` est toléré — si aucun fichier backend n'a changé, le job `backend` est skippé et `ci-success` passe quand même.

**Non-bloquants** (continue-on-error ou hors ci-success) :
- `codeql` — alertes visibles dans Security tab, ne bloque pas le merge
- `sonar` — `continue-on-error: true`, désactivable en supprimant le secret `SONAR_TOKEN`
- `trivy` — ne tourne que sur main, hors ci-success
- `lighthouse-ci` — `continue-on-error: true` pendant la phase de baseline

---

## Ce que lefthook fait en local (avant le push)

| Hook | Ce qui tourne | Bloquant |
| ---- | ------------- | -------- |
| `pre-commit` | `gitleaks --staged` | oui |
| `commit-msg` | `commitlint` | oui |
| `pre-push` | backend tests + lint, frontend lint + typecheck + tests | oui, sur les fichiers modifiés seulement |

Le pre-push skippera les jobs frontend si seul le backend a changé (et vice-versa).
Lefthook est la première ligne — si ça passe en local, la CI a de bonnes chances de passer.

---

## Lecons apprises

**Scopes commitlint** : les scopes sont une liste fermée dans `commitlint.config.mjs`. Ajouter un scope en cours de projet (ex. `profiles`) nécessite de mettre à jour cette liste **avant** de commiter avec ce scope, sinon le job `commitlint` de la PR échoue en erreur bloquante.

**Scope-enum actuel** :
`auth`, `posts`, `chat`, `notif`, `profile`, `profiles`, `follows`, `groups`, `events`, `search`, `audit`, `rgpd`, `db`, `api`, `web`, `ws`, `ops`, `adr`, `specs`, `docs`, `ci`, `deps`, `scaffold`, `release`, `learnings`

**footer-leading-blank** : avertissement non bloquant — une ligne vide manque avant le footer d'un commit (ex. `Co-authored-by`, `Closes #n`). Warning, pas error.

**SonarCloud limitation free plan** : n'analyse que la branche `main`. Le job `sonar` est restreint à `github.ref == 'refs/heads/main'` pour éviter des erreurs silencieuses sur les autres branches.

**Protected branch main** : push direct interdit. Toujours passer par une PR. Un merge local suivi d'un `git push origin main` sera rejeté par GitHub même si le merge réussit localement.
