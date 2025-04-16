package services

import (
	"context"
	"encoding/json"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
)

func toJsonString(v any) (*string, error) {
	if v == nil {
		return nil, nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	x := string(data)
	return &x, nil
}

func (s Service) CreateProductBilling(
	ctx context.Context,
	user gmodel.User,
	billing_type model.ProductBillingType,
	product gmodel.Product,
	new_data any,
	old_data any,
) (res model.ProductBilling, err error) {
	db := s.DbOrTxQueryable()
	cur_rate := model.ProductBillingRate{}
	rate_qb := table.ProductBillingRate.
		SELECT(table.ProductBillingRate.AllColumns).
		FROM(table.ProductBillingRate).
		WHERE(table.ProductBillingRate.Type.EQ(postgres.NewEnumValue(billing_type.String()))).
		LIMIT(1)
	if err := rate_qb.QueryContext(ctx, db, &cur_rate); err != nil {
		return model.ProductBilling{}, err
	}

	new_data_json, err := toJsonString(new_data)
	if err != nil {
		return model.ProductBilling{}, err
	}
	old_data_json, err := toJsonString(old_data)
	if err != nil {
		return model.ProductBilling{}, err
	}

	qb := table.ProductBilling.
		INSERT(
			table.ProductBilling.UserID,
			table.ProductBilling.ProductID,
			table.ProductBilling.Rate,
			table.ProductBilling.BillingRateType,
			table.ProductBilling.NewData,
			table.ProductBilling.OldData,
		).MODEL(model.ProductBilling{
			UserID: user.ID,
			ProductID: product.ID,
			Rate: cur_rate.Rate,
			BillingRateType: billing_type,
			NewData: new_data_json,
			OldData: old_data_json,
		}).RETURNING(table.ProductBilling.AllColumns)
	if err := qb.QueryContext(ctx, db, &res); err != nil {
		return model.ProductBilling{}, err
	}
	return res, nil
}

func (s Service) FindProductBillingByUser(ctx context.Context, paginator_input gmodel.PaginatorInput, user gmodel.User) (res gmodel.PaginatedProductBilling, err error) {
	where_clause := table.ProductBilling.UserID.EQ(postgres.Int(user.ID))
	my_table := table.ProductBilling.
		INNER_JOIN(table.User, table.User.ID.EQ(table.ProductBilling.UserID)).
		INNER_JOIN(table.Product, table.Product.ID.EQ(table.ProductBilling.ProductID)).
		INNER_JOIN(table.Category, table.Category.ID.EQ(table.Product.CategoryID))

	paginator, err := s.Paginate(ctx, paginator_input, my_table, table.ProductBilling.ID, where_clause)
	if err != nil {
		return gmodel.PaginatedProductBilling{}, nil
	}

	qb := table.ProductBilling.
		SELECT(
			table.ProductBilling.ID,
			table.ProductBilling.ProductID,
			table.ProductBilling.UserID,
			table.ProductBilling.CreatedAt,
			table.ProductBilling.Rate,
			table.ProductBilling.PaidAt,
			table.ProductBilling.BillingRateType,
			table.User.ID,
			table.User.Name,
			table.User.Avatar,
			table.User.Active,
			table.Product.AllColumns,
			table.Category.AllColumns,
		).
		FROM(my_table).
		WHERE(where_clause).
		ORDER_BY(table.ProductBilling.CreatedAt).
		LIMIT(int64(paginator.Limit)).
		OFFSET(int64(paginator.Offset))
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &res.Data); err != nil {
		return gmodel.PaginatedProductBilling{}, err
	}
	res.Paginator = &paginator.Paginator
	return res, nil
}
