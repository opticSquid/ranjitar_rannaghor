package billing

import (
	"context"
	"time"
)

func GetBillReportService(ctx context.Context, userID int, startDate, endDate time.Time) (BillReport, error) {
	return FetchBillReportFromDB(ctx, userID, startDate, endDate)
}
