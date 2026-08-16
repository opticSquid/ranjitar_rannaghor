package orderbook

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	_ "time/tzdata"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
)

func getCreationTime(logDate time.Time) time.Time {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		fmt.Println("Error loading location:", err)
	}
	logDate = logDate.In(loc)
	y, m, d := logDate.Date()
	now := time.Now().In(loc)
	created_at := time.Date(y, m, d, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), loc)
	return created_at.UTC()
}

func constructCreationTime(dateVar time.Time, timeVar time.Time) time.Time {
	y, m, d := dateVar.Date()
	h, M, s := timeVar.Clock()
	return time.Date(y, m, d, h, M, s, 0, timeVar.Location())
}

func CalculateTotalCost(log NewOrderRequest, prices map[string]float64) float64 {
	mealPrice := 0.0
	if log.HasMainMeal {
		mealPrice = prices["standard"]
		if log.IsSpecial {
			mealPrice = prices["special"]
		}
	}
	totalCost := mealPrice + (float64(log.ExtraRiceQty) * prices["rice"]) + (float64(log.ExtraRotiQty) * prices["roti"]) + (float64(log.ExtraChickenQty) * prices["chicken"]) + (float64(log.ExtraFishQty) * prices["fish"]) + (float64(log.ExtraEggQty) * prices["egg"]) + (float64(log.ExtraVegetableQty) * prices["vegetable"])
	return totalCost
}

func createOrderService(ctx context.Context, r NewOrderRequest) (NewOrderResponse, error) {
	var orderTime time.Time
	switch r.MealType {
	case LUNCH:
		orderTime = time.Date(r.EntryDate.Year(), r.EntryDate.Month(), r.EntryDate.Day(), 7, 30, 0, 0, time.UTC)
	case DINNER:
		orderTime = time.Date(r.EntryDate.Year(), r.EntryDate.Month(), r.EntryDate.Day(), 14, 30, 0, 0, time.UTC)
	default:
		return NewOrderResponse{}, fmt.Errorf("Invalid Meal Type %s. %w", r.MealType, ErrInvalidMealType)
	}
	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return NewOrderResponse{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	doesUserExist, err := checkUserExist(tx, ctx, r.UserId)
	if err != nil {
		return NewOrderResponse{}, fmt.Errorf("failed to check user exist: %w", err)
	}
	if !doesUserExist {
		return NewOrderResponse{}, fmt.Errorf("user does not exist. %w", ErrUserDoesNotExist)
	}
	var itemIds = make([]int, 0, len(r.OrderDetails))
	for _, detail := range r.OrderDetails {
		itemIds = append(itemIds, detail.ItemId)
	}
	invalidItemIds, err := checkMenuItemsExist(tx, ctx, itemIds)
	if err != nil {
		return NewOrderResponse{}, fmt.Errorf("failed to check item exist: %w", err)
	}
	if len(invalidItemIds) > 0 {
		slog.Info("invalid item ids found", "invalid_item_ids", invalidItemIds)
		return NewOrderResponse{}, fmt.Errorf("invalid item ids found. %w", ErrMenuItemDoesNotExist)
	}
	order := order{
		userId:      r.UserId,
		orderTs:     orderTime,
		mealType:    r.MealType,
		totalAmount: 0.0,
		status:      PENDING,
	}
	prices, err := fetchPrices(tx, ctx, itemIds, orderTime)
	if err != nil {
		return NewOrderResponse{}, fmt.Errorf("failed to fetch prices: %w", err)
	}
	orderId, err := createInitialOrder(tx, ctx, order)
	if err != nil {
		return NewOrderResponse{}, fmt.Errorf("failed to create initial order: %w", err)
	}
	var orderItems = make([]orderItem, 0, len(r.OrderDetails))
	total := 0.0
	for _, detail := range r.OrderDetails {
		item := orderItem{
			orderId:  orderId,
			itemId:   detail.ItemId,
			priceId:  prices[detail.ItemId].priceId,
			quantity: detail.Quantity,
			subTotal: float64(detail.Quantity) * prices[detail.ItemId].price,
		}
		orderItems = append(orderItems, item)
		total += item.subTotal
	}
	if err := createOrderDetails(tx, ctx, orderItems); err != nil {
		return NewOrderResponse{}, fmt.Errorf("failed to create order details: %w", err)
	}
	txn := walletTransaction{
		userId:  r.UserId,
		orderId: orderId,
		txnType: DELIVERY,
		amount:  total,
		txnTs:   orderTime,
	}
	if err := createTransaction(tx, ctx, txn); err != nil {
		fmt.Println("failed to create transaction:", err)
		return NewOrderResponse{}, err
	}
	order.orderId = orderId
	order.totalAmount = total
	order.status = COMPLETED
	if err = createFinalizedOrder(tx, ctx, order); err != nil {
		fmt.Println("failed to create finalized order:", err)
		return NewOrderResponse{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return NewOrderResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return NewOrderResponse{OrderId: orderId, Status: order.status}, nil
}

func DeleteDailyEntryService(ctx context.Context, logID int) (float64, error) {
	return DeleteDailyEntryFromDB(ctx, logID)
}

func UpdateDailyEntryService(ctx context.Context, logID int, req NewOrderRequest) (float64, error) {
	// Let repository compute new total cost using the original creation timestamp
	return UpdateDailyEntryInDB(ctx, logID, req)
}

func GetDailyEntriesService(ctx context.Context, date time.Time, userID int) ([]DailyLog, error) {
	return FetchDailyEntries(ctx, date, userID)
}
