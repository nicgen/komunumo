# Feature - Profils (Personne et Association)

- Status : `Approved`
- Owner : nic
- Last updated : 2026-05-02
- Linked ADRs : ADR-0001, ADR-0003, ADR-0006

## Objectif

Permettre aux comptes (member ou association) de présenter une identité publique ou semi-privée, conforme RGPD et RGAA.

## Acteurs

- Personne authentifiée (kind=member).
- Association authentifiée (kind=association).
- Visiteur non authentifié (lecture des profils publics uniquement).

## Règles métier

- Un `Account` a exactement 1 `Member` OU 1 `Association` (jamais les deux), créé au moment de l'inscription.
- `accounts` ne contient que les credentials (email, password_hash, status, kind). Les données d'identité vivent dans `members` ou `associations`.
- Champs Personne (`members`) : first_name, last_name, birth_date, nickname, about_me (≤ 500 car), avatar_path.
- Champs Association (`associations`) : legal_name, siren (9 chiffres, optionnel), rna (W + 9 chiffres, optionnel), postal_code, about (≤ 2000 car), logo_path.
- Âge minimum inscription Personne : **≥ 18 ans** (RGPD — mineur sans accord parental, V1). La gestion de l'accord parental est déférée en V2.
- Visibility : `public` (tout le monde), `members_only` (connecté uniquement), `private` (followers acceptés uniquement).
- Un seul des deux, siren ou rna, est obligatoire pour une Association (les deux peuvent être renseignés).
- Modifier email : déclenche revalidation par token + audit log.

## Décisions de cadrage Phase 2

| Décision | Retenue | Justification |
|----------|---------|---------------|
| Localisation PII (first_name, last_name, birth_date) | `members` (MLD) | Séparation nette auth/identité ; MLD fait autorité |
| Valeurs status | `active`, `suspended`, `deleted` | Alignement MLD + RGPD article 17 (soft-delete) |
| Visibility | `public`, `members_only`, `private` | 3 niveaux de visibilité |
| Âge minimum | 18 ans V1 | Mineur sans accord parental hors scope V1 |
| Avatar AVIF | Déféré V2 | Constitution principe II : stocker l'original en V1 |
| Chemin API | `/api/v1/...` | Convention existante |

## Migration Phase 1 → Phase 2

La migration `0002_profiles` doit :
1. Ajouter `kind TEXT NOT NULL DEFAULT 'member' CHECK(kind IN ('member','association'))` sur `accounts`.
2. Modifier les valeurs de `status` : renommer `verified` → `active`, `disabled` → `suspended`, ajouter la valeur `deleted`.
3. Créer la table `members` avec les colonnes PII migrées depuis `accounts` (first_name, last_name, birth_date).
4. Créer la table `associations`.
5. Créer la table `memberships`.
6. Migrer les lignes existantes : `INSERT INTO members(account_id, first_name, last_name, birth_date, visibility) SELECT id, first_name, last_name, date_of_birth, 'public' FROM accounts`.
7. Supprimer les colonnes `first_name`, `last_name`, `date_of_birth`, `email_canonical`, `last_login_at` de `accounts` (recréation de la table — SQLite).

## Scénarios Gherkin

```gherkin
Feature: Inscription Personne

  Scenario: Inscription réussie d'une Personne
    Given un visiteur sur POST /api/v1/auth/register/member
    When il soumet email, password, first_name, last_name, birth_date (>= 18 ans)
    Then un compte kind=member + status=pending_verification est créé
    And une ligne members est créée avec visibility=public
    And un email de vérification est envoyé
    And l'audit log enregistre "account_created"

  Scenario: Inscription avec âge insuffisant
    Given un visiteur avec birth_date < 18 ans
    When il soumet le formulaire
    Then la réponse est 422 avec "vous devez avoir au moins 18 ans"

Feature: Inscription Association

  Scenario: Inscription réussie d'une Association
    Given un visiteur sur POST /api/v1/auth/register/association
    When il soumet email, password, legal_name, postal_code, first_name, last_name, birth_date du créateur
    Then un compte kind=association + status=pending_verification est créé
    And une ligne associations est créée
    And une ligne memberships est créée avec role=owner, status=active
    And un email de vérification est envoyé
    And l'audit log enregistre "account_created"

  Scenario: SIREN invalide
    Given un visiteur sur POST /api/v1/auth/register/association
    When il soumet siren="123"
    Then la réponse est 422 avec "siren must be 9 digits"

Feature: Profil

  Scenario: Personne complète son profil
    Given je suis connecté en tant que member status=active
    When je PATCH /api/v1/me/profile avec {nickname:"Léa", about_me:"..."}
    Then la réponse est 200
    And l'audit log contient action="profile.updated"

  Scenario: Visiteur consulte un profil public
    Given un profil member avec visibility=public
    When GET /api/v1/accounts/{id}/profile sans auth
    Then la réponse est 200 sans PII sensibles (birth_date absent)

  Scenario: Visiteur bloqué sur profil privé
    Given un profil member avec visibility=private
    When GET /api/v1/accounts/{id}/profile sans auth
    Then la réponse est 404

  Scenario: Association renseigne RNA invalide
    Given je suis connecté en tant qu'association
    When je PATCH /api/v1/me/profile avec rna="X123"
    Then la réponse est 422 avec "rna must start with W followed by 9 digits"
```

## API

- `POST /api/v1/auth/register/member` — inscription Personne (reporté Phase 1, livré Phase 2).
- `POST /api/v1/auth/register/association` — inscription Association (reporté Phase 1, livré Phase 2).
- `GET /api/v1/me/profile` — profil complet du compte connecté (member ou association selon kind).
- `PATCH /api/v1/me/profile` — mise à jour profil (champs selon kind).
- `GET /api/v1/accounts/{id}/profile` — profil public (visibilité respectée).
- `POST /api/v1/me/avatar` — upload avatar (multipart, ≤ 2 Mo, stockage original, AVIF déféré V2).
- `GET /api/v1/auth/me` — mise à jour pour inclure `kind` dans la réponse.

## Critères d'acceptation

- [ ] Migration 0002 sans perte de données (accounts existants migrés vers members).
- [ ] Validation SIREN (9 chiffres) et RNA (W + 9 chiffres) côté serveur.
- [ ] Avatar : stockage original uniquement, pas de processing AVIF en V1.
- [ ] Visibility respectée : profil private → 404 pour visiteur non autorisé.
- [ ] Champs PII (birth_date) jamais retournés à un visiteur non autorisé.
- [ ] Audit log sur account_created, profile.updated.
- [ ] GET /api/v1/auth/me inclut `kind`.
- [ ] status=active, suspended, deleted opérationnels dans le code.

## Hors scope Phase 2

- Accord parental pour mineurs — V2.
- Avatar AVIF processing — V2.
- Modification d'email avec re-vérification — V2.
- OAuth/OIDC — V2.

## Liens

- `docs/specs/04-data/mld.md` — tables `accounts`, `members`, `associations`, `memberships`.
- `docs/specs/05-quality/security.md` V2.4 (PII protection).
- `docs/adr/0001-architecture-hexagonale-go.md`.
- `docs/adr/0003-sqlite-wal-sqlc.md`.
