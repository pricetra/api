package services

import (
	"context"

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
	created_user_table, updated_user_table, user_cols := s.CreatedAndUpdatedUserTable()
	qb := table.Store.SELECT(
		table.Store.AllColumns,
		user_cols...,
	).FROM(
		table.Store.
			LEFT_JOIN(created_user_table, created_user_table.ID.EQ(table.Store.CreatedByID)).
			LEFT_JOIN(updated_user_table, updated_user_table.ID.EQ(table.Store.UpdatedByID)),
	).
	ORDER_BY(table.Store.CreatedAt.DESC())
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &stores)
	return stores, err
}
