# Feature - Notifications agrégées

## Objectif

Centraliser les évènements pertinents (follow, message, RSVP, post, modération) avec **agrégation** pour réduire le bruit cognitif et la charge réseau (argument eco + a11y).

## Règles métier

- 1 notification = (recipientId, kind, payloadJSON, createdAt, readAt, aggregateCount).
- Agrégation : si 2 événements même `kind` même `aggregationKey` < 5 minutes -> increment `aggregate_count` + update `created_at`.
- `aggregationKey` dépend du kind (ex: `msg.received -> conversationId`, `follow.requested -> targetId`).
- Email digest (mode opt-in) : 1/h max, regroupant les notifications non lues.
- Push web (V2 uniquement, MVP = polling/WS in-app).

## Kinds initiaux

- `follow.requested`, `follow.accepted`.
- `msg.received` (hors-ligne).
- `event.published`, `event.reminder` (cron 24h avant).
- `post.commented` (post à moi).
- `asso.invited`, `asso.request.received`.
- `mod.warning`, `mod.removed_content`.

## Scénarios Gherkin

```gherkin
Scenario: Agrégation messages
  Given Bob hors-ligne
  When Alice envoie 5 messages dans la même conversation en 2 minutes
  Then Bob a 1 notification msg.received avec aggregate_count=5

Scenario: Mark all read
  When je POST /v1/notifications/read-all
  Then toutes mes notifications ont read_at != null
```

## API

- `GET /v1/notifications?status=unread&limit=20`.
- `POST /v1/notifications/{id}/read`.
- `POST /v1/notifications/read-all`.
- `GET /v1/notifications/preferences`, `PATCH /v1/notifications/preferences`.

## Liens

- `04-data/mld.md` table `notifications` + `notification_preferences`.
- F2 dans `02-fonctionnalites-innovantes.md` (resources/studies).
