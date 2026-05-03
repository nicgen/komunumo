package account_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/domain/account"
)

var now = time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)

func TestCanonicalizeEmail(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"user@example.com", "user@example.com", false},
		{"USER@EXAMPLE.COM", "user@example.com", false},
		{"Léa@Example.COM", "léa@example.com", false},
		// NFKC: ﬁ (U+FB01 LATIN SMALL LIGATURE FI) → fi
		{"\ufb01@example.com", "fi@example.com", false},
		{"notanemail", "", true},
		{"@nodomain.com", "", true},
		{"user@", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := account.CanonicalizeEmail(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, account.ErrEmailMalformed)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestNew_EmailCanonical(t *testing.T) {
	a, err := account.New("id1", "Léa@Example.COM", now)
	require.NoError(t, err)
	assert.Equal(t, "Léa@Example.COM", a.Email)
	assert.Equal(t, "léa@example.com", a.EmailCanonical)
}

func TestNew_InvalidEmail(t *testing.T) {
	_, err := account.New("id1", "notanemail", now)
	require.ErrorIs(t, err, account.ErrEmailMalformed)
}

func TestNew_InitialStatus(t *testing.T) {
	a, err := account.New("id1", "user@example.com", now)
	require.NoError(t, err)
	assert.Equal(t, account.StatusPendingVerification, a.Status)
	assert.Equal(t, account.KindMember, a.Kind)
	assert.Equal(t, now, a.CreatedAt)
	assert.Equal(t, now, a.UpdatedAt)
}

func TestVerify_Transitions(t *testing.T) {
	later := now.Add(time.Hour)

	t.Run("pending → active OK", func(t *testing.T) {
		a, _ := account.New("id1", "user@example.com", now)
		require.NoError(t, a.Verify(later))
		assert.Equal(t, account.StatusActive, a.Status)
		assert.Equal(t, later, a.UpdatedAt)
	})

	t.Run("active → active is invalid transition", func(t *testing.T) {
		a, _ := account.New("id1", "user@example.com", now)
		_ = a.Verify(later)
		err := a.Verify(later)
		require.ErrorIs(t, err, account.ErrInvalidTransition)
	})

	t.Run("suspended → active is invalid transition", func(t *testing.T) {
		a, _ := account.New("id1", "user@example.com", now)
		_ = a.Verify(later)
		_ = a.Disable(later)
		err := a.Verify(later)
		require.ErrorIs(t, err, account.ErrInvalidTransition)
	})
}

func TestDisable_Transitions(t *testing.T) {
	later := now.Add(time.Hour)

	t.Run("active → suspended OK", func(t *testing.T) {
		a, _ := account.New("id1", "user@example.com", now)
		_ = a.Verify(later)
		require.NoError(t, a.Disable(later))
		assert.Equal(t, account.StatusSuspended, a.Status)
	})

	t.Run("pending → suspended OK", func(t *testing.T) {
		a, _ := account.New("id1", "user@example.com", now)
		require.NoError(t, a.Disable(later))
		assert.Equal(t, account.StatusSuspended, a.Status)
	})

	t.Run("suspended → suspended is invalid transition", func(t *testing.T) {
		a, _ := account.New("id1", "user@example.com", now)
		_ = a.Disable(later)
		err := a.Disable(later)
		require.ErrorIs(t, err, account.ErrInvalidTransition)
	})
}
