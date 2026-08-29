package reports

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
)

func GenerateBillReportService(ctx context.Context, r GenerateBillRequest) (GenerateBillResponse, error) {
	if r.StartDate.Before(r.EndDate) {
		slog.Error("start date must be before end date", "start_date", r.StartDate, "end_date", r.EndDate)
		return GenerateBillResponse{}, fmt.Errorf("start date must be before end date.  %w", ErrDateRange)
	}
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
	res := GenerateBillResponse{
		UserId:    r.UserId,
		StartDate: r.StartDate,
		EndDate:   r.EndDate,
	}
	startTs := time.Date(r.StartDate.Year(), r.StartDate.Month(), r.StartDate.Day(), 0, 0, 0, 0, time.UTC)
	endTs := time.Date(r.EndDate.Year(), r.EndDate.Month(), r.EndDate.Day(), 23, 59, 59, 0, time.UTC)
	prevBal, err := fetchBalance(tx, ctx, r.UserId, startTs.AddDate(0, 0, -1))
	if err != nil {
		return GenerateBillResponse{}, fmt.Errorf("failed to fetch previous balance: %w", err)
	}
	res.PrevStartingDayBalance = prevBal
	currBal, err := fetchBalance(tx, ctx, r.UserId, endTs)
	if err != nil {
		return GenerateBillResponse{}, fmt.Errorf("failed to fetch current balance: %w", err)
	}
	res.Balance = currBal

	userDetails, err := getUserDetails(tx, ctx, r.UserId)
	if err != nil {
		return GenerateBillResponse{}, fmt.Errorf("failed to get user details: %w", err)
	}
	res.UserDetails = userDetails
	if err := tx.Commit(ctx); err != nil {
		return GenerateBillResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}
}
