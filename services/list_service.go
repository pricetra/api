package services

import (
	"context"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
)

func (s Service) CreateList(ctx context.Context, user_id int64, name string, list_type gmodel.ListType) (list gmodel.List, err error) {
	if list_type == gmodel.ListTypeFavorites {
		return gmodel.List{}, fmt.Errorf("cannot create duplicate favorites list")
	}
	if list_type == gmodel.ListTypeWatchList {
		return gmodel.List{}, fmt.Errorf("cannot create duplicate watch list")
	}
	var db_list_type model.ListType
	if err := db_list_type.Scan(list_type); err != nil {
		return gmodel.List{}, fmt.Errorf("invalid list type")
	}

	qb := table.List.INSERT(
		table.List.UserID,
		table.List.Name,
		table.List.Type,
	).MODEL(model.List{
		UserID: user_id,
		Name: name,
		Type: db_list_type,
	}).RETURNING(table.List.AllColumns)
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &list); err != nil {
		return gmodel.List{}, err
	}
	return list, nil
}

func (s Service) FindListByIdAndUserId(ctx context.Context, list_id int64, user_id int64) (list gmodel.List, err error) {
	qb := table.List.
		SELECT(table.List.AllColumns).
		FROM(table.List).
		WHERE(
			table.List.ID.EQ(postgres.Int(list_id)).
			AND(table.List.UserID.EQ(postgres.Int(user_id))),
		).LIMIT(1)
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &list); err != nil {
		return gmodel.List{}, err
	}
	return list, nil
}

func (s Service) FindAllListsByUserId(ctx context.Context, user gmodel.User, list_type *gmodel.ListType) (lists []gmodel.List, err error) {
	db := s.DbOrTxQueryable()
	where_clause := table.List.UserID.EQ(postgres.Int(user.ID))
	if list_type != nil {
		where_clause = where_clause.
			AND(table.List.Type.EQ(postgres.NewEnumValue(list_type.String())))
	}
	qb := table.List.
		SELECT(table.List.AllColumns).
		FROM(table.List).
		WHERE(where_clause).
		ORDER_BY(table.List.CreatedAt.ASC())
	if err = qb.QueryContext(ctx, db, &lists); err != nil {
		return nil, err
	}

	// TODO: Ideally this should all be taken care on the db side...
	for i, list := range lists {
		// Products
		product_list_qb := table.ProductList.
			SELECT(
				table.ProductList.AllColumns,
				table.Product.AllColumns,
				table.Category.AllColumns,
				table.Stock.AllColumns,
				table.Price.AllColumns,
			).
			FROM(
				table.ProductList.
					INNER_JOIN(table.Product, table.Product.ID.EQ(table.ProductList.ProductID)).
					INNER_JOIN(table.Category, table.Category.ID.EQ(table.Product.CategoryID)).
					LEFT_JOIN(table.Stock, table.Stock.ID.EQ(table.ProductList.StockID)).
					LEFT_JOIN(table.Price, table.Price.ID.EQ(table.Stock.LatestPriceID)),
			).
			WHERE(table.ProductList.ListID.EQ(postgres.Int(list.ID))).
			ORDER_BY(table.ProductList.CreatedAt.DESC())
		if err := product_list_qb.QueryContext(ctx, db, &lists[i].ProductList); err != nil {
			return nil, err
		}

		// Cleanup zero-value struct for stock. For some reason even when stock_id is null
		// Jet maps it as a zero-valued gmodel.Stock{}
		// TODO: Take care of this on the db side
		for j, pl := range lists[i].ProductList {
			if pl.StockID != nil {
				continue
			}
			lists[i].ProductList[j].Stock = nil
		}

		// Branches
		branch_list_qb := table.BranchList.
			SELECT(
				table.BranchList.AllColumns,
				table.Branch.AllColumns,
				table.Store.AllColumns,
				table.Address.AllColumns,
			).
			FROM(
				table.BranchList.
					INNER_JOIN(table.Branch, table.Branch.ID.EQ(table.BranchList.BranchID)).
					INNER_JOIN(table.Store, table.Store.ID.EQ(table.Branch.StoreID)).
					INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)),
			).
			WHERE(table.BranchList.ListID.EQ(postgres.Int(list.ID))).
			ORDER_BY(table.BranchList.CreatedAt.DESC())
		if err := branch_list_qb.QueryContext(ctx, db, &lists[i].BranchList); err != nil {
			return nil, err
		}
	}
	return lists, nil
}

