# Phase 1 - Ports (interfaces côté `internal/ports`)

Contrats Go que la couche `application/auth/*` consomme. Toutes les implémentations vivent côté `internal/adapters/*` et sont injectées en composition root (`cmd/server/main.go`). Aucune méthode ne renvoie un type SQL — tous les types métier sont définis dans `internal/domain/*`.

Toutes les interfaces respectent :
- `context.Context` en premier paramètre.
- Erreurs en dernier retour ; les erreurs métier sont des `error` typés du domaine (`domain/account.ErrEmailTaken`, `domain/session.ErrSessionExpired`, etc.) — jamais d'erreur SQL nue.
- Pas d'allocation de transaction côté ports : les use cases passent un `Tx` opaque géré par un port `UnitOfWork` (cf. ci-dessous).

## `ports.AccountRepository`

```go
package ports

import (
    "context"
    "time"

    "komunumo/internal/domain/account"
)

type AccountRepository interface {
    // Create insère un compte status=pending_verification.
    // Retourne ErrEmailTaken si email_canonical déjà présent.
    Create(ctx context.Context, a *account.Account) error

    // FindByEmailCanonical retourne (nil, nil) si non trouvé (pas une erreur).
    FindByEmailCanonical(ctx context.Context, emailCanonical string) (*account.Account, error)

    FindByID(ctx context.Context, id string) (*account.Account, error)

    // UpdateStatus passe pending_verification → verified.
    UpdateStatus(ctx context.Context, id string, status account.Status, at time.Time) error

    // UpdatePasswordHash change le hash et écrit updated_at.
    UpdatePasswordHash(ctx context.Context, id, hash string, at time.Time) error

    // TouchLastLogin met à jour last_login_at = at.
    TouchLastLogin(ctx context.Context, id string, at time.Time) error
}
```

## `ports.SessionRepository`

```go
type SessionRepository interface {
    Create(ctx context.Context, s *session.Session) error

    // FindByID retourne (nil, nil) si non trouvé.
    // Doit ignorer les sessions expirées (expires_at <= now).
    FindByID(ctx context.Context, id string, now time.Time) (*session.Session, error)

    TouchLastSeen(ctx context.Context, id string, at time.Time) error

    Delete(ctx context.Context, id string) error

    // DeleteAllForAccount révoque toutes les sessions d'un compte
    // (utilisé après changement de mot de passe).
    DeleteAllForAccount(ctx context.Context, accountID string) error

    // DeleteExpired supprime les sessions dont expires_at <= now.
    // Appelé par un cron daily (post-V1).
    DeleteExpired(ctx context.Context, now time.Time) (int64, error)
}
```

## `ports.TokenRepository`

Mutualise `email_verifications` et `password_resets`. Le `kind` est typé pour éviter les confusions.

```go
type TokenKind string

const (
    TokenKindEmailVerification TokenKind = "email_verification"
    TokenKindPasswordReset     TokenKind = "password_reset"
)

type TokenRepository interface {
    // Create insère un token. tokenHash = SHA-256 du token brut envoyé par email.
    Create(ctx context.Context, t *token.Token) error

    // FindActiveByHash retourne le token s'il est non consommé et non expiré, sinon nil.
    FindActiveByHash(ctx context.Context, kind TokenKind, tokenHash string, now time.Time) (*token.Token, error)

    // Consume marque consumed_at=at de manière atomique (CAS).
    // Retourne ErrTokenAlreadyConsumed si déjà consommé.
    Consume(ctx context.Context, kind TokenKind, id string, at time.Time) error

    // RevokeActiveForAccount marque tous les tokens actifs d'un compte comme consommés
    // (utilisé quand on regénère un email de vérification ou de reset).
    RevokeActiveForAccount(ctx context.Context, kind TokenKind, accountID string, at time.Time) error
}
```

## `ports.AuditRepository`

