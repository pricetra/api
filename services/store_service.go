package services

import (
	"context"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
)

func (s Service) CreateStore(ctx context.Context, user gmodel.User, input gmodel.CreateStore) (store gmodel.Store, err error) {
	if err := s.StructValidator.StructCtx(ctx, input); err != nil {
		return store, err
	}

	qb := table.Store.
		INSERT(
			table.Store.AllColumns.Except(
				table.Store.ID,
				table.Store.CreatedAt,
				table.Store.UpdatedAt,
			),
		).
		MODEL(struct{
			gmodel.CreateStore
			CreatedByID *int64
			UpdatedByID *int64
		}{
			CreateStore: input,
			CreatedByID: &user.ID,
			UpdatedByID: &user.ID,
		}).
		RETURNING(table.Store.AllColumns)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &store)
	return store, err
}

func (s Service) GetAllStores(ctx context.Context) (stores []gmodel.Store, err error) {
	created_user_table, updated_user_table, cols := s.CreatedAndUpdatedUserTable()

	stock_count_col_name := "stock_count"
	stock_count_col := postgres.COUNT(table.Stock.ID)
	cols = append(cols, stock_count_col.AS(stock_count_col_name))
	qb := table.Store.SELECT(
		table.Store.AllColumns,
		cols...,
	).FROM(
		table.Store.
			LEFT_JOIN(table.Stock, table.Stock.StoreID.EQ(table.Store.ID)).
			LEFT_JOIN(created_user_table, created_user_table.ID.EQ(table.Store.CreatedByID)).
			LEFT_JOIN(updated_user_table, updated_user_table.ID.EQ(table.Store.UpdatedByID)),
	).
	GROUP_BY(table.Store.ID, created_user_table.ID, updated_user_table.ID).
	ORDER_BY(stock_count_col.DESC(), table.Store.CreatedAt.ASC())
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &stores)
	return stores, err
}

func (s Service) FindStore(ctx context.Context, id int64) (store gmodel.Store, err error) {
	created_user_table, updated_user_table, user_cols := s.CreatedAndUpdatedUserTable()
	qb := table.Store.SELECT(
		table.Store.AllColumns,
		user_cols...,
	).FROM(
		table.Store.
			LEFT_JOIN(created_user_table, created_user_table.ID.EQ(table.Store.CreatedByID)).
			LEFT_JOIN(updated_user_table, updated_user_table.ID.EQ(table.Store.UpdatedByID)),
	).
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
