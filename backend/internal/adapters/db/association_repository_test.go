package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"komunumo/backend/internal/adapters/db"
	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/association"
)

func TestAssociationRepository_CreateAndFind(t *testing.T) {
	conn := openTestDB(t)
	accRepo := db.NewAccountRepository(conn)
	assoRepo := db.NewAssociationRepository(conn)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	// Prerequisite: Account
	acc := &account.Account{
		ID:             "acc1",
		Email:          "asso@example.com",
		EmailCanonical: "asso@example.com",
		Kind:           account.KindAssociation,
		Status:         account.StatusPendingVerification,
	}
	require.NoError(t, accRepo.Create(context.Background(), acc))

	asso, _ := association.New("acc1", "Les Amis du Code", "75011", now)
	asso.SIREN = "123456789"
	asso.About = "Promotion du code open source"
	asso.Visibility = association.VisibilityPublic

	require.NoError(t, assoRepo.Create(context.Background(), asso))

	found, err := assoRepo.FindByAccountID(context.Background(), "acc1")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, asso.LegalName, found.LegalName)
	assert.Equal(t, asso.SIREN, found.SIREN)
	assert.Equal(t, asso.PostalCode, found.PostalCode)
	assert.Equal(t, asso.About, found.About)
	assert.Equal(t, asso.Visibility, found.Visibility)
}

func TestAssociationRepository_Update(t *testing.T) {
	conn := openTestDB(t)
	accRepo := db.NewAccountRepository(conn)
	assoRepo := db.NewAssociationRepository(conn)
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	acc := &account.Account{ID: "acc1", Email: "asso@example.com", EmailCanonical: "asso@example.com", Kind: account.KindAssociation, Status: account.StatusPendingVerification}
	require.NoError(t, accRepo.Create(context.Background(), acc))

	asso, _ := association.New("acc1", "Les Amis du Code", "75011", now)
	require.NoError(t, assoRepo.Create(context.Background(), asso))

	asso.About = "Updated description"
	asso.PostalCode = "75012"
	asso.Visibility = association.VisibilityMembersOnly
	require.NoError(t, assoRepo.Update(context.Background(), asso))

	found, err := assoRepo.FindByAccountID(context.Background(), "acc1")
	require.NoError(t, err)
	assert.Equal(t, "Updated description", found.About)
	assert.Equal(t, "75012", found.PostalCode)
	assert.Equal(t, association.VisibilityMembersOnly, found.Visibility)
}