```go
type AuditRepository interface {
    // Append insère une ligne dans audit_log. Append-only — pas de Update/Delete.
    Append(ctx context.Context, e *audit.Event) error
}
```

## `ports.EmailSender`

```go
type EmailSender interface {
    // SendVerification envoie un email de vérification avec le lien
    // https://app.komunumo.fr/verify-email/confirm?token=<rawToken>.
    SendVerification(ctx context.Context, to string, displayName string, rawToken string) error

    // SendPasswordReset envoie un email de reset avec le lien
    // https://app.komunumo.fr/reset-password/confirm?token=<rawToken>.
    SendPasswordReset(ctx context.Context, to string, displayName string, rawToken string) error

    // SendPasswordChanged notifie le propriétaire après un changement réussi.
    // Aucun token. Pas de lien d'action — uniquement informatif.
    SendPasswordChanged(ctx context.Context, to string, displayName string) error
}
```

## `ports.PasswordHasher`

```go
type PasswordHasher interface {
    // Hash retourne le hash bcrypt cost 12.
    Hash(plaintext string) (string, error)

    // Verify retourne (true, nil) si match.
    // Retourne (false, nil) si mismatch — pas d'erreur (UX vs erreur typée).
    // Retourne (false, err) uniquement si le hash est corrompu/illisible.
    Verify(hash, plaintext string) (bool, error)
}
```

## `ports.Clock`

Permet des tests déterministes ; toutes les use cases dépendent d'un `Clock`, jamais de `time.Now()` directement.

```go
type Clock interface {
    Now() time.Time
}
```

## `ports.TokenGenerator`

Isolé pour permettre une substitution déterministe en test.

```go
type TokenGenerator interface {
    // NewRawToken renvoie 32 octets aléatoires encodés base64 URL-safe (~43 chars).
    NewRawToken() (string, error)

    // HashToken renvoie le SHA-256 hex du raw token (pour stockage).
    HashToken(raw string) string

    // NewID renvoie un UUID v7 sous forme canonique 36 chars.
    NewID() string
}
```

## `ports.RateLimiter`

```go
type RateLimiter interface {
    // Allow retourne (true, 0) si l'action est autorisée,
    // (false, retryAfter) si bloquée.
    // key combine action + identifiant (ip ou account_id) :
    //   "login:ip:1.2.3.4"
    //   "register:ip:1.2.3.4"
    //   "password_reset_request:email:hash"
    Allow(ctx context.Context, key string) (allowed bool, retryAfter time.Duration)
}
```

## `ports.UnitOfWork`

Pour les use cases nécessitant atomicité multi-tables (ex: register = INSERT account + INSERT email_verification + INSERT audit_log).

```go
type UnitOfWork interface {
    // Do exécute fn dans une transaction SQLite.
    // Rollback si fn renvoie une erreur ou panique.
    Do(ctx context.Context, fn func(ctx context.Context) error) error
}
```

L'implémentation SQLite stocke la `Tx` dans le `context.Context` ; les repositories `*Repository` lisent ce `Tx` via une clé de contexte privée. Si absent, ils utilisent la pool DB par défaut (mode auto-commit).

## Composition (vue d'ensemble)

```text
                ┌─────────────────────┐
                │ application/auth/*  │
                │  Register / Login   │
                │  VerifyEmail / etc. │
                └─────────┬───────────┘
                          │ dépend de
              ┌───────────┴───────────────────────────┐
              ▼                                       ▼
       ports.* (interfaces)                  domain.* (entités)
              ▲                                       ▲
              │ implémenté par                        │
   ┌──────────┴───────────┐                           │
   ▼                      ▼                           │
adapters/db/*       adapters/email/*  adapters/password/*
adapters/clock/*    adapters/http/*
```

Cette structure garantit qu'un changement de driver SQLite, de provider email ou d'algo de hash n'impacte ni le domaine ni les use cases — uniquement la composition root. Conforme Constitution principe III.
