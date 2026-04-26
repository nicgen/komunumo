package audit

import "time"

type EventType string

const (
	EventAccountCreated         EventType = "account.created"
	EventAccountEmailVerified   EventType = "account.email_verified"
	EventAuthLoginSuccess       EventType = "auth.login_success"
	EventAuthLoginFailed        EventType = "auth.login_failed"
	EventAuthPasswordResetReq   EventType = "auth.password_reset_requested"
	EventAuthPasswordChanged    EventType = "auth.password_changed"
	EventAuthLogout             EventType = "auth.logout"
	EventAuthSessionExpired     EventType = "auth.session_expired"
)

type Event struct {
	ID         string
	OccurredAt time.Time
	Type       EventType
	AccountID  *string
	EmailHash  *string
	IP         string
	UserAgent  string
	Metadata   map[string]any
}
