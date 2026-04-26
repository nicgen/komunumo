# ADR-0011 - Stockage des fichiers utilisateurs sur volume Docker local

- Statut : Accepté
- Date : 2026-04-26
- Décideur : nic

## Contexte

L'application accepte des uploads utilisateurs limités au MVP : avatars (Personne, Association) et 1 image par post. Limite par fichier : 2 Mo en entrée, redimensionné serveur à 1280px max et converti AVIF (cf. `05-quality/eco.md`). Volume estimé sur 4 mois : < 5 Go pour 500 utilisateurs.

Le serveur cible héberge déjà Traefik v2.11 et plusieurs services Docker dans `/home/nic/docker`. Aucun service S3-compatible n'y est actuellement déployé. Les solutions externes (Scaleway Object Storage, Cloudflare R2) ajoutent de la complexité (clés d'accès, SDK, signed URLs, facturation) pour un volume modeste au MVP.

## Décision

**Volume Docker nommé `komunumo_uploads` monté dans le conteneur backend Go**, exposé en lecture par Traefik via une route statique dédiée (`/uploads/*`).

Configuration `docker-compose` (extrait) :
```yaml
services:
  komunumo_api:
    volumes:
      - komunumo_uploads:/app/uploads
volumes:
  komunumo_uploads:
    driver: local
    driver_opts:
      type: none
      device: /home/nic/docker/data/komunumo/uploads
      o: bind
```

Convention de nommage des fichiers : `<entityType>/<entityId>/<fileId>.<ext>` (ex : `posts/01J9.../9d2a.avif`). Aucun nom utilisateur, donc pas d'IDOR par énumération.

Sauvegarde : `restic` snapshot quotidien chiffré vers stockage externe (script existant à étendre, cf. SRE backlog).

## Alternatives écartées

- **Scaleway Object Storage** : excellent choix souverain (Paris) et S3-compatible. Coût ~0,01 €/Go/mois, donc ~5 c€/mois en MVP. Écarté pour le MVP car ajoute SDK AWS, signed URLs, et un secret à gérer. **Plan d'évolution V2 documenté.**
- **Cloudflare R2** : zero egress, mais Cloudflare = entreprise US (CLOUD Act). Contradiction avec l'argumentaire souveraineté.
- **Volume monté hors Docker (NFS, S3FS-FUSE)** : trop de complexité pour un MVP solo.
- **Stockage en base SQLite (BLOB)** : viable en théorie pour un volume aussi faible, mais alourdit le fichier DB, ralentit les sauvegardes, et complique le service HTTP statique.

## Conséquences

- (+) Mise en œuvre triviale (volume Docker, montage Traefik static).
- (+) Aucun coût additionnel, aucun secret externe à gérer.
- (+) Latence locale, pas de round-trip réseau.
- (+) Sauvegarde unifiée avec le reste des données du serveur.
- (-) **Pas de réplication.** Perte du serveur = perte des fichiers (mitigation : sauvegarde restic chiffrée externe quotidienne).
- (-) **Pas de scalabilité horizontale.** Un seul nœud peut servir les fichiers (acceptable, le backend est mono-instance en MVP).
- (-) Migration vers S3 nécessaire si > 50 Go ou multi-noeud. **Stratégie d'évolution V2 :**
  1. Adapter le port `FileStorage` (déjà abstrait dans `internal/ports/`) à un adapter S3 (Scaleway).
  2. Script de migration `cmd/migrate-files-to-s3` (lit le volume, écrit S3, met à jour les chemins en DB).
  3. Bascule en bleu/vert (URL pointe vers nouvelle origine).

## Argumentaire jury

> "Pour le MVP, le stockage des fichiers utilisateurs est confié à un volume Docker local sur le serveur Traefik. Ce choix est dicté par la simplicité (zéro intégration tierce) et par le faible volume attendu (< 5 Go sur 4 mois). Le port `FileStorage` est cependant abstrait dans la couche application : la migration vers Scaleway Object Storage (ou tout backend S3-compatible souverain) est anticipée et documentée comme étape de mise à l'échelle V2, déclenchée à 50 Go ou si le service passe en multi-nœud."

## Références

- ADR-0001 (architecture hexagonale, port `FileStorage`).
- ADR-0009 (déploiement hybride).
- Scaleway Object Storage - https://www.scaleway.com/fr/object-storage/
