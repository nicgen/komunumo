# ADR-0014 - coder/websocket avec hub en mémoire (remplace ADR-0005)

- Statut : Accepté
- Date : 2026-04-26
- Décideur : nic
- Remplace : [ADR-0005](./0005-gorilla-websocket-hub-memoire.md)

## Contexte

ADR-0005 retenait initialement `gorilla/websocket`. Vérification au moment de l'implémentation : la bibliothèque a été mise en mode maintenance par ses auteurs en 2022. Bien qu'un fork communautaire ait repris le repo en 2024, l'activité reste lente et les CVE potentielles s'accumulent. Pour un projet critique sécurité, on préfère une bibliothèque activement développée.

`coder/websocket` (anciennement `nhooyr.io/websocket`) est :
- Activement maintenu par l'équipe Coder.
- Plus moderne (API context-aware, support natif `context.Context` pour annulation).
- Conçu pour Go 1.18+, sans dépendances externes.
- Intègre nativement `wsjson` pour les messages JSON.
- Compatible avec `net/http.Hijacker` standard.

## Décision

Implémenter le **hub WebSocket en mémoire** dans le backend Go, basé sur **`github.com/coder/websocket`** :

- Une `goroutine` "reader" et une "writer" par client (lecture via `conn.Read(ctx)`).
- Un `Hub` central qui maintient la table des clients connectés (`sync.Map[accountID]*Client`).
- Des "rooms" pour les groupes (DM = room virtuelle entre 2 comptes).
- Authentification via cookie de session validé à l'upgrade HTTP.
- Heartbeat ping/pong toutes les 30s via `conn.Ping(ctx)`, déconnexion à 60s sans pong.
- Persistance asynchrone des messages dans SQLite via channel bufferisé.

API library typique :
```go
import "github.com/coder/websocket"

c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
    Subprotocols: []string{"komunumo.v1"},
    OriginPatterns: []string{"app.local.hello-there.net"},
})
if err != nil { /* ... */ }
defer c.CloseNow()

ctx, cancel := context.WithTimeout(r.Context(), time.Hour)
defer cancel()

for {
    var msg ClientMessage
    if err := wsjson.Read(ctx, c, &msg); err != nil {
        return // connection closed or error
    }
    // ...
}
```

## Alternatives écartées

- **`gorilla/websocket`** : maintenance lente, plus dans la liste de "go-recommendations" depuis 2023. Cf. ADR-0005.
- **`nbio/nbhttp`** : haute performance epoll, mais API plus complexe pour un MVP.
- **Server-Sent Events (SSE)** : unidirectionnel, ne couvre pas le typing/présence client -> serveur.
- **Long-polling HTTP** : surcoût réseau important, contraire à l'argument eco.
- **NATS / Redis Pub/Sub dès le MVP** : sur-dimensionné. Voir doc `studies/04-perspectives-scalabilite.md` pour la phase 2 multi-instances.
- **Solution managée (Pusher, Ably)** : verrouillage tiers, surcoût, pas de souveraineté.

## Conséquences

- (+) Bibliothèque activement maintenue.
- (+) API context-aware native, plus simple à intégrer dans un service Go moderne.
- (+) Latence minimale (pas de hop réseau supplémentaire).
- (+) ~50k connexions par instance Go atteignables sur un VPS modeste.
- (+) Aucune dépendance externe à exploiter en MVP.
- (+) Validation `OriginPatterns` native (anti-CSWSH par défaut).
- (-) Mono-instance : redéploiement = coupure WS. Mitigation : reconnexion auto côté client (exponential backoff, max 60s).
- (-) Pas de scale horizontal sans refonte. Documenté en perspective (NATS comme bus inter-instances).
- (-) Moins de tutoriels que gorilla/websocket. Mitigation : la doc officielle Coder est complète, et l'API est plus simple.

## Argumentaire jury

> "Le choix initial s'était porté sur gorilla/websocket, bibliothèque historique de la communauté Go. La vérification au moment de l'implémentation a montré que la maintenance est ralentie depuis 2022 (mise en archive temporaire, repris en mode bénévole). Pour un projet à forte exigence sécurité (chat, notifications, données associatives), j'ai documenté un nouvel ADR qui pivote vers coder/websocket, activement maintenu et conçu autour de `context.Context`. La bascule est triviale (~30 minutes de code) car la couche métier dépend d'un port `WSConnection`, pas de la lib directement."

## Références

- coder/websocket - https://github.com/coder/websocket
- Comparison nhooyr vs gorilla - https://github.com/coder/websocket/blob/master/docs/architecture.md
- ADR-0005 (remplacé).
- ADR-0001 (port `WSConnection` abstrait).