func (s Service) DeleteList(ctx context.Context, user gmodel.User, list_id int64) (list gmodel.List, err error) {
	list, err = s.FindListByIdAndUserId(ctx, list_id, user.ID)
	if err != nil {
		return gmodel.List{}, fmt.Errorf("invalid list")
	}
	qb := table.List.
		DELETE().
		WHERE(
			table.List.ID.EQ(postgres.Int(list_id)).
			AND(table.List.UserID.EQ(postgres.Int(user.ID))),
		).
		RETURNING(table.List.AllColumns)
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &list); err != nil {
		return gmodel.List{}, err
	}
	return list, nil
}

func (s Service) FindProductListWithProductId(
	ctx context.Context,
	user gmodel.User,
	list_id int64,
	product_id int64,
	stock_id *int64,
) (product_list gmodel.ProductList, err error) {
	where_clause := table.ProductList.ListID.EQ(postgres.Int(list_id)).
			AND(table.ProductList.UserID.EQ(postgres.Int(user.ID))).
			AND(table.ProductList.ProductID.EQ(postgres.Int(product_id)))
	if stock_id != nil {
		where_clause = where_clause.AND(
			table.ProductList.StockID.EQ(postgres.Int(*stock_id)),
		)
	}
	qb := table.ProductList.
		SELECT(table.ProductList.AllColumns).
		FROM(table.ProductList).
		WHERE(where_clause).
		LIMIT(1)
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &product_list); err != nil {
		return gmodel.ProductList{}, err
	}
	return product_list, nil
}

func (s Service) FindProductListsByUserAndProductId(
	ctx context.Context,
	user gmodel.User,
	product_id int64,
) (product_lists []gmodel.ProductList, err error) {
	qb := table.ProductList.
		SELECT(
			table.ProductList.AllColumns,
			table.List.Type.AS("product_list.list_type"),
		).
		FROM(
			table.ProductList.
				INNER_JOIN(table.List, table.List.ID.EQ(table.ProductList.ListID)),
		).
		WHERE(
			table.ProductList.UserID.EQ(postgres.Int(user.ID)).
			AND(table.ProductList.ProductID.EQ(postgres.Int(product_id))),
		).
		ORDER_BY(table.ProductList.CreatedAt.DESC())
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &product_lists); err != nil {
		return nil, err
	}
	return product_lists, nil
}

func (s Service) FindBranchListWithBranchId(
	ctx context.Context,
	user gmodel.User,
	list_id int64,
	branch_id int64,
) (branch_list gmodel.BranchList, err error) {
	qb := table.BranchList.
		SELECT(table.BranchList.AllColumns).
		FROM(table.BranchList).
		WHERE(
			table.BranchList.ListID.EQ(postgres.Int(list_id)).
			AND(table.BranchList.UserID.EQ(postgres.Int(user.ID))).
			AND(table.BranchList.BranchID.EQ(postgres.Int(branch_id))),
		).LIMIT(1)
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &branch_list); err != nil {
		return gmodel.BranchList{}, err
	}
	return branch_list, nil
}

func (s Service) FindProductListById( ctx context.Context, user gmodel.User, product_list_id int64) (product_list gmodel.ProductList, err error) {
	qb := table.ProductList.
		SELECT(table.ProductList.AllColumns).
		FROM(table.ProductList).
		WHERE(
			table.ProductList.ID.EQ(postgres.Int(product_list_id)).
			AND(table.ProductList.UserID.EQ(postgres.Int(user.ID))),
		).LIMIT(1)
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &product_list); err != nil {
		return gmodel.ProductList{}, err
	}
	return product_list, nil
}

func (s Service) FindBranchListById( ctx context.Context, user gmodel.User, branch_list_id int64) (product_list gmodel.BranchList, err error) {
	qb := table.BranchList.
		SELECT(table.BranchList.AllColumns).
		FROM(table.BranchList).
		WHERE(
			table.BranchList.ID.EQ(postgres.Int(branch_list_id)).
			AND(table.BranchList.UserID.EQ(postgres.Int(user.ID))),
		).LIMIT(1)
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &product_list); err != nil {
		return gmodel.BranchList{}, err
	}
	return product_list, nil
}

