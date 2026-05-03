package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/adapters/db"
	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/ports"
)

func TestMembershipRepository_CreateAndFind(t *testing.T) {
	conn := openTestDB(t)
	accRepo := db.NewAccountRepository(conn)
	membershipRepo := db.NewMembershipRepository(conn)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	// Prerequisite: Accounts
	memberAcc := &account.Account{ID: "member1", Email: "m@test.com", EmailCanonical: "m@test.com", Kind: account.KindMember, Status: account.StatusPendingVerification}
	assoAcc := &account.Account{ID: "asso1", Email: "a@test.com", EmailCanonical: "a@test.com", Kind: account.KindAssociation, Status: account.StatusPendingVerification}
	require.NoError(t, accRepo.Create(context.Background(), memberAcc))
	require.NoError(t, accRepo.Create(context.Background(), assoAcc))

	m := &ports.Membership{
		ID:                   "ms1",
		MemberAccountID:      "member1",
		AssociationAccountID: "asso1",
		Role:                 "owner",
		Status:               "active",
		JoinedAt:             now.Format(time.RFC3339),
	}

	require.NoError(t, membershipRepo.Create(context.Background(), m))

	found, err := membershipRepo.FindByAccountIDs(context.Background(), "member1", "asso1")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, m.ID, found.ID)
	assert.Equal(t, m.Role, found.Role)
	assert.Equal(t, m.Status, found.Status)
}
