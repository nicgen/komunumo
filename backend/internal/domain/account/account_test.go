package account_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/domain/account"
)

var (
	now = time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	dob = now.AddDate(-20, 0, 0)
)

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

func TestNew_AgeValidation(t *testing.T) {
	t.Run("exactly 16 is allowed", func(t *testing.T) {
		dob16 := now.AddDate(-16, 0, 0)
		a, err := account.New("id1", "user@example.com", "Léa", "Dupont", dob16, now)
		require.NoError(t, err)
		assert.NotNil(t, a)
	})

	t.Run("15 years 364 days is rejected", func(t *testing.T) {
		dob15 := now.AddDate(-16, 0, 1)
		_, err := account.New("id1", "user@example.com", "Léa", "Dupont", dob15, now)
		require.ErrorIs(t, err, account.ErrAgeBelow16)
	})

	t.Run("30 years is allowed", func(t *testing.T) {
		dob30 := now.AddDate(-30, 0, 0)
		a, err := account.New("id1", "user@example.com", "Léa", "Dupont", dob30, now)
		require.NoError(t, err)
		assert.NotNil(t, a)
	})
}

func TestNew_EmailCanonical(t *testing.T) {
	a, err := account.New("id1", "Léa@Example.COM", "Léa", "Dupont", dob, now)
	require.NoError(t, err)
	assert.Equal(t, "Léa@Example.COM", a.Email)
	assert.Equal(t, "léa@example.com", a.EmailCanonical)
}

func TestNew_InvalidEmail(t *testing.T) {
	_, err := account.New("id1", "notanemail", "Léa", "Dupont", dob, now)
	require.ErrorIs(t, err, account.ErrEmailMalformed)
}

func TestNew_InitialStatus(t *testing.T) {
	a, err := account.New("id1", "user@example.com", "Léa", "Dupont", dob, now)
	require.NoError(t, err)
	assert.Equal(t, account.StatusPendingVerification, a.Status)
	assert.Equal(t, now, a.CreatedAt)
	assert.Equal(t, now, a.UpdatedAt)
}

func TestVerify_Transitions(t *testing.T) {
	later := now.Add(time.Hour)

	t.Run("pending → verified OK", func(t *testing.T) {
		a, _ := account.New("id1", "user@example.com", "Léa", "Dupont", dob, now)
		require.NoError(t, a.Verify(later))
		assert.Equal(t, account.StatusVerified, a.Status)
		assert.Equal(t, later, a.UpdatedAt)
	})

	t.Run("verified → verified is invalid transition", func(t *testing.T) {
		a, _ := account.New("id1", "user@example.com", "Léa", "Dupont", dob, now)
		_ = a.Verify(later)
		err := a.Verify(later)
		require.ErrorIs(t, err, account.ErrInvalidTransition)
	})

	t.Run("disabled → verified is invalid transition", func(t *testing.T) {
		a, _ := account.New("id1", "user@example.com", "Léa", "Dupont", dob, now)
		_ = a.Verify(later)
		_ = a.Disable(later)
		err := a.Verify(later)
		require.ErrorIs(t, err, account.ErrInvalidTransition)
	})
}

func TestDisable_Transitions(t *testing.T) {
	later := now.Add(time.Hour)

	t.Run("verified → disabled OK", func(t *testing.T) {
		a, _ := account.New("id1", "user@example.com", "Léa", "Dupont", dob, now)
		_ = a.Verify(later)
		require.NoError(t, a.Disable(later))
		assert.Equal(t, account.StatusDisabled, a.Status)
	})

	t.Run("pending → disabled OK", func(t *testing.T) {
		a, _ := account.New("id1", "user@example.com", "Léa", "Dupont", dob, now)
		require.NoError(t, a.Disable(later))
		assert.Equal(t, account.StatusDisabled, a.Status)
	})

	t.Run("disabled → disabled is invalid transition", func(t *testing.T) {
		a, _ := account.New("id1", "user@example.com", "Léa", "Dupont", dob, now)
		_ = a.Disable(later)
		err := a.Disable(later)
		require.ErrorIs(t, err, account.ErrInvalidTransition)
	})
}
