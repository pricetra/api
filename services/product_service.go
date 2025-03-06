package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
	"github.com/pricetra/api/types"
)

const UPCItemdb_API = "https://api.upcitemdb.com/prod"

func (s Service) CreateProduct(ctx context.Context, input gmodel.CreateProduct) (product gmodel.Product, err error) {
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
		).
		MODEL(input).
		RETURNING(table.Product.AllColumns)

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
	qb := table.Product.
		SELECT(table.Product.AllColumns).
		FROM(table.Product).
		ORDER_BY(table.Product.CreatedAt.DESC())
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &products)
	return products, err
}
