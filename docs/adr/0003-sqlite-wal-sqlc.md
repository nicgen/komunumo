# ADR-0003 - SQLite + WAL + sqlc comme couche persistance

- Statut : Accepté
- Date : 2026-04-26
- Décideur : nic

## Contexte

Le brief CDA exige une base **relationnelle SQL**. La cible AssoLink (associations + TPE/PME locales, < 10k MAU pilote) ne nécessite pas un SGBD réseau. La performance SQLite en WAL atteint 50k+ écritures/sec, largement au-dessus de la charge prévue. Le projet doit rester sobre énergétiquement (argument différenciant) et simple à exploiter.

## Décision

Utiliser **SQLite** en mode **WAL** (`PRAGMA journal_mode=WAL; synchronous=NORMAL`) avec **sqlc** pour générer des fonctions Go typées à partir de fichiers `.sql` versionnés. Migrations gérées par `golang-migrate`. Driver `modernc.org/sqlite` (pure Go, sans CGO) pour faciliter le build cross-platform en CI.

## Alternatives écartées

- **PostgreSQL direct** : sur-dimensionné pour la charge cible, surcoût hébergement, complexité opérationnelle (sauvegarde, monitoring, replicas) injustifiée pour le MVP. Sera réévalué si LiteFS / read-replicas SQLite ne suffisent plus (cf. doc `studies/04-perspectives-scalabilite.md`).
- **MySQL / MariaDB** : aucun avantage net sur Postgres pour notre cas, écosystème Go moins fourni.
- **GORM ou autre ORM Go** : performance moindre, génère du SQL imprévisible. sqlc reste plus simple à auditer.

## Conséquences

- (+) Sobriété énergétique : pas de serveur DB séparé, un fichier sur disque.
- (+) ACID complet, FTS5 pour la recherche full-text (cf. F3), JSON1 pour colonnes flexibles.
- (+) Sauvegarde triviale (cron + cp + chiffrement).
- (+) sqlc empêche les erreurs SQL/types à la compilation. Sécurité : pas de string concaténation possible.
- (-) Un seul écrivain à la fois. Acceptable pour notre cible. Mitigation : transactions courtes, file de jobs pour les écritures longues.
- (-) Pas de réplication native. Mitigation perspectives : Litestream (S3) puis LiteFS si besoin.

## Références

- SQLite "When to use" - https://www.sqlite.org/whentouse.html
- sqlc - https://docs.sqlc.dev/
- modernc.org/sqlite (pure Go) - https://pkg.go.dev/modernc.org/sqlite
