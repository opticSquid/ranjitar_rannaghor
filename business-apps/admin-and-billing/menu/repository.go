package menu

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (c MenuItemCategory) MarshalText() ([]byte, error) {
	return []byte(c), nil
}

func (c *MenuItemCategory) UnmarshalText(text []byte) error {
	val := MenuItemCategory(text)
	switch val {
	case COMBO_THALI, A_LA_CARTE:
		*c = val
		return nil
	default:
		return fmt.Errorf("invalid MenuItemCategory value: %s", string(text))
	}
}

func checkMenuItemExistanceByName(tx pgx.Tx, ctx context.Context, name string) (bool, error) {
	var doesExist bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1
		FROM MENU_ITEMS
		WHERE ITEM_NAME = $1`, name).Scan(&doesExist)
	if err != nil {
		return false, err
	}
	return doesExist, nil
}

func checkMenuItemExistanceById(tx pgx.Tx, ctx context.Context, id int) (bool, error) {
	var doesExist bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1
		FROM MENU_ITEMS
		WHERE ITEM_ID = $1`, id).Scan(&doesExist)
	if err != nil {
		return false, err
	}
	return doesExist, nil
}

func insertMenuItem(tx pgx.Tx, ctx context.Context, m menuItem) (int, error) {
	var itemId int
	err := tx.QueryRow(ctx, `
			INSERT INTO MENU_ITEMS (ITEM_NAME, CATEGORY, IS_ACTIVE)
			VALUES ($1, $2, $3)
			RETURNING ITEM_ID
		`, m.itemName, m.category, m.isActive).Scan(itemId)
	if err != nil {
		return -1, err
	}
	return itemId, nil
}

func updateMenuItemDetails(tx pgx.Tx, ctx context.Context, m MenuItemUpdateRequest) error {
	res, err := tx.Exec(ctx, `
		UPDATE MENU_ITEMS SET ITEM_NAME = $1, CATEGORY = $2, IS_ACTIVE = $3 WHERE ITEM_ID = $4`, m.Name, m.Category, m.IsActive, m.ItemId)
	if err != nil {
		return err
	}
	if res.RowsAffected() > 1 {
		return errors.New("more than 1 row affected by update, something is wrong")
	}
	return nil
}

func insertMenuItemPrice(tx pgx.Tx, ctx context.Context, p menuPriceSchedule) (int, error) {
	var priceId int
	err := tx.QueryRow(ctx, `
		INSERT INTO MENU_PRICE_SCHEDULE (ITEM_ID, PRICE, EFFECTIVE_FROM)
		VALUES ($1, $2, $3)
		RETURNING PRICE_ID
		`, p.itemId, p.price, p.effectiveFrom).Scan(priceId)
	if err != nil {
		return -1, err
	}
	return priceId, nil
}

func fetchAllMenuItems(tx pgx.Tx, ctx context.Context) ([]MenuItemResponse, error) {
	rows, err := tx.Query(ctx, `SELECT DISTINCT
		ON (I.ITEM_ID) I.ITEM_ID,
		I.ITEM_NAME,
		I.CATEGORY,
		I.IS_ACTIVE,
		P.PRICE,
		P.EFFECTIVE_FROM
	FROM
		PUBLIC.MENU_ITEMS I
		JOIN PUBLIC.MENU_PRICE_SCHEDULE P ON I.ITEM_ID = P.ITEM_ID
	WHERE
		P.EFFECTIVE_FROM <= NOW()
	ORDER BY
		I.ITEM_ID,
		P.EFFECTIVE_FROM DESC;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []MenuItemResponse
	for rows.Next() {
		var item MenuItemResponse
		err := rows.Scan(&item.ItemId, &item.ItemName, &item.Category, &item.IsActive, &item.LatestPrice, &item.EffectiveFrom)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func fetchSingleMenuItem(tx pgx.Tx, ctx context.Context, id int) (MenuItemResponse, error) {
	menuItem := MenuItemResponse{
		ItemId: id,
	}
	err := tx.QueryRow(ctx, `SELECT
		MI.ITEM_ID,
		MI.ITEM_NAME,
		MI.CATEGORY,
		MI.IS_ACTIVE,
		MPS.PRICE,
		MPS.EFFECTIVE_FROM
	FROM
		PUBLIC.MENU_ITEMS MI
		CROSS JOIN LATERAL (
			SELECT
				PRICE,
				EFFECTIVE_FROM
			FROM
				PUBLIC.MENU_PRICE_SCHEDULE MPS
			WHERE
				MPS.ITEM_ID = MI.ITEM_ID
			ORDER BY
				MPS.EFFECTIVE_FROM DESC
			LIMIT
				1
		) MPS
	WHERE
		MI.ITEM_ID = $1;`, id).Scan(&menuItem.ItemName, &menuItem.Category, &menuItem.IsActive, &menuItem.LatestPrice, &menuItem.EffectiveFrom)
	if err != nil {
		return menuItem, err
	}
	return menuItem, nil
}

func deleteMenuItem(tx pgx.Tx, ctx context.Context, id int) error {
	rows, err := tx.Exec(ctx, `DELETE FROM PUBLIC.MENU_PRICE_SCHEDULE WHERE ITEM_ID = $1`, id)
	if err != nil {
		return err
	}
	if rows.RowsAffected() == 0 {
		return fmt.Errorf("no prices for the menu item found to delete. %w", ErrMenuItemDoesNotExist)
	}
	rows, err = tx.Exec(ctx, `DELETE FROM PUBLIC.MENU_ITEMS WHERE ITEM_ID = $1`, id)
	if err != nil {
		return err
	}
	if rows.RowsAffected() == 0 {
		return fmt.Errorf("no menu item found to delete. %w", ErrMenuItemDoesNotExist)
	}
	return nil
}

func fetchPriceHistory(tx pgx.Tx, ctx context.Context, id int) ([]PriceHistoryResponse, error) {
	rows, err := tx.Query(ctx, `SELECT PRICE, EFFECTIVE_FROM FROM PUBLIC.MENU_PRICE_SCHEDULE WHERE ITEM_ID = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PriceHistoryResponse
	for rows.Next() {
		var item PriceHistoryResponse
		err := rows.Scan(&item.Price, &item.EffectiveFrom)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
