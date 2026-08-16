package orderbook

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	_ "time/tzdata"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
)

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

func deleteOrderService(ctx context.Context, orderId int) error {
	// check order exists then accordingly issue a refund
	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	doesExist, err := checkForExistingOrder(tx, ctx, orderId)
	if err != nil {
		return fmt.Errorf("failed to check for existing order: %w", err)
	}
	if !doesExist {
		return fmt.Errorf("order does not exist. %w", ErrOrderDoesNotExist)
	}
	order, err := fetcOrder(tx, ctx, orderId)
	if err != nil {
		return fmt.Errorf("failed to fetch order: %w", err)
	}
	var refundTxnTs time.Time
	switch order.mealType {
	case LUNCH:
		refundTxnTs = time.Date(order.orderTs.Year(), order.orderTs.Month(), order.orderTs.Day(), 8, 30, 0, 0, time.UTC)
	case DINNER:
		refundTxnTs = time.Date(order.orderTs.Year(), order.orderTs.Month(), order.orderTs.Day(), 15, 30, 0, 0, time.UTC)
	}

	walletTxn := walletTransaction{
		userId:  order.userId,
		orderId: orderId,
		txnType: REFUND,
		amount:  order.totalAmount,
		txnTs:   refundTxnTs,
	}
	err = createTransaction(tx, ctx, walletTxn)
	if err != nil {
		return fmt.Errorf("failed to create refund transaction: %w", err)
	}
	err = deleteOrderDetails(tx, ctx, orderId)
	if err != nil {
		return fmt.Errorf("failed to delete order details: %w", err)
	}
	order.status = CANCELLED
	err = createCancelOrder(tx, ctx, order)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func updateOrderService(ctx context.Context, orderId int, r NewOrderRequest) (NewOrderResponse, error) {
	err := deleteOrderService(ctx, orderId)
	if err != nil {
		return NewOrderResponse{}, fmt.Errorf("failed to delete old order: %w", err)
	}
	orderRes, err := createOrderService(ctx, r)
	if err != nil {
		return NewOrderResponse{}, fmt.Errorf("failed to create new order: %w", err)
	}
	return orderRes, nil
}
