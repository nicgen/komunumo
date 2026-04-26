# Phase 0 - Research: Authentification utilisateur

Décisions techniques consolidées pour la feature `auth`. Chaque décision référence l'ADR ou le principe de Constitution qui la motive. Aucune `NEEDS CLARIFICATION` n'a survécu à la Phase /specify.

## D-001 : Driver SQLite pure Go (`modernc.org/sqlite`)

- **Décision**: utiliser `modernc.org/sqlite` (driver Go natif, sans CGO).
- **Rationale**: pas de dépendance C → builds reproductibles, compilation cross-platform triviale (Linux Scaleway + macOS dev), déploiement en image Docker `scratch` ou `distroless` envisageable. Performance suffisante pour ~10 k req/s en lecture (largement au-dessus de la cible MVP).
- **Alternatives considérées**:
  - `mattn/go-sqlite3` — performances marginalement supérieures mais nécessite CGO ; rejet pour la complexité de build.
  - `crawshaw.io/sqlite` — bas niveau et plus difficile d'usage ; rejet pour l'ergonomie sqlc.
- **Contraintes opérationnelles**: activer le mode WAL au démarrage (`PRAGMA journal_mode=WAL`) et `PRAGMA foreign_keys=ON` pour le respect des FK (cf. ADR-0003).

## D-002 : Génération de code SQL → Go via `sqlc`

- **Décision**: tous les accès base passent par du code généré par `sqlc` à partir de fichiers SQL versionnés dans `backend/internal/adapters/db/queries/`.
- **Rationale**: type-safety à la compilation, lisibilité du SQL natif, zéro magie ORM, parfaitement aligné avec la Constitution principe III (domaine indépendant de la techno).
- **Alternatives considérées**:
  - GORM — ORM dynamique, runtime errors, charge cognitive ; rejet pour le test-first et la perf.
  - `database/sql` brut — verbeux, beaucoup de code répétitif ; sqlc enlève le boilerplate sans cacher la requête.
- **Workflow**: `make generate` lance `sqlc generate` ; le code généré est commité au dépôt (lecture par les LLMs, simplicité onboarding).

## D-003 : Sessions cookies HttpOnly vs JWT

- **Décision**: sessions opaques côté serveur, matérialisées par un cookie `__Host-session` (HttpOnly, Secure, SameSite=Strict, Path=/, sans Domain). Cf. **ADR-0004**.
- **Rationale**: révocation immédiate côté serveur (JWT nécessiterait blacklist + temps de propagation), absence de leak via JS (XSS-safe), conforme OWASP ASVS L2. SameSite=Strict bloque CSRF cross-site sans token additionnel sur cookie nu, mais un double-submit CSRF token est exigé en plus pour les POST sensibles (défense en profondeur).
- **Alternatives considérées**:
  - JWT signé — complexifie la révocation, augmente la taille des requêtes, payload dans le token = leak potentiel.
  - Sessions + JWT signature — over-engineering pour un MVP solo dev.

## D-004 : Algorithme de hashage = bcrypt cost 12

- **Décision**: `golang.org/x/crypto/bcrypt` avec cost 12.
- **Rationale**: bcrypt reste recommandé par OWASP en 2026 ; cost 12 cible ~250 ms par hash sur Scaleway DEV1-S, ce qui est en équilibre entre sécurité (résistance brute force) et UX (login p95 < 400 ms). Test unitaire mesurera le temps réel sur la machine cible et alertera si < 200 ms ou > 500 ms.
- **Alternatives considérées**:
  - argon2id — plus moderne, mais paramétrage plus subtil, et la lib `golang.org/x/crypto/argon2` n'expose pas un wrapper aussi ergonomique. À reconsidérer en V2 si argon2id devient le standard de l'industrie.
  - scrypt — moins répandu, moins documenté dans l'écosystème Go.

## D-005 : Tokens à usage unique (vérification + reset) — stockage haché

