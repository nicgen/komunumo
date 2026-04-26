# Feature - Messagerie temps réel

## Objectif

Permettre conversations 1:1 et canal d'asso, en temps réel via WebSocket, avec persistance SQLite.

## Règles métier

- 2 types : `direct` (2 participants exactement) et `association` (canal d'une asso, lecture par membres actifs).
- Un message a 1..2000 caractères, sanitize.
- Pas d'édition ni suppression message en MVP (immutabilité = simplicité + audit).
- Présence `online/away/offline` exposée via WS broadcast.
- `chat.typing` debounce 500ms client.
- Rate limit : 60 messages/minute/compte.

## Contrat WebSocket

Voir `03-api/websocket.md` (frames JSON `{type, ...}`).

## Scénarios Gherkin

```gherkin
Scenario: Envoi message direct
  Given Alice et Bob ont une conversation directe id=conv1
  And les deux sont connectés au WS
  When Alice envoie {type:"chat.send", convId:"conv1", content:"Hi"}
  Then Bob reçoit {type:"chat.message", senderId:"alice", content:"Hi"}
  And Alice reçoit l'echo

Scenario: Canal asso lecture
  Given je suis membre actif de "Repair Café"
  When je GET /v1/conversations/asso-repair-cafe/messages?cursor=
  Then je vois les 50 derniers messages

Scenario: Hors-ligne -> notification
  Given Bob est offline
  When Alice envoie un message
  Then une notification "msg.received" est agrégée pour Bob
```

## API HTTP

- `POST /v1/conversations` body {kind, participants ou associationId}.
- `GET /v1/conversations` (mes conversations).
- `GET /v1/conversations/{id}/messages?cursor=...&limit=50`.
- `GET /v1/ws/upgrade` (handshake WebSocket).

## Critères d'acceptation

- [ ] Heartbeat ping/pong 30s, déconnexion auto à 60s sans pong.
- [ ] Reconnexion WS auto côté client (backoff exponentiel max 60s).
- [ ] Tests d'intégration : 2 clients, envoi/réception, drop connection.
- [ ] Rate limit testé (envoi 100 msg en 1s -> erreur 429 sur dernier paquet).

## Liens

- `04-data/mld.md` tables `conversations`, `conversation_members`, `messages`.
- `03-api/websocket.md` détail des frames.
- `02-features/notifications.md` agrégation msg hors-ligne.
