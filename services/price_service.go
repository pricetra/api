package services

import (
	"context"
	"fmt"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"

	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
)

func (s Service) CreatePrice(ctx context.Context, user gmodel.User, input gmodel.CreatePrice) (price gmodel.Price, err error) {
	if err = s.StructValidator.StructCtx(ctx, input); err != nil {
		return gmodel.Price{}, fmt.Errorf("invalid input: %w", err)
	}

	if s.TX, err = s.DB.BeginTx(ctx, nil); err != nil {
		return gmodel.Price{}, fmt.Errorf("could not begin transaction")
	}
	defer s.TX.Rollback()

	product, err := s.FindProductById(ctx, input.ProductID)
	if err != nil {
		return gmodel.Price{}, fmt.Errorf("could not find product")
	}
	branch, err := s.FindBranchById(ctx, input.BranchID)
	if err != nil {
		return gmodel.Price{}, fmt.Errorf("could not find branch")
	}

	if !input.Sale && input.OriginalPrice != nil {
		// if original price is provided but not on sale then ignore it
		input.OriginalPrice = nil
	}

	if input.Sale && input.OriginalPrice != nil {
		if *input.OriginalPrice <= input.Amount {
			return gmodel.Price{}, fmt.Errorf("original price must be greater than the current price")
		}

		// if original price is provided then add that first as an entry
		_, err = s.CreatePrice(ctx, user, gmodel.CreatePrice{
			ProductID: input.ProductID,
			Amount: *input.OriginalPrice,
			BranchID: input.BranchID,
			CurrencyCode: input.CurrencyCode,
		})
		if err != nil {
			return gmodel.Price{}, fmt.Errorf("could not create original price entry: %w", err)
		}
	}

	stock, err := s.FindOrCreateStock(ctx, user, product.ID, branch.ID, branch.StoreID)
	if err != nil {
		return gmodel.Price{}, err
	}

	if input.Sale {
		// if original price is not provided use latest stock price
		if input.OriginalPrice == nil && stock.LatestPrice != nil {
			input.OriginalPrice = &stock.LatestPrice.Amount
		}
	
		if input.ExpiresAt == nil {
			next_week := time.Now().Add(time.Hour * 24 * 7)
			input.ExpiresAt = &next_week
		}
	
		if input.ExpiresAt.Before(time.Now()) {
			return gmodel.Price{}, fmt.Errorf("expiration date cannot be in the past")
		}
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
		table.Price.CreatedAt,
		table.Price.UpdatedAt,
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
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}).RETURNING(table.Price.AllColumns)
	if err = qb.QueryContext(ctx, s.TX, &price); err != nil {
		return gmodel.Price{}, err
	}

	updated_stock, err := s.UpdateStockWithLatestPrice(ctx, user, stock.ID, price.ID)
	if err != nil {
		return gmodel.Price{}, fmt.Errorf("could not update stock with latest price")
	}
	if err = s.TX.Commit(); err != nil {
		return gmodel.Price{}, fmt.Errorf("could not commit transaction")
	}
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

func (s Service) FindPrices(ctx context.Context, product_id int64, branch_id int64) (prices []gmodel.Price, err error) {
	created_by_user, _, _ := s.CreatedAndUpdatedUserTable()
	qb := table.Price.
		SELECT(
			table.Price.AllColumns,
			created_by_user.ID,
			created_by_user.Name,
			created_by_user.Avatar,
		).
		FROM(table.Price.
			LEFT_JOIN(created_by_user, created_by_user.ID.EQ(table.Price.CreatedByID)),
		).
		WHERE(
			postgres.AND(
				table.Price.ProductID.EQ(postgres.Int(product_id)),
				table.Price.BranchID.EQ(postgres.Int(branch_id)),
			),
		).
		ORDER_BY(table.Price.ID.DESC())
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &prices); err != nil {
		return nil, err
	}
	return prices, nil
}

