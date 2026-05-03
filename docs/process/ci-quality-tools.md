# Outils qualité en intégration continue

Chaque outil de la CI répond à une exigence précise définie dans les specs (`docs/specs/05-quality/`).
Ce document liste les modules actifs, ce qu'ils vérifient, et les gaps encore ouverts.

---

## Architecture de la pipeline

```
push / PR
  └── changes (dorny/paths-filter)      ← évite les jobs inutiles
        ├── backend  ──► backend job
        ├── frontend ──► frontend job
        └── docs     ──► docs job
                              ↓
                        trivy (main only)
                              ↓
                        ci-success (quality gate)
```

Les hooks `lefthook` complètent la CI côté développeur (pre-commit, commit-msg, pre-push).

---

## Modules actifs

### Conventional Commits — `commitlint`

**Trigger :** chaque PR via `wagoid/commitlint-github-action`.
**Lefthook :** `commit-msg` via `@commitlint/cli` (exécuté localement avant push).
**Vérifie :** format `type(scope): sujet` selon la config `commitlint.config.mjs` à la racine.
**Norme :** `@commitlint/config-conventional` + règles projet (types, scopes, longueur d'en-tête ≤ 100).
**Pourquoi :** traçabilité du changelog, lisibilité des PR, qualification pour la soutenance.

---

### Détection de secrets — `gitleaks`

**Trigger :** hook `pre-commit` lefthook sur les fichiers stagés (`protect --staged`).
**Config :** `.gitleaks.toml` — règles par défaut + allowlist sur `*_test.go`, `docs/`, `*.example`.
**Vérifie :** clés API, tokens, mots de passe dans le diff avant qu'il entre dans l'historique git.
**Norme :** `security.md` — V6.4.1 "Secrets non en code".
**Limite :** ne scanne pas l'historique existant (à faire manuellement une fois : `gitleaks detect`).

---

### Qualité Go — `golangci-lint`

**Trigger :** job `backend`, sur tout push si backend modifié.
**Vérifie :** formatage (`gofmt`), imports, complexité cyclomatique, erreurs non vérifiées, variables inutilisées, etc. — ensemble de linters Go configurés via `.golangci.yml`.
**Norme :** conventions Go 1.24, architecture hexagonale (imports entre couches).

---

### SAST Go — `gosec`

**Trigger :** job `backend`.
**Sortie :** rapport SARIF uploadé dans GitHub Security tab.
**Vérifie :** vulnérabilités applicatives Go : injections SQL, chemins de fichiers non nettoyés, crypto faible (MD5/SHA1), `G#` rules.
**Norme :** `security.md` — V6.2.5 "Pas de MD5/SHA1", OWASP ASVS L1.

---

### CVE dépendances Go — `govulncheck`

**Trigger :** job `backend`.
**Vérifie :** vulnérabilités connues (CVE) dans les modules Go du `go.sum`, en analysant les chemins d'appel effectifs (pas seulement la présence du module).
**Norme :** `security.md` — tableau "Outils de scan automatique en CI".

---

### Tests Go avec détecteur de race — `go test -race`

**Trigger :** job `backend`.
**Vérifie :** correctness des tests unitaires + data races dans les goroutines (critique pour le hub WebSocket).
**Norme :** couverture de code (`coverage.out` uploadé en artefact).

---

### Lint TypeScript — `eslint`

**Trigger :** job `frontend` (`pnpm lint`).
**Vérifie :** règles ESLint Next.js (`eslint-config-next`) — accessibilité de base, hooks React, imports.
**Norme :** conventions TypeScript 5.6+, Next.js 16.

---

### Typage strict — `tsc --noEmit`

**Trigger :** job `frontend` (`pnpm typecheck`).
**Vérifie :** cohérence des types TypeScript sans produire de build — détecte les régressions de types invisibles à l'exécution.
**Norme :** TypeScript 5.6+ strict mode.

---

### Tests frontend — Vitest

**Trigger :** job `frontend` (`pnpm test`).
**Vérifie :** tests unitaires et d'intégration des composants React.
**Norme :** couverture des cas métier frontend.

---

### Build Next.js

**Trigger :** job `frontend` (`pnpm build`).
**Vérifie :** compilation complète, erreurs SSR/RSC, pages statiques générables.
**Norme :** valide que le déploiement Vercel ne sera pas cassé.

---

### Lint markdown — `markdownlint-cli2`

**Trigger :** job `docs`, si `docs/**` modifié.
**Vérifie :** formatage Markdown cohérent (titres, listes, longueur de ligne).
**Norme :** lisibilité du dossier de soutenance.

---

### Validation OpenAPI — `@redocly/cli`

**Trigger :** job `docs`.
**Vérifie :** conformité du fichier `docs/specs/03-api/openapi.yaml` à la spec OpenAPI 3.1.0 — types corrects, références `$ref` résolues, schémas valides.
**Norme :** `docs/specs/03-api/` — contrat d'API frontend/backend.

---

### Validation Mermaid — `@mermaid-js/mermaid-cli`

**Trigger :** job `docs`.
**Vérifie :** que les fichiers `.mmd` (`docs/diagrams/`, `docs/specs/04-data/`) sont parsables et rendables.
**Norme :** `docs/specs/04-data/mcd.mmd` — MCD inclus dans le dossier de soutenance.

---

### Scan image/filesystem — `trivy`

**Trigger :** job `trivy`, uniquement sur push vers `main`.
**Vérifie :** CVE CRITICAL et HIGH dans le filesystem (Dockerfiles, dépendances lockfile).
**Norme :** `security.md` — tableau "Outils de scan automatique en CI".

---

### CVE dépendances JS — `pnpm audit`

**Trigger :** job `frontend`, après `pnpm install`.
**Vérifie :** vulnérabilités connues (CVE) dans les dépendances npm au niveau HIGH et CRITICAL.
**Norme :** `security.md` — tableau "Outils de scan automatique en CI".

---

### Accessibilité statique — `eslint-plugin-jsx-a11y`

**Trigger :** job `frontend` (`pnpm lint`) + hook `pre-push` lefthook.
**Vérifie :** 9 règles RGAA critiques en `error` : `alt-text`, `aria-props`, `aria-proptypes`, `aria-unsupported-elements`, `role-has-required-aria-props`, `role-supports-aria-props`, `no-access-key`, `interactive-supports-focus`, `label-has-associated-control`.
**Norme :** `a11y.md` — RGAA 4.1, thèmes 1 (images), 7 (scripts), 11 (formulaires).

---

### Performance & accessibilité runtime — `lighthouse-ci`

**Trigger :** job `frontend`, après `pnpm build`. Démarre `pnpm start` et mesure `http://localhost:3000/`.
**Config :** `frontend/.lighthouserc.yml`. Seuils actuels (à affiner après baselines) :
- Performance ≥ 0.7 → warn
- Accessibilité ≥ 0.8 → error
- Best practices ≥ 0.8 → warn

**Norme :** `a11y.md` (score ≥ 95 cible finale), `eco.md` (budgets perf par page).
**Note :** `continue-on-error: true` le temps d'établir les baselines. Retirer ce flag une fois les scores stables.

---

### SAST GitHub natif — `codeql`

**Trigger :** tout push et PR. Job matrix : `go` + `javascript-typescript`.
**Sortie :** alertes dans GitHub Security → Code scanning alerts.
**Norme :** `security.md` — "CodeQL | SAST GitHub natif".
**Note :** ne bloque pas `ci-success` — les alertes sont traitées indépendamment.

---

### Qualité globale & surveillance code LLM — `sonar`

**Trigger :** après succès des jobs `backend` et `frontend`.
**Vérifie :** code smells, duplication, bugs, couverture agrégée backend + frontend. Utile pour surveiller la qualité du code généré par LLM (cohérence, duplication, complexité cyclomatique).
**Config :** `sonar-project.properties` à la racine. Secret `SONAR_TOKEN` requis dans GitHub repo settings.
**Norme :** ADR-0008.
**Note :** `continue-on-error: true` — se désactive proprement si `SONAR_TOKEN` absent. Pour désactiver définitivement : commenter le job ou supprimer le secret.
**Prérequis SonarCloud :** désactiver "Automatic Analysis" dans Administration → Analysis Method.

---

## Gaps V2 (post-soutenance)

| Outil | Spec source | Raison du report |
| ----- | ----------- | ---------------- |
| `axe-core` + Playwright | `a11y.md` | Nécessite des tests E2E à écrire |
| `pa11y-ci` | `a11y.md` | Couvert partiellement par lighthouse-ci |
| Contraste WCAG AA tokens | `a11y.md` | Outillage à identifier |
| `Semgrep` | `security.md` | Optionnel, CodeQL couvre l'essentiel |
| OWASP ZAP nightly | `security.md` | Workflow `zap.yml` à créer en V2 |
| `eco.yml` + EcoIndex | `eco.md` | Lighthouse-ci couvre les perf budgets pour la soutenance |

---

## Vue d'ensemble par spec

| Spec | Couverture | Taux |
| ---- | ---------- | ---- |
| `security.md` | gitleaks, gosec, govulncheck, pnpm audit, trivy, CodeQL — manque : ZAP, Semgrep | ~85% |
| `a11y.md` | jsx-a11y (lint), lighthouse-ci (runtime) — manque : axe-core, pa11y-ci | ~50% |
| `eco.md` | lighthouse-ci perf budgets, paths-filter — manque : eco.yml, EcoIndex | ~40% |
| Conventional Commits | commitlint CI + lefthook local | 100% |
| Qualité Go | golangci-lint, gosec, govulncheck, go test -race, SonarCloud | 100% |
| Qualité frontend | eslint + jsx-a11y, tsc, vitest, build, SonarCloud | 100% |
| Docs / contrats | markdownlint, redocly, mermaid-cli | 100% |
