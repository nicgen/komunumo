# Architecture Decision Records

Format : [MADR 4.0](https://adr.github.io/madr/) - Markdown Any Decision Records.

## Convention

- Un fichier par décision : `NNNN-titre-court.md`.
- Numérotés en continu, jamais renommés.
- Statut : `Proposé` | `Accepté` | `Déprécié` | `Remplacé par ADR-XXXX`.
- Un ADR n'est **jamais** modifié rétroactivement. Pour changer d'avis, on rédige un nouvel ADR qui remplace le précédent.

## Index

| ID | Titre | Statut |
|----|-------|--------|
| [0001](./0001-architecture-hexagonale-go.md) | Architecture hexagonale pour le backend Go | Accepté |
| [0002](./0002-nextjs-app-router-rsc.md) | Next.js 16 App Router et React Server Components | Accepté |
| [0003](./0003-sqlite-wal-sqlc.md) | SQLite + WAL + sqlc comme couche persistance | Accepté |
| [0004](./0004-sessions-cookies-vs-jwt.md) | Sessions cookies HttpOnly plutôt que JWT seul | Accepté |
| [0005](./0005-gorilla-websocket-hub-memoire.md) | gorilla/websocket avec hub en mémoire | Accepté |
| [0006](./0006-tailwind-shadcn-ui.md) | Tailwind v4 et shadcn/ui sur Radix UI | Accepté |
| [0007](./0007-mermaid-modelisation.md) | Mermaid pour la modélisation UML et ER | Accepté |
| [0008](./0008-github-actions-sonarcloud.md) | GitHub Actions et SonarCloud pour la CI/CD | Accepté |
| [0009](./0009-vercel-frontend-traefik-backend.md) | Vercel pour frontend, Traefik+VPS pour backend | Accepté |
| [0010](./0010-pandoc-eisvogel-dossier.md) | Markdown + Pandoc + Eisvogel pour le dossier | Accepté |
| [0011](./0011-stockage-fichiers-volume-docker.md) | Stockage fichiers utilisateurs sur volume Docker local | Accepté |
| [0012](./0012-email-transactionnel-brevo.md) | Email transactionnel via Brevo | Accepté |
