package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
	"github.com/pricetra/api/types"
)

const UPCItemdb_API = "https://api.upcitemdb.com/prod"

func (s Service) CreateProduct(ctx context.Context, user gmodel.User, input gmodel.CreateProduct, source *model.ProductSourceType) (product gmodel.Product, err error) {
	if err := s.StructValidator.StructCtx(ctx, input); err != nil {
		return product, err
	}

	var source_val model.ProductSourceType
	if source == nil {
		source_val = model.ProductSourceType_Pricetra
	} else {
		source_val = *source
	}
	qb := table.Product.
		INSERT(
			table.Product.Name,
			table.Product.Image,
			table.Product.Description,
			table.Product.URL,
			table.Product.Brand,
			table.Product.Code,
			table.Product.Color,
			table.Product.Model,
			table.Product.Category,
			table.Product.Weight,
			table.Product.LowestRecordedPrice,
			table.Product.HighestRecordedPrice,
			table.Product.Source,
			table.Product.CreatedByID,
			table.Product.UpdatedByID,
		).
		MODEL(struct{
			gmodel.CreateProduct
			Source model.ProductSourceType
			CreatedByID *int64
			UpdatedByID *int64
		}{
			CreateProduct: input,
			Source: source_val,
			CreatedByID: &user.ID,
			UpdatedByID: &user.ID,
		}).
		RETURNING(table.Product.AllColumns)

	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &product)
	return product, err
}

func (s Service) FindProductById(ctx context.Context, id int64) (product gmodel.Product, err error) {
	qb := table.Product.
		SELECT(table.Product.AllColumns).
		FROM(table.Product).
		WHERE(table.Product.ID.EQ(postgres.Int(id))).
		LIMIT(1)
	
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &product)
	return product, err
}

func (s Service) FindProductWithCode(ctx context.Context, barcode string) (product gmodel.Product, err error) {
	qb := table.Product.
		SELECT(table.Product.AllColumns).
		FROM(table.Product).
		WHERE(table.Product.Code.EQ(postgres.String(barcode))).
		LIMIT(1)
	
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &product)
	return product, err
}

func (s Service) UPCItemDbLookupWithUpcCode(upc string) (result types.UPCItemDbJsonResult, err error) {
	res, err := http.Get(fmt.Sprintf("%s/trial/lookup?upc=%s", UPCItemdb_API, upc))
	if err != nil {
		return result, err
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return types.UPCItemDbJsonResult{}, err
	}

	if result.Code != "OK" {
		message := ""
		if result.Message != nil {
			message = *result.Message
		}
		return types.UPCItemDbJsonResult{}, fmt.Errorf("%s - %s", result.Code, message)
	}
	return result, nil
}

func (s Service) FindAllProducts(ctx context.Context) (products []gmodel.Product, err error) {
	created_user_table, updated_user_table, user_cols := s.CreatedAndUpdatedUserTable()
	qb := table.Product.
		SELECT(table.Product.AllColumns, user_cols...).
		FROM(
			table.Product.
				LEFT_JOIN(created_user_table, created_user_table.ID.EQ(table.Product.CreatedByID)).
				LEFT_JOIN(updated_user_table, updated_user_table.ID.EQ(table.Product.UpdatedByID)),
		).
		ORDER_BY(table.Product.CreatedAt.DESC())
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &products)
	return products, err
}

func (s Service) UpdateProductById(ctx context.Context) {}