func (s Service) AddProductToList(
	ctx context.Context,
	user gmodel.User,
	list_id int64,
	product_id int64,
	stock_id *int64,
) (product_list gmodel.ProductList, err error) {
	list, err := s.FindListByIdAndUserId(ctx, list_id, user.ID)
	if err != nil {
		return gmodel.ProductList{}, fmt.Errorf("invalid list")
	}
	if list.Type == gmodel.ListTypeWatchList && stock_id == nil {
		return gmodel.ProductList{}, fmt.Errorf("watch list requires a stock associated with the product")
	}
	product, err := s.FindProductById(ctx, product_id)
	if err != nil {
		return gmodel.ProductList{}, fmt.Errorf("invalid product")
	}

	// if stock_id is defined then check if it's valid
	if stock_id != nil {
		my_stock, err := s.FindStockById(ctx, *stock_id)
		if err != nil {
			return gmodel.ProductList{}, fmt.Errorf("invalid stock")
		}
		if my_stock.ProductID != product.ID {
			return gmodel.ProductList{}, fmt.Errorf("stock does not represent product")
		}
	}

	// return product_list if it already exists in list
	existing_product_list, _ := s.FindProductListWithProductId(ctx, user, list.ID, product_id, stock_id)
	if existing_product_list != (gmodel.ProductList{}) {
		return existing_product_list, nil
	}
	qb := table.ProductList.INSERT(
		table.ProductList.UserID,
		table.ProductList.ListID,
		table.ProductList.ProductID,
		table.ProductList.StockID,
	).MODEL(model.ProductList{
		UserID: user.ID,
		ListID: list.ID,
		ProductID: product.ID,
		StockID: stock_id,
	}).RETURNING(table.ProductList.AllColumns)
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &product_list); err != nil {
		return gmodel.ProductList{}, err
	}
	return product_list, nil
}

func (s Service) RemoveProductFromList(ctx context.Context, user gmodel.User, list_id int64, product_list_id int64) (product_list gmodel.ProductList, err error) {
	list, err := s.FindListByIdAndUserId(ctx, list_id, user.ID)
	if err != nil {
		return gmodel.ProductList{}, fmt.Errorf("invalid list")
	}

	product_list, err = s.FindProductListById(ctx, user, product_list_id)
	if err != nil {
		return gmodel.ProductList{}, fmt.Errorf("could not find product list")
	}
	if product_list.ListID != list.ID {
		return gmodel.ProductList{}, fmt.Errorf("list does not match product list provided")
	}

	qb := table.ProductList.
		DELETE().
		WHERE(table.ProductList.ID.EQ(postgres.Int(product_list.ID))).
		RETURNING(table.ProductList.AllColumns)
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &product_list); err != nil {
		return gmodel.ProductList{}, err
	}
	return product_list, nil
}

func (s Service) RemoveProductFromListWitProductId(ctx context.Context, user gmodel.User, list_id int64, product_id int64, stock_id *int64) (product_list gmodel.ProductList, err error) {
	list, err := s.FindListByIdAndUserId(ctx, list_id, user.ID)
	if err != nil {
		return gmodel.ProductList{}, fmt.Errorf("invalid list")
	}

	product, err := s.FindProductById(ctx, product_id)
	if err != nil {
		return gmodel.ProductList{}, fmt.Errorf("invalid product")
	}

	where_clause := table.ProductList.ListID.EQ(postgres.Int(list.ID)).
		AND(table.ProductList.ProductID.EQ(postgres.Int(product.ID))).
		AND(table.ProductList.UserID.EQ(postgres.Int(user.ID)))
	if stock_id != nil {
		where_clause = where_clause.AND(
			table.ProductList.StockID.EQ(postgres.Int(*stock_id)),
		)
	}
	qb := table.ProductList.
		DELETE().
		WHERE(where_clause).
		RETURNING(table.ProductList.AllColumns)
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &product_list); err != nil {
		return gmodel.ProductList{}, err
	}
	return product_list, nil
}

func (s Service) AddBranchToList(
	ctx context.Context,
	user gmodel.User,
	list_id int64,
	branch_id int64,
) (branch_list gmodel.BranchList, err error) {
	list, err := s.FindListByIdAndUserId(ctx, list_id, user.ID)
	if err != nil {
		return gmodel.BranchList{}, fmt.Errorf("invalid list")
	}

	branch, err := s.FindBranchById(ctx, branch_id)
	if err != nil {
		return gmodel.BranchList{}, fmt.Errorf("could not find branch")
	}

	existing_branch_list, err := s.FindBranchListWithBranchId(ctx, user, list.ID, branch_id)
	if err == nil {
		return existing_branch_list, nil
	}
	qb := table.BranchList.INSERT(
		table.BranchList.UserID,
		table.BranchList.ListID,
		table.BranchList.BranchID,
	).MODEL(model.BranchList{
		UserID: user.ID,
		ListID: list.ID,
		BranchID: branch.ID,
	}).RETURNING(table.BranchList.AllColumns)
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &branch_list); err != nil {
		return gmodel.BranchList{}, err
	}
	return branch_list, nil
}

