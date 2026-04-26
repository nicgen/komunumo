# Feature - Suivi (Follows) et demandes

## Objectif

Permettre à un compte d'en suivre un autre (Personne ou Association). Selon la visibilité du suivi, accepter automatiquement ou via validation.

## Règles métier

- Un Follow a un état : `pending`, `accepted`, `declined`, `revoked`.
- Si `target.visibility = public` -> `accepted` direct.
- Si `target.visibility != public` -> `pending`, accepteur reçoit notification.
- Pas d'auto-suivi (followerId != targetId).
- Unicité (followerId, targetId).
- Unfollow = soft (status=revoked) pour préserver historique audit.

## Scénarios Gherkin

```gherkin
Scenario: Suivre une asso publique
  Given je suis connecté en tant que member
  And l'asso "Repair Café" a visibility=public
  When je POST /v1/follows {targetId:"<asso>"}
  Then le follow a status="accepted"
  And la notification "asso.new_follower" est créée pour l'asso

Scenario: Demander à suivre un member privé
  Given le member cible a visibility=members_only
  When je POST /v1/follows
  Then status="pending"
  When la cible POST /v1/follows/{id}/accept
  Then status="accepted"
```

## API

- `POST /v1/follows` body {targetId}.
- `GET /v1/follows/incoming?status=pending` (demandes reçues).
- `POST /v1/follows/{id}/accept`.
- `POST /v1/follows/{id}/decline`.
- `DELETE /v1/follows/{id}` (unfollow).

## Liens

- `04-data/mld.md` table `follows`.
- `02-features/notifications.md` kind=`follow.requested`, `follow.accepted`.
