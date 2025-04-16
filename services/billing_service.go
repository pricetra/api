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
