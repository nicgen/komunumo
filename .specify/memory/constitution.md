# Constitution AssoLink / Komunumo

Ce document énonce les principes non négociables du projet. Toute spécification, plan, tâche ou code produit (humain ou LLM) doit s'y conformer. Les décisions techniques détaillées vivent dans `docs/adr/`. Les exigences fonctionnelles vivent dans `docs/specs/`. Cette constitution arbitre quand elles se contredisent.

## Core Principles

### I. Souveraineté numérique

Toute donnée utilisateur reste hébergée en France ou en UE (CNIL, RGPD article 44). Frontend Vercel région `cdg1`, backend Scaleway DEV1-S Paris, email Brevo (FR), pas de fournisseur cloud hors UE pour le chemin critique. Le modèle de données reste compatible avec une fédération ActivityPub future (identifiants `user@domain`, URLs canoniques) mais celle-ci n'est pas livrée en V1. Aucune dépendance à un service propriétaire dont le départ casserait l'application — chaque brique a un plan B documenté.

### II. Sobriété numérique (RGESN) et Accessibilité (RGAA)

**Accessibilité — RGAA 4.1.** Les parcours critiques (inscription, connexion, création de post, inscription à un événement, messagerie) atteignent et maintiennent le **niveau AAA**. Le reste de l'application atteint au minimum **AA**. Le choix de **shadcn/ui sur Radix UI** est motivé par son socle accessible par défaut (gestion ARIA, focus, navigation clavier, conformité WAI-ARIA Authoring Practices) ; toute déviation à un primitif Radix est documentée. Toute PR touchant un parcours critique inclut une vérification `axe-core` + `lighthouse-ci` dans la CI, et toute régression d'accessibilité est un blocker de merge équivalent à un test cassé. Les contrastes, focus visibles, navigation clavier et compatibilité lecteur d'écran (NVDA + Orca) sont vérifiés à la main sur les parcours critiques avant chaque démo. Une **déclaration de conformité RGAA** est livrée avec le dossier de soutenance.

**Sobriété — RGESN.** Page d'accueil sous **0,3 gCO2eq** par visite (calcul EcoIndex, score A). Budget : ~50 KB HTML+CSS critiques, pas de JavaScript bloquant le rendu, pas d'images en haut de page non essentielles. Les images sont systématiquement servies en AVIF avec fallback WebP, lazy-loaded, et dimensionnées par les attributs `width`/`height`. Un mode sobriété (texte seul, désactivation des images et animations) est disponible et persisté en cookie. Toute feature ajoute son propre budget de poids dans sa spec et la CI échoue si dépassement.

### III. Architecture hexagonale et test-first sur le domaine

Le backend Go suit une **architecture hexagonale** (ports & adapters) : `internal/domain` (entités + règles métier pures, zéro dépendance externe), `internal/application` (use cases), `internal/ports` (interfaces), `internal/adapters/{http,ws,db,email,storage}` (implémentations). **Le code du domaine est écrit après ses tests, non l'inverse** : tout PR touchant `internal/domain` ou `internal/application` doit présenter un commit `test(scope): …` antérieur au commit `feat(scope): …` correspondant dans l'historique de la branche. Cette discipline est **vérifiable en direct dans `git log`** lors de la soutenance et fait partie de l'argumentaire — la transparence prime sur la perfection : toute exception (debug, prototype, refactor de tests existants) est tracée explicitement dans le message du commit (`feat(scope): … [no-test-first: motif]`). Les adapters sont couverts par des tests d'intégration. Couverture cible : `domain` ≥ 90 %, `application` ≥ 80 %, global ≥ 70 %.

### IV. Spec-driven development

Toute feature non triviale (≥ 1 jour-homme) commence par une spec dans `docs/specs/02-features/`, suit le template `_template.md`, et est validée avant d'écrire du code. Le workflow Speckit (`/speckit-specify` → `/speckit-clarify` → `/speckit-plan` → `/speckit-tasks` → `/speckit-implement`) est l'outil par défaut. Les ADRs (`docs/adr/`) capturent les décisions architecturales en MADR 4.0 ; aucune décision structurelle non triviale n'est prise sans ADR. Les changements de spec post-implémentation sont autorisés mais doivent référencer le commit qui les opère.

### V. Sécurité par défaut

