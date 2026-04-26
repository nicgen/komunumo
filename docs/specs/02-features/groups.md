# Feature - Adhésions associations (memberships)

## Objectif

Modéliser l'appartenance d'une Personne à une Association avec rôles et cycle de vie.

## Règles métier

- Rôles : `owner` (≥ 1, propriétaire historique, transférable), `admin`, `member`.
- Statuts : `pending` (demande), `active`, `suspended`, `left`.
- Workflow d'invitation : asso -> POST invite -> notification au member -> acceptation = active.
- Workflow demande : member -> POST request -> admin asso valide -> active.
- Owner ne peut pas se retirer s'il est seul owner ; doit transférer d'abord.
- Une suspension par admin : member ne peut plus poster dans le canal asso ni voir contenu `association_members`.

## Scénarios Gherkin

```gherkin
Scenario: Invitation acceptée
  Given je suis owner de "Repair Café"
  When je POST /v1/associations/{id}/invitations {memberAccountId}
  Then une notification "asso.invited" est créée
  When le member POST /v1/memberships/{id}/accept
  Then status="active" et role="member"

Scenario: Transfert d'ownership
  Given je suis owner unique
  When je POST /v1/memberships/{id}/transfer-ownership {newOwnerId}
  Then mon role devient "admin"
  And l'autre devient "owner"
  And l'audit log trace l'opération
```

## API

- `POST /v1/associations/{id}/invitations`.
- `POST /v1/associations/{id}/requests` (member candidate).
- `POST /v1/memberships/{id}/accept`.
- `POST /v1/memberships/{id}/decline`.
- `POST /v1/memberships/{id}/suspend` (admin).
- `POST /v1/memberships/{id}/transfer-ownership`.
- `DELETE /v1/memberships/{id}` (leave).

## Liens

- `04-data/mld.md` tables `memberships`, `membership_invitations`.
- `02-features/audit-log.md` toutes opérations sensibles tracées.
