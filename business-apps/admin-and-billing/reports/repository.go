package reports

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func checkUserExist(tx pgx.Tx, ctx context.Context, userId int) (bool, error) {
	var doesExist bool
	err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM PUBLIC.USERS WHERE USER_ID = $1)", userId).Scan(&doesExist)
	if err != nil {
		return false, err
	}
	return doesExist, nil
}