Aucun secret en clair dans le dépôt (vérifié par git-secrets ou équivalent en pre-commit). Les sessions sont des cookies `HttpOnly`, `Secure`, `SameSite=Strict`. Les mots de passe sont hashés avec **bcrypt cost ≥ 12**. Toute action sensible (suppression, export RGPD, modification de rôle) est journalisée dans une **table d'audit append-only** (`audit_log`) — INSERT-only, jamais UPDATE/DELETE, contrainte vérifiée par trigger SQLite. Cette piste d'audit (Audit Trail) constitue le socle V1. Une évolution **V2 documentée en ADR** propose un chaînage cryptographique HMAC-SHA256 (chaque entrée hache la précédente, clé côté serveur) pour atteindre la **non-répudiation** et détecter toute altération a posteriori — pratique courante en banque/assurance, présentée à l'oral comme perspective d'industrialisation. Toute dépendance ajoutée passe `govulncheck` (Go) ou `pnpm audit` (Node) ; tout ajout en dépendance directe nécessite une justification dans le commit. La CI bloque sur vulnérabilité CRITICAL ou HIGH non corrigeable.

### VI. Workflow Git industrialisé

Conventional Commits obligatoires (vérifiés par commitlint). Branches courtes (≤ 3 jours) au format `<type>/<scope>-<slug>`. Aucun push direct sur `main` — toute modification passe par PR avec CI verte. Les ADRs et specs sont **versionnés et immuables** : on ne réécrit pas l'histoire d'un ADR accepté, on en rédige un nouveau qui le remplace. Squash-merge par défaut pour garder un historique lisible. Override admin admis pour le développement solo mais documenté en commentaire de PR.

## Standards techniques

- **Stack** : Go 1.24 + coder/websocket + SQLite WAL + sqlc | Next.js 16 + Tailwind v4 + shadcn/ui | Docker + Traefik v2.11 + Cloudflare DNS-01.
- **Persistance** : SQLite avec WAL et FTS5 ; migrations versionnées via golang-migrate ; pas d'ORM côté Go (`sqlc` seul).
- **API** : OpenAPI 3.1 source de vérité (`docs/specs/03-api/openapi.yaml`) ; tout endpoint a un test de contrat.
- **Frontend** : Server Components par défaut, `"use client"` justifié au cas par cas. Pas de state global tant qu'un Context local suffit.
- **Modélisation jury** : Looping pour Merise (MCD/MLD/MPD), Mermaid pour UML (cas d'usage, classes, séquences). Voir ADR-0013.

## Workflow de développement

- **Daily** : ouvrir `docs/specs/` ou `.specify/` avant `backend/` ou `frontend/` ; écrire la spec ou la mettre à jour avant de coder.
- **PR** : titre = Conventional Commit ; description = lien vers la spec/ADR concerné ; checklist `Test plan` obligatoire pour toute feature.
- **Review** : en solo, l'auteur relit son propre diff à J+1 minimum avant de merger (sauf hotfix). Override admin documenté dans le commentaire de merge.
- **Quality gates en CI** :
  1. `commitlint` (Conventional Commits)
  2. `markdownlint` (docs)
  3. `golangci-lint` + `gosec` + `govulncheck` + tests Go avec couverture
  4. ESLint + tsc + tests Vitest + build Next.js
  5. `axe-core` + `lighthouse-ci` sur parcours critiques (post-MVP)
  6. `trivy` (scan filesystem + container) sur push `main`

## Governance

Cette constitution **prime** sur toute autre directive (sauf instruction explicite et tracée du formateur ENI ou du jury CDA). Toute modification de cette constitution :
1. Fait l'objet d'une PR dédiée portant le scope `(constitution)` en commit.
2. Incrémente le numéro de version selon SemVer (MAJOR pour suppression d'un principe, MINOR pour ajout, PATCH pour clarification).
3. Met à jour `Last Amended`.
4. Liste les ADRs/specs/code à mettre en cohérence dans la description de PR.

Les LLMs (Claude Code, autres) lisent ce document **en début de session** et refusent toute requête qui le contredit en l'absence d'instruction explicite de l'utilisateur. Une dérogation ponctuelle (POC, expérimentation) est tracée en commentaire dans le code ou la spec concernée.

**Version**: 1.0.0 | **Ratified**: 2026-04-26 | **Last Amended**: 2026-04-26
