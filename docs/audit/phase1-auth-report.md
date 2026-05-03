# Rapport d'Audit — AssoLink Phase 1 (Auth)

## Résumé exécutif
L'implémentation de la Phase 1 est **techniquement solide** sur l'architecture (hexagonale pure) et la sécurité bas niveau (bcrypt, audit log). Cependant, il manque des pans entiers de la spécification fonctionnelle (modèle Association vs Membre) et des règles de sécurité critiques (rate limiting sur l'inscription). L'accessibilité est négligée malgré les exigences AAA de la Constitution.

## 1. Constitution

| Principe | Statut | Justification |
| ---------- | --------- | --------------- |
| **I. Hexagonale** | **CONFORME** | Structure `domain`/`application`/`ports`/`adapters` respectée, aucune fuite de dépendance. |
| **II. Test-first** | **PARTIEL** | Respecté pour US1 (`5be78c1`), mais absent pour US2, Password Reset et GET /me (pas de commit test antérieur dans le log). |
| **III. Couverture** | **PARTIEL** | Domain: >95% (OK). Application: ~77% (Cible 80%). Global: 42% (Cible 70%). `internal/domain/session` n'est pas testé. |
| **IV. Sécurité** | **PARTIEL** | bcrypt 12 OK. Audit log OK (Append-only via triggers). Rate limit: OK sur login, ABSENT sur inscription. Cookie: Lax (Constitution demandait Strict). |
| **V. Conventional Commits** | **CONFORME** | Format `type(scope): message` bien suivi dans l'historique de la branche `feat/001-user-auth`. |

## 2. Endpoints Auth

| Endpoint | Attendu | Statut | Écarts |
| ---------- | --------- | -------- | -------- |
| `POST /api/v1/auth/register` | Compte + session pending + email + audit | **PARTIEL** | Un seul type de compte (Member). Pas de distinction Association/Member. Pas de rate limit au niveau service. |
| `POST /api/v1/auth/login` | Session cookie + audit | **CONFORME** | Cookie `SameSite=Lax` au lieu de `Strict`. |
| `POST /api/v1/auth/logout` | Session delete + cookie invalid | **CONFORME** | |
| `POST /api/v1/auth/verify-email` | Token hashé + expiration + verify | **CONFORME** | |
| `POST /api/v1/auth/resend-verification` | Renvoie email + status 200 | **CONFORME** | |
| `POST /api/v1/auth/password-reset/request` | Token 32 octets + hash + 200 incond. | **CONFORME** | |
| `POST /api/v1/auth/password-reset/confirm` | Bcrypt update + invalid sessions | **CONFORME** | |
| `GET /api/v1/auth/me` | Retourne profil connecté | **CONFORME** | |

**Règles métier :**
- Password >= 12 chars : **CONFORME** (Validation domaine stricte avec complexité).
- Age >= 16 ans : **CONFORME** (Validé dans `account.New`).
- Email doublon (pas de leak) : **CONFORME** (Service renvoie 200 via `SendAccountAlreadyExists`).
- Sessions (30j + rotation) : **CONFORME**.

## 3. Frontend

| Page | URL attendue | Présente ? | Remarques |
| ------ | ------------- | ----------- | ----------- |
| Inscription | `/register` | OUI | `/register/association` et `/register/member` fusionnés (non conforme à la spec). |
| Connexion | `/login` | OUI | |
| Vérif email envoyée | `/verify-email/sent` | OUI | |
| Vérif email confirm | `/verify-email/confirm` | OUI | |
| Mot de passe oublié | `/forgot-password` | OUI | |
| Reset confirmation | `/reset-password` | OUI | Situé techniquement sous `/reset-password/confirm`. |

**Observations :**
- **Shadcn/ui** : Utilisé pour tous les formulaires.
- **Accessibilité** : **NON CONFORME**. Absence de `aria-describedby` reliant les messages d'erreur aux inputs.
- **Validation** : **PARTIEL**. La validation client (Zod) ne vérifie pas la règle de l'âge (>= 16 ans).
- **Config** : **NON CONFORME**. Utilise des URLs relatives en dur (`/api/v1/...`) au lieu de la variable `NEXT_PUBLIC_API_URL` pourtant définie en `.env`.

## 4. Tests

- **Backend** : 115 tests passent. Échec de compilation sur le dossier `scratch` (mains multiples) à ignorer.
- **Frontend** : **TRÈS FAIBLE**. Un seul fichier de tests unitaires (`register.test.tsx`) pour tout le module Auth.
- **CI** : **CONFORME**. Workflow complet incluant `gosec`, `govulncheck` et `commitlint`.

## 5. Écarts et recommandations

1. **BLOQUANT** : Implémenter les tables `associations` et `memberships` et les flux d'inscription distincts.
2. **BLOQUANT** : Activer le `RateLimiter` dans le service `Register` (le port est présent mais l'appel `Allow` manque).
3. **IMPORTANT** : Mettre en conformité l'accessibilité (ARIA) pour atteindre le niveau AAA sur ces parcours critiques.
4. **IMPORTANT** : Augmenter la couverture de tests sur la couche `internal/application` et tester le domaine `session`.
5. **MINEUR** : Basculer le cookie de session en `SameSite=Strict` conformément à la Constitution.

## 6. Verdict Phase 1
[ ] VALIDÉE — prête pour Phase 2
[X] CONDITIONNELLEMENT VALIDÉE — points 1, 2 et 3 à corriger avant Phase 2
[ ] NON VALIDÉE — retravailler avant de continuer
