# ADR-0009 - Vercel pour le frontend, Scaleway DEV1-S pour le backend (démo), Traefik local pour le dev

- Statut : Accepté
- Date : 2026-04-26
- Révisé : 2026-04-26 (clarification de la chaîne dev/démo, le développeur n'a pas de VPS personnel)
- Décideur : nic

## Contexte

L'application doit être :
- **Développée en local** confortablement (TLS valide pour tester cookies cross-subdomain et WebSocket).
- **Démontrée au jury** depuis une URL publique, stable, durant la fenêtre de soutenance.

Le développeur **ne possède pas de VPS personnel** : Traefik (v2.11 + Cloudflare DNS-01 + 1Password) tourne sur sa machine de développement (`/home/nic/docker`) uniquement pour fournir un reverse proxy et un certificat TLS valide en localhost. Cette infra n'a pas vocation à servir le trafic externe permanent.

Le frontend Next.js bénéficierait du DX Vercel (preview PR, builds rapides, edge cache) tout en étant gratuit pour ce projet. Le backend Go avec WebSocket persistant ne tient pas dans une fonction serverless, il faut une VM ou un conteneur long-running.

La cible (associations, données potentiellement sensibles) impose de pouvoir argumenter la souveraineté.

## Décision

Architecture de déploiement à **trois étages** :

### 1. Développement local (jour le jour)

- Backend Go en `go run ./cmd/server` ou conteneur Docker, écoutant sur `:8080`.
- Frontend Next.js en `pnpm dev`, écoutant sur `:3000`.
- **Traefik local** (déjà configuré dans `/home/nic/docker`) sert de reverse proxy avec certificat TLS valide via Cloudflare DNS-01, exposant :
  - `https://app.local.hello-there.net` -> `localhost:3000`.
  - `https://api.local.hello-there.net` -> `localhost:8080`.
- DNS Cloudflare : entrées `*.local.hello-there.net` pointant vers `127.0.0.1` (résolution interne uniquement) ou IP locale réseau.
- Cookie de session avec `Domain=.local.hello-there.net` partagé entre `app.local` et `api.local`.

### 2. Preview / staging (à chaque PR)

- Frontend : **Vercel preview deployment** automatique sur PR (URL `pr-XXX-komunumo.vercel.app`).
- Backend : pointage configurable via variable d'environnement Vercel `NEXT_PUBLIC_API_URL`. En PR on peut pointer vers l'instance Scaleway de soutenance ou vers une instance dédiée de staging (selon budget).

### 3. Démo / soutenance (1 mois)

- **Frontend Next.js 16 sur Vercel** (free hobby plan, région `cdg1` Paris).
- **Backend Go + SQLite sur Scaleway DEV1-S** (instance VM Paris, ~3 €/mois HT, souverain FR).
  - URL publique : `https://api.komunumo.hello-there.net` (ou autre subdomain au choix, distinct de `local.`).
  - Certificat TLS via Caddy ou Traefik en standalone sur la VM (Let's Encrypt HTTP-01).
  - Volume bloc Scaleway monté pour `/data` (SQLite + uploads, cf. ADR-0011).
- **Email transactionnel** : Brevo (cf. ADR-0012).
- **Secrets** : 1Password personnel + GitHub Secrets pour la CI. Les secrets sont injectés au déploiement via `op run` côté local et via l'environnement de la VM côté démo.
- **Communication** : Next.js -> API en HTTPS, cookie de session avec `Domain=.hello-there.net` ou `__Host-session` selon stratégie finalisée.
- CORS strictement configuré côté Go (allow uniquement `app.komunumo.hello-there.net`).

### Argumentaire jury

> "L'environnement de développement utilise Traefik en local pour offrir un certificat TLS valide et reproduire la chaîne réseau de production. La démonstration est servie depuis une instance Scaleway DEV1-S à Paris, fournisseur français — choix cohérent avec le discours de souveraineté du projet, à 3 € pour la durée du jury. Vercel hébergera le frontend statique-first ; aucune donnée utilisateur n'y transite, seuls les rendus HTML et le routage."

## Alternatives écartées

- **Tout sur Vercel** : ne supporte pas un binaire Go long-running avec WebSocket persistant. Fonctions limitées en durée, hub WS impossible.
- **Tout sur Clever Cloud (FR)** : excellent côté souveraineté, mais Clever Cloud Cellar (S3) + scaler ajoute une plateforme à apprendre, et les tarifs free tier sont moins clairs que Scaleway pour un déploiement court d'1 mois.
- **Render.com / Fly.io free** : Render free met l'instance en sleep après 15 min d'inactivité, cold start ~30s, **incompatible avec un hub WebSocket** qui aurait besoin de rester chaud. Fly.io n'a plus de free tier en 2024-2026.
- **Koyeb free tier** : viable techniquement (1 Nano gratuit, supporte Go + WS), mais société américaine. Argument souveraineté affaibli.
- **Cloudflare Tunnel depuis machine perso** : zéro coût, expose le PC du dev. Risque démo (PC qui s'éteint, qui se met en veille, qui plante). Acceptable comme **plan B**, pas comme plan A.
- **OVH Public Cloud Discovery** : équivalent Scaleway, retenu si Scaleway pose problème (équivalent fonctionnel, ~2 €/mois).
- **Tout en self-hosted permanent** : nécessite un VPS ou un serveur dédié 24/7, coût mensuel récurrent injustifié pour un projet de soutenance.

## Conséquences

- (+) DX Vercel sur frontend (preview PR, edge cache, builds rapides).
- (+) Backend démo souverain (Scaleway = entreprise française, datacenter Paris).
- (+) Coût total démo très bas (~3 € HT pour 1 mois), justifiable pour la soutenance.
- (+) Cookie cross-subdomain fonctionne (deux sous-domaines du même domaine racine).
- (+) Réutilisation de l'infra Traefik perso pour le dev, **sans dépendance permanente**.
- (-) **Trois environnements** à gérer (local, staging Vercel, démo Scaleway). Mitigation : config par variables d'env, scripts de bootstrap documentés dans `ops/`.
- (-) Latence Vercel CDG -> Scaleway PAR : faible (~5-15 ms), acceptable.
- (-) Vercel = entreprise US (CLOUD Act). Mitigation : frontend = rendus + routage, aucune PII utilisateur stockée. Données = backend Scaleway FR.
- (-) Démantèlement à prévoir post-soutenance pour ne pas payer indéfiniment. Inscrit dans la checklist de fin de projet.

## Références

- Vercel régions Europe - https://vercel.com/docs/edge-network/regions
- Scaleway DEV1-S - https://www.scaleway.com/fr/instances/dev1/
- Traefik local existant : `/home/nic/docker/docker-compose.yml`
- DINUM Doctrine "Cloud au Centre" - https://www.numerique.gouv.fr/doctrine/
- ADR-0011 (stockage fichiers).
- ADR-0012 (email Brevo).
