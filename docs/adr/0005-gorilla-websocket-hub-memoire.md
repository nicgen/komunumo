# ADR-0005 - gorilla/websocket avec hub en mémoire

- Statut : Accepté
- Date : 2026-04-26
- Décideur : nic

## Contexte

AssoLink propose du chat (DM et groupe), des notifications temps réel et de la présence granulaire (typing, online). Le MVP tourne sur un seul process Go derrière Traefik. La cible (< 10k MAU pilote) reste largement sous la capacité d'un hub mémoire monolithique.

## Décision

Implémenter un **hub WebSocket en mémoire** dans le backend Go, basé sur **gorilla/websocket** :

- Une `goroutine` "reader" et une "writer" par client.
- Un `Hub` central qui maintient la table des clients connectés (`sync.Map[userID]*Client`).
- Des "rooms" pour les groupes (DM = room virtuelle entre 2 users).
- Authentification via cookie de session validé à l'upgrade HTTP.
- Heartbeat ping/pong toutes les 30s, déconnexion à 60s sans pong.
- Persistance asynchrone des messages dans SQLite via channel bufferisé.

## Alternatives écartées

- **nhooyr.io/websocket / coder/websocket** : API plus moderne, mais moins de tutoriels et d'exemples pour un MVP rapide. À envisager en refonte V2.
- **Server-Sent Events (SSE)** : unidirectionnel, ne couvre pas le typing/présence client -> serveur.
- **Long-polling HTTP** : surcoût réseau important, contraire à l'argument eco.
- **NATS / Redis Pub/Sub dès le MVP** : sur-dimensionné. Voir doc `studies/04-perspectives-scalabilite.md` pour la phase 2 multi-instances.
- **Solution managée (Pusher, Ably)** : verrouillage tiers, surcoût, pas de souveraineté.

## Conséquences

- (+) Latence minimale (pas de hop réseau supplémentaire).
- (+) ~50k connexions par instance Go atteignables sur un VPS modeste.
- (+) Pas de dépendance externe à exploiter en MVP.
- (+) Conforme à la liste des packages typiquement autorisés en formation (gorilla/websocket).
- (-) Mono-instance : redéploiement = coupure WS. Mitigation : reconnexion auto côté client (exponential backoff).
- (-) Pas de scale horizontal sans refonte. Documenté en perspective (NATS comme bus inter-instances).

## Références

- gorilla/websocket - https://pkg.go.dev/github.com/gorilla/websocket
- "Building a real-time chat in Go" patterns standards
- `studies/04-perspectives-scalabilite.md`