func (s Service) RemoveBranchFromList(ctx context.Context, user gmodel.User, list_id int64, branch_list_id int64) (branch_list gmodel.BranchList, err error) {
	list, err := s.FindListByIdAndUserId(ctx, list_id, user.ID)
	if err != nil {
		return gmodel.BranchList{}, fmt.Errorf("invalid list")
	}

	branch_list, err = s.FindBranchListById(ctx, user, branch_list_id)
	if err != nil {
		return gmodel.BranchList{}, fmt.Errorf("could not find branch list")
	}
	if branch_list.ListID != list.ID {
		return gmodel.BranchList{}, fmt.Errorf("list does not match branch list provided")
	}

	qb := table.BranchList.
		DELETE().
		WHERE(table.BranchList.ID.EQ(postgres.Int(branch_list.ID))).
		RETURNING(table.BranchList.AllColumns)
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &branch_list); err != nil {
		return gmodel.BranchList{}, err
	}
	return branch_list, nil
}

func (s Service) GetProductListsWithListId(
	ctx context.Context,
	user gmodel.User,
	list_id int64,
) (product_lists []gmodel.ProductList, err error) {
	if _, err = s.FindListByIdAndUserId(ctx, list_id, user.ID); err != nil {
		return nil, fmt.Errorf("invalid list")
	}

	qb := table.ProductList.
		SELECT(
			table.ProductList.AllColumns,
			table.Product.AllColumns,
			table.Category.AllColumns,
			table.Stock.AllColumns,
			table.Price.AllColumns,
			table.Branch.AllColumns,
			table.Address.AllColumns,
			table.Store.AllColumns,
		).
		FROM(
			table.ProductList.
				INNER_JOIN(table.Product, table.Product.ID.EQ(table.ProductList.ProductID)).
				INNER_JOIN(table.Category, table.Category.ID.EQ(table.Product.CategoryID)).
				LEFT_JOIN(table.Stock, table.Stock.ID.EQ(table.ProductList.StockID)).
				LEFT_JOIN(table.Price, table.Price.ID.EQ(table.Stock.LatestPriceID)).
				LEFT_JOIN(table.Branch, table.Branch.ID.EQ(table.Stock.BranchID)).
				LEFT_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
				LEFT_JOIN(table.Store, table.Store.ID.EQ(table.Stock.StoreID)),
		).
		WHERE(
			table.ProductList.ListID.EQ(postgres.Int(list_id)).
			AND(table.ProductList.UserID.EQ(postgres.Int(user.ID))),
		).
		ORDER_BY(table.ProductList.CreatedAt.DESC())
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &product_lists); err != nil {
		return nil, err
	}
	for i := range product_lists {
		if product_lists[i].StockID != nil {
			continue
		}
		product_lists[i].Stock = nil
	}
	return product_lists, nil
}

func (s Service) GetBranchListsWithListId(
	ctx context.Context,
	user gmodel.User,
	list_id int64,
) (branch_lists []gmodel.BranchList, err error) {
	if _, err = s.FindListByIdAndUserId(ctx, list_id, user.ID); err != nil {
		return nil, fmt.Errorf("invalid list")
	}

	qb := table.BranchList.
		SELECT(
			table.BranchList.AllColumns,
			table.Branch.AllColumns,
			table.Store.AllColumns,
			table.Address.AllColumns,
		).
		FROM(
			table.BranchList.
				INNER_JOIN(table.Branch, table.Branch.ID.EQ(table.BranchList.BranchID)).
				INNER_JOIN(table.Store, table.Store.ID.EQ(table.Branch.StoreID)).
				INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)),
		).
		WHERE(
			table.BranchList.ListID.EQ(postgres.Int(list_id)).
			AND(table.BranchList.UserID.EQ(postgres.Int(user.ID))),
		).
		ORDER_BY(table.BranchList.CreatedAt.DESC())
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &branch_lists); err != nil {
		return nil, err
	}
	return branch_lists, nil
}
