package services

import (
	"database/sql"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/go-playground/validator/v10"
	"github.com/pricetra/api/types"
	"googlemaps.github.io/maps"

	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
)

type Service struct {
	DB *sql.DB
	TX *sql.Tx
	StructValidator *validator.Validate
	Tokens *types.Tokens
	Cloudinary *cloudinary.Cloudinary
	ExpoPushClient *expo.PushClient
	GoogleMapsClient *maps.Client
	GoogleVisionApiClient *vision.ImageAnnotatorClient
}

// Returns a transaction if present.
// Otherwise returns the Database connection instance as qrm.Queryable
func (s *Service) DbOrTxQueryable() qrm.Queryable {
	if s.TX != nil {
		return s.TX
	}
	return s.DB
}
