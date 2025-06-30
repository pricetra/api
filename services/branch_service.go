package services

import (
	"context"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
)

func (s Service) CreateBranchFromFullAddress(ctx context.Context, user gmodel.User, store_id int64, full_address string) (branch gmodel.Branch, err error) {
	store, err := s.FindStore(ctx, store_id)
	if err != nil {
		return gmodel.Branch{}, fmt.Errorf("store id is invalid")
	}

	address, err := s.FullAddressToCreateAddress(ctx, full_address)
	if err != nil {
		return gmodel.Branch{}, err
	}

	return s.CreateBranch(ctx, user, gmodel.CreateBranch{
		Name: fmt.Sprintf("%s %s", store.Name, address.City),
		Address: &address,
		StoreID: store_id,
	})
}

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
	branch.Address = &address
	return branch, err
}

func (s Service) FindBranchById(ctx context.Context, id int64) (branch gmodel.Branch, err error) {
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
		WHERE(table.Branch.ID.EQ(postgres.Int(id))).
		LIMIT(1)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &branch)
	return branch, err
}

func (s Service) FindBranchesByStoreId(
	ctx context.Context,
	store_id int64,
	paginator_input gmodel.PaginatorInput,
	search *string,
	location *gmodel.LocationInput,
) (branches gmodel.PaginatedBranches, err error) {
	created_user_table, updated_user_table, user_cols := s.CreatedAndUpdatedUserTable()
	tables := table.Branch.
		INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
		INNER_JOIN(table.Country, table.Country.Code.EQ(table.Address.CountryCode)).
		LEFT_JOIN(created_user_table, created_user_table.ID.EQ(table.Branch.CreatedByID)).
		LEFT_JOIN(updated_user_table, updated_user_table.ID.EQ(table.Branch.UpdatedByID))

	columns := []postgres.Projection{
		table.Address.AllColumns,
		table.Country.Name,
	}
	columns = append(columns, user_cols...)

	where_clause := table.Branch.StoreID.EQ(postgres.Int(store_id))
	order_by := []postgres.OrderByClause{}
	if search != nil {
		full_text_components := s.BuildFullTextSearchQueryComponents(*search)
		columns = append(columns, full_text_components.RankColumn)
		where_clause = where_clause.AND(full_text_components.WhereClause)
		order_by = append(order_by, full_text_components.OrderByClause.DESC())
	}
	if location != nil {
		dist := s.GetDistanceCols(location.Latitude, location.Longitude, location.RadiusMeters)
		columns = append(columns, dist.DistanceColumn)
		order_by = append(order_by, postgres.FloatColumn(dist.DistanceColumnName).ASC())
		where_clause = where_clause.AND(dist.DistanceWhereClauseWithRadius)
	}

	sql_paginator, err := s.Paginate(ctx, paginator_input, tables, table.Product.ID, where_clause)
	if err != nil {
		// Return empty result
		return gmodel.PaginatedBranches{
			Branches: []*gmodel.Branch{},
			Paginator: &gmodel.Paginator{},
		}, nil
	}

	order_by = append(order_by, table.Branch.CreatedAt.DESC())
	qb := table.Branch.
		SELECT(table.Branch.AllColumns, columns...).
		FROM(tables).
		WHERE(where_clause).
		ORDER_BY(order_by...).
		LIMIT(int64(sql_paginator.Limit)).
		OFFSET(int64(sql_paginator.Offset))
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

func (s Service) FindBranchesByCoordinates(ctx context.Context, lat float64, lon float64, radius_meters int) (branches []gmodel.Branch, err error) {
	created_user_table, updated_user_table, user_cols := s.CreatedAndUpdatedUserTable()
	distance_cols := s.GetDistanceCols(lat, lon, &radius_meters)
	columns := []postgres.Projection{
		table.Store.AllColumns,
		table.Address.AllColumns,
		table.Country.Name,
		distance_cols.DistanceColumn,
	}
	columns = append(columns, user_cols...)
	qb := table.Branch.
		SELECT(
			table.Branch.AllColumns,
			columns...,
		).
		FROM(
			table.Branch.
				INNER_JOIN(table.Store, table.Store.ID.EQ(table.Branch.StoreID)).
				INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
				INNER_JOIN(table.Country, table.Country.Code.EQ(table.Address.CountryCode)).
				LEFT_JOIN(created_user_table, created_user_table.ID.EQ(table.Branch.CreatedByID)).
				LEFT_JOIN(updated_user_table, updated_user_table.ID.EQ(table.Branch.UpdatedByID)),
		).
		WHERE(distance_cols.DistanceWhereClauseWithRadius).
		ORDER_BY(
			postgres.FloatColumn(distance_cols.DistanceColumnName).ASC(),
			table.Branch.ID.ASC(),
		)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &branches)
	return branches, err
}

