package account_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/domain/account"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{"valid strong password", "SecurePass123!", nil},
		{"exactly 12 chars", "Abcde12345!X", nil},
		{"too short 11 chars", "Abcde1234!X", account.ErrPasswordTooShort},
		{"no uppercase", "abcdefghij1!", account.ErrPasswordTooWeak},
		{"no lowercase", "ABCDEFGHIJ1!", account.ErrPasswordTooWeak},
		{"no digit", "AbcdefghijkL!", account.ErrPasswordTooWeak},
		{"no special char", "Abcdefghij12", account.ErrPasswordTooWeak},
		{"empty password", "", account.ErrPasswordTooShort},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := account.ValidatePassword(tt.password)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidatePassword_AllClassesRequired(t *testing.T) {
	// Verify all four classes are individually required
	cases := []struct {
		name string
		pass string
	}{
		{"missing uppercase", "abcdefghij1!"},
		{"missing lowercase", "ABCDEFGHIJ1!"},
		{"missing digit", "AbcdefghijkL!"},
		{"missing special", "AbcdefghijkL1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := account.ValidatePassword(c.pass)
			assert.Error(t, err)
		})
	}
}
