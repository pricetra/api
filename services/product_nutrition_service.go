package services

import (
	"context"

	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
)

func (s Service) CreateProductNutrition(ctx context.Context, product_id int64, input model.ProductNutrition) (res model.ProductNutrition, err error) {
	qb := table.ProductNutrition.
		INSERT(
			table.ProductNutrition.ProductID,
			table.ProductNutrition.IngredientText,
			table.ProductNutrition.IngredientList,
			table.ProductNutrition.Nutriments,
			table.ProductNutrition.ServingSize,
			table.ProductNutrition.ServingSizeValue,
			table.ProductNutrition.ServingSizeUnit,
			table.ProductNutrition.OpenfoodfactsUpdatedAt,
			table.ProductNutrition.Vegan,
			table.ProductNutrition.Vegetarian,
			table.ProductNutrition.GlutenFree,
			table.ProductNutrition.LactoseFree,
			table.ProductNutrition.Halal,
			table.ProductNutrition.Kosher,
		).
		VALUES(input).
		RETURNING(table.ProductNutrition.AllColumns)
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &res); err != nil {
		return model.ProductNutrition{}, err
	}
	return res, nil
}
