# Qualité - Sécurité

Référentiel : **OWASP ASVS 4.0 niveau L1** (cible MVP) avec quelques contrôles L2 ciblés.

## Authentification

| Contrôle | Mise en œuvre |
|----------|---------------|
| V2.1.1 - Pas d'envoi du mot de passe en clair en logs | `slog` avec redaction PII, lint vérifie absence de `slog.Info("...password=%s",...)` |
| V2.1.2 - Mot de passe ≥ 12 caractères | Validation client (Zod) + serveur (validator) |
| V2.1.7 - Pas de mots de passe communs | TODO V2 : check offline contre top-100k haveibeenpwned |
| V2.2.1 - Pas de questions secrètes | Non implémenté (anti-pattern) |
| V2.4.1 - Hash bcrypt cost ≥ 12 | `golang.org/x/crypto/bcrypt` cost 12, mesuré ~250ms en CI |
| V2.5.6 - Reset token expirant 30 min | Implémenté |
| V2.5.7 - Token reset à usage unique | Marqué `consumed_at` après usage |

## Sessions

| Contrôle | Mise en œuvre |
|----------|---------------|
| V3.2.1 - Session ID 128 bits | 32 octets random base64 |
| V3.4.1 - Cookie HttpOnly | Oui |
| V3.4.2 - Cookie Secure | Oui |
| V3.4.3 - Cookie SameSite Lax | Oui |
| V3.4.5 - Idle timeout | Pas en MVP, expiration fixe 30j ; rotation à chaque login |
| V3.5.1 - Logout invalide la session côté serveur | Oui (DELETE FROM sessions) |
| V3.5.3 - Sessions actives invalidées sur changement mot de passe | Oui |

## Autorisation

| Contrôle | Mise en œuvre |
|----------|---------------|
| V4.1.1 - Vérification d'autorisation à chaque endpoint | Middleware `RequireAuth` + assertion par use case |
| V4.1.3 - Principe du moindre privilège | Rôles `owner/admin/member` |
| V4.2.1 - IDOR mitigé | UUID v7 imprévisibles + check ownership systématique |

## Validation entrée

| Contrôle | Mise en œuvre |
|----------|---------------|
| V5.1.3 - Allow-list pour les enums | CHECK constraints + validator `oneof=...` |
| V5.1.5 - Refus des HTML dans les champs texte | Sanitizer `bluemonday` (UGC policy) |
| V5.2.5 - Limites de taille | maxLength dans OpenAPI + validator |
| V5.3.4 - Encoding contextuel | React échappe par défaut, pas de `dangerouslySetInnerHTML` |

## Cryptographie

| Contrôle | Mise en œuvre |
|----------|---------------|
| V6.2.1 - Algos approuvés | bcrypt, HMAC-SHA256, AES-GCM si chiffrement |
| V6.2.5 - Pas de MD5/SHA1 | Lint `gosec` |
| V6.4.1 - Secrets non en code | 1Password via `op run` |

## Logging

| Contrôle | Mise en œuvre |
|----------|---------------|
| V7.1.1 - Pas de PII en logs | Helper `slog.With("email_hash", hash(email))` jamais d'email brut |
| V7.2.1 - Logs structurés | `log/slog` JSON handler |
| V7.3.1 - Logs append-only pour audit | Table `audit_log` avec trigger anti-update + HMAC chaîné (F6) |
| V7.4.2 - Pas de logs côté client | `console.log` interdit en prod (lint) |

## Communication

| Contrôle | Mise en œuvre |
|----------|---------------|
| V9.1.1 - TLS 1.2+ uniquement | Traefik conf TLS 1.2 min, 1.3 préféré |
| V9.1.2 - HSTS strict | Header HSTS `max-age=31536000; includeSubDomains; preload` |
| V9.2.1 - Certificats validés (Let's Encrypt + Cloudflare DNS-01) | Existant |

## Configuration

| Contrôle | Mise en œuvre |
|----------|---------------|
| V14.1.1 - Process de build documenté | Dockerfile + GitHub Actions versionnés |
| V14.4.1 - CSP | `Content-Security-Policy: default-src 'self'; img-src 'self' data:; ...` |
| V14.4.2 - X-Frame-Options DENY | Oui |
| V14.4.3 - Referrer-Policy strict-origin-when-cross-origin | Oui |
| V14.4.4 - X-Content-Type-Options nosniff | Oui |
| V14.5.1 - CORS strict | Allow uniquement `app.local.hello-there.net` ; `Access-Control-Allow-Credentials: true` |

## Anti-abus

| Mesure | Implémentation |
|--------|----------------|
| Rate limiting global | Token bucket par IP en mémoire (Go) |
| Rate limiting auth | 5 inscriptions/IP/h, 10 logins/IP/15min |
| Rate limiting WS | 60 msg/min/compte |
| CSRF | Double-submit token sur POST sensibles |
| Brute force protection | Compteur d'échecs login par compte (lock 30min après 10 échecs) |

## Outils de scan automatique en CI

| Outil | Rôle |
|-------|------|
| `gosec` | SAST Go |
| `govulncheck` | CVE dans deps Go |
| `npm audit` (ou `pnpm audit`) | CVE deps JS |
| `Trivy` | Scan images Docker |
| `CodeQL` | SAST GitHub natif |
| `Semgrep` | Patterns custom (optionnel) |

## Pen-test minimum avant soutenance

- OWASP ZAP baseline scan en CI nightly.
- Manual : tester les 10 OWASP Top 10 sur les parcours critiques (1h dédiée S4).
