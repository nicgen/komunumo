# Specifications - AssoLink

Specs Speckit qui pilotent le développement assisté LLM.

## Pourquoi des specs

1. Verrouillent le périmètre face à la dérive LLM ("refais selon la spec X").
2. Permettent de générer des **tests E2E à partir des critères d'acceptation** (Gherkin).
3. Servent de **source de vérité** pour le dossier de soutenance.

## Structure

```
specs/
  00-vision.md            # Pitch, personas, mission, contraintes
  01-domain.md            # Glossaire métier, invariants
  02-features/            # Une feature = un fichier
    auth.md
    posts.md
    groups.md
    chat.md
    notifications.md
    profile.md
    follows.md
    events.md
    search.md
    audit-log.md
    rgpd-export.md
  03-api/
    openapi.yaml          # Contrat REST
    websocket.md          # Contrat WS
  04-data/
    mcd.mmd               # Modèle conceptuel (Mermaid ER)
    mld.md                # Modèle logique (tables + clés)
    mpd.sql               # Modèle physique (DDL SQLite)
  05-quality/
    a11y.md               # Critères RGAA cible et automatisation
    eco.md                # Critères RGESN cible et mesures
    security.md           # Contrôles ASVS sélectionnés
    performance.md        # Budgets perf
    tests-strategy.md     # Stratégie de tests
```

## Convention de spec feature

Chaque feature suit le format :

```markdown
# Feature - <Nom>

## Objectif (1-2 phrases)
## Personas concernés
## User stories
  - En tant que ... je veux ... afin de ...
## Critères d'acceptation (Gherkin)
  Scenario: ...
    Given ... When ... Then ...
## Règles métier
## Permissions / autorisation
## Modèle de données impacté
## Endpoints API impactés
## Considérations RGAA / éco / sécurité
## Hors scope
## Status (Draft, Approved, Implemented)
```
