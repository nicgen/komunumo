# ADR-0008 - GitHub Actions et SonarCloud pour la CI/CD

- Statut : Accepté
- Date : 2026-04-26
- Décideur : nic

## Contexte

Le brief CDA exige une **pipeline CI/CD** automatisant tests et déploiement, et un **outil de qualité** statique. L'organisation doit simuler une équipe (PR review, branches protégées) bien que le développeur soit seul. La pipeline doit être démontrable au jury, idéalement avec des résultats datés et reproductibles.

## Décision

Utiliser **GitHub Actions** pour l'orchestration CI/CD, et **SonarCloud** (gratuit pour public repos) pour l'analyse qualité. Pipeline minimum :

1. Lint frontend (Biome) + lint backend (`go vet`, `staticcheck`, `gosec`).
2. Tests frontend (Vitest) + tests backend (`go test ./...`).
3. SonarCloud scan (avec quality gate progressive).
4. Build images Docker multi-stage (frontend + backend).
5. Scan Trivy sur images.
6. Push registry GHCR.
7. Audit pa11y + Lighthouse sur staging.
8. Déploiement staging (push `main`) ou prod (tag `v*`).

Branches protégées sur `main`, PR obligatoire, conventional commits, squash merge.

## Alternatives écartées

- **Jenkins self-hosted** : hébergement et maintenance d'un Jenkins coûtent en temps que tu n'as pas. Démontrable mais surcharge MVP.
- **GitLab CI** : excellent, mais demande de migrer le repo vers GitLab. Non justifié.
- **CircleCI / Buildkite** : payants, complexité non justifiée.
- **SonarQube self-hosted** : doable mais charge ops importante. SonarCloud free pour open source suffit.
- **CodeClimate** : alternative valable, communauté Go moins fournie.

## Conséquences

- (+) Pipeline démontrable au jury via les badges README et l'historique des runs.
- (+) Quality gate bloquante en PR : couverture min, no new bugs, no new vulns.
- (+) Trivy + gosec + npm audit + govulncheck = défense en profondeur sécurité.
- (+) Pas d'infra à maintenir.
- (-) Verrouillage GitHub. Mitigation : actions standardisées, migrables vers GitLab CI si besoin.
- (-) Quotas free GitHub Actions sur orgs. Mitigation : repo perso ou public, quotas suffisants.

## Références

- GitHub Actions docs - https://docs.github.com/en/actions
- SonarCloud - https://www.sonarsource.com/products/sonarcloud/
- Trivy - https://trivy.dev/
