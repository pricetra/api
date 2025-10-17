package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/Goldziher/go-utils/sliceutils"
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

	address, err := s.FindOrCreateAddress(ctx, &user, *input.Address)
	if err != nil {
		return branch, err
	}

	if _, err := s.FindBranchByStoreIdAndAddressId(ctx, input.StoreID, address.ID); err == nil {
		return gmodel.Branch{}, fmt.Errorf("branch with this store and address already exists")
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
	columns := []postgres.Projection{
		table.Address.AllColumns,
		table.Country.Name,
	}
	qb := table.Branch.
		SELECT(
			table.Branch.AllColumns,
			columns...,
		).
		FROM(
			table.Branch.
				INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
				INNER_JOIN(table.Country, table.Country.Code.EQ(table.Address.CountryCode)),
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
) (res gmodel.PaginatedBranches, err error) {
	tables := table.Branch.
		INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
		INNER_JOIN(table.Country, table.Country.Code.EQ(table.Address.CountryCode))

	columns := []postgres.Projection{
		table.Address.AllColumns,
		table.Country.Name,
	}

	where_clause := table.Branch.StoreID.EQ(postgres.Int(store_id))
	order_by := []postgres.OrderByClause{}
	if search != nil && len(strings.TrimSpace(*search)) > 0 {
		full_text_components := s.BuildFullTextSearchQueryComponents(table.Address.SearchVector, *search)
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

	sql_paginator, err := s.Paginate(ctx, paginator_input, tables, table.Branch.ID, where_clause)
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
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &res.Branches); err != nil {
		return gmodel.PaginatedBranches{}, err
	}
	
	res.Paginator = &sql_paginator.Paginator
	return res, err
}

func (s Service) PaginatedBranches(
	ctx context.Context,
	paginator_input gmodel.PaginatorInput,
	search *string,
	location *gmodel.LocationInput,
	branch_ids ...int64,
) (res gmodel.PaginatedBranches, err error) {
	tables := table.Branch.
		INNER_JOIN(table.Store, table.Store.ID.EQ(table.Branch.StoreID)).
		INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
		INNER_JOIN(table.Country, table.Country.Code.EQ(table.Address.CountryCode))

	columns := []postgres.Projection{
		table.Store.AllColumns,
		table.Address.AllColumns,
		table.Country.Name,
	}

	where_clause := postgres.Bool(true)
	order_by := []postgres.OrderByClause{}
	if search != nil && len(strings.TrimSpace(*search)) > 0 {
		full_text_components := s.BuildFullTextSearchQueryComponents(table.Address.SearchVector, *search)
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
	if len(branch_ids) > 0 {
		ids := sliceutils.Map(branch_ids, func(val int64, index int, arr []int64) postgres.Expression {
			return postgres.Int(val)
		})
		where_clause = where_clause.AND(table.Branch.ID.IN(ids...))
	}

	sql_paginator, err := s.Paginate(ctx, paginator_input, tables, table.Branch.ID, where_clause)
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
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &res.Branches); err != nil {
		return gmodel.PaginatedBranches{}, err
	}
	
	res.Paginator = &sql_paginator.Paginator
	return res, err
}

func (s Service) FindBranchByBranchIdAndStoreId(ctx context.Context, branch_id int64, store_id int64) (branch gmodel.Branch, err error) {
	columns := []postgres.Projection{
		table.Address.AllColumns,
		table.Country.Name,
	}
	qb := table.Branch.
		SELECT(
			table.Branch.AllColumns,
			columns...,
		).
		FROM(
			table.Branch.
				INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
				INNER_JOIN(table.Country, table.Country.Code.EQ(table.Address.CountryCode)),
		).
		WHERE(
			table.Branch.StoreID.EQ(postgres.Int(store_id)).
			AND(table.Branch.ID.EQ(postgres.Int(branch_id))),
		).
		LIMIT(1)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &branch)
	return branch, err
}

