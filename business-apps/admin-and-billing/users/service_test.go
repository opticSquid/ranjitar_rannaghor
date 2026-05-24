package users

import (
	"context"
	"testing"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/testdb"
	"github.com/stretchr/testify/require"
)

func TestCreateUserService_DefaultRole(t *testing.T) {
	testdb.ResetData()
	u := &User{Name: "SvcUser", Plan: "standard"}
	err := CreateUserService(context.Background(), u)
	require.NoError(t, err)
	require.Equal(t, "normal", u.Role)
}
