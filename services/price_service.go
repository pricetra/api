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

func (s Service) CreatePrice(ctx context.Context, user gmodel.User, input gmodel.CreatePrice) (price gmodel.Price, err error) {
	product, err := s.FindProductById(ctx, input.ProductID)
	if err != nil {
		return gmodel.Price{}, fmt.Errorf("could not find product")
	}
	branch, err := s.FindBranchById(ctx, input.BranchID)
	if err != nil {
		return gmodel.Price{}, fmt.Errorf("could not find branch")
	}
	stock, err := s.FindOrCreateStock(ctx, user, product.ID, branch.ID, branch.StoreID)
	if err != nil {
		return gmodel.Price{}, err
	}

	// if original price is provided then add that first as an entry
	if input.Sale && input.OriginalPrice != nil {
		s.CreatePrice(ctx, user, gmodel.CreatePrice{
			ProductID: input.ProductID,
			Amount: *input.OriginalPrice,
			BranchID: input.BranchID,
			CurrencyCode: input.CurrencyCode,
		})
	}

	if input.Sale && input.ExpiresAt == nil {
		next_week := time.Now().Add(time.Hour * 24 * 7)
		input.ExpiresAt = &next_week
	}

	currency_code := "USD"
	if input.CurrencyCode != nil {
		currency_code = *input.CurrencyCode
	}
	qb := table.Price.INSERT(
		table.Price.Amount,
		table.Price.CurrencyCode,
		table.Price.ProductID,
		table.Price.BranchID,
		table.Price.StoreID,
		table.Price.StockID,
		table.Price.Sale,
		table.Price.OriginalPrice,
		table.Price.Condition,
		table.Price.UnitType,
		table.Price.ImageID,
		table.Price.ExpiresAt,
		table.Price.CreatedByID,
		table.Price.UpdatedByID,
	).MODEL(model.Price{
		Amount: input.Amount,
		CurrencyCode: currency_code,
		ProductID: product.ID,
		StoreID: branch.StoreID,
		BranchID: branch.ID,
		StockID: stock.ID,
		Sale: input.Sale,
		OriginalPrice: input.OriginalPrice,
		Condition: input.Condition,
		UnitType: input.UnitType,
		ImageID: input.ImageID,
		ExpiresAt: input.ExpiresAt,
		CreatedByID: &user.ID,
		UpdatedByID: &user.ID,
	}).RETURNING(table.Price.AllColumns)
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &price); err != nil {
		return gmodel.Price{}, err
	}

	updated_stock, _ := s.UpdateStockWithLatestPrice(ctx, user, stock.ID, price.ID)
	price.Stock = &updated_stock
	price.Product = &product
	price.Branch = &branch
	price.Store = branch.Store
	return price, nil
}

func (s Service) LatestPriceForProduct(ctx context.Context, product_id int64, branch_id int64) (price gmodel.Price, err error) {
	qb := table.Price.
		SELECT(table.Price.AllColumns).
		FROM(table.Price).
		WHERE(
			postgres.AND(
				table.Price.ProductID.EQ(postgres.Int(product_id)),
				table.Price.BranchID.EQ(postgres.Int(branch_id)),
			),
		).
		ORDER_BY(table.Price.ID.DESC()).
		LIMIT(1)
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &price); err != nil {
		return gmodel.Price{}, err
	}
	return price, nil
}
