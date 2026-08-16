package orderbook

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (c MealType) MarshalText() ([]byte, error) {
	return []byte(c), nil
}

func (c *MealType) UnmarshalText(text []byte) error {
	val := MealType(text)
	switch val {
	case LUNCH, DINNER:
		*c = val
		return nil
	default:
		return fmt.Errorf("invalid meal type value: %s", string(text))
	}
}

func (c OrderStatus) MarshalText() ([]byte, error) {
	return []byte(c), nil
}

func (c *OrderStatus) UnmarshalText(text []byte) error {
	val := OrderStatus(text)
	switch val {
	case COMPLETED, PENDING, CANCELLED:
		*c = val
		return nil
	default:
		return fmt.Errorf("invalid order status value: %s", string(text))
	}
}

func (c TxnType) MarshalText() ([]byte, error) {
	return []byte(c), nil
}

func (c *TxnType) UnmarshalText(text []byte) error {
	val := TxnType(text)
	switch val {
	case RECHARGE, REFUND, DELIVERY:
		*c = val
		return nil
	default:
		return fmt.Errorf("invalid transaction type value: %s", string(text))
	}
}

func checkUserExist(tx pgx.Tx, ctx context.Context, userId int) (bool, error) {
	var doesExist bool
	err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM PUBLIC.USERS WHERE USER_ID = $1)", userId).Scan(&doesExist)
	if err != nil {
		return false, err
	}
	return doesExist, nil
}

func checkMenuItemsExist(tx pgx.Tx, ctx context.Context, menuItemIds []int) ([]int, error) {
	rows, err := tx.Query(ctx, `SELECT
		INPUT_ID
	FROM
		UNNEST($1::INT[]) AS INPUT_ID
	WHERE
		NOT EXISTS (
			SELECT
				1
			FROM
				PUBLIC.MENU_ITEMS
			WHERE
				ITEM_ID = INPUT_ID
		)`, menuItemIds)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var invalidItemIds []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan missing item ID: %w", err)
		}
		invalidItemIds = append(invalidItemIds, id)
	}
	// Check for any error encountered during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating validation rows: %w", err)
	}
	return invalidItemIds, nil
}

func fetchPrices(tx pgx.Tx, ctx context.Context, itemIds []int, ts time.Time) (map[int]price, error) {
	rows, err := tx.Query(ctx, `SELECT DISTINCT
		ON (ITEM_ID) ITEM_ID,
		PRICE_ID,
		PRICE
	FROM
		PUBLIC.MENU_PRICE_SCHEDULE
	WHERE
		ITEM_ID = ANY ($1)
		AND EFFECTIVE_FROM <= $2
	ORDER BY
		ITEM_ID,
		EFFECTIVE_FROM DESC;`, itemIds, ts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices %w", err)
	}
	defer rows.Close()

	prices := make(map[int]price, len(itemIds))
	for rows.Next() {
		var itemId int
		var price price
		if err := rows.Scan(&itemId, &price.priceId, &price.price); err != nil {
			return nil, fmt.Errorf("failed to scan price %w", err)
		}
		prices[itemId] = price
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to fetch prices %w", rows.Err())
	}
	if len(prices) != len(itemIds) {
		return nil, fmt.Errorf("expected %d prices, got %d", len(itemIds), len(prices))
	}
	return prices, nil
}

func createInitialOrder(tx pgx.Tx, ctx context.Context, order order) (int, error) {
	var orderId int
	err := tx.QueryRow(ctx, `INSERT INTO
		PUBLIC.ORDERS (
			USER_ID,
			ORDER_TIMESTAMP,
			MEAL_TYPE,
			TOTAL_AMOUNT,
			STATUS
		)
	VALUES
		($1, $2, $3, $4, $5)
	RETURNING
		ORDER_ID`, order.userId, order.orderTs, order.mealType, order.totalAmount, order.status).Scan(&orderId)
	if err != nil {
		return 0, err
	}
	return orderId, nil
}

