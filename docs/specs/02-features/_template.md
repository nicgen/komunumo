# Feature - <Nom de la feature>

- Status : `Draft` | `Approved` | `Implemented`
- Owner : nic
- Last updated : 2026-MM-DD
- Linked ADRs : ADR-XXXX

## Objectif

Une à deux phrases décrivant la valeur métier.

## Personas concernés

- P1 (Anne, présidente)
- P2 (Karim, gérant TPE)
- ...

## User stories

- En tant que **<persona>**, je veux **<action>** afin de **<bénéfice>**.

## Critères d'acceptation (Gherkin)

```gherkin
Feature: <nom>

  Scenario: <cas nominal>
    Given <pré-condition>
    When <action>
    Then <résultat attendu>
    And <résultat additionnel>

  Scenario: <cas d'erreur>
    Given <pré-condition>
    When <action invalide>
    Then <message d'erreur attendu>
    And <état système inchangé>
```

## Règles métier

- Règle 1
- Règle 2

## Permissions

| Rôle | Action |
|------|--------|
| ... | ... |

## Modèle de données impacté

Tables touchées : `users`, `posts`, ...

## Endpoints API impactés

- `POST /api/v1/<resource>`
- `GET /api/v1/<resource>/{id}`

## Considérations

### RGAA
- Critères concernés : 9.1, 11.1, ...
- Tests automatisés : pa11y sur le parcours complet.

### Éco-conception
- Budget poids : ...
- Budget requêtes : ...

### Sécurité
- Authentification requise : oui/non
- Rate limit : N requêtes / minute / utilisateur
- Validation entrée : Zod côté client + validator côté serveur
- Logs : événement loggé en `slog` niveau info ; PII redaction

## Hors scope

- ...

## Open questions

- ...
