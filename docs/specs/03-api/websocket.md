# Contrat WebSocket - AssoLink

URL : `wss://api.local.hello-there.net/v1/ws`

## Authentification

L'upgrade HTTP doit porter le cookie `session_id`. Le serveur valide la session ; rejette avec 401 sinon.

## Format des messages

Tous les messages (entrants et sortants) sont JSON :
```json
{ "type": "<event_type>", "ts": "<RFC3339>", "data": { ... } }
```

## Événements client -> serveur

| Type | Payload | Effet |
| ------ | --------- | ------- |
| `chat.send` | `{ conversation_id, content }` | Envoie un message |
| `chat.typing` | `{ conversation_id, typing: bool }` | Annonce typing |
| `chat.read` | `{ conversation_id, until_message_id }` | Marque lu |
| `presence.set` | `{ status: "active"\|"idle"\|"away" }` | MAJ présence |
| `subscribe` | `{ channel: "conversation:<id>" }` | S'abonne à un flux |
| `unsubscribe` | `{ channel: "..." }` | Désabonnement |

## Événements serveur -> client

| Type | Payload | Quand |
| ------ | --------- | ------- |
| `chat.message` | `{ conversation_id, message: {...} }` | Nouveau message |
| `chat.typing` | `{ conversation_id, account_id, typing }` | Typing tier |
| `chat.read` | `{ conversation_id, account_id, until }` | Read receipt |
| `presence.update` | `{ account_id, status, last_seen }` | Changement présence |
| `notification.new` | `{ notification: {...} }` | Notification (déjà agrégée si F2) |
| `error` | `{ code, message }` | Erreur applicative |
| `pong` | `{}` | Heartbeat réponse |

## Heartbeat

- Serveur envoie `ping` (frame WS standard) toutes les 30s.
- Si pas de `pong` après 60s, déconnexion.
- Côté client : ping de niveau applicatif `{type:"ping"}` autorisé en cas de blocage proxy.

## Limites

- 1 connexion par compte par défaut. La nouvelle déconnecte les anciennes (configurable plus tard).
- Rate limit : 60 messages `chat.send` / minute / compte.
- Taille max d'un message : 8 KB.

## Erreurs

Codes :
- `auth.required` (4001)
- `auth.invalid` (4002)
- `rate.limited` (4029)
- `payload.invalid` (4400)
- `permission.denied` (4403)
- `server.error` (5000)
