package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/graph/gmodel"
)

// Docs: https://www.upcitemdb.com/api/explorer#!/lookup/get_trial_lookup
const UPCItemdb_API_ROOT = "https://api.upcitemdb.com/prod"

func (s Service) GetUPCItemdbApiUrl() string {
	if s.Tokens.UPCitemdbUserKey == "" {
		return UPCItemdb_API_ROOT + "/trial"
	}
	return UPCItemdb_API_ROOT + "/v1"
}

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

func (Service) FetchUPCItemdb(endpoint string) (result UPCItemDbJsonResult, err error) {
	res, err := http.Get(fmt.Sprintf("%s%s", UPCItemdb_API, endpoint))
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

func (s Service) UPCItemDbLookupWithUpcCode(upc string) (result UPCItemDbJsonResult, err error) {
	return s.FetchUPCItemdb(fmt.Sprintf("/lookup?upc=%s", upc))
}

func (s Service) UPCItemdbSearch(search gmodel.SaveExternalProductInput, offset *int) (result UPCItemDbJsonResult, err error) {
	// See https://www.upcitemdb.com/api/explorer#!/search/get_trial_search
	query_params := url.Values{}
	query_params.Add("s", search.Search)
	if offset != nil {
		query_params.Add("offset", fmt.Sprint(*offset))
	}
	if search.Brand != nil {
		query_params.Add("brand", *search.Brand)
	}
	if search.Category != nil {
		query_params.Add("category", *search.Category)
	}
	return s.FetchUPCItemdb("/search?" + query_params.Encode())
}

func (s Service) UPCItemdbSaveSearchProducts(ctx context.Context, user gmodel.User, search gmodel.SaveExternalProductInput) (products []*gmodel.Product, err error) {
	source := model.ProductSourceType_Upcitemdb
	offset := 0
	for i := 0; i < search.NumPagesToQuery; i++ {
		if i != 0 && offset == 0 {
			break
		}

		log.Println("Iteration: ", i)
		results, err := s.UPCItemdbSearch(search, &offset)
		if err != nil {
			return nil, err
		}

		log.Printf("%+v\n", results)

		offset = results.Offset
		for _, result := range results.Items {
			if result.Upc == "" || result.Brand == "" || result.Title == "" || len(result.Images) == 0 {
				log.Printf("skipping %+v\n", result)
				continue
			}
			if s.BarcodeExists(ctx, result.Upc) {
				log.Println(result.Upc, "already exists. skipping.")
				continue
			}
			input := result.ToCreateProduct(ctx, s, nil)
			input.Name = strings.ToTitle(input.Name)
			product, err := s.CreateProduct(ctx, user, input, &source)
			if err != nil {
				log.Println("could not add product", err)
				continue
			}
			// Upload image...
			if input.Image != nil {
				_, err := s.ImageUrlUpload(ctx, *input.Image, uploader.UploadParams{
					PublicID: product.Code,
					Tags:     []string{"PRODUCT"},
				})
				if err != nil {
					log.Println("could not upload remote product image URL.", err.Error())
				}
			}
			products = append(products, &product)
		}
		time.Sleep(10 * time.Second)
	}
	return products, nil
}
