# Feature - Audit log HMAC chaîné (F6)

## Objectif

Garantir l'**intégrité** et la **traçabilité** des actions sensibles. Sceau cryptographique chaîné (HMAC-SHA256) inspiré de la blockchain mais en base SQL.

## Règles métier

- Table `audit_log` append-only (trigger `BEFORE UPDATE/DELETE` -> `RAISE(ABORT)`).
- Chaque ligne contient :
  - `id` BIGSERIAL.
  - `actor_id` (peut être null si action système).
  - `action` (string ex: `account.register`, `post.delete`, `membership.suspend`).
  - `target_type`, `target_id`.
  - `payload_json` (taille ≤ 4 ko, sans PII brute).
  - `prev_hash` (hash de la ligne précédente).
  - `hash` = HMAC-SHA256(secret, prev_hash || actor_id || action || target_type || target_id || payload_json || at).
  - `at` timestamp UTC.
- Secret stocké dans 1Password, injecté à l'exécution.
- Endpoint admin `GET /v1/audit/verify` re-calcule la chaîne et signale toute incohérence.

## Actions tracées (liste minimum)

| Action | Déclencheur |
| -------- | ------------- |
| `account.register` | UC1, UC2 |
| `account.verified` | UC4 |
| `login.ok` / `login.failed` | UC3 |
| `post.create` / `post.delete` | UC9 |
| `membership.invite` / `accept` / `suspend` / `transfer` | UC13, UC14 |
| `event.create` / `event.cancel` | UC16 |
| `mod.remove_content` / `mod.warn` | UC25 |
| `rgpd.export` / `account.delete` | UC23, UC24 |

## Scénarios Gherkin

```gherkin
Scenario: Vérification de la chaîne
  Given 1000 entrées dans audit_log
  When je GET /v1/audit/verify (admin uniquement)
  Then la réponse est {ok:true, lastVerifiedId:1000}

Scenario: Tentative d'altération
  Given une UPDATE manuelle sur audit_log
  Then la base lève RAISE(ABORT) - opération bloquée
```

## API

- `GET /v1/audit/verify` (rôle admin instance).
- `GET /v1/audit/me` (mon historique d'actions personnelles, RGPD).

## Liens

- `05-quality/security.md` V7.3.1.
- F6 dans `02-fonctionnalites-innovantes.md`.