func createOrderDetails(tx pgx.Tx, ctx context.Context, orderItems []orderItem) error {
	rows := [][]interface{}{}
	for _, item := range orderItems {
		rows = append(rows, []interface{}{item.orderId, item.itemId, item.priceId, item.quantity, item.subTotal})
	}
	// bulk insert using copy from operation
	copyCount, err := tx.CopyFrom(ctx,
		pgx.Identifier{"PUBLIC.ORDER_ITEMS"},
		[]string{"ORDER_ID", "ITEM_ID", "PRICE_ID", "QUANTITY", "SUBTOTAL"},
		pgx.CopyFromRows(rows))
	if err != nil {
		return err
	}
	if copyCount != int64(len(rows)) {
		return fmt.Errorf("expected to copy %d rows, got %d", len(rows), copyCount)
	}
	return nil
}

func createTransaction(tx pgx.Tx, ctx context.Context, txn walletTransaction) error {
	res, err := tx.Exec(ctx, `INSERT INTO
		PUBLIC.WALLET_TRANSACTIONS (
			USER_ID,
			TXN_TYPE,
			AMOUNT,
			TXN_TIMESTAMP,
			ORDER_ID
		)
	VALUES
		($1, $2, $3, $4, $5)
	`, txn.userId, txn.txnType, txn.amount, txn.txnTs, txn.orderId)
	if err != nil {
		return err
	}
	if res.RowsAffected() != 1 {
		return fmt.Errorf("expected to insert 1 row, got %d", res.RowsAffected())
	}
	return nil
}

func createFinalizedOrder(tx pgx.Tx, ctx context.Context, order order) error {
	res, err := tx.Exec(ctx, `UPDATE PUBLIC.ORDERS SET TOTAL_AMOUNT = $1, STATUS = $2 WHERE ORDER_ID = $3`, order.totalAmount, order.status, order.orderId)
	if err != nil {
		return err
	}
	if res.RowsAffected() != 1 {
		return fmt.Errorf("expected to update 1 row, got %d", res.RowsAffected())
	}
	return nil
}

func checkForExistingOrder(tx pgx.Tx, ctx context.Context, orderId int) (bool, error) {
	var doesExist bool
	err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM PUBLIC.ORDERS WHERE ORDER_ID = $1)", orderId).Scan(&doesExist)
	if err != nil {
		return false, err
	}
	return doesExist, nil
}

func fetcOrder(tx pgx.Tx, ctx context.Context, orderId int) (order, error) {
	var or order
	err := tx.QueryRow(ctx, `
		SELECT USER_ID, TOTAL_AMOUNT, ORDER_TIMESTAMP, MEAL_TYPE FROM PUBLIC.ORDERS WHERE ORDER_ID = $1`, orderId).Scan(&or.userId, &or.totalAmount, &or.orderTs, &or.mealType)
	if err != nil {
		return order{}, err
	}
	return or, nil
}

func deleteOrderDetails(tx pgx.Tx, ctx context.Context, orderId int) error {
	res, err := tx.Exec(ctx, `DELETE FROM PUBLIC.ORDER_ITEMS WHERE ORDER_ID = $1`, orderId)
	if err != nil {
		return err
	}
	if rowsAffected := res.RowsAffected(); rowsAffected == 0 {
		return fmt.Errorf("no order details found for order id: %d", orderId)
	}
	return nil
}

func createCancelOrder(tx pgx.Tx, ctx context.Context, order order) error {
	res, err := tx.Exec(ctx, `UPDATE PUBLIC.ORDERS SET STATUS = $1 WHERE ORDER_ID = $2`, order.status, order.orderId)
	if err != nil {
		return err
	}
	if rowsAffected := res.RowsAffected(); rowsAffected == 0 {
		return fmt.Errorf("no order found for order id: %d", order.orderId)
	}
	return nil
}
