package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
	"github.com/pricetra/api/utils"
	"googlemaps.github.io/maps"
)

func (s Service) AddressExists(
	ctx context.Context, 
	lat float64, 
	lon float64,
) bool {
	query_builder := table.Address.
		SELECT(table.Address.ID).
		FROM(table.Address).
		WHERE(
			table.Address.Latitude.EQ(postgres.Float(lat)).
			AND(table.Address.Longitude.EQ(postgres.Float(lon))),
		).LIMIT(1)
	var address struct{
		ID int64 `sql:"primary_key"`
	}
	db := s.DbOrTxQueryable()
	err := query_builder.QueryContext(ctx, db, &address)
	return err == nil
}

func (s Service) FindAddressByCoords(
	ctx context.Context, 
	lat float64, 
	lon float64,
) (address gmodel.Address, err error) {
	qb := table.Address.
		SELECT(table.Address.AllColumns).
		FROM(table.Address).
		WHERE(
			table.Address.Latitude.EQ(postgres.Float(lat)).
			AND(table.Address.Longitude.EQ(postgres.Float(lon))),
		).LIMIT(1)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &address)
	return address, err
}

func (s Service) CreateAddress(ctx context.Context, user *gmodel.User, input gmodel.CreateAddress) (address gmodel.Address, err error) {
	if err = s.StructValidator.StructCtx(ctx, input); err != nil {
		return gmodel.Address{}, err
	}
	var country_code model.CountryCodeAlpha2 = model.CountryCodeAlpha2(input.CountryCode)
	if country_code.Scan(input.CountryCode) != nil {
		return gmodel.Address{}, fmt.Errorf("could not scan country code")
	}
	if s.AddressExists(ctx, input.Latitude, input.Longitude) {
		return gmodel.Address{}, fmt.Errorf("address at location already exists")
	}

	qb := table.Address.INSERT(
		table.Address.Latitude,
		table.Address.Longitude,
		table.Address.MapsLink,
		table.Address.FullAddress,
		table.Address.Street,
		table.Address.City,
		table.Address.AdministrativeDivision,
		table.Address.CountryCode,
		table.Address.ZipCode,
		table.Address.CreatedByID,
		table.Address.UpdatedByID,
	).MODEL(model.Address{
		Latitude: input.Latitude,
		Longitude: input.Longitude,
		MapsLink: input.MapsLink,
		FullAddress: input.FullAddress,
		Street: input.Street,
		City: input.City,
		AdministrativeDivision: input.AdministrativeDivision,
		CountryCode: country_code,
		ZipCode: int32(input.ZipCode),
		CreatedByID: &user.ID,
		UpdatedByID: &user.ID,
	}).RETURNING(table.Address.AllColumns)

	db := s.DbOrTxQueryable()
	if err := qb.QueryContext(ctx, db, &address); err != nil {
		return gmodel.Address{}, err
	}
	return address, nil
}

func (s Service) FindOrCreateAddress(
	ctx context.Context,
	user *gmodel.User,
	input gmodel.CreateAddress,
) (address gmodel.Address, err error) {
	address, err = s.FindAddressByCoords(ctx, input.Latitude, input.Longitude)
	if err == nil {
		return address, nil
	}
	return s.CreateAddress(ctx, user, input)
}

type DistanceColumns struct {
	AddressCoordinatesColumnName string // Column name for "address"."coordinates"
	DistanceColumnName string // "address"."distance"
	DistanceColumn postgres.Projection // Distance from. uses ST_Distance("address"."coordinates", 'SRID=4326;POINT($lat $lon)'::geometry)
	DistanceWhereClauseWithRadius postgres.BoolExpression // Where clause to determine if address is within lat, lon, and radius
}

func (s Service) GetDistanceCols(lat float64, lon float64, radius_meters *int) DistanceColumns {
	var d DistanceColumns
	d.AddressCoordinatesColumnName = utils.BuildFullTableName(table.Address.Coordinates)
	d.DistanceColumnName = fmt.Sprintf("%s.%s", table.Address.Coordinates.TableName(), "distance")
	d.DistanceColumn = postgres.RawString(
		fmt.Sprintf(
			"ST_Distance(%s, 'SRID=4326;POINT(%f %f)'::geometry)",
			d.AddressCoordinatesColumnName,
			lon,
			lat,
		),
	).AS(d.DistanceColumnName)

	if radius_meters == nil {
		d.DistanceWhereClauseWithRadius = postgres.Bool(true)
	} else {
		d.DistanceWhereClauseWithRadius = postgres.RawBool(
			fmt.Sprintf(
				"ST_DWithin(%s, 'POINT(%f %f)'::geometry, %d, TRUE)",
				d.AddressCoordinatesColumnName,
				lon,
				lat,
				*radius_meters,
			),
		)
	}
	return d
}

// Uses Google Maps Geocode API to go from a raw address to gmodel.CreateAddress.
// On success, the address object will have the correct coordinates, and all other values.
func (s Service) FullAddressToCreateAddress(ctx context.Context, full_address string) (address gmodel.CreateAddress, err error) {
	data, err := s.GoogleMapsClient.Geocode(ctx, &maps.GeocodingRequest{
		Address: full_address,
	})
	if err != nil || len(data) == 0 {
		return gmodel.CreateAddress{}, fmt.Errorf("could not parse raw address")
	}

	res := data[0];
	address = gmodel.CreateAddress{
		Latitude: res.Geometry.Location.Lat,
		Longitude: res.Geometry.Location.Lng,
		MapsLink: fmt.Sprintf(
			"https://www.google.com/maps/search/?api=1&query=%f%%2C%f&query_place_id=%s", 
			res.Geometry.Location.Lat, 
			res.Geometry.Location.Lng, 
			res.PlaceID,
		),
		FullAddress: res.FormattedAddress,
	}
	street_components := []string{}
	for _, component := range res.AddressComponents {
		// Address component types: https://developers.google.com/maps/documentation/geocoding/requests-geocoding#address-types
		switch component.Types[0] {
		case "street_number", "route":
			if len(component.ShortName) > 0 {
				street_components = append(street_components, component.ShortName)
			}
		case "locality":
			address.City = component.LongName
		case "administrative_area_level_1":
			address.AdministrativeDivision = component.LongName
		case "country":
			address.CountryCode = component.ShortName
		case "postal_code":
			zip_code, err := strconv.Atoi(component.LongName)
			if err != nil {
				return gmodel.CreateAddress{}, err
			}
			address.ZipCode = zip_code
		}
	}

	if len(street_components) != 0 {
		street := strings.Join(street_components, " ")
		address.Street = &street
	}
	return address, nil
}
