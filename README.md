# AssoLink / Komunumo

Réseau social local pour associations, TPE/PME et bénévoles, accessible (RGAA AAA sur les parcours critiques) et éco-conçu (RGESN).

- **AssoLink** : nom de projet (interne, dossier CDA, code, ce dépôt).
- **Komunumo** : nom de produit (interface utilisateur). Esperanto pour "communauté".

Projet de soutenance pour la certification française **CDA - Concepteur Développeur d'Applications** (RNCP37873, niveau 6, ENI).

## Stack

| Couche | Choix |
|--------|-------|
| Backend | Go 1.24, hexagonal, gorilla/websocket, sqlc |
| Persistance | SQLite WAL + FTS5, migrations golang-migrate |
| Frontend | Next.js 16 (App Router, RSC), Tailwind v4, shadcn/ui |
| Auth | Sessions cookies HttpOnly + bcrypt cost 12 |
| Déploiement | Vercel (front) + Traefik v2.11 + Docker (back, France) |
| Email | Brevo (transactionnel, FR) |
| Stockage | Volume Docker local (MVP) -> Scaleway Object Storage (V2) |
| Qualité | golangci-lint, gosec, govulncheck, axe-core, lighthouse-ci, SonarCloud |
| Dossier | Markdown + Pandoc + Eisvogel |

Détails dans `docs/adr/`.

## Documentation

- **ADRs** : `docs/adr/` (12 décisions architecturales en MADR).
- **Specs** : `docs/specs/` (vision, domaine, features, API, données, qualité).
- **Diagrammes** : `docs/diagrams/` (UML use cases, classes, séquences).
- **Process Git** : `docs/process/git-workflow.md`.

## Démarrage rapide (à venir)

```bash
# Backend
cd backend && go run ./cmd/server

# Frontend
cd frontend && pnpm install && pnpm dev
```

## URLs cibles

- Production frontend : https://app.hello-there.net
- Production backend : https://api.hello-there.net
- WebSocket : wss://api.hello-there.net/v1/ws

## Licence

Code propriétaire de Nicolas Genin (2026). À reconsidérer post-soutenance (AGPL-3.0 envisagée).

## Contact

- Auteur : nic
- Dépôt : https://github.com/nicgen/komunumo
