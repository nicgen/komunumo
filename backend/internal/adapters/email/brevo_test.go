package email_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/adapters/email"
)

func newTestBrevo(t *testing.T, handler http.HandlerFunc) *email.BrevoSender {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return email.NewBrevoSender(email.BrevoConfig{
		APIKey:      "test-key-noop",
		FromEmail:   "noreply@komunumo.fr",
		FromName:    "Komunumo",
		BaseURL:     srv.URL,
		AppBaseURL:  "https://app.komunumo.fr",
	})
}

func TestBrevo_SendVerification_200(t *testing.T) {
	var captured map[string]any
	srv := newTestBrevo(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "test-key-noop", r.Header.Get("api-key"))
		_ = json.NewDecoder(r.Body).Decode(&captured)
		w.WriteHeader(http.StatusCreated)
	})

	err := srv.SendVerification(context.Background(), "lea@example.com", "Léa", "raw-token-123")
	require.NoError(t, err)

	require.NotNil(t, captured)
	to, ok := captured["to"].([]any)
	require.True(t, ok)
	require.Len(t, to, 1)
}

func TestBrevo_SendVerification_4xx(t *testing.T) {
	srv := newTestBrevo(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":"bad_request","message":"invalid email"}`))
	})

	err := srv.SendVerification(context.Background(), "bad", "Name", "token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "400")
}

func TestBrevo_SendVerification_5xx(t *testing.T) {
	srv := newTestBrevo(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	err := srv.SendVerification(context.Background(), "lea@example.com", "Léa", "token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestBrevo_SendAccountAlreadyExists_200(t *testing.T) {
	var called bool
	srv := newTestBrevo(t, func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	})

	err := srv.SendAccountAlreadyExists(context.Background(), "lea@example.com", "Léa")
	require.NoError(t, err)
	assert.True(t, called)
}

func TestBrevo_SendPasswordReset_200(t *testing.T) {
	srv := newTestBrevo(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	err := srv.SendPasswordReset(context.Background(), "lea@example.com", "Léa", "raw-token")
	require.NoError(t, err)
}

func TestBrevo_SendPasswordChanged_200(t *testing.T) {
	srv := newTestBrevo(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	err := srv.SendPasswordChanged(context.Background(), "lea@example.com", "Léa")
	require.NoError(t, err)
}
