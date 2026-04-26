# Feature - Profils (Personne et Association)

## Objectif

Permettre aux comptes (member ou association) de présenter une identité publique ou semi-privée, conforme RGPD et RGAA.

## Acteurs

- Personne authentifiée (kind=member).
- Association authentifiée (kind=association).
- Visiteur non authentifié (lecture des profils publics uniquement).

## Règles métier

- Un `Account` a 0..1 `Member` OU 0..1 `Association` (jamais les deux).
- Champs Personne : firstName, lastName, birthDate (≥ 13 ans à l'inscription, RGPD), nickname, aboutMe (≤ 500 car), avatar.
- Champs Association : legalName, siren (9 chiffres) OU rna (W + 9 chiffres), postalCode, about (≤ 2000 car), logo.
- Visibility : `public`, `members_only` (logged in), `private` (followers acceptés uniquement).
- Modifier email : déclenche revalidation par token + log audit.

## Scénarios Gherkin (extraits)

```gherkin
Scenario: Personne complète son profil
  Given je suis connecté en tant que member status=active
  When je PATCH /v1/profile avec {nickname:"Léa", aboutMe:"..."}
  Then la réponse est 200
  And l'audit log contient action="profile.updated"

Scenario: Asso renseigne SIREN invalide
  Given je suis connecté en tant qu'association
  When je PATCH /v1/profile avec siren="123"
  Then la réponse est 422
  And la réponse contient "siren must be 9 digits"
```

## API

- `GET /v1/profile/me` (auth requis).
- `PATCH /v1/profile` (auth requis).
- `GET /v1/profile/{accountId}` (visibilité respectée).
- `POST /v1/profile/avatar` (multipart, ≤ 2 Mo, redim serveur 512x512).

## Critères d'acceptation

- [ ] Validation SIREN/RNA serveur.
- [ ] Avatar AVIF généré côté serveur.
- [ ] Champs PII jamais retournés à un visiteur non autorisé.
- [ ] Audit log sur modifications.

## Liens

- `04-data/mld.md` table `members`, `associations`.
- `05-quality/security.md` V2.4 (PII protection).