- **Décision**: générer un token aléatoire de 32 octets (`crypto/rand`), encoder en base64 URL-safe pour l'envoi par email, stocker uniquement le SHA-256 du token côté serveur.
- **Rationale**: si la base est compromise, l'attaquant ne peut pas réutiliser les tokens (one-way hash). Standard OAuth/PASETO pour les codes d'autorisation. Pas besoin d'une PBKDF (les tokens ont 256 bits d'entropie native).
- **Alternatives considérées**:
  - Stockage en clair — rejeté pour la sécurité.
  - JWT signé — complexifie inutilement, et empêche la révocation côté serveur (conflit avec D-003).

## D-006 : Email transactionnel = Brevo (FR)

- **Décision**: envoyer les emails via l'API Brevo (cf. **ADR-0012**) en synchrone côté backend (pas de queue en V1).
- **Rationale**: Brevo offre 300 emails/jour gratuits, hébergement FR (souveraineté), API simple, déliverabilité bonne. Le free tier ajoute "Sent via Brevo" en footer ; acceptable en MVP.
- **Alternatives considérées**:
  - SES — hébergement US, rejeté par souveraineté.
  - SMTP local + Postfix — opérationnellement coûteux pour solo dev.
  - Mailgun, Sendgrid — non-FR.
- **Mode dégradé**: en cas d'erreur Brevo, l'inscription échoue de manière transactionnelle (pas de compte sans email envoyé) ; le compteur d'erreurs sera surveillé en V2 (queue Brevo + retry asynchrone).

## D-007 : Audit log = table append-only avec trigger SQLite

- **Décision**: table `audit_log` simple INSERT-only, contrainte par triggers SQLite empêchant tout UPDATE/DELETE.
- **Rationale**: respecte Constitution principe V (audit trail V1) sans la complexité du chaînage HMAC (positionné en V2 via futur ADR). Suffisant pour un MVP solo : la non-répudiation n'est pas un blocker pour la soutenance, l'argument "perspective d'industrialisation" est un atout oral.
- **Alternatives considérées**:
  - Chaînage HMAC-SHA256 immédiat — surdimensionné pour V1, complexe à expliquer au jury, alourdit la revue de code.
  - Pas d'audit log — viole Constitution V et FR-011.
- **Implémentation trigger**:
  ```sql
  CREATE TRIGGER audit_log_no_update BEFORE UPDATE ON audit_log
  BEGIN SELECT RAISE(ABORT, 'audit_log is append-only'); END;
  CREATE TRIGGER audit_log_no_delete BEFORE DELETE ON audit_log
  BEGIN SELECT RAISE(ABORT, 'audit_log is append-only'); END;
  ```

## D-008 : Rate limiting = token bucket en mémoire (single-node)

- **Décision**: implémenter le rate limit en mémoire dans le backend Go via une lib légère (`golang.org/x/time/rate` ou impl maison ~50 LoC) — par IP **et** par compte.
- **Rationale**: backend mono-instance en V1, pas besoin de Redis. Simple, performant, observable via les métriques exposées plus tard. Permettra une évolution V2 vers Redis quand on déploiera plusieurs instances.
- **Alternatives considérées**:
  - Redis sur Scaleway — coût + complexité opérationnelle ; rejeté pour le V1.
  - Pas de rate limit — viole FR-009 et OWASP.
- **Limites** : 5 logins échoués / IP / 15 min → blocage 30 min ; 5 inscriptions / IP / heure → blocage 1 h.

## D-009 : CSRF = double-submit cookie sur tous les POST authentifiés

- **Décision**: génération côté serveur d'un CSRF token, posé en cookie `__Host-csrf` (HttpOnly=false pour lecture JS) et exigé en header `X-CSRF-Token` pour tous les POST sauf `/login` et `/register` (qui ne sont pas authentifiés au moment de l'appel).
- **Rationale**: SameSite=Strict couvre la majorité des CSRF cross-site, mais le double-submit ajoute une défense en profondeur recommandée OWASP. Ne nécessite pas de stockage côté serveur (stateless).
- **Alternatives considérées**:
  - Synchronizer token (stocké en session) — cher en lecture/écriture par requête.
  - Pas de CSRF — non recommandé OWASP même avec SameSite=Strict.

