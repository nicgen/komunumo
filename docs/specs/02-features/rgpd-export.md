# Feature - Export RGPD et suppression de compte (F9)

## Objectif

Garantir les droits RGPD : portabilité (article 20), accès (article 15), effacement (article 17).

## Règles métier

- Export JSON + médias dans une archive ZIP signée (HMAC).
- Génération asynchrone (job in-process queue, livrable < 24h MVP, < 1h en pratique).
- Téléchargement via lien temporaire (24h, à usage unique).
- Suppression de compte : soft delete (anonymisation des contributions) + suppression physique des PII après 30 jours (purge cron).
- Posts/commentaires conservés mais signés "Compte supprimé" (anonymisation).
- Audit log : entrée `account.delete` conservée 5 ans (intérêt légitime, mention RGPD).

## Contenu de l'export

```
export-<accountId>-<date>.zip
  manifest.json        (version, generatedAt, hmac)
  account.json         (PII : email, profil, préférences)
  posts.json           (mes posts)
  comments.json        (mes commentaires)
  messages.json        (conversations + messages)
  follows.json         (mes follows et followers)
  events.json          (mes RSVP)
  audit-log.json       (mes actions)
  media/
    avatar.avif
    posts/<postId>.avif
    ...
```

## Scénarios Gherkin

```gherkin
Scenario: Demande d'export
  When je POST /v1/rgpd/export-request
  Then la réponse est 202 et job=in_progress
  When le job se termine
  Then je reçois un email avec lien /v1/rgpd/download/<token>

Scenario: Suppression de compte
  When je DELETE /v1/account avec confirmation password
  Then mon compte passe en status=deleted
  And mes posts deviennent "Compte supprimé"
  And dans 30j la purge supprime les PII restantes
  And l'audit log contient action="account.delete"
```

## API

- `POST /v1/rgpd/export-request`.
- `GET /v1/rgpd/exports` (mes demandes).
- `GET /v1/rgpd/download/{token}`.
- `DELETE /v1/account` body {password}.

## Liens

- `05-quality/security.md` V8 (data protection).
- F9 dans `02-fonctionnalites-innovantes.md`.
- ADR-0003 (rétention SQLite).
