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

func (s Service) CreateStock(ctx context.Context, user gmodel.User, input gmodel.CreateStock) (stock gmodel.Stock, err error) {
	qb := table.Stock.INSERT(
		table.Stock.ProductID,
		table.Stock.BranchID,
		table.Stock.StoreID,
		table.Stock.CreatedByID,
		table.Stock.UpdatedByID,
	).MODEL(model.Stock{
		ProductID: input.ProductID,
		BranchID: input.BranchID,
		StoreID: input.StoreID,
		CreatedByID: &user.ID,
		UpdatedByID: &user.ID,
	}).RETURNING(table.Stock.AllColumns);
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &stock); err != nil {
		return gmodel.Stock{}, err
	}
	return stock, nil
}

func (s Service) FindStockById(ctx context.Context, id int64) (stock gmodel.Stock, err error) {
	qb := table.Stock.
		SELECT(table.Stock.AllColumns).
		FROM(table.Stock).
		WHERE(table.Stock.ID.EQ(postgres.Int(id))).
		LIMIT(1)
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &stock); err != nil {
		return gmodel.Stock{}, err
	}
	return stock, nil
}

func (s Service) FindStock(
	ctx context.Context,
	product_id int64,
	branch_id int64,
	store_id int64,
) (stock gmodel.Stock, err error) {
	qb := table.Stock.
		SELECT(table.Stock.AllColumns).
		FROM(table.Stock).
		WHERE(postgres.AND(
			table.Stock.ProductID.EQ(postgres.Int(product_id)),
			table.Stock.BranchID.EQ(postgres.Int(branch_id)),
			table.Stock.StoreID.EQ(postgres.Int(store_id)),
		)).
		LIMIT(1)
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &stock); err != nil {
		return gmodel.Stock{}, err
	}
	return stock, nil
}

func (s Service) FindOrCreateStock(
	ctx context.Context,
	user gmodel.User,
	product_id int64,
	branch_id int64,
	store_id int64,
) (stock gmodel.Stock, err error) {
	stock, err = s.FindStock(ctx, product_id, branch_id, store_id)
	if err == nil {
		return stock, nil
	}
	create_stock_input := gmodel.CreateStock{
		ProductID: product_id,
		BranchID: branch_id,
		StoreID: store_id,
	}
	stock, err = s.CreateStock(ctx, user, create_stock_input)
	if err != nil {
		return gmodel.Stock{}, err
	}
	return stock, nil
}

func (s Service) UpdateStockWithLatestPrice(ctx context.Context, user gmodel.User, stock_id int64, price_id int64) (updated_stock gmodel.Stock, err error) {
	qb := table.Stock.
		UPDATE(
			table.Stock.LatestPriceID,
			table.Stock.UpdatedByID,
			table.Stock.UpdatedAt,
		).MODEL(model.Stock{
			LatestPriceID: &price_id,
			UpdatedByID: &user.ID,
			UpdatedAt: time.Now(),
		}).WHERE(
			table.Stock.ID.EQ(postgres.Int(stock_id)),
		).RETURNING(table.Stock.AllColumns)
	db := s.DbOrTxQueryable()
	if err = qb.QueryContext(ctx, db, &updated_stock); err != nil {
		return gmodel.Stock{}, err
	}
	return updated_stock, nil
}

func (s Service) GetStocksForProduct(ctx context.Context, product_id int64, location *gmodel.LocationInput) (stocks []gmodel.Stock, err error) {
	if !s.ProductExists(ctx, product_id) {
		return nil, fmt.Errorf("invalid product id")
	}

	created_at_table, updated_at_table, cols := s.CreatedAndUpdatedUserTable()
	cols = append(
		cols,
		table.Branch.AllColumns,
		table.Store.AllColumns,
		table.Address.AllColumns,
		table.Price.AllColumns,
	)
	where_clause := table.Product.ID.EQ(postgres.Int(product_id))
	order_by := []postgres.OrderByClause{}
	if location != nil {
		l := *location
		d := s.GetDistanceCols(l.Latitude, l.Longitude, l.RadiusMeters)
		cols = append(cols, d.DistanceColumn)
		where_clause = where_clause.AND(d.DistanceWhereClauseWithRadius)
		order_by = append(order_by, postgres.FloatColumn(d.DistanceColumnName).ASC())
	}
	order_by = append(
		order_by,
		table.Price.CreatedAt.DESC(),
		table.Stock.CreatedAt.DESC(),
	)
	qb := table.Stock.
		SELECT(table.Stock.AllColumns, cols...).
		FROM(table.Stock.
			INNER_JOIN(table.Product, table.Product.ID.EQ(table.Stock.ProductID)).
			INNER_JOIN(table.Branch, table.Branch.ID.EQ(table.Stock.BranchID)).
			INNER_JOIN(table.Store, table.Store.ID.EQ(table.Stock.StoreID)).
			INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
			INNER_JOIN(table.Price, table.Price.ID.EQ(table.Stock.LatestPriceID)).
			LEFT_JOIN(created_at_table, created_at_table.ID.EQ(table.Price.CreatedByID)).
			LEFT_JOIN(updated_at_table, updated_at_table.ID.EQ(table.Price.UpdatedByID)),
		).
		WHERE(where_clause).
		ORDER_BY(order_by...)
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &stocks); err != nil {
		return nil, err
	}
	return stocks, nil
}
