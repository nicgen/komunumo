# Process - Git workflow professionnel

## Objectif

Le projet est porté par un seul développeur mais doit **simuler un environnement d'équipe** pour répondre aux exigences CDA (industrialisation, traçabilité, qualité). L'historique Git fait partie intégrante de la preuve méthodologique présentée au jury.

## Modèle de branchement

**GitHub Flow simplifié** (pas de develop, pas de release branch).

```
main      ────●─────●─────●─────●─────●──── (toujours déployable)
              │     │     │     │     │
feature/    ──┴──   │     │     │     │
fix/              ──┴──   │     │     │
chore/                  ──┴──   │     │
docs/                         ──┴──   │
                                    ──┴──
```

- `main` : branche protégée, toujours déployable. Aucun push direct.
- `feature/<scope>-<short-name>` : nouvelle fonctionnalité (ex: `feature/auth-register`).
- `fix/<scope>-<short-name>` : correction de bug (ex: `fix/posts-pagination`).
- `chore/<scope>` : tâches techniques sans impact fonctionnel (deps, tooling, refactor).
- `docs/<scope>` : documentation seule.
- `spike/<topic>` : exploration jetable. Mergée ou supprimée, jamais en `main`.

## Convention de commits

**Conventional Commits 1.0** (https://www.conventionalcommits.org/fr/v1.0.0/).

```
<type>(<scope>): <description courte>

[corps optionnel : pourquoi, comment, références]

[footer optionnel : BREAKING CHANGE, Refs: #123, Co-authored-by:]
```

**Types acceptés** : `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `perf`, `build`, `ci`, `style`.

**Scopes courants** : `auth`, `posts`, `chat`, `notif`, `db`, `api`, `web`, `ops`, `adr`, `specs`.

Exemples valides :
```
feat(auth): add email verification flow with token expiry

Implements UC4 from specs/02-features/auth.md. Token TTL 24h, single use.
Refs ADR-0004.

Closes #42
```

```
fix(chat): drop heartbeat after 60s without pong

Previously hung clients lingered indefinitely in the hub.
```

```
docs(adr): add ADR-0011 about file storage on docker volume
```

Validation locale : `commitlint` via `husky` (pre-commit hook).

## Workflow par feature

1. **Issue GitHub** créée avec template `feature` (lié à la spec Speckit correspondante).
2. **Branche** `feature/<scope>-<slug>` créée depuis `main` à jour.
3. **Commits atomiques** suivant Conventional Commits.
4. **Push** régulier (au moins en fin de journée), même WIP. Sécurise le travail.
5. **Pull Request** ouverte tôt (statut "Draft" si non terminée).
6. **CI verte** obligatoire (tests, lint, build, scans).
7. **Self-review** systématique avant de passer en "Ready for review".
8. **Squash & merge** ou **Rebase & merge** selon la qualité de l'historique :
   - Branche < 5 commits propres et atomiques -> rebase merge (préserve l'histoire).
   - Branche bruyante (WIP, fixup, typos) -> squash merge.
   - **Pas de merge commit** sur `main` (historique linéaire).
9. **Branche supprimée** automatiquement après merge (config repo).

## Protections de la branche `main`

À configurer dans GitHub Settings > Branches :

- [x] Require a pull request before merging.
- [x] Require approvals: 1 (en pratique, self-approve via review d'IA assistante documentée).
- [x] Dismiss stale pull request approvals when new commits are pushed.
- [x] Require status checks to pass before merging :
  - `lint` (golangci-lint, eslint).
  - `test-go`, `test-web`.
  - `build-go`, `build-web`.
  - `security-scan` (gosec, govulncheck, npm audit, trivy).
  - `sonar` (Quality Gate).
- [x] Require branches to be up to date before merging.
- [x] Require conversation resolution before merging.
- [x] Require linear history (no merge commits).
- [x] Do not allow bypassing the above settings.
- [ ] Require signed commits (V2 si je configure GPG).

## Tags et releases

**SemVer** : `vMAJOR.MINOR.PATCH`.

- `v0.1.0` : fin S0 (cadrage + setup CI/CD).
- `v0.x.y` : itérations MVP par sprint.
- `v1.0.0` : soutenance jury.

Tags annotés signés (`git tag -a -s vX.Y.Z -m "message"`) si GPG configuré.
Release GitHub avec changelog auto-généré (script `cliff` ou `release-please` en V2).

## Process points-clefs (étapes de versionnement obligatoires)

À chaque jalon clé du planning, **un commit ou un tag** doit matérialiser l'avancement :

| Jalon | Action Git |
|-------|-----------|
| Fin de cadrage S0 | tag `v0.1.0`, commit `chore: initial documentation skeleton` |
| Fin de squelette technique S0 | tag `v0.2.0`, commit `feat(scaffold): backend and frontend skeletons` |
| Fin S1 (auth + profils) | tag `v0.3.0` |
| Fin S2 (posts + follows + feed) | tag `v0.4.0` |
| Fin S3 (chat + events + notif) | tag `v0.5.0` |
| Fin S4 (audit + RGPD + a11y AAA) | tag `v0.6.0` |
| Soutenance | tag `v1.0.0` |

À chaque jalon : ouvrir une **Release GitHub** avec changelog rédigé (10 min, mais artefact de jury précieux).

## Hygiène quotidienne

- `git pull --rebase` avant tout début de session.
- Une seule fonctionnalité par PR, pas de "fourre-tout".
- Commits message en **anglais** (norme open source), discussions PR/issues en **français** acceptées.
- `.gitignore` rigoureux : pas de binaires, pas de `node_modules`, pas de `.env`.
- Secrets jamais commités. Pre-commit hook `gitleaks` actif.

## Références

- Conventional Commits - https://www.conventionalcommits.org/fr/v1.0.0/
- SemVer - https://semver.org/lang/fr/
- ADR-0008 (CI/CD GitHub Actions).
- `dossier/` chapitre Industrialisation.
