# ADR-0004 - Sessions cookies HttpOnly plutôt que JWT seul

- Statut : Accepté
- Date : 2026-04-26
- Décideur : nic

## Contexte

L'application a besoin d'authentifier les utilisateurs sur des parcours web (Next.js -> API Go) avec une expérience fluide ("rester connecté") et sécurisée. Le brief mentionne sessions, cookies, JWT comme outils possibles. La cible (associations, données potentiellement sensibles) impose une posture sécurité conservative et conforme OWASP.

## Décision

Utiliser des **sessions côté serveur** persistées en SQLite, identifiées par un **cookie HttpOnly + Secure + SameSite=Lax** sur le browser (`session_id`). Le mot de passe est haché avec **bcrypt cost ≥ 12**. Les JWT seront introduits **plus tard, exclusivement** pour l'API mobile/externe (V2), sans remplacer les sessions.

Cookies fixés avec :
- `HttpOnly` (pas d'accès JavaScript, anti-XSS).
- `Secure` (HTTPS uniquement).
- `SameSite=Lax` (anti-CSRF par défaut, compatible navigation classique).
- Expiration : 30 jours, rotation à chaque connexion.
- Domain : `.<domaine>` pour cross-subdomain front <-> api.

CSRF token additionnel (double-submit cookie) sur mutations.

## Alternatives écartées

- **JWT seul stocké en localStorage** : vulnérable XSS, impossibilité de révoquer côté serveur. **Anti-pattern OWASP.**
- **JWT en cookie HttpOnly** : la révocation reste compliquée (besoin d'une blacklist côté serveur, ce qui annule l'avantage stateless du JWT). Autant rester sessions.
- **OAuth/OIDC externe (Google, Apple)** : utile en V2, hors scope MVP. Crée une dépendance tiers.
- **Magic links email seulement** : UX inhabituelle, pas adapté au public asso.

## Conséquences

- (+) Révocation immédiate côté serveur (suppression de la session).
- (+) Conforme OWASP Session Management Cheat Sheet.
- (+) Pas de fuite de tokens côté client (HttpOnly).
- (+) Compatible RGPD (cookie strictement nécessaire, pas de bandeau requis).
- (-) Léger surcoût : une lecture SQLite par requête authentifiée. Mitigation : cache mémoire LRU sur les sessions actives (TTL 60s).
- (-) Cookie cross-subdomain à configurer correctement (domain `.<domaine>`). Mitigation : ADR-0009 documente le setup.

## Références

- OWASP Session Management Cheat Sheet - https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
- OWASP ASVS V3 (Session Management) - https://owasp.org/www-project-application-security-verification-standard/
