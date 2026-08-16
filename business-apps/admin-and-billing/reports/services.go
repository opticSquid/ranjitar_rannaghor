package reports

import (
	"context"
	"fmt"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
)

func GenerateBillReportService(ctx context.Context, r GenerateBillRequest) (GenerateBillResponse, error) {
	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return GenerateBillResponse{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	doesExist, err := checkUserExist(tx, ctx, r.UserId)
	if err != nil {
		return GenerateBillResponse{}, fmt.Errorf("failed to check user exist: %w", err)
	}
	if !doesExist {
		return GenerateBillResponse{}, fmt.Errorf("user does not exist. %W", ErrUserDoesNotExist)
	}

	if err := tx.Commit(ctx); err != nil {
		return GenerateBillResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}
}
