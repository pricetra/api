package services

import (
	"context"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
	"github.com/pricetra/api/utils"
)

func (s Service) CreateStore(ctx context.Context, user gmodel.User, input gmodel.CreateStore) (store gmodel.Store, err error) {
	if err := s.StructValidator.StructCtx(ctx, input); err != nil {
		return gmodel.Store{}, err
	}
	if input.LogoBase64 == nil && input.LogoFile == nil {
		return gmodel.Store{}, fmt.Errorf("logo file or base64 is required")
	}
	if input.LogoBase64 != nil && !utils.IsValidBase64Image(*input.LogoBase64) {
		return gmodel.Store{}, fmt.Errorf("invalid base64 image provided")
	}

	qb := table.Store.
		INSERT(
			table.Store.Name,
			table.Store.Website,
			table.Store.Logo,
			table.Store.CreatedByID,
			table.Store.UpdatedByID,
		).
		MODEL(struct{
			gmodel.CreateStore
			Logo string
			CreatedByID *int64
			UpdatedByID *int64
		}{
			CreateStore: input,
			Logo: uuid.NewString(),
			CreatedByID: &user.ID,
			UpdatedByID: &user.ID,
		}).
		RETURNING(table.Store.AllColumns)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &store)
	return store, err
}

func (s Service) GetAllStores(ctx context.Context) (stores []gmodel.Store, err error) {
	stock_count_col_name := "stock_count"
	stock_count_col := postgres.COUNT(table.Stock.ID)
	qb := table.Store.SELECT(
		table.Store.AllColumns,
		stock_count_col.AS(stock_count_col_name),
	).FROM(
		table.Store.
			LEFT_JOIN(table.Stock, table.Stock.StoreID.EQ(table.Store.ID)),
	).
	GROUP_BY(table.Store.ID).
	ORDER_BY(stock_count_col.DESC(), table.Store.CreatedAt.ASC())
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &stores)
	return stores, err
}

func (s Service) PaginatedStores(ctx context.Context, paginator_input gmodel.PaginatorInput, search *string) (paginated_stores gmodel.PaginatedStores, err error) {
	where_clause := postgres.Bool(true)
	if search != nil && *search != "" {
		where_clause = where_clause.AND(
			postgres.RawBool(
				fmt.Sprintf("%s ILIKE $store_name", utils.BuildFullTableName(table.Store.Name)), 
				map[string]any{
					"$store_name": "%" + *search + "%",
				},
			),
		)
	}
	tables := table.Store.
			LEFT_JOIN(table.Stock, table.Stock.StoreID.EQ(table.Store.ID))

	sql_paginator, err := s.Paginate(
		ctx,
		paginator_input,
		table.Store,
		table.Store.ID,
		where_clause,
	)
	if err != nil {
		return gmodel.PaginatedStores{
			Stores: []*gmodel.Store{},
			Paginator: &gmodel.Paginator{},
		}, nil
	}

	stock_count_col_name := "stock_count"
	stock_count_col := postgres.COUNT(table.Stock.ID)
	qb := table.Store.
		SELECT(
			table.Store.AllColumns,
			stock_count_col.AS(stock_count_col_name),
		).
		FROM(tables).
		WHERE(where_clause).
		GROUP_BY(table.Store.ID,).
		ORDER_BY(stock_count_col.DESC(), table.Store.CreatedAt.ASC()).
		LIMIT(int64(sql_paginator.Limit)).
		OFFSET(int64(sql_paginator.Offset))
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &paginated_stores.Stores); err != nil {
		return gmodel.PaginatedStores{}, err
	}

	paginated_stores.Paginator = &sql_paginator.Paginator
	return paginated_stores, nil
}

func (s Service) FindStore(ctx context.Context, id int64) (store gmodel.Store, err error) {
	qb := table.Store.SELECT(table.Store.AllColumns).
	FROM(table.Store).
	WHERE(table.Store.ID.EQ(postgres.Int(id))).
	LIMIT(1)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &store)
	return store, err
}

func (s Service) StoreExists(ctx context.Context, id int64) bool {
	qb := table.Store.
		SELECT(table.Store.ID).
		FROM(table.Store).
		WHERE(table.Store.ID.EQ(postgres.Int(id))).
		LIMIT(1)
	var store struct{
		ID int64 `sql:"primary_key"`
	}
	err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &store)
	return err == nil
}
