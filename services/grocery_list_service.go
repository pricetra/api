package services

import (
	"context"
	"fmt"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
)

func (s Service) GetDefaultGroceryList(ctx context.Context, user gmodel.User) (grocery_list gmodel.GroceryList, err error) {
	qb := table.GroceryList.
		SELECT(table.GroceryList.AllColumns).
		FROM(table.GroceryList).
		WHERE(
			table.GroceryList.Default.EQ(postgres.Bool(true)).
				AND(table.GroceryList.UserID.EQ(postgres.Int(user.ID))),
		)
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &grocery_list); err != nil {
		return gmodel.GroceryList{}, err
	}
	return grocery_list, nil
}

func (s Service) GetGroceryList(ctx context.Context, user gmodel.User, id int64) (grocery_list gmodel.GroceryList, err error) {
	qb := table.GroceryList.
		SELECT(table.GroceryList.AllColumns).
		FROM(table.GroceryList).
		WHERE(
			table.GroceryList.ID.EQ(postgres.Int(id)).
				AND(table.GroceryList.UserID.EQ(postgres.Int(user.ID))),
		)
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &grocery_list); err != nil {
		return gmodel.GroceryList{}, err
	}
	return grocery_list, nil
}

func (s Service) GetGroceryListWithItems(ctx context.Context, user gmodel.User, id int64) (grocery_list gmodel.GroceryList, err error) {
	qb := table.GroceryList.
		SELECT(
			table.GroceryList.AllColumns,
			table.GroceryListItem.AllColumns,
			table.Product.AllColumns,
			table.Category.AllColumns,
		).
		FROM(
			table.GroceryList.
				LEFT_JOIN(table.GroceryListItem, table.GroceryListItem.GroceryListID.EQ(table.GroceryList.ID)).
				LEFT_JOIN(table.Product, table.Product.ID.EQ(table.GroceryListItem.ProductID)).
				LEFT_JOIN(table.Category, table.Category.ID.EQ(table.Product.CategoryID)),
		).
		WHERE(
			table.GroceryList.ID.EQ(postgres.Int(id)).
				AND(table.GroceryList.UserID.EQ(postgres.Int(user.ID))),
		).
		ORDER_BY(table.GroceryListItem.CreatedAt.ASC())
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &grocery_list); err != nil {
		return gmodel.GroceryList{}, err
	}
	return grocery_list, nil
}

func (s Service) GetGroceryLists(ctx context.Context, user gmodel.User) (grocery_lists []gmodel.GroceryList, err error) {
	qb := table.GroceryList.
		SELECT(table.GroceryList.AllColumns).
		FROM(table.GroceryList).
		WHERE(table.GroceryList.UserID.EQ(postgres.Int(user.ID))).
		ORDER_BY(table.GroceryList.CreatedAt.DESC())
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &grocery_lists); err != nil {
		return nil, err
	}
	return grocery_lists, nil
}

func (s Service) CreateGroceryListItem(
	ctx context.Context,
	user gmodel.User,
	grocery_list_id int64,
	input gmodel.CreateGroceryListItemInput,
) (grocery_list_item gmodel.GroceryListItem, err error) {
	grocery_list, err := s.GetGroceryList(ctx, user, grocery_list_id)
	if err != nil {
		return gmodel.GroceryListItem{}, fmt.Errorf("grocery list not found")
	}

	if input.Quantity == nil || *input.Quantity < 1 {
		quantity := 1
		input.Quantity = &quantity
	}
	if input.Unit == nil {
		unit := "item"
		input.Unit = &unit
	}
	qb := table.GroceryListItem.INSERT(
			table.GroceryListItem.GroceryListID,
			table.GroceryListItem.UserID,
			table.GroceryListItem.ProductID,
			table.GroceryListItem.Quantity,
			table.GroceryListItem.Unit,
			table.GroceryListItem.Category,
			table.GroceryListItem.Weight,
			table.GroceryListItem.Completed,
		).MODEL(model.GroceryListItem{
			GroceryListID: grocery_list.ID,
			UserID: user.ID,
			ProductID: input.ProductID,
			Quantity: int32(*input.Quantity),
			Unit: input.Unit,
			Category: input.Category,
			Weight: input.Weight,
			Completed: false,
		}).RETURNING(table.GroceryListItem.AllColumns)
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &grocery_list_item); err != nil {
		return gmodel.GroceryListItem{}, err
	}
	return grocery_list_item, nil
}

func (s Service) GetGroceryListItems(ctx context.Context, user gmodel.User, grocery_list_id int64) (grocery_list_items []gmodel.GroceryListItem, err error) {
	if _, err := s.GetGroceryList(ctx, user, grocery_list_id); err != nil {
		return nil, fmt.Errorf("grocery list not found")
	}

	qb := table.GroceryListItem.
		SELECT(
			table.GroceryListItem.AllColumns,
			table.Product.AllColumns,
			table.Category.AllColumns,
		).
		FROM(
			table.GroceryListItem.
				LEFT_JOIN(table.Product, table.Product.ID.EQ(table.GroceryListItem.ProductID)).
				LEFT_JOIN(table.Category, table.Category.ID.EQ(table.Product.CategoryID)),
		).
		WHERE(
			table.GroceryListItem.UserID.EQ(postgres.Int(user.ID)).
				AND(table.GroceryListItem.GroceryListID.EQ(postgres.Int(grocery_list_id))),
		).
		ORDER_BY(table.GroceryListItem.ID.DESC())
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &grocery_list_items); err != nil {
		return nil, err
	}
	return grocery_list_items, nil
}

