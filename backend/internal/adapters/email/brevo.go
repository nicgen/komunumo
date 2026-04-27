package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type BrevoConfig struct {
	APIKey     string
	FromEmail  string
	FromName   string
	BaseURL    string
	AppBaseURL string
}

type BrevoSender struct {
	cfg    BrevoConfig
	client *http.Client
}

func NewBrevoSender(cfg BrevoConfig) *BrevoSender {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.brevo.com/v3"
	}
	return &BrevoSender{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type brevoEmail struct {
	Sender  map[string]string   `json:"sender"`
	To      []map[string]string `json:"to"`
	Subject string              `json:"subject"`
	HTMLContent string          `json:"htmlContent"`
}

func (s *BrevoSender) send(ctx context.Context, to, displayName, subject, htmlContent string) error {
	payload := brevoEmail{
		Sender:  map[string]string{"email": s.cfg.FromEmail, "name": s.cfg.FromName},
		To:      []map[string]string{{"email": to, "name": displayName}},
		Subject: subject,
		HTMLContent: htmlContent,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("brevo: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.BaseURL+"/smtp/email", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("brevo: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", s.cfg.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("brevo: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("brevo: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (s *BrevoSender) SendVerification(ctx context.Context, to, displayName, rawToken string) error {
	link := fmt.Sprintf("%s/verify-email/confirm?token=%s", s.cfg.AppBaseURL, rawToken)
	html := fmt.Sprintf(`<!DOCTYPE html><html lang="fr"><body>
<p>Bonjour %s,</p>
<p>Cliquez sur le lien ci-dessous pour vérifier votre adresse email :</p>
<p><a href="%s">Vérifier mon adresse email</a></p>
<p>Ce lien expire dans 24 heures.</p>
</body></html>`, displayName, link)
	return s.send(ctx, to, displayName, "Vérifiez votre adresse email – Komunumo", html)
}

func (s *BrevoSender) SendAccountAlreadyExists(ctx context.Context, to, displayName string) error {
	html := fmt.Sprintf(`<!DOCTYPE html><html lang="fr"><body>
<p>Bonjour %s,</p>
<p>Nous avons reçu une demande d'inscription avec cette adresse email, qui est déjà associée à un compte Komunumo.</p>
<p>Si vous avez oublié votre mot de passe, vous pouvez le réinitialiser depuis la page de connexion.</p>
</body></html>`, displayName)
	return s.send(ctx, to, displayName, "Tentative d'inscription – Komunumo", html)
}

func (s *BrevoSender) SendPasswordReset(ctx context.Context, to, displayName, rawToken string) error {
	link := fmt.Sprintf("%s/reset-password/confirm?token=%s", s.cfg.AppBaseURL, rawToken)
	html := fmt.Sprintf(`<!DOCTYPE html><html lang="fr"><body>
<p>Bonjour %s,</p>
<p>Cliquez sur le lien ci-dessous pour réinitialiser votre mot de passe :</p>
<p><a href="%s">Réinitialiser mon mot de passe</a></p>
<p>Ce lien expire dans 30 minutes.</p>
</body></html>`, displayName, link)
	return s.send(ctx, to, displayName, "Réinitialisation de mot de passe – Komunumo", html)
}

func (s *BrevoSender) SendPasswordChanged(ctx context.Context, to, displayName string) error {
	html := fmt.Sprintf(`<!DOCTYPE html><html lang="fr"><body>
<p>Bonjour %s,</p>
<p>Votre mot de passe Komunumo a été modifié avec succès.</p>
<p>Si vous n'êtes pas à l'origine de ce changement, contactez-nous immédiatement.</p>
</body></html>`, displayName)
	return s.send(ctx, to, displayName, "Votre mot de passe a été modifié – Komunumo", html)
}
