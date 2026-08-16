package menu

import (
	"context"
	"fmt"
	"time"

	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
)

func createMenuItemService(ctx context.Context, r NewMenuItemRequest) (MenuItemResponse, error) {
	item := menuItem{
		itemName: r.Name,
		isActive: true,
	}
	itemPriceEntry := menuPriceSchedule{
		price: r.Price,
	}

	// Data validation
	switch r.Category {
	case "combo_thali":
		item.category = COMBO_THALI
	case "a_la_carte":
		item.category = A_LA_CARTE
	default:
		return MenuItemResponse{}, fmt.Errorf("menu item category is invalid. Category: %s.  %w", r.Category, ErrInvalidMenuCategory)
	}
	curTime := time.Now().UTC()
	if r.EffectiveFrom.UTC().Before(curTime) {
		return MenuItemResponse{}, fmt.Errorf("effective_from value can not be in the past of current time. current timestamp (utc): %v, effective_from value (utc): %v; %w", curTime, r.EffectiveFrom.UTC(), ErrEffectiveFromValueOfPast)
	}
	itemPriceEntry.effectiveFrom = r.EffectiveFrom

	// persisting data
	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return MenuItemResponse{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	//check if an item with the same name already exists in menu then fail
	doesExist, err := checkMenuItemExistanceByName(tx, ctx, item.itemName)
	if err != nil {
		return MenuItemResponse{}, fmt.Errorf("failed to check menu item existence: %w", err)
	}
	if doesExist {
		return MenuItemResponse{}, fmt.Errorf("menu item with name %s already exists. %w", item.itemName, ErrMenuItemExists)
	}

	item.itemId, err = insertMenuItem(tx, ctx, item)

	if err != nil || item.itemId == -1 {
		return MenuItemResponse{}, fmt.Errorf("failed to create menu item: %w", err)
	}

	// relating data
	// item.itemId will be populated if above transaction succeeds
	itemPriceEntry.itemId = item.itemId

	itemPriceEntry.priceId, err = insertMenuItemPrice(tx, ctx, itemPriceEntry)
	if err != nil {
		return MenuItemResponse{}, fmt.Errorf("failed to insert menu item price: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return MenuItemResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return MenuItemResponse{
		ItemId:        itemPriceEntry.itemId,
		ItemName:      item.itemName,
		Category:      item.category,
		LatestPrice:   itemPriceEntry.price,
		EffectiveFrom: itemPriceEntry.effectiveFrom,
	}, nil
}

func getMenuItemsService(ctx context.Context) ([]MenuItemResponse, error) {
	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	menuItems, err := fetchAllMenuItems(tx, ctx)
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return menuItems, nil
}

func updateMenuItemService(ctx context.Context, r MenuItemUpdateRequest) (MenuItemResponse, error) {
	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return MenuItemResponse{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	doesExist, err := checkMenuItemExistanceById(tx, ctx, r.ItemId)
	if err != nil {
		return MenuItemResponse{}, fmt.Errorf("failed to check menu item existence: %w", err)
	}
	if !doesExist {
		return MenuItemResponse{}, fmt.Errorf("menu item does not exist %w", ErrMenuItemDoesNotExist)
	}
	err = updateMenuItemDetails(tx, ctx, r)
	if err != nil {
		return MenuItemResponse{}, fmt.Errorf("failed to update menu item details: %w", err)
	}
	menuItem, err := fetchSingleMenuItem(tx, ctx, r.ItemId)
	if err != nil {
		return MenuItemResponse{}, fmt.Errorf("failed to fetch menu item: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return MenuItemResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return menuItem, nil
}

func updateMenuItemPriceService(ctx context.Context, r PriceUpdateRequest) (PriceUpdateResponse, error) {
	curTime := time.Now().UTC()
	if r.EffectiveFrom.UTC().Before(curTime) {
		return PriceUpdateResponse{}, fmt.Errorf("effective_from value can not be in the past of current time. current timestamp (utc): %v, effective_from value (utc): %v; %w", curTime, r.EffectiveFrom.UTC(), ErrEffectiveFromValueOfPast)
	}

	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return PriceUpdateResponse{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	doesExist, err := checkMenuItemExistanceById(tx, ctx, r.ItemId)
	if err != nil {
		return PriceUpdateResponse{}, fmt.Errorf("failed to check menu item existence: %w", err)
	}
	if !doesExist {
		return PriceUpdateResponse{}, fmt.Errorf("menu item does not exist %w", ErrMenuItemDoesNotExist)
	}
	menuPriceHistory := menuPriceSchedule{
		itemId:        r.ItemId,
		price:         r.NewPrice,
		effectiveFrom: r.EffectiveFrom,
	}
	menuPriceHistory.priceId, err = insertMenuItemPrice(tx, ctx, menuPriceHistory)
	if err != nil {
		return PriceUpdateResponse{}, fmt.Errorf("failed to insert menu item price: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return PriceUpdateResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return PriceUpdateResponse{
		Price:         r.NewPrice,
		EffectiveFrom: r.EffectiveFrom,
	}, nil
}

func deleteMenuItemService(ctx context.Context, id int) error {
	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	err = deleteMenuItem(tx, ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete menu item: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func getPriceHistoryService(ctx context.Context, itemID int) ([]PriceHistoryResponse, error) {
	dbPool := database.GetDbConn()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	prices, err := fetchPriceHistory(tx, ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch price history: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return prices, nil
}