## D-010 : Frontend = formulaires avec dégradation gracieuse

- **Décision**: tous les formulaires `/register`, `/login`, `/reset-password` fonctionnent **sans JavaScript** via soumission HTML standard (`<form action method=post>`). React Hooks pour la validation côté client est une amélioration progressive (label d'erreur en temps réel), pas un prérequis fonctionnel.
- **Rationale**: aligne sobriété (RGESN) + accessibilité (RGAA AAA pour lecteurs d'écran anciens / mode JS désactivé) + résilience (CDN ou JS bundle inaccessible ne casse pas l'auth). Force aussi un design API REST propre (le backend doit accepter `application/x-www-form-urlencoded` en plus de JSON).
- **Alternatives considérées**:
  - SPA full-JS — viole RGAA AAA (lecteurs d'écran anciens), viole RGESN (poids JS), antipatterns pour des formulaires simples.
- **Conséquence**: les handlers HTTP backend acceptent **deux Content-Types** : `application/x-www-form-urlencoded` (HTML natif) et `application/json` (fetch JS). Mêmes règles, mêmes réponses sémantiques (303 redirect vs 200 JSON).

## D-011 : Identifiants UUID v7

- **Décision**: tous les IDs d'entités (`accounts.id`, `sessions.id`, etc.) sont des **UUID v7** stockés en TEXT (format canonique 36 chars).
- **Rationale**: UUID v7 inclut un timestamp triable lexicographiquement → bon comportement de cache + index B-tree. Format compatible URL et JSON sans encodage. Permet la fédération future (`user@domain` + UUID stable).
- **Alternatives considérées**:
  - INTEGER autoincrement — leak de cardinalité (énumération possible).
  - UUID v4 — pas de localité dans l'index, plus de fragmentation.
  - ULID — pas de support standard Go ; UUID v7 est désormais la RFC 9562.
- **Lib**: `github.com/google/uuid` v1.6+ (supporte v7 depuis fin 2024).

## D-012 : Frontend → Backend = proxy Next.js rewrites

- **Décision**: le frontend Next.js déclare un rewrite `/api/:path*` → `https://api.local.hello-there.net/api/:path*` (dev) / `https://api.komunumo.fr/api/:path*` (cible démo) dans `next.config.ts`. En production Vercel, on utilise `vercel.ts` avec `routes.rewrite`.
- **Rationale**: les cookies `__Host-session` exigent que le frontend et le backend partagent le même hôte (ou que le frontend voie le backend via son propre hôte). Le rewrite Vercel/Next.js règle ce problème sans CORS.
- **Alternatives considérées**:
  - CORS strict + cookie cross-origin — incompatible avec `__Host-` prefix.
  - Cookies `Domain=.komunumo.fr` — plus permissif, mais perd certaines protections de `__Host-`.

## Synthèse des décisions

| ID | Décision | ADR / Principe |
|----|----------|----------------|
| D-001 | Driver SQLite pure Go | ADR-0003 |
| D-002 | sqlc pour requêtes typées | ADR-0003 |
| D-003 | Sessions cookies HttpOnly opaques | ADR-0004 |
| D-004 | bcrypt cost 12 | Constitution V |
| D-005 | Tokens à usage unique hachés SHA-256 | Constitution V |
| D-006 | Email Brevo synchrone | ADR-0012 |
| D-007 | Audit log append-only via triggers SQLite | Constitution V |
| D-008 | Rate limit token bucket en mémoire | FR-009 + OWASP |
| D-009 | CSRF double-submit cookie | Constitution V |
| D-010 | Formulaires sans JS obligatoire | Constitution II |
| D-011 | IDs en UUID v7 | Federation-ready |
| D-012 | Frontend rewrites au lieu de CORS | ADR-0009 |

Toutes les décisions sont compatibles avec la Constitution v1.0.0 et les ADRs 0001–0014. Aucune `NEEDS CLARIFICATION` ne reste ouverte.
