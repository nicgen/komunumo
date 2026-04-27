package ports

import "context"

type EmailSender interface {
	SendVerification(ctx context.Context, to, displayName, rawToken string) error
	SendAccountAlreadyExists(ctx context.Context, to, displayName string) error
	SendPasswordReset(ctx context.Context, to, displayName, rawToken string) error
	SendPasswordChanged(ctx context.Context, to, displayName string) error
}
