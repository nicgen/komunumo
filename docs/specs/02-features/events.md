# Feature - Événements associatifs

## Objectif

Permettre aux associations de créer des événements et aux comptes de répondre (RSVP).

## Règles métier

- Seuls les `admin` ou `owner` d'une asso créent un événement.
- `startsAt < endsAt`, `endsAt > now` à la création.
- `capacity` optionnelle ; si atteinte, RSVP `going` rejetés -> file `waitlist`.
- 3 réponses possibles : `going`, `maybe`, `no`.
- Visibilité événement = visibilité de l'asso (public, followers, members).
- Filtrage par `postalCode` (préfixe ou rayon code postal).

## Scénarios Gherkin

```gherkin
Scenario: Création événement
  Given je suis admin de "Ressourcerie 91"
  When je POST /v1/events {title, startsAt, endsAt, postalCode}
  Then la réponse est 201
  And mes followers reçoivent notification "event.published"

Scenario: RSVP avec capacité atteinte
  Given un événement avec capacity=10 et 10 going
  When je POST /v1/events/{id}/rsvp {response:"going"}
  Then je suis ajouté en waitlist
  And la réponse contient {position:1}
```

## API

- `POST /v1/events`, `PATCH /v1/events/{id}`, `DELETE /v1/events/{id}`.
- `GET /v1/events?postalCode=91&from=...&to=...`.
- `POST /v1/events/{id}/rsvp` body {response}.
- `GET /v1/events/{id}/rsvps` (admin asso uniquement).

## Liens

- `04-data/mld.md` tables `events`, `event_rsvps`.
