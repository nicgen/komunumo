# Feature Specification: Authentification utilisateur

**Feature Branch**: `feat/001-user-auth`
**Created**: 2026-04-26
**Status**: Draft
**Input**: User description: "Authentification utilisateur AssoLink/Komunumo. Parcours end-to-end couvrant: inscription par email + mot de passe (validation email obligatoire, RGAA AAA), connexion avec sessions cookies HttpOnly/Secure/SameSite=Strict (bcrypt cost 12+), reset de mot de passe par email Brevo, déconnexion (invalidation session côté serveur), middleware RequireAuth pour les routes protégées."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Inscription d'un compte personnel avec vérification email (Priority: P1)

Un visiteur (Léa, bénévole occasionnelle) souhaite créer un compte personnel pour suivre des associations et postuler à des annonces. Elle saisit son email, son prénom, son nom, sa date de naissance et un mot de passe robuste. Elle accepte les CGU et la politique de confidentialité. Elle reçoit un email de vérification, clique sur le lien, et son compte devient actif.

**Why this priority**: Sans inscription, aucune autre fonctionnalité ne peut être utilisée. Première friction perçue par tout nouvel utilisateur, parcours critique RGAA AAA. La vérification email obligatoire prévient les comptes jetables et garantit un canal de communication fiable pour les notifications associatives.

**Independent Test**: Un visiteur sans compte arrive sur `/register`, complète le formulaire avec des données valides, reçoit l'email de vérification dans une boîte de test, clique sur le lien et se voit confirmer la création de son compte. Vérifiable au navigateur + boîte mail de test (Mailpit en local, journal Brevo en démo).

**Acceptance Scenarios**:

1. **Given** un visiteur sur `/register`, **When** il saisit un email valide, prénom, nom, date de naissance ≥ 16 ans, mot de passe ≥ 12 caractères et accepte les CGU, **Then** un compte est créé en statut "pending_verification", un email de vérification est envoyé, et l'utilisateur est redirigé vers une page "vérifiez votre email".
2. **Given** un compte en statut "pending_verification" avec un token de vérification valide non expiré, **When** l'utilisateur clique sur le lien de l'email, **Then** le compte passe en statut "verified" et l'utilisateur est invité à se connecter.
3. **Given** un visiteur soumet le formulaire avec un email déjà utilisé par un compte existant, **When** la requête est traitée, **Then** la réponse est identique au cas nominal (HTTP 200, page "vérifiez votre email") pour ne pas révéler l'existence du compte, et un email "tentative d'inscription sur compte existant" est envoyé à l'adresse concernée.
4. **Given** un utilisateur de lecteur d'écran (NVDA, Orca), **When** il navigue le formulaire au clavier seul, **Then** chaque champ est annoncé avec son label, les erreurs sont annoncées via `aria-live`, et l'ordre de tabulation suit le sens de lecture.

---

### User Story 2 - Connexion avec session persistante (Priority: P1)

