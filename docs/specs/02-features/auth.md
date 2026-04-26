# Feature - Authentification

- Status : `Approved`
- Owner : nic
- Last updated : 2026-04-26
- Linked ADRs : ADR-0004, ADR-0009

## Objectif

Permettre à un utilisateur de créer un compte (Member ou Association), se connecter, rester connecté et se déconnecter, en respectant les bonnes pratiques OWASP et RGPD.

## Personas concernés

- P1 (Anne, présidente d'asso) - inscrit son association.
- P2 (Karim, gérant TPE) - inscrit son organisation.
- P3 (Léa, bénévole) - crée un compte personnel.
- P4 (Maurice, senior) - crée un compte personnel, parcours doit être 100% clavier.

## User stories

- En tant que **visiteur**, je veux **m'inscrire en tant que Personne** afin de **pouvoir suivre des assos et postuler à des annonces**.
- En tant que **visiteur**, je veux **inscrire mon Association** afin de **publier des annonces et organiser des événements**.
- En tant qu'**utilisateur inscrit**, je veux **me connecter** afin d'**accéder à mes données**.
- En tant qu'**utilisateur connecté**, je veux **me déconnecter** afin de **protéger mon compte sur appareil partagé**.
- En tant qu'**utilisateur**, je veux **récupérer mon mot de passe** afin de **garder l'accès en cas d'oubli**.

## Critères d'acceptation (Gherkin)

```gherkin
Feature: Inscription

  Scenario: Inscription d'une Personne avec données valides
    Given un visiteur sur /register/personne
    When il saisit email valide, prénom, nom, date de naissance >= 16 ans, mot de passe >= 12 caractères
    And il accepte les CGU et la politique de confidentialité
    And il soumet le formulaire
    Then un compte Member est créé avec status "pending_verification"
    And un email de vérification est envoyé via <EMAIL_PROVIDER>
    And l'utilisateur est redirigé vers /verify-email/sent
    And l'audit log enregistre l'événement "account_created"

  Scenario: Inscription avec email déjà utilisé
    Given un visiteur sur /register/personne
    And un compte existe déjà avec cet email
    When il soumet le formulaire
    Then la réponse est 200 (pas de leak d'existence)
    And aucun nouveau compte n'est créé
    And un email "tentative d'inscription sur compte existant" est envoyé

  Scenario: Inscription d'une Association
    Given un visiteur sur /register/association
    When il saisit nom moral, email association, optionnellement SIREN, code postal, mot de passe valide
    Then un compte Association est créé en status "pending_verification"
    And la Personne créatrice (lui-même, déjà inscrit ou avec compte créé en parallèle) est ajoutée comme Membership avec rôle "owner"

Feature: Connexion

  Scenario: Connexion réussie
    Given un compte vérifié existe avec email "anne@example.org" et mot de passe correct
    When l'utilisateur soumet ces credentials sur /login
    Then une session est créée en base
    And un cookie session_id (HttpOnly, Secure, SameSite=Lax, Domain=.hello-there.net) est posé
    And l'utilisateur est redirigé vers /home
    And l'audit log enregistre "login_success"

  Scenario: Connexion échouée (mauvais mot de passe)
    Given un compte existe
    When l'utilisateur saisit un mauvais mot de passe
    Then la réponse est 401 avec un message générique "identifiants incorrects"
    And aucune information sur l'existence du compte n'est révélée
    And un compteur d'échecs est incrémenté pour cet email
    And après 5 échecs en 15 min, l'IP est rate-limitée 30 min

  Scenario: Connexion sur compte non vérifié
    Given un compte en status "pending_verification"
    When l'utilisateur tente de se connecter
    Then la réponse est 403 avec un lien pour renvoyer l'email de vérification

Feature: Déconnexion

  Scenario: Déconnexion volontaire
    Given un utilisateur connecté
    When il clique sur "Se déconnecter"
    Then la session est supprimée en base
    And le cookie session_id est invalidé (Max-Age=0)
    And il est redirigé vers /

Feature: Réinitialisation de mot de passe

  Scenario: Demande de réinitialisation
    Given un compte existe avec email "anne@example.org"
    When le visiteur soumet cet email sur /reset-password
    Then un token de reset (jeton 32 octets, hash en base, expire 30 min) est créé
    And un email avec lien /reset-password/confirm?token=... est envoyé
    And la réponse est 200 quel que soit l'existence du compte

  Scenario: Confirmation de réinitialisation
    Given un token de reset valide non expiré
    When l'utilisateur soumet un nouveau mot de passe valide
    Then le hash du mot de passe est mis à jour (bcrypt cost 12)
    And toutes les sessions actives de ce compte sont invalidées
    And le token est marqué consommé
    And un email "votre mot de passe a été modifié" est envoyé
```

## Règles métier

- Mot de passe : ≥ 12 caractères, vérification contre haveibeenpwned (top-100k offline) en V2 (TODO V2).
- Email : vérification regex simple côté client + format RFC côté serveur.
- Date de naissance : utilisateur doit avoir ≥ 16 ans (RGPD France).
- Vérification email obligatoire avant toute action publique.
- Sessions : durée 30 jours, rotation à chaque login, invalidation sur changement de mot de passe.
- Rate limit : 5 inscriptions / IP / heure, 10 logins / IP / 15 min.

## Permissions

| Action | Anonyme | Pending verification | Vérifié |
|--------|---------|---------------------|---------|
| S'inscrire | Oui | - | - |
| Se connecter | Oui (échec si pending) | - | Oui |
| Vérifier email | Oui (via lien) | Oui | - |
| Demander reset | Oui | Oui | Oui |
| Modifier email | - | - | Oui (re-vérification requise) |

## Modèle de données impacté

Tables : `accounts`, `members`, `associations`, `memberships`, `sessions`, `email_verifications`, `password_resets`, `audit_log`.

## Endpoints API impactés

- `POST /api/v1/auth/register/member`
- `POST /api/v1/auth/register/association`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/verify-email`
- `POST /api/v1/auth/resend-verification`
- `POST /api/v1/auth/password-reset/request`
- `POST /api/v1/auth/password-reset/confirm`
- `GET /api/v1/auth/me`

## Considérations

### RGAA
- Formulaires avec labels explicites (critère 11.1, 11.2).
- Messages d'erreur associés aux champs (`aria-describedby`, critère 11.10).
- Soumission au clavier (critère 12.x).
- Indicateur de force de mot de passe accessible (annonce ARIA live polite).

### Éco-conception
- Page d'inscription < 100 KB JS.
- Pas de dépendance externe (CAPTCHA Google interdit, alternative : challenge serveur léger ou hCaptcha auto-hébergé V2).

### Sécurité
- bcrypt cost 12 (mesurer en CI : doit prendre ~250ms).
- Cookie `__Host-session` recommandé (impose Secure, Path=/, pas de Domain - alternative : `Domain=.hello-there.net` si besoin cross-subdomain).
- CSRF token sur tous les POST sensibles (double-submit pattern).
- Headers : CSP strict, HSTS, X-Frame-Options DENY.
- Rate limiting via middleware Go (token bucket par IP + par compte).
- Logs sans PII : email haché en `slog` (`email_hash=sha256(email)`).

## Hors scope

- OAuth/OIDC providers (Google, Apple) - V2.
- 2FA TOTP - V2.
- Magic links - V2.
- Email change avec re-vérification - V2 (édition profil basique en V1).

## Open questions

- Choix final `<EMAIL_PROVIDER>` (Brevo recommandé).
- Cookie `Domain=.hello-there.net` : besoin du domaine final pour valider (cf. Q1 ouverte).
