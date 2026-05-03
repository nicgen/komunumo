package association_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/domain/association"
)

func TestNewAssociation_OK(t *testing.T) {
	a, err := association.New("acc-1", "Les Amis du Code", "75011", time.Now())
	require.NoError(t, err)
	assert.Equal(t, "acc-1", a.AccountID)
	assert.Equal(t, "Les Amis du Code", a.LegalName)
	assert.Equal(t, "75011", a.PostalCode)
	assert.Equal(t, association.VisibilityPublic, a.Visibility)
}

func TestNewAssociation_EmptyLegalName(t *testing.T) {
	_, err := association.New("acc-1", "", "75011", time.Now())
	require.ErrorIs(t, err, association.ErrInvalidLegalName)
}

func TestNewAssociation_EmptyPostalCode(t *testing.T) {
	_, err := association.New("acc-1", "Test", "", time.Now())
	require.ErrorIs(t, err, association.ErrInvalidPostalCode)
}

func TestValidateSIREN_OK(t *testing.T) {
	require.NoError(t, association.ValidateSIREN("123456789"))
}

func TestValidateSIREN_TooShort(t *testing.T) {
	require.ErrorIs(t, association.ValidateSIREN("12345678"), association.ErrInvalidSIREN)
}

func TestValidateSIREN_Letters(t *testing.T) {
	require.ErrorIs(t, association.ValidateSIREN("12345678A"), association.ErrInvalidSIREN)
}

func TestValidateSIREN_Empty(t *testing.T) {
	// Empty SIREN is allowed (optional field)
	require.NoError(t, association.ValidateSIREN(""))
}

func TestValidateRNA_OK(t *testing.T) {
	require.NoError(t, association.ValidateRNA("W123456789"))
}

func TestValidateRNA_MissingW(t *testing.T) {
	require.ErrorIs(t, association.ValidateRNA("A123456789"), association.ErrInvalidRNA)
}

func TestValidateRNA_TooShort(t *testing.T) {
	require.ErrorIs(t, association.ValidateRNA("W12345678"), association.ErrInvalidRNA)
}

func TestValidateRNA_Empty(t *testing.T) {
	// Empty RNA is allowed (optional field)
	require.NoError(t, association.ValidateRNA(""))
}

func TestAssociation_SetAbout_TooLong(t *testing.T) {
	a, _ := association.New("acc-1", "Test", "75011", time.Now())
	long := strings.Repeat("a", 2001)
	err := a.SetAbout(long)
	require.ErrorIs(t, err, association.ErrAboutTooLong)
}

func TestAssociation_SetAbout_OK(t *testing.T) {
	a, _ := association.New("acc-1", "Test", "75011", time.Now())
	require.NoError(t, a.SetAbout("Nous aidons les développeurs."))
	assert.Equal(t, "Nous aidons les développeurs.", a.About)
}