func (s Service) SendPriceChangePushNotifications(ctx context.Context, users []gmodel.User, new_price gmodel.Price, old_price gmodel.Price) (res []expo.PushResponse, err error) {
	if len(users) == 0 {
		return res, nil
	}

	product, err := s.FindProductById(ctx, new_price.ProductID)
	if err != nil {
		return nil, err
	}
	data := map[string]string{
		"priceId": fmt.Sprint(new_price.ID),
		"productId": fmt.Sprint(product.ID),
		"stockId": fmt.Sprint(new_price.StockID),
		"priceAmount": fmt.Sprintf("$%.2f", new_price.Amount),
		"priceSale": fmt.Sprint(new_price.Sale),
	}
	var title, body string
	if new_price.Sale {
		title = "Sale reported on your watched product"
		body = fmt.Sprintf(
			"%s is reported to be on sale $%.2f",
			product.Name,
			new_price.Amount,
		)
		if new_price.ExpiresAt != nil {
			body += fmt.Sprintf(". Valid until %s", new_price.ExpiresAt.Format("January 1"))
		}
		if new_price.Condition != nil {
			body += fmt.Sprint("*", *new_price.Condition)
		}
	} else if new_price.Amount > old_price.Amount {
		title = "Price increase reported on your watched product"
		body = fmt.Sprintf(
			"%s is reported to have increased from $%.2f to $%.2f",
			product.Name,
			old_price.Amount,
			new_price.Amount,
		)
	} else if new_price.Amount < old_price.Amount {
		title = "Price dropped on your watched product"
		body = fmt.Sprintf(
			"%s is reported to have decreased from $%.2f to $%.2f",
			product.Name,
			old_price.Amount,
			new_price.Amount,
		)
	} else {
		// Price was never changed. So skip notifications
		return []expo.PushResponse{}, nil
	}

	notifications := []expo.PushMessage{}
	for _, user := range users {
		if user.ExpoPushToken == nil {
			continue
		}
		notifications = append(notifications, expo.PushMessage{
			To: []expo.ExponentPushToken{expo.ExponentPushToken(*user.ExpoPushToken)},
			Badge: 0,
			Title: title,
			Body: body,
			Data: data,
		})
	}
	res, err = s.ExpoPushClient.PublishMultiple(notifications)
	if err != nil {
		s.CreatePushNotificationEntry(ctx, notifications, expo.PushResponse{
			Status: "error",
			Message: err.Error(),
		})
		return []expo.PushResponse{}, err
	}

	s.CreatePushNotificationEntry(ctx, notifications, res)
	return res, nil
}

func (s Service) PaginatedPrices(
	ctx context.Context,
	product_id int64,
	stock_id int64,
	paginator_input gmodel.PaginatorInput,
	filters *gmodel.PriceHistoryFilter,
) (res gmodel.PaginatedPriceHistory, err error) {
	created_by_user, updated_by_user, user_cols := s.CreatedAndUpdatedUserTable()
	tables := table.Price.
		LEFT_JOIN(created_by_user, created_by_user.ID.EQ(table.Price.CreatedByID)).
		LEFT_JOIN(updated_by_user, updated_by_user.ID.EQ(table.Price.UpdatedByID))
	where_clause := postgres.AND(
		table.Price.ProductID.EQ(postgres.Int(product_id)),
		table.Price.StockID.EQ(postgres.Int(stock_id)),
	)
	sql_paginator, err := s.Paginate(ctx, paginator_input, tables, table.Price.ID, where_clause)
	if err != nil {
		return gmodel.PaginatedPriceHistory{
			Prices: []*gmodel.Price{},
			Paginator: &gmodel.Paginator{},
		}, nil
	}

	order_by := table.Price.CreatedAt.DESC()
	if filters != nil {
		if filters.OrderBy != nil {
			switch *filters.OrderBy {
				case gmodel.OrderByTypeAsc:
					order_by = table.Price.ID.ASC()
				case gmodel.OrderByTypeDesc:
					order_by = table.Price.ID.DESC()
				default:
					order_by = table.Price.ID.DESC()
			}
		}
	}
	qb := table.Price.
		SELECT(
			table.Price.AllColumns,
			user_cols...,
		).
		FROM(tables).
		WHERE(where_clause).
		ORDER_BY(order_by).
		LIMIT(int64(sql_paginator.Limit)).
		OFFSET(int64(sql_paginator.Offset))
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &res.Prices); err != nil {
		return gmodel.PaginatedPriceHistory{}, err
	}

	res.Paginator = &sql_paginator.Paginator
	return res, nil
}
