package users

import (
	"context"
	"testing"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/testdb"
	"github.com/stretchr/testify/require"
)

func TestInsertAndFetchUser(t *testing.T) {
	testdb.ResetData()
	u := &User{Name: "RepoUser", Plan: "standard", Role: "normal"}
	err := InsertUserToDB(context.Background(), u)
	require.NoError(t, err)
	require.True(t, u.UserID > 0)

	users, err := FetchUsersFromDB(context.Background())
	require.NoError(t, err)
	require.Len(t, users, 1)
}
