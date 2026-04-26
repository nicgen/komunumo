# Feature - Posts et commentaires

## Objectif

Diffuser de courts contenus textuels (+1 image optionnelle) auprès de followers ou de membres d'asso.

## Règles métier

- `Post.visibility` : `public`, `followers`, `association_members` (si auteur=asso), `private_list` (sélection explicite de comptes).
- Texte 1..2000 caractères, sanitize via `bluemonday.UGCPolicy()`.
- 1 image max, AVIF généré serveur, alt obligatoire (RGAA 1.1).
- Un commentaire dépend d'un post visible par l'auteur du commentaire.
- Suppression auteur : soft delete (`deleted_at`) + audit log.
- Suppression modérateur asso : possible si post posté dans le contexte asso.

## Scénarios Gherkin

```gherkin
Scenario: Publication d'un post avec image
  Given je suis connecté en tant que member
  When je POST /v1/posts avec content="Hello" et image
  Then la réponse est 201
  And le post apparait dans le feed de mes followers

Scenario: Lecture feed
  Given je suis connecté
  When je GET /v1/feed?cursor=
  Then je reçois mes posts + ceux des comptes que je suis (ordre antechronologique)
  And la pagination cursor est "eyJsYXN0SWQiOiJ..."
```

## API

- `POST /v1/posts` (multipart si image).
- `GET /v1/posts/{id}` (visibilité respectée).
- `DELETE /v1/posts/{id}`.
- `GET /v1/feed?cursor=...&limit=20`.
- `POST /v1/posts/{id}/comments`.
- `GET /v1/posts/{id}/comments?cursor=...`.

## Critères d'acceptation

- [ ] Sanitization HTML (test injection `<script>`).
- [ ] Visibilité respectée (test `private_list` exclut les non listés).
- [ ] Pagination cursor stable.
- [ ] Index `(author_id, created_at DESC)` performant à 100k posts.

## Liens

- `04-data/mld.md` tables `posts`, `comments`, `post_audience`.
- `02-features/search.md` index FTS5 sur posts.
