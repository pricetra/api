package services

import (
	"context"

	"github.com/pricetra/api/graph/gmodel"
)

// func (s Service) GetGroceryList(ctx context.Context, user gmodel.User, id int64) (gmodel.GroceryList, error) {
// 	// Implementation for fetching a grocery list by ID
// 	return nil, nil
// }

func (s Service) CreateGroceryListItem(
	ctx context.Context,
	user gmodel.User,
	grocery_list_id int64,
	input gmodel.CreateGroceryListItemInput,
) (grocery_list_item gmodel.GroceryListItem, err error) {
	return gmodel.GroceryListItem{}, nil
}