Un utilisateur déjà inscrit et vérifié (Anne, présidente d'association) revient sur l'application le lendemain et souhaite accéder à son tableau de bord. Elle saisit son email et son mot de passe, et accède à son espace privé. Sa session est maintenue pendant 30 jours par défaut, et un cookie `HttpOnly` empêche tout vol via JavaScript.

**Why this priority**: Sans connexion, aucun utilisateur récurrent ne peut revenir à son contenu. Action effectuée plusieurs fois par jour par les utilisateurs actifs. La sécurité de cette étape conditionne la confiance dans toute la plateforme.

**Independent Test**: Un compte vérifié de seed data se connecte via `/login`, vérifie qu'il accède à `/home`, et que `GET /api/v1/auth/me` renvoie ses informations. Le cookie de session est inspectable dans les DevTools et présente bien les attributs `HttpOnly`, `Secure`, `SameSite=Strict`.

**Acceptance Scenarios**:

1. **Given** un compte vérifié avec email et mot de passe corrects, **When** l'utilisateur soumet `/login`, **Then** une session est créée en base, un cookie `__Host-session` est posé avec attributs `HttpOnly`, `Secure`, `SameSite=Strict`, et l'utilisateur est redirigé vers `/home`.
2. **Given** un utilisateur soumet un mauvais mot de passe, **When** la requête est traitée, **Then** la réponse est HTTP 401 avec un message générique "identifiants incorrects" qui ne révèle pas l'existence du compte, et le compteur d'échecs est incrémenté.
3. **Given** un compte en statut "pending_verification", **When** l'utilisateur tente de se connecter avec credentials corrects, **Then** la réponse est HTTP 403 avec un lien pour renvoyer l'email de vérification.
4. **Given** une même IP a échoué 5 fois en 15 minutes, **When** une 6e tentative arrive, **Then** la réponse est HTTP 429 et la résolution n'est rouverte qu'après 30 minutes.

---

### User Story 3 - Réinitialisation de mot de passe oublié (Priority: P2)

Un utilisateur (Maurice, senior) a oublié son mot de passe. Il saisit son email sur `/reset-password`, reçoit un email contenant un lien sécurisé, et choisit un nouveau mot de passe. Toutes ses sessions actives sont invalidées par mesure de sécurité, et il est invité à se reconnecter.

**Why this priority**: Sans reset, un mot de passe oublié devient un compte perdu — friction inacceptable et obstacle à la rétention. Toutefois moins fréquent que login/registration, donc P2.

**Independent Test**: Saisir un email d'un compte de seed sur `/reset-password`, récupérer le lien dans la boîte mail de test, ouvrir le lien, saisir un nouveau mot de passe, puis vérifier que la connexion échoue avec l'ancien mot de passe et réussit avec le nouveau.

**Acceptance Scenarios**:

1. **Given** un visiteur sur `/reset-password`, **When** il soumet un email, **Then** la réponse est HTTP 200 quel que soit l'existence du compte (anti-énumération), et si le compte existe un email avec lien `/reset-password/confirm?token=...` est envoyé. Le token expire 30 minutes après émission.
2. **Given** un token de reset valide et non expiré, **When** l'utilisateur soumet un nouveau mot de passe valide, **Then** le hash du mot de passe est mis à jour, toutes les sessions actives du compte sont invalidées, le token est marqué consommé, et un email "votre mot de passe a été modifié" est envoyé.
3. **Given** un token de reset expiré ou déjà consommé, **When** l'utilisateur tente de l'utiliser, **Then** la réponse est HTTP 410 avec un message clair et un lien pour redemander un reset.

---

### User Story 4 - Déconnexion volontaire (Priority: P2)

Un utilisateur (Maurice, sur un poste partagé en médiathèque) souhaite se déconnecter explicitement à la fin de sa session. Il clique sur "Se déconnecter" dans le menu principal. Sa session est immédiatement invalidée côté serveur et le cookie est supprimé côté navigateur.

**Why this priority**: Indispensable sur appareils partagés, attendu par tout utilisateur web depuis 25 ans. Mais la connexion expire automatiquement après 30 jours, donc moins critique que login.

**Independent Test**: Se connecter, cliquer "Se déconnecter", vérifier que `GET /api/v1/auth/me` renvoie HTTP 401 et que le cookie est absent du navigateur. Re-tenter avec l'ancien `session_id` (interception réseau) doit échouer.

**Acceptance Scenarios**:

1. **Given** un utilisateur connecté, **When** il déclenche la déconnexion via le menu, **Then** la session est supprimée côté serveur, le cookie `__Host-session` est invalidé (`Max-Age=0`), et l'utilisateur est redirigé vers la page d'accueil publique.
2. **Given** un utilisateur déconnecté, **When** il accède à une route protégée par `RequireAuth`, **Then** il est redirigé vers `/login` avec un paramètre `next` qui le ramènera à la page demandée après authentification.

---

### Edge Cases

- **Email indisponible (panne Brevo)** : la création du compte échoue de manière transactionnelle (pas de compte sans email envoyé) ; un message clair invite à réessayer ; un compteur d'erreurs alerte l'opérateur si > 1 % d'échec sur 5 minutes.
- **Token de vérification ou de reset expiré** : message explicite + bouton "renvoyer un email", sans révéler si le compte existe.
- **Concurrence sur réinitialisation** : si deux tokens de reset sont émis pour le même compte, seul le dernier reste valide ; les précédents sont marqués consommés.
- **Adresse email contenant des caractères Unicode (IDN)** : acceptée selon RFC 6531 ; normalisation NFKC avant stockage et avant comparaison.
- **Soumission JavaScript désactivé (mode sobriété, lecteur d'écran ancien)** : tous les formulaires fonctionnent sans JavaScript via soumission HTML standard ; la validation côté client est progressive.
- **Tentative de fixation de session (session fixation)** : le `session_id` est régénéré à chaque login réussi ; les anciens cookies présentés sont ignorés.
- **CSRF sur POST sensibles** : double-submit cookie pattern obligatoire sur tous les POST hors `/login` et `/register` (qui ne nécessitent pas d'auth préalable).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Le système DOIT permettre à un visiteur de créer un compte personnel via email et mot de passe sans étape externe (pas d'OAuth en V1).
- **FR-002**: Le système DOIT envoyer un email de vérification contenant un lien à usage unique avant l'activation du compte ; aucune action authentifiée ne peut être effectuée tant que le compte n'est pas vérifié.
- **FR-003**: Le système DOIT exiger un mot de passe d'au moins 12 caractères et refuser les mots de passe figurant dans une liste noire interne (top 100 mots de passe communs).
- **FR-004**: Le système DOIT hasher les mots de passe avec un algorithme adaptatif moderne (bcrypt cost ≥ 12) ; aucun mot de passe n'est stocké en clair ni journalisé.
- **FR-005**: Le système DOIT créer une session côté serveur lors d'une connexion réussie et matérialiser cette session via un cookie `HttpOnly`, `Secure`, `SameSite=Strict`.
- **FR-006**: Le système DOIT invalider toutes les sessions actives d'un compte lors d'un changement de mot de passe.
- **FR-007**: Les utilisateurs DOIVENT pouvoir initier une procédure de réinitialisation de mot de passe via leur adresse email, qui produit un token à usage unique expirant après 30 minutes.
- **FR-008**: Les utilisateurs DOIVENT pouvoir se déconnecter explicitement, ce qui supprime la session côté serveur et invalide le cookie côté navigateur.
- **FR-009**: Le système DOIT appliquer un rate limiting sur les endpoints d'inscription, de connexion et de demande de reset (5 échecs en 15 minutes par IP → blocage 30 minutes).
- **FR-010**: Le système DOIT répondre de manière indistinguable à une requête sur un compte existant ou inexistant pour les flux d'inscription, de connexion et de demande de reset (anti-énumération).
- **FR-011**: Toute action sensible (création de compte, vérification, connexion, échec de connexion, changement de mot de passe, déconnexion) DOIT être journalisée dans une table d'audit append-only avec horodatage, identifiant de compte et type d'événement, sans contenir de mot de passe ni d'email en clair (email haché SHA-256).
- **FR-012**: Le système DOIT exposer une route protégée `/api/v1/auth/me` qui renvoie l'identité de l'utilisateur courant, ou HTTP 401 si la session est absente, expirée ou révoquée.
- **FR-013**: Tous les formulaires d'authentification DOIVENT être utilisables au clavier seul, annoncer les erreurs via `aria-live`, atteindre un contraste AAA sur tous les textes critiques, et fonctionner sans JavaScript (dégradation gracieuse).
- **FR-014**: Le système DOIT refuser l'inscription d'un utilisateur ayant déclaré une date de naissance < 16 ans (RGPD France, recueil du consentement parental hors scope V1).
- **FR-015**: Le système DOIT envoyer tous les emails transactionnels (vérification, reset, alerte changement de mot de passe) via le fournisseur configuré (Brevo en V1) ; aucun envoi direct depuis le serveur applicatif.

### Key Entities *(include if feature involves data)*

- **Account** : représente un utilisateur authentifiable. Attributs métier : identifiant unique, email, statut (`pending_verification`, `verified`, `disabled`), date de création, date de dernière connexion. Le hash de mot de passe et les attributs internes ne sont pas exposés.
- **Session** : représente une connexion active. Attributs métier : identifiant opaque, compte associé, date de création, date d'expiration (30 jours par défaut), informations de provenance (IP, User-Agent) à des fins d'audit.
- **Email Verification Token** : jeton à usage unique permettant de prouver la possession de l'email. Attributs : compte, hash du jeton, date de création, date d'expiration (24 h), date de consommation.
- **Password Reset Token** : jeton à usage unique permettant de réinitialiser un mot de passe. Attributs : compte, hash du jeton, date de création, date d'expiration (30 min), date de consommation.
- **Audit Log Entry** : enregistrement append-only d'un événement sensible. Attributs : horodatage, type d'événement, compte concerné, hash de l'email, source (IP, User-Agent), corrélation avec la session courante. Aucune donnée en clair.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 95 % des nouveaux utilisateurs complètent l'inscription et la vérification email en moins de 3 minutes lors des tests utilisateurs avec les personas cibles (Anne, Léa, Maurice).
- **SC-002**: 90 % des utilisateurs réussissent à se connecter du premier coup lors des tests utilisateurs (mémorisation correcte de leurs credentials, formulaires non-ambigus).
- **SC-003**: 100 % des parcours d'authentification (inscription, connexion, reset, déconnexion) atteignent le niveau **RGAA AAA** sur les contrôles automatisés (`axe-core`, `lighthouse-ci`) et sur l'audit manuel NVDA + Orca.
- **SC-004**: Aucun mot de passe en clair ni email non haché n'apparaît dans les logs applicatifs, vérifié par recherche pattern dans les sorties `slog` sur 1 000 requêtes simulées.
- **SC-005**: Le rate limiting est efficace : après 5 tentatives de login échouées, le système bloque effectivement les 100 requêtes suivantes pendant la fenêtre de 30 minutes (test automatisé en CI).
- **SC-006**: Le poids cumulé HTML + CSS + JS critique des pages `/register`, `/login`, `/reset-password` reste sous 100 KB chacune (RGESN, vérifié en CI via `lighthouse-ci`).
- **SC-007**: Le temps de réponse médian des endpoints `/login` et `/register` est inférieur à 400 ms (incluant le hashage bcrypt cost 12, qui prend à lui seul ~250 ms).
- **SC-008**: La table d'audit contient une entrée correcte pour chacun des événements suivants après leur déclenchement : `account_created`, `email_verified`, `login_success`, `login_failed`, `password_reset_requested`, `password_changed`, `logout` (vérifié par tests d'intégration).

## Assumptions

- L'utilisateur dispose d'une adresse email valide qu'il peut consulter dans les minutes suivant l'inscription. Les utilisateurs sans email ne sont pas pris en charge en V1.
- Le fournisseur d'email transactionnel (Brevo en V1, cf. ADR-0012) est disponible avec un délai de livraison médian < 30 secondes ; un mode dégradé (file d'attente serveur) est hors scope V1.
- L'utilisateur a JavaScript activé pour bénéficier d'une expérience optimale, mais tous les parcours critiques fonctionnent en HTML pur (mode sobriété, lecteurs d'écran anciens).
- Les tokens de vérification et de reset sont transmis exclusivement par email ; pas de SMS ni de notifications push en V1.
- La 2FA (TOTP, WebAuthn) et les fournisseurs OAuth/OIDC (Google, Apple, France Connect) sont **explicitement hors scope V1** et reportés à V2 (cf. spec hand-rolled `docs/specs/02-features/auth.md`).
- La rotation de session à chaque login et l'invalidation cross-session lors du changement de mot de passe sont jugées suffisantes contre la fixation de session ; aucune protection plus avancée (device binding) n'est implémentée en V1.
- L'âge minimum d'inscription est fixé à 16 ans (RGPD France, recueil du consentement parental hors scope) ; un module dédié pour mineurs est reporté à une évolution post-MVP.
- Le système d'audit log V1 est une table append-only avec contrainte de non-modification (trigger SQLite) ; le chaînage cryptographique HMAC-SHA256 est positionné en évolution V2 (cf. Constitution v1.0.0, principe V).
