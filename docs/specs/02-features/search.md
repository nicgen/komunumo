# Feature - Recherche full-text

## Objectif

Recherche unifiée sur posts, associations, événements via SQLite **FTS5**.

## Règles métier

- Index FTS5 séparés par entité (`posts_fts`, `associations_fts`, `events_fts`).
- Tokenizer `unicode61` avec retrait des accents.
- Re-indexation via triggers `AFTER INSERT/UPDATE/DELETE`.
- Visibilité respectée : on ne renvoie que les contenus que le viewer peut voir (filtrage post-FTS au use-case).
- Pagination par cursor.

## Scénarios Gherkin

```gherkin
Scenario: Recherche par mot-clé
  Given des posts contenant "réparation"
  When je GET /v1/search?q=réparation&type=posts
  Then je reçois les résultats classés par BM25
  And la latence p95 < 50ms à 100k posts

Scenario: Filtrage géographique association
  When je GET /v1/search?q=ressourcerie&type=associations&postalCode=91
  Then les résultats sont filtrés sur préfixe code postal
```

## API

- `GET /v1/search?q=...&type=posts|associations|events&postalCode=...&cursor=...`.

## Critères d'acceptation

- [ ] Tests : recherche "café" trouve "Café", "cafés", "CAFÉ" (case + accents).
- [ ] Test de charge léger : 100k posts, requête sous 100ms p95.
- [ ] Aucun résultat privé (test viewer non autorisé).

## Liens

- `04-data/mld.md` section FTS5.
- F3 dans `02-fonctionnalites-innovantes.md`.