func (s Service) FindBranchByStoreIdAndAddressId(ctx context.Context, store_id int64, address_id int64) (branch gmodel.Branch, err error) {
	qb := table.Branch.
		SELECT(table.Branch.AllColumns).
		FROM(table.Branch).
		WHERE(
			table.Branch.StoreID.EQ(postgres.Int(store_id)).
				AND(table.Branch.AddressID.EQ(postgres.Int(address_id))),
		).LIMIT(1)
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &branch); err != nil {
		return gmodel.Branch{}, err
	}
	return branch, nil
}

func (s Service) FindBranchesByCoordinates(ctx context.Context, lat float64, lon float64, radius_meters int) (branches []gmodel.Branch, err error) {
	distance_cols := s.GetDistanceCols(lat, lon, &radius_meters)
	columns := []postgres.Projection{
		table.Store.AllColumns,
		table.Address.AllColumns,
		table.Country.Name,
		distance_cols.DistanceColumn,
	}
	qb := table.Branch.
		SELECT(
			table.Branch.AllColumns,
			columns...,
		).
		FROM(
			table.Branch.
				INNER_JOIN(table.Store, table.Store.ID.EQ(table.Branch.StoreID)).
				INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
				INNER_JOIN(table.Country, table.Country.Code.EQ(table.Address.CountryCode)),
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

func (s Service) BranchesWithProducts(
	ctx context.Context,
	paginator_input gmodel.PaginatorInput,
	product_limit int,
	search *gmodel.ProductSearch,
) (res gmodel.PaginatedBranches, err error) {
	if search == nil {
		return gmodel.PaginatedBranches{}, fmt.Errorf("filters is required")
	}
	if search.BranchIds == nil && search.Location == nil {
		return gmodel.PaginatedBranches{}, fmt.Errorf("branch ids or location filters are required")
	}

	branch_where_clause, _, _ := s.ProductFiltersBuilder(search)
	branch_ids_qb_col := []postgres.Projection{}
	branch_ids_qb_group_by := []postgres.GroupByClause{table.Branch.ID}
	if search.Location != nil {
		d := s.GetDistanceCols(search.Location.Latitude, search.Location.Longitude, search.Location.RadiusMeters)
		branch_ids_qb_col = append(branch_ids_qb_col, d.DistanceColumn)
		branch_ids_qb_group_by = append(branch_ids_qb_group_by, postgres.FloatColumn(d.DistanceColumnName))
	}
	branch_ids_qb := table.Branch.
		SELECT(table.Branch.ID, branch_ids_qb_col...).
		FROM(
			table.Branch.
				INNER_JOIN(table.Store, table.Store.ID.EQ(table.Branch.StoreID)).
				INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
				INNER_JOIN(table.Stock, table.Stock.BranchID.EQ(table.Branch.ID)).
				INNER_JOIN(table.Price, table.Price.ID.EQ(table.Stock.LatestPriceID)).
				INNER_JOIN(table.Product, table.Product.ID.EQ(table.Price.ProductID)).
				INNER_JOIN(table.Category, table.Category.ID.EQ(table.Product.CategoryID)),
		).
		WHERE(branch_where_clause).
		GROUP_BY(branch_ids_qb_group_by...)
	branch_ids_table := branch_ids_qb.AsTable("branch_ids_table")
	sql_paginator, err := s.Paginate(
		ctx,
		paginator_input, 
		branch_ids_table,
		table.Branch.ID.From(branch_ids_table),
		nil,
	)
	if err != nil {
		// Return empty result
		return gmodel.PaginatedBranches{
			Branches: []*gmodel.Branch{},
			Paginator: &gmodel.Paginator{},
		}, nil
	}

	var distance_cols *DistanceColumns
	if search.Location != nil {
		d := s.GetDistanceCols(search.Location.Latitude, search.Location.Longitude, search.Location.RadiusMeters)
		distance_cols = &d
	}

	created_user_table, updated_user_table, cols := s.CreatedAndUpdatedUserTable()
	cols = append(
		cols,
		table.Category.AllColumns,
		table.Stock.AllColumns,
		table.Store.AllColumns,
		table.Branch.AllColumns,
		table.Price.AllColumns,
		table.Address.AllColumns,
	)
	search.Location = nil // disable location filtering
	search.BranchID = nil // disable branch filtering
	search.BranchIds = nil // disable branch filtering
	search.StoreID = nil // disable store filtering

	where_clause, order_by, filter_cols := s.ProductFiltersBuilder(search)

	order_by_with_distance := []postgres.OrderByClause{}
	if distance_cols != nil {
		cols = append(cols, distance_cols.DistanceColumn)
		order_by_with_distance = append(order_by_with_distance, postgres.FloatColumn(distance_cols.DistanceColumnName).ASC())
	}
	order_by = append(
		order_by,
		table.Price.CreatedAt.DESC(),
		table.Product.Views.DESC(),
	)
	cols = append(cols, filter_cols...)

	// Represents all the paginated branch ids
	if len(order_by_with_distance) > 0 {
		branch_ids_qb = branch_ids_qb.ORDER_BY(order_by_with_distance...)
	}
	paginated_branch_ids_table := branch_ids_qb.
		LIMIT(int64(sql_paginator.Limit)).
		OFFSET(int64(sql_paginator.Offset)).
		AsTable("paginated_branch_ids_table")

	// Subquery/CTE (Common table expression) to get stocks with row numbers
	// this allows us to limit the number of products per branch
	row_num_col_name := "rn"
	row_number_col := postgres.
		ROW_NUMBER().
		OVER(
			postgres.
				PARTITION_BY(table.Stock.BranchID).
				ORDER_BY(order_by...),
		).AS(row_num_col_name)
	stock_cte := postgres.CTE("stock_cte")
	stock_sub_query_cols := append(filter_cols, row_number_col)
	stock_sub_query := table.Stock.
			SELECT(
				table.Stock.ID,
				stock_sub_query_cols...,
			).
			FROM(
				paginated_branch_ids_table.
					INNER_JOIN(table.Stock, table.Stock.BranchID.EQ(table.Branch.ID.From(paginated_branch_ids_table))).
					INNER_JOIN(table.Price, table.Price.ID.EQ(table.Stock.LatestPriceID)).
					INNER_JOIN(table.Product, table.Product.ID.EQ(table.Stock.ProductID)).
					INNER_JOIN(table.Category, table.Category.ID.EQ(table.Product.CategoryID)),
			).
			WHERE(where_clause)

	final_order_by := []postgres.OrderByClause{}
	final_order_by = append(final_order_by, order_by_with_distance...)
	final_order_by = append(final_order_by, order_by...)
	qb := postgres.
		WITH(stock_cte.AS(stock_sub_query))(
			// Main query to select products joining with the Stock CTE
			table.Product.
				SELECT(table.Product.AllColumns, cols...).
				FROM(
					stock_cte.
						INNER_JOIN(table.Stock, table.Stock.ID.EQ(table.Stock.ID.From(stock_cte))).
						INNER_JOIN(table.Product, table.Product.ID.EQ(table.Stock.ProductID)).
						INNER_JOIN(table.Category, table.Category.ID.EQ(table.Product.CategoryID)).
						INNER_JOIN(table.Store, table.Store.ID.EQ(table.Stock.StoreID)).
						INNER_JOIN(table.Branch, table.Branch.ID.EQ(table.Stock.BranchID)).
						INNER_JOIN(table.Price, table.Price.ID.EQ(table.Stock.LatestPriceID)).
						INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
						LEFT_JOIN(created_user_table, created_user_table.ID.EQ(table.Price.CreatedByID)).
						LEFT_JOIN(updated_user_table, updated_user_table.ID.EQ(table.Price.UpdatedByID)),
				).WHERE(
					postgres.AND(
						table.Stock.ID.IN(table.Stock.ID.From(stock_cte)),
						postgres.IntegerColumn(row_num_col_name).LT_EQ(postgres.Int(int64(product_limit))),
					),
				).ORDER_BY(final_order_by...),
		)
	var branches []*gmodel.Branch
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &branches); err != nil {
		return gmodel.PaginatedBranches{
			Branches: []*gmodel.Branch{},
			Paginator: &gmodel.Paginator{},
		}, nil
	}

	res.Branches = branches
	res.Paginator = &sql_paginator.Paginator
	return res, nil
}
