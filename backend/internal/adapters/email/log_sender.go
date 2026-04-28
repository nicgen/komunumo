package email

import (
	"context"
	"fmt"
	"log/slog"
)

type LogSender struct {
	appBaseURL string
}

func NewLogSender(appBaseURL string) *LogSender {
	return &LogSender{appBaseURL: appBaseURL}
}

func (s *LogSender) SendVerification(ctx context.Context, to, displayName, rawToken string) error {
	link := fmt.Sprintf("%s/verify-email/confirm?token=%s", s.appBaseURL, rawToken)
	slog.Info("EMAIL: Verification", "to", to, "link", link)
	fmt.Printf("\n--- EMAIL SENT ---\nTo: %s\nSubject: Verify Email\nLink: %s\n------------------\n\n", to, link)
	return nil
}

func (s *LogSender) SendAccountAlreadyExists(ctx context.Context, to, displayName string) error {
	slog.Info("EMAIL: Account Already Exists", "to", to)
	return nil
}

func (s *LogSender) SendPasswordReset(ctx context.Context, to, displayName, rawToken string) error {
	link := fmt.Sprintf("%s/reset-password/confirm?token=%s", s.appBaseURL, rawToken)
	slog.Info("EMAIL: Password Reset", "to", to, "link", link)
	fmt.Printf("\n--- EMAIL SENT ---\nTo: %s\nSubject: Password Reset\nLink: %s\n------------------\n\n", to, link)
	return nil
}

func (s *LogSender) SendPasswordChanged(ctx context.Context, to, displayName string) error {
	slog.Info("EMAIL: Password Changed", "to", to)
	return nil
}
