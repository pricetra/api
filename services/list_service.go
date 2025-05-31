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
	where_clause := table.List.UserID.EQ(postgres.Int(user.ID))
	if list_type != nil {
		where_clause = where_clause.
			AND(table.List.Type.EQ(postgres.NewEnumValue(list_type.String())))
	}
	branch_list_branch_table := table.Branch.AS("branch_list_branch")
	qb := table.List.
		SELECT(
			table.List.AllColumns,
			table.ProductList.AllColumns,
			table.Product.AllColumns,
			table.Category.AllColumns,
			table.Stock.AllColumns,
			table.Price.AllColumns,
			table.BranchList.AllColumns,
			branch_list_branch_table.AllColumns,
			table.Store.AllColumns,
			table.Address.AllColumns,
		).
		FROM(table.List.
			LEFT_JOIN(table.ProductList, table.ProductList.ListID.EQ(table.List.ID)).
			LEFT_JOIN(table.Product, table.Product.ID.EQ(table.ProductList.ProductID)).
			LEFT_JOIN(table.Category, table.Category.ID.EQ(table.Product.CategoryID)).
			LEFT_JOIN(table.Stock, table.Stock.ID.EQ(table.ProductList.StockID)).
			LEFT_JOIN(table.Price, table.Price.ID.EQ(table.Stock.LatestPriceID)).
			LEFT_JOIN(table.BranchList, table.BranchList.ListID.EQ(table.List.ID)).
			LEFT_JOIN(branch_list_branch_table, branch_list_branch_table.ID.EQ(table.BranchList.BranchID)).
			LEFT_JOIN(table.Store, table.Store.ID.EQ(branch_list_branch_table.StoreID)).
			LEFT_JOIN(table.Address, table.Address.ID.EQ(branch_list_branch_table.AddressID)),
		).
		WHERE(where_clause).
		ORDER_BY(
			table.List.CreatedAt.ASC(),
			table.ProductList.CreatedAt.DESC(),
			table.BranchList.CreatedAt.DESC(),
		)
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &lists); err != nil {
		return nil, err
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
) (product_list gmodel.ProductList, err error) {
	qb := table.ProductList.
		SELECT(table.ProductList.AllColumns).
		FROM(table.ProductList).
		WHERE(
			table.ProductList.ListID.EQ(postgres.Int(list_id)).
			AND(table.ProductList.UserID.EQ(postgres.Int(user.ID))).
			AND(table.ProductList.ProductID.EQ(postgres.Int(product_id))),
		).LIMIT(1)
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &product_list); err != nil {
		return gmodel.ProductList{}, err
	}
	return product_list, nil
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
	existing_product_list, _ := s.FindProductListWithProductId(ctx, user, list.ID, product_id)
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