func (s Service) GetGroceryListItem(ctx context.Context, user gmodel.User, grocery_list_item_id int64) (grocery_list_item gmodel.GroceryListItem, err error) {
	qb := table.GroceryListItem.
		SELECT(
			table.GroceryListItem.AllColumns,
			table.Product.AllColumns,
			table.Category.AllColumns,
		).
		FROM(
			table.GroceryListItem.
				LEFT_JOIN(table.Product, table.Product.ID.EQ(table.GroceryListItem.ProductID)).
				LEFT_JOIN(table.Category, table.Category.ID.EQ(table.Product.CategoryID)),
		).
		WHERE(
			table.GroceryListItem.UserID.EQ(postgres.Int(user.ID)).
				AND(table.GroceryListItem.ID.EQ(postgres.Int(grocery_list_item_id))),
		).LIMIT(1)
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &grocery_list_item); err != nil {
		return gmodel.GroceryListItem{}, err
	}
	return grocery_list_item, nil
}

func (s Service) UpdateGroceryListItem(
	ctx context.Context,
	user gmodel.User,
	grocery_list_item_id int64,
	input gmodel.CreateGroceryListItemInput,
) (grocery_list_item gmodel.GroceryListItem, err error) {
	grocery_list_item, err = s.GetGroceryListItem(ctx, user, grocery_list_item_id)
	if err != nil {
		return gmodel.GroceryListItem{}, fmt.Errorf("grocery list not found")
	}

	var product *gmodel.Product
	cols := postgres.ColumnList{}
	if input.ProductID != nil {
		p, err := s.FindProductById(ctx, *input.ProductID)
		if err != nil {
			return gmodel.GroceryListItem{}, fmt.Errorf("product not found")
		}

		product = &p
		cols = append(cols, table.GroceryListItem.ProductID)
	}
	if input.Quantity != nil && *input.Quantity > 0 {
		cols = append(cols, table.GroceryListItem.Quantity)
		grocery_list_item.Quantity = *input.Quantity
	}
	if input.Unit != nil {
		cols = append(cols, table.GroceryListItem.Unit)
	}
	if input.Category != nil {
		cols = append(cols, table.GroceryListItem.Category)
	}
	if input.Weight != nil {
		cols = append(cols, table.GroceryListItem.Weight)
	}
	if input.Completed != nil {
		cols = append(cols, table.GroceryListItem.Completed)
		grocery_list_item.Completed = *input.Completed
	}
	if len(cols) > 0 {
		cols = append(cols, table.GroceryListItem.UpdatedAt)
	}
	qb := table.GroceryListItem.
		UPDATE(cols).
		MODEL(model.GroceryListItem{
			ProductID: input.ProductID,
			Quantity: int32(grocery_list_item.Quantity),
			Unit: input.Unit,
			Category: input.Category,
			Weight: input.Weight,
			Completed: grocery_list_item.Completed,
			UpdatedAt: time.Now(),
		}).
		WHERE(
			table.GroceryListItem.ID.EQ(postgres.Int(grocery_list_item_id)).
				AND(table.GroceryListItem.UserID.EQ(postgres.Int(user.ID))),
		).
		RETURNING(table.GroceryListItem.AllColumns)
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &grocery_list_item); err != nil {
		return gmodel.GroceryListItem{}, err
	}
	grocery_list_item.Product = product
	return grocery_list_item, nil
}

func (s Service) DeleteGroceryListItem(
	ctx context.Context,
	user gmodel.User,
	grocery_list_item_id int64,
) (grocery_list_item gmodel.GroceryListItem, err error) {
	grocery_list_item, err = s.GetGroceryListItem(ctx, user, grocery_list_item_id)
	if err != nil {
		return gmodel.GroceryListItem{}, fmt.Errorf("grocery list not found")
	}

	qb := table.GroceryListItem.
		DELETE().
		WHERE(
			table.GroceryListItem.ID.EQ(postgres.Int(grocery_list_item_id)).
				AND(table.GroceryListItem.UserID.EQ(postgres.Int(user.ID))),
		)
	if _, err = qb.ExecContext(ctx, s.DB); err != nil {
		return gmodel.GroceryListItem{}, err
	}
	return grocery_list_item, nil
}

func (s Service) CountGroceryListItems(
	ctx context.Context, 
	user gmodel.User,
	grocery_list_id *int64,
	include_completed *bool,
) int {
	where_clause := table.GroceryList.UserID.EQ(postgres.Int(user.ID))
	if grocery_list_id != nil {
		where_clause = where_clause.AND(
			table.GroceryList.ID.EQ(postgres.Int(*grocery_list_id)),
		)
	}
	if include_completed == nil || *include_completed == false {
		where_clause = where_clause.AND(table.GroceryListItem.Completed.IS_FALSE())
	}
	qb := table.GroceryList.
		SELECT(postgres.COUNT(table.GroceryListItem.ID).AS("count")).
		FROM(
			table.GroceryList.
				INNER_JOIN(
					table.GroceryListItem,
					table.GroceryListItem.GroceryListID.EQ(table.GroceryList.ID),
				),
		).
		WHERE(where_clause)
	var res struct{Count int}
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &res); err != nil {
		return 0
	}
	return res.Count
}
