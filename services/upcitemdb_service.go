package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pricetra/api/graph/gmodel"
)

type UPCItemDbJsonResultItem struct {
	Ean string `json:"ean"`
	Title string `json:"title,omitempty"`
	Upc string `json:"upc"`
	Gtin *string `json:"gtin,omitempty"`
	Asin *string `json:"asin,omitempty"`
	Description string `json:"description,omitempty"`
	Brand string `json:"brand,omitempty"`
	Model *string `json:"model,omitempty"`
	Color *string `json:"color,omitempty"`
	Weight *string `json:"weight,omitempty"`
	Category *string `json:"category,omitempty"`
	LowestRecordedPrice *float64 `json:"lowest_recorded_price,omitempty"`
	HighestRecordedPrice *float64 `json:"highest_recorded_price,omitempty"`
	Images []string `json:"images,omitempty"`
	Offers []any `json:"offers,omitempty"`
	Elid *string `json:"elid,omitempty"`
}

type UPCItemDbJsonResult struct {
	Code string `json:"code"`
	Total int `json:"total"`
	Offset int `json:"offset"`
	Items []UPCItemDbJsonResultItem `json:"items"`
	Message *string `json:"message,omitempty"`
}

func (ob UPCItemDbJsonResultItem) ToCreateProduct(ctx context.Context, service Service, upc *string) gmodel.CreateProduct {
	image := ""
	if len(ob.Images) > 0 {
		image = ob.Images[0]
	}
	barcode := ob.Upc
	if upc != nil {
		barcode = *upc
	}
	ob_category := "Uncategorized"
	if ob.Category != nil && strings.TrimSpace(*ob.Category) != "" {
		ob_category = *ob.Category
	}
	category, err := service.CategoryRecursiveInsert(ctx, ob_category)
	if err != nil {
		panic(err)
	}
	return gmodel.CreateProduct{
		Name: ob.Title,
		Image: &image,
		Description: ob.Description,
		Brand: ob.Brand,
		Code: barcode,
		Color: ob.Color,
		Model: ob.Model,
		CategoryID: category.ID,
		Weight: ob.Weight,
		LowestRecordedPrice: ob.LowestRecordedPrice,
		HighestRecordedPrice: ob.HighestRecordedPrice,
	}
}

func (s Service) UPCItemDbLookupWithUpcCode(upc string) (result UPCItemDbJsonResult, err error) {
	res, err := http.Get(fmt.Sprintf("%s/trial/lookup?upc=%s", UPCItemdb_API, upc))
	if err != nil {
		return result, err
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return UPCItemDbJsonResult{}, err
	}

	if result.Code != "OK" {
		message := ""
		if result.Message != nil {
			message = *result.Message
		}
		return UPCItemDbJsonResult{}, fmt.Errorf("%s - %s", result.Code, message)
	}
	return result, nil
}
