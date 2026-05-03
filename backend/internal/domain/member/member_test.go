package member_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/domain/member"
)

var now = time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)

func TestNewMember_OK(t *testing.T) {
	dob := now.AddDate(-20, 0, 0).Format("2006-01-02")
	m, err := member.New("acc-1", "Léa", "Martin", dob, now)
	require.NoError(t, err)
	assert.Equal(t, "acc-1", m.AccountID)
	assert.Equal(t, "Léa", m.FirstName)
	assert.Equal(t, "Martin", m.LastName)
	assert.Equal(t, dob, m.BirthDate)
	assert.Equal(t, member.VisibilityPublic, m.Visibility)
}

func TestNewMember_TooYoung(t *testing.T) {
	dob17 := now.AddDate(-17, 0, 1).Format("2006-01-02")
	_, err := member.New("acc-1", "Léa", "Martin", dob17, now)
	require.ErrorIs(t, err, member.ErrTooYoung)
}

func TestNewMember_Exactly18(t *testing.T) {
	dob18 := now.AddDate(-18, 0, 0).Format("2006-01-02")
	m, err := member.New("acc-1", "Léa", "Martin", dob18, now)
	require.NoError(t, err)
	assert.NotNil(t, m)
}

func TestNewMember_EmptyFirstName(t *testing.T) {
	dob := now.AddDate(-20, 0, 0).Format("2006-01-02")
	_, err := member.New("acc-1", "", "Martin", dob, now)
	require.ErrorIs(t, err, member.ErrInvalidName)
}

func TestNewMember_EmptyLastName(t *testing.T) {
	dob := now.AddDate(-20, 0, 0).Format("2006-01-02")
	_, err := member.New("acc-1", "Léa", "", dob, now)
	require.ErrorIs(t, err, member.ErrInvalidName)
}

func TestNewMember_BadBirthDate(t *testing.T) {
	_, err := member.New("acc-1", "Léa", "Martin", "not-a-date", now)
	require.ErrorIs(t, err, member.ErrInvalidBirthDate)
}

func TestMember_SetAboutMe_TooLong(t *testing.T) {
	dob := now.AddDate(-20, 0, 0).Format("2006-01-02")
	m, err := member.New("acc-1", "Léa", "Martin", dob, now)
	require.NoError(t, err)

	long := strings.Repeat("a", 501)
	err = m.SetAboutMe(long, now)
	require.ErrorIs(t, err, member.ErrAboutMeTooLong)
}

func TestMember_SetAboutMe_OK(t *testing.T) {
	dob := now.AddDate(-20, 0, 0).Format("2006-01-02")
	m, _ := member.New("acc-1", "Léa", "Martin", dob, now)

	require.NoError(t, m.SetAboutMe("Passionnée de code", now))
	assert.Equal(t, "Passionnée de code", m.AboutMe)
}

func TestMember_SetAboutMe_Empty(t *testing.T) {
	dob := now.AddDate(-20, 0, 0).Format("2006-01-02")
	m, _ := member.New("acc-1", "Léa", "Martin", dob, now)
	require.NoError(t, m.SetAboutMe("", now))
	assert.Equal(t, "", m.AboutMe)
}
