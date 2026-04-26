# ADR-0001 - Architecture hexagonale pour le backend Go

- Statut : Accepté
- Date : 2026-04-26
- Décideur : nic

## Contexte

Le brief CDA exige une architecture en **3 couches** (présentation, métier, accès données). Le backend Go doit pouvoir être testé sans dépendance forte à SQLite ou au transport HTTP, et permettre une évolution future (changement de DB, ajout d'API mobile, fédération ActivityPub) sans réécrire le métier.

## Décision

Adopter une **architecture hexagonale légère** (ports & adapters) :

```
internal/
  domain/         # entités et règles pures (User, Asso, Post...)
  application/    # use cases (CreatePost, FollowUser...)
  ports/          # interfaces (UserRepo, EventBus, PasswordHasher)
  adapters/
    http/         # handlers REST
    ws/           # handlers WebSocket
    sqlite/       # repos concrets
    auth/         # sessions, bcrypt
cmd/server/main.go  # composition root
```

Les use cases ne dépendent que des **interfaces** des ports, jamais des implémentations.

## Alternatives écartées

- **MVC classique** : confond couche métier et couche transport, rend les tests dépendants du framework HTTP.
- **Clean Architecture stricte** (couches concentriques 4+) : sur-dimensionné pour 4 semaines de MVP, trop de cérémonie.
- **Architecture en oignon** : très proche de l'hexagonale, mais terminologie moins standard dans la communauté Go.

## Conséquences

- (+) Tests unitaires des use cases sans DB ni HTTP (mocks des ports).
- (+) Transition future SQLite -> Postgres ne touche que `adapters/sqlite/`.
- (+) Cadre clair pour répondre aux questions du jury sur la séparation des préoccupations.
- (-) Plus de fichiers et d'interfaces qu'un MVC. Mitigation : interfaces définies au plus près de leur consommateur (Go idiomatic), pas dans un package dédié.
- (-) Légère duplication entre entités domaine et DTOs HTTP. Mitigation : assumée pour découpler.

## Références

- "Hexagonal Architecture" - Alistair Cockburn, https://alistair.cockburn.us/hexagonal-architecture/
- "Standard Go Project Layout" - https://github.com/golang-standards/project-layout
