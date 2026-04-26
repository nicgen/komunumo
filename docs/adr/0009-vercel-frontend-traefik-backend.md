# ADR-0009 - Vercel pour le frontend, Traefik+VPS pour le backend

- Statut : Accepté
- Date : 2026-04-26
- Décideur : nic

## Contexte

L'application doit être déployée pour la démo et le contrôle jury. Le candidat dispose déjà d'une infrastructure **Traefik v2.11 + Cloudflare DNS-01 + 1Password** sur VPS personnel (`/home/nic/docker`). La cible (associations, données potentiellement sensibles) impose de pouvoir argumenter la souveraineté tout en respectant les délais MVP. Le frontend Next.js bénéficierait du DX Vercel pour livrer vite.

## Décision

Architecture de déploiement hybride :

- **Frontend Next.js 16 sur Vercel** (région Paris `cdg1` ou Francfort `fra1`).
- **Backend Go + SQLite sur VPS personnel** derrière Traefik existant. Service exposé sur `api.hello-there.net` via labels Traefik et certificat ACME Cloudflare DNS-01.
- **Stockage fichiers** : volume Docker nommé `komunumo_uploads` monté sur le serveur Traefik (cf. ADR-0011).
- **Email transactionnel** : Brevo (cf. ADR-0012).
- **Secrets** : 1Password injectés via `op run` au démarrage des conteneurs (pattern existant).
- **Communication** : Next.js -> API en HTTPS, cookie de session avec `Domain=.hello-there.net` partagé entre `app.hello-there.net` et `api.hello-there.net`. CORS strictement configuré côté Go.

Argumentaire jury : "Vercel pour la rapidité de mise sur le marché du frontend statique-first ; backend critique et données utilisateurs sur infra contrôlée en France, derrière Traefik que je gère personnellement. Migration complète sur Clever Cloud envisagée en V2."

## Alternatives écartées

- **Tout sur Vercel** : Next.js OK, mais Vercel ne fait pas tourner un binaire Go long-running ni un hub WebSocket persistant de manière satisfaisante (Functions limitées en durée). Inadapté.
- **Tout sur Clever Cloud** : très défendable côté souveraineté mais plus de friction DX en MVP, sans bénéfice immédiat.
- **Tout sur le VPS Traefik** (Next.js auto-hébergé) : viable, mais pas de CDN edge, builds lents, pas de previews PR. Mauvais tradeoff pour 4 semaines.
- **Scaleway Container** : alternative valable au VPS pour le backend, mais ajoute une plateforme à apprendre.

## Conséquences

- (+) DX Vercel sur frontend (preview PR, edge cache, builds rapides).
- (+) Souveraineté maintenue sur les données (DB et backend en France, sur infra perso).
- (+) Réutilisation immédiate de l'infra Traefik existante.
- (+) Cookie cross-subdomain fonctionne (deux sous-domaines du même domaine).
- (-) Deux plateformes de déploiement à orchestrer. Mitigation : pipelines séparées GH Actions, secrets séparés.
- (-) Latence Vercel -> VPS Go (variable). Mitigation : VPS en Europe, mesures Lighthouse régulières.
- (-) Vercel = entreprise US, soumise CLOUD Act. Mitigation : frontend ne stocke pas de données utilisateur, juste rendu. Argumentaire défendable.

## Références

- Setup Traefik existant : `/home/nic/docker/docker-compose.yml`
- Vercel régions Europe - https://vercel.com/docs/edge-network/regions
- DINUM Doctrine "Cloud au Centre" - https://www.numerique.gouv.fr/doctrine/
