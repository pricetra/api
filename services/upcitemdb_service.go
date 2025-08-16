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
	"github.com/pricetra/api/utils"
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
		category, _ = service.FindCategoryByExactName(ctx, ob_category)
	}
	var weight *string
	if ob.Weight != nil {
		weight = utils.ParseWeight(*ob.Weight)
	}
	return gmodel.CreateProduct{
		Name: ob.Title,
		ImageURL: &image,
		Description: ob.Description,
		Brand: ob.Brand,
		Code: barcode,
		Color: ob.Color,
		Model: ob.Model,
		CategoryID: category.ID,
		Weight: weight,
		LowestRecordedPrice: ob.LowestRecordedPrice,
		HighestRecordedPrice: ob.HighestRecordedPrice,
	}
}

func (s Service) FetchUPCItemdb(ctx context.Context, endpoint string) (result UPCItemDbJsonResult, err error) {
	client := http.Client{}
	url := fmt.Sprintf("%s%s", s.GetUPCItemdbApiUrl(), endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return UPCItemDbJsonResult{}, err
	}

	if s.Tokens.UPCitemdbUserKey != "" {
		req.Header = http.Header{
			"user_key": {s.Tokens.UPCitemdbUserKey},
		}
	}
	res, err := client.Do(req)
	if err != nil {
		return UPCItemDbJsonResult{}, err
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return UPCItemDbJsonResult{}, err
	}

	if result.Code != "OK" {
		return UPCItemDbJsonResult{}, fmt.Errorf(result.Code)
	}
	return result, nil
}

func (s Service) UPCItemDbLookupWithUpcCode(ctx context.Context, upc string) (result UPCItemDbJsonResult, err error) {
	return s.FetchUPCItemdb(ctx, "/lookup?upc=" + upc)
}

func (s Service) UPCItemdbSearch(ctx context.Context, search gmodel.SaveExternalProductInput, offset *int) (result UPCItemDbJsonResult, err error) {
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
	return s.FetchUPCItemdb(ctx, "/search?" + query_params.Encode())
}

func (s Service) UPCItemdbSaveSearchProducts(ctx context.Context, user gmodel.User, search gmodel.SaveExternalProductInput) (res gmodel.SearchResult, err error) {
	source := model.ProductSourceType_Upcitemdb
	offset := 0
	if search.Offset != nil {
		offset = *search.Offset
	}
	for i := 0; i < search.NumPagesToQuery; i++ {
		if i != 0 && offset == 0 {
			fmt.Println("added ", res.Total, ". done.")
			break
		}

		log.Println("iteration: ", i + 1)
		results, err := s.UPCItemdbSearch(ctx, search, &offset)
		if err != nil {
			if err.Error() == "TOO_FAST" {
				log.Println(err.Error(), "waiting 10 seconds")
				i = i - 1
				time.Sleep(10 * time.Second)
				continue
			}
			log.Println(err.Error())
			return res, nil
		}

		offset = results.Offset
		for _, result := range results.Items {
			res.Total += 1

			if result.Upc == "" || result.Brand == "" || result.Title == "" || result.Category == nil || len(result.Images) == 0 {
				res.Failed += 1
				log.Printf("skipping %+v\n", result)
				continue
			}
			log.Println(result.Upc)
			if s.BarcodeExists(ctx, result.Upc) {
				res.Failed += 1
				log.Println("already exists. skipping.")
				continue
			}
			input := result.ToCreateProduct(ctx, s, nil)
			product, err := s.CreateProduct(ctx, user, input, &source)
			if err != nil {
				res.Failed += 1
				log.Printf("could not add product. %s. %+v. %+v", err, result, input)
				continue
			}
			// Upload image...
			if input.ImageURL != nil {
				_, err := s.ImageUrlUpload(ctx, *input.ImageURL, uploader.UploadParams{
					PublicID: product.Code,
					Tags:     []string{"PRODUCT"},
				})
				if err != nil {
					log.Println("could not upload remote product image URL.", err.Error())
				}
			}
			res.Added += 1
		}
		log.Printf("Waiting.... Next offset: %d\n\n", offset)
		time.Sleep(5 * time.Second)
	}
	return res, nil
}