func (s Service) AllFavoriteBranchProductPrices(ctx context.Context, user gmodel.User, product_id int64) (branches []gmodel.BranchListWithPrices, err error) {
	if !s.ProductExists(ctx, product_id) {
		return nil, fmt.Errorf("invalid product")
	}

	db := s.DbOrTxQueryable()
	branch_list := table.BranchList.AS("branch_list_with_prices")
	qb := table.List.SELECT(
			branch_list.AllColumns,
			table.Branch.AllColumns,
			table.Address.AllColumns,
			table.Store.AllColumns,
			table.Stock.AllColumns,
			table.Price.AllColumns,
		).
		FROM(table.List.
			INNER_JOIN(branch_list, branch_list.ListID.EQ(table.List.ID)).
			INNER_JOIN(table.Branch, table.Branch.ID.EQ(branch_list.BranchID)).
			INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
			INNER_JOIN(table.Store, table.Store.ID.EQ(table.Branch.StoreID)).
			LEFT_JOIN(table.Stock, 
				table.Stock.ProductID.EQ(postgres.Int(product_id)).
				AND(table.Stock.BranchID.EQ(table.Branch.ID)),
			).
			LEFT_JOIN(table.Price, table.Price.ID.EQ(table.Stock.LatestPriceID)),
		).
		WHERE(
			table.List.UserID.EQ(postgres.Int(user.ID)).
			AND(table.List.Type.EQ(postgres.NewEnumValue(
				model.ListType_Favorites.String(),
			))),
		).
		ORDER_BY(branch_list.CreatedAt.DESC())
	if err := qb.QueryContext(ctx, db, &branches); err != nil {
		return nil, err
	}

	for i, bl := range branches {
		if bl.Stock != nil && bl.Stock.ID != 0 {
			continue
		}
		if bl.Branch == nil || bl.Branch.Address == nil {
			return nil, fmt.Errorf("unexpected error")
		}
		// since a stock is not present we can use
		// stocks from other branches
		dist := 32187 // 20 miles
		distance_cols := s.GetDistanceCols(bl.Branch.Address.Latitude, bl.Branch.Address.Longitude, &dist)
		branch_qb := table.Branch.
			SELECT(postgres.AVG(table.Price.Amount).AS("avg")).
			FROM(
				table.Branch.
					INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
					INNER_JOIN(
						table.Stock, 
						table.Stock.BranchID.EQ(table.Branch.ID).
							AND(table.Stock.ProductID.EQ(postgres.Int(product_id))),
					).
					INNER_JOIN(table.Price, table.Price.ID.EQ(table.Stock.LatestPriceID)),
			).
			WHERE(
				table.Branch.StoreID.EQ(postgres.Int(bl.Branch.StoreID)).
				AND(table.Address.AdministrativeDivision.EQ(
					postgres.String(bl.Branch.Address.AdministrativeDivision),
				)).
				AND(distance_cols.DistanceWhereClauseWithRadius),
			)
		avg := struct{Avg float64}{}
		if err := branch_qb.QueryContext(ctx, db, &avg); err != nil {
			return nil, err
		}
		branches[i].Stock = nil
		if avg.Avg != 0 {
			branches[i].ApproximatePrice = &avg.Avg
		}
	}
	return branches, nil
}
