package services

import (
	"context"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
)

func (s Service) CreateBranch(ctx context.Context, user gmodel.User, input gmodel.CreateBranch) (branch gmodel.Branch, err error) {
	if err := s.StructValidator.StructCtx(ctx, input); err != nil {
		return branch, err
	}
	if !s.StoreExists(ctx, input.StoreID) {
		return branch, fmt.Errorf("store id is invalid")
	}

	s.TX, err = s.DB.BeginTx(ctx, nil)
	if err != nil {
		return branch, err
	}
	defer s.TX.Rollback()

	address, err := s.CreateAddress(ctx, &user, *input.Address)
	if err != nil {
		return branch, err
	}

	qb := table.Branch.INSERT(
		table.Branch.Name,
		table.Branch.AddressID,
		table.Branch.StoreID,
		table.Branch.CreatedByID,
		table.Branch.UpdatedByID,
	).MODEL(model.Branch{
		Name: input.Name,
		AddressID: address.ID,
		StoreID: input.StoreID,
		CreatedByID: &user.ID,
		UpdatedByID: &user.ID,
	}).
	RETURNING(table.Branch.AllColumns)
	
	if err := qb.QueryContext(ctx, s.TX, &branch); err != nil {
		return branch, err
	}

	if err := s.TX.Commit(); err != nil {
		return branch, err
	}
	return branch, err
} 

func (s Service) FindBranchesByStoreId(ctx context.Context, store_id int64) (branches []gmodel.Branch, err error) {
	created_user_table, updated_user_table, user_cols := s.CreatedAndUpdatedUserTable()
	columns := []postgres.Projection{
		table.Address.AllColumns,
		table.Country.Name,
	}
	columns = append(columns, user_cols...)
	qb := table.Branch.
		SELECT(
			table.Branch.AllColumns,
			columns...,
		).
		FROM(
			table.Branch.
				INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
				INNER_JOIN(table.Country, table.Country.Code.EQ(table.Address.CountryCode)).
				LEFT_JOIN(created_user_table, created_user_table.ID.EQ(table.Branch.CreatedByID)).
				LEFT_JOIN(updated_user_table, updated_user_table.ID.EQ(table.Branch.UpdatedByID)),
		).
		WHERE(table.Branch.StoreID.EQ(postgres.Int(store_id))).
		ORDER_BY(table.Branch.CreatedAt.DESC())
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &branches)
	return branches, err
}

func (s Service) FindBranchByBranchIdAndStoreId(ctx context.Context, branch_id int64, store_id int64) (branch gmodel.Branch, err error) {
	created_user_table, updated_user_table, user_cols := s.CreatedAndUpdatedUserTable()
	columns := []postgres.Projection{
		table.Address.AllColumns,
		table.Country.Name,
	}
	columns = append(columns, user_cols...)
	qb := table.Branch.
		SELECT(
			table.Branch.AllColumns,
			columns...,
		).
		FROM(
			table.Branch.
				INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
				INNER_JOIN(table.Country, table.Country.Code.EQ(table.Address.CountryCode)).
				LEFT_JOIN(created_user_table, created_user_table.ID.EQ(table.Branch.CreatedByID)).
				LEFT_JOIN(updated_user_table, updated_user_table.ID.EQ(table.Branch.UpdatedByID)),
		).
		WHERE(
			table.Branch.StoreID.EQ(postgres.Int(store_id)).
			AND(table.Branch.ID.EQ(postgres.Int(branch_id))),
		).
		LIMIT(1)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &branch)
	return branch, err
}
