package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	if s.BarcodeExists(ctx, input.Code) {
		return product, fmt.Errorf("barcode already exists in the database. please use the update method")
	}

	var source_val model.ProductSourceType
	if source == nil {
		source_val = model.ProductSourceType_Pricetra
	} else {
		source_val = *source
	}

	if input.Image == nil {
		if input.ImageFile != nil {
			image := fmt.Sprintf("%s/%s", CLOUDINARY_UPLOAD_BASE, input.Code)
			input.Image = &image
		} else {
			image := ""
			input.Image = &image
		}
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

func (s Service) BarcodeExists(ctx context.Context, barcode string) bool {
	qb := table.Product.
		SELECT(table.Product.Code.AS("code")).
		FROM(table.Product).
		WHERE(table.Product.Code.EQ(postgres.String(barcode))).
		LIMIT(1)
	var product struct{
		Code string
	}
	err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &product)
	return err == nil
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

func (s Service) UpdateProductById(ctx context.Context, user gmodel.User, id int64, input gmodel.UpdateProduct) (updated_product gmodel.Product, err error) {
	if err := s.StructValidator.StructCtx(ctx, input); err != nil {
		return updated_product, err
	}
	
	product, err := s.FindProductById(ctx, id)
	if err != nil {
		return updated_product, fmt.Errorf("product with id does not exist")
	}

	cols := postgres.ColumnList{}
	code := product.Code
	if input.Name != nil && *input.Name != product.Name {
		cols = append(cols, table.Product.Name)
	}
	if input.Description != nil && *input.Description != product.Description {
		cols = append(cols, table.Product.Description)
	}
	if input.URL != nil && product.URL != nil && *input.URL != *product.URL {
		cols = append(cols, table.Product.URL)
	}
	if input.Brand != nil && *input.Brand != product.Brand {
		cols = append(cols, table.Product.Brand)
	}
	if input.Code != nil && *input.Code != product.Code {
		if s.BarcodeExists(ctx, *input.Code) {
			return updated_product, fmt.Errorf("new barcode is already in use")
		}
		code = *input.Code
		cols = append(cols, table.Product.Code)
	}
	if input.Color != nil && product.Color != nil && *input.Color != *product.Color {
		cols = append(cols, table.Product.Color)
	}
	if input.Model != nil && product.Model != nil && *input.Model != *product.Model {
		cols = append(cols, table.Product.Model)
	}
	if input.Category != nil && product.Category != nil && *input.Category != *product.Category {
		cols = append(cols, table.Product.Category)
	}
	if input.Weight != nil && product.Weight != nil && *input.Weight != *product.Weight {
		cols = append(cols, table.Product.Weight)
	}
	if input.LowestRecordedPrice != nil && product.LowestRecordedPrice != nil && *input.LowestRecordedPrice != *product.LowestRecordedPrice {
		cols = append(cols, table.Product.LowestRecordedPrice)
	}
	if input.HighestRecordedPrice != nil && product.HighestRecordedPrice != nil && *input.HighestRecordedPrice != *product.HighestRecordedPrice {
		cols = append(cols, table.Product.HighestRecordedPrice)
	}
	if input.ImageFile != nil {
		image := fmt.Sprintf("%s/%s", CLOUDINARY_UPLOAD_BASE, code)
		input.Image = &image
		cols = append(cols, table.Product.Image)
	}

	if len(cols) == 0 {
		return product, nil
	}
	cols = append(cols, table.Product.UpdatedByID, table.Product.UpdatedAt)
	qb := table.Product.
		UPDATE(cols).
		MODEL(struct{
			gmodel.UpdateProduct
			UpdatedById *int64
			UpdatedAt time.Time
		}{
			UpdateProduct: input,
			UpdatedById: &user.ID,
			UpdatedAt: time.Now(),
		}).
		RETURNING(table.Product.AllColumns)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &updated_product)
	if err != nil {
		return updated_product, err
	}
	return updated_product, nil
}
