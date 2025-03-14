package types

import "github.com/pricetra/api/graph/gmodel"

type SendGridTemplates struct {
	EmailVerification string `json:"emailVerification"`
}

type SendGridTokens struct {
	ApiKey string `json:"apiKey"`
	RecoveryCode string `json:"recoveryCode,omitempty"`
	ApiKeyId string `json:"apiKeyId,omitempty"`
	Templates SendGridTemplates `json:"templates,omitempty"`
}

type CloudinaryTokens struct {
	ApiKey string `json:"apiKey"`
	ApiSecret string `json:"apiSecret"`
	CloudName string `json:"cloudName"`
}

type Tokens struct {
	JwtKey string `json:"jwtKey"`
	SendGrid SendGridTokens `json:"sendGrid,omitempty"`
	Cloudinary CloudinaryTokens `json:"cloudinary"`
	// To add other tokens create a struct and add them here,
	// make sure to also update ./tokens.json
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

func (ob UPCItemDbJsonResultItem) ToCreateProduct(upc *string) gmodel.CreateProduct {
	image := ""
	if len(ob.Images) > 0 {
		image = ob.Images[0]
	}
	barcode := ob.Upc
	if upc != nil {
		barcode = *upc
	}
	return gmodel.CreateProduct{
		Name: ob.Title,
		Image: &image,
		Description: ob.Description,
		Brand: ob.Brand,
		Code: barcode,
		Color: ob.Color,
		Model: ob.Model,
		Category: ob.Category,
		Weight: ob.Weight,
		LowestRecordedPrice: ob.LowestRecordedPrice,
		HighestRecordedPrice: ob.HighestRecordedPrice,
	}
}

type UPCItemDbJsonResult struct {
	Code string `json:"code"`
	Total int `json:"total"`
	Offset int `json:"offset"`
	Items []UPCItemDbJsonResultItem `json:"items"`
	Message *string `json:"message,omitempty"`
}
