package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
	"github.com/pricetra/api/types"
	"github.com/pricetra/api/utils"
)

func (s Service) CreateProductNutrition(ctx context.Context, product_id int64, input model.ProductNutrition) (res gmodel.ProductNutrition, err error) {
	if input.ProductID == 0 {
		input.ProductID = product_id
	}

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
		MODEL(input).
		RETURNING(table.ProductNutrition.AllColumns)
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &res); err != nil {
		return gmodel.ProductNutrition{}, err
	}
	return res, nil
}

func (s Service) UpdateProductNutrition(ctx context.Context, product_id int64, input model.ProductNutrition) (res gmodel.ProductNutrition, err error) {
	if input.ProductID == 0 {
		input.ProductID = product_id
	}

	qb := table.ProductNutrition.
		UPDATE(
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
		MODEL(input).
		WHERE(table.ProductNutrition.ProductID.EQ(postgres.Int(product_id))).
		RETURNING(table.ProductNutrition.AllColumns)
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &res); err != nil {
		return gmodel.ProductNutrition{}, err
	}
	return res, nil
}

func (s Service) FindProductNutrition(ctx context.Context, product_id int64) (p gmodel.ProductNutrition, err error) {
	qb := table.ProductNutrition.
		SELECT(table.ProductNutrition.AllColumns).
		WHERE(table.ProductNutrition.ProductID.EQ(postgres.Int(product_id))).
		LIMIT(1)
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &p); err != nil {
		return gmodel.ProductNutrition{}, err
	}
	return p, nil
}

func (s Service) FetchOpenFoodFactsDataAndMapToProductNutrition(ctx context.Context, upc string) (model.ProductNutrition, error) {
	product_facts, err := s.OpenFoodFactsClient.Product(upc)
	if err != nil {
		return model.ProductNutrition{}, err
	}

	var serving_weight_unit, serving_size *string
	var serving_weight_value *float64
	if serving_weight_comps, err := utils.ParseWeightIntoStruct(product_facts.ServingSize); err == nil {
		serving_quantity, err := product_facts.ServingQuantity.Float64()
		if err != nil || serving_weight_comps.Weight != serving_quantity {
			return model.ProductNutrition{}, fmt.Errorf("serving size parsing mismatch");
		}
		serving_weight_value = &serving_quantity
		serving_weight_unit = &serving_weight_comps.WeightType
		serving_size = &product_facts.ServingSize
	}

	var nutriments *string
	if nutriment_json, err := json.Marshal(product_facts.Nutriments); err == nil {
		nutriment_json_str := string(nutriment_json)
		nutriments = &nutriment_json_str
	}

	ingredients := make([]string, len(product_facts.Ingredients))
	var vegan, vegetarian, gluten_free, lactose_free, halal, kosher *bool
	for i, ingredient := range product_facts.Ingredients {
		ingredients[i] = ingredient.Text

		if ingredient.Vegan != "" {
			vegan = types.IngredientLabelType(ingredient.Vegan).ToBool()
		}
		if ingredient.Vegetarian != "" {
			vegetarian = types.IngredientLabelType(ingredient.Vegetarian).ToBool()
		}

		// TODO: check for gluten free, lactose free, halal, and kosher
	}
	ingredients_pg_array := utils.ToPostgresArray(ingredients)

	return model.ProductNutrition{
		IngredientText: &product_facts.IngredientsText,
		IngredientList: &ingredients_pg_array,
		Nutriments: nutriments,
		ServingSizeValue: serving_weight_value,
		ServingSizeUnit: serving_weight_unit,
		ServingSize: serving_size,
		OpenfoodfactsUpdatedAt: product_facts.LastModifiedTime.Time,
		Vegan: vegan,
		Vegetarian: vegetarian,
		GlutenFree: gluten_free,
		LactoseFree: lactose_free,
		Halal: halal,
		Kosher: kosher,
	}, nil
}

// Given a product, fetch its nutrition data from OpenFoodFacts and store it in the database.
// If the product nutrition data already exists, return the row directly
func (s Service) ProcessOpenFoodFactsData(ctx context.Context, product gmodel.Product) (product_nutrition gmodel.ProductNutrition, err error) {
	if pn, err := s.FindProductNutrition(ctx, product.ID); err == nil {
		return pn, nil
	}

	product_nutrition_model, err := s.FetchOpenFoodFactsDataAndMapToProductNutrition(ctx, product.Code)
	if err != nil {
		return gmodel.ProductNutrition{}, err
	}
	return s.CreateProductNutrition(ctx, product.ID, product_nutrition_model)
}

func (s Service) UpdateOpenFoodFactsDataForProduct(ctx context.Context, product gmodel.Product) (product_nutrition gmodel.ProductNutrition, err error) {
	product_nutrition_model, err := s.FetchOpenFoodFactsDataAndMapToProductNutrition(ctx, product.Code)
	if err != nil {
		return gmodel.ProductNutrition{}, err
	}
	return s.UpdateProductNutrition(ctx, product.ID, product_nutrition_model)
}
