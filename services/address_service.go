package services

import (
	"context"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
)

func (service Service) AddressExists(
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
	db := service.DbOrTxQueryable()
	err := query_builder.QueryContext(ctx, db, &address)
	return err == nil
}

func (service Service) CreateAddress(ctx context.Context, user *gmodel.User, input gmodel.CreateAddress) (gmodel.Address, error) {
	var country_code model.CountryCodeAlpha2 = model.CountryCodeAlpha2(input.CountryCode)
	if country_code.Scan(input.CountryCode) != nil {
		return gmodel.Address{}, fmt.Errorf("could not scan country code")
	}
	if service.AddressExists(ctx, input.Latitude, input.Longitude) {
		return gmodel.Address{}, fmt.Errorf("address at location already exists")
	}

	qb := table.Address.INSERT(
		table.Address.Latitude,
		table.Address.Longitude,
		table.Address.MapsLink,
		table.Address.FullAddress,
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
		City: input.City,
		AdministrativeDivision: input.AdministrativeDivision,
		CountryCode: country_code,
		ZipCode: int32(input.ZipCode),
		CreatedByID: &user.ID,
		UpdatedByID: &user.ID,
	}).RETURNING(table.Address.AllColumns)

	db := service.DbOrTxQueryable()
	var address gmodel.Address
	if err := qb.QueryContext(ctx, db, &address); err != nil {
		return gmodel.Address{}, err
	}
	return address, nil
}

type DistanceColumns struct {
	AddressCoordinatesColumnName string // Column name for "address"."coordinates"
	DistanceColumnName string // "address"."distance"
	DistanceColumn postgres.Projection // Distance from. uses ST_Distance("address"."coordinates", 'SRID=4326;POINT($lat $lon)'::geometry)
	DistanceWhereClauseWithRadius postgres.BoolExpression // Where clause to determine if address is within lat, lon, and radius
}

func (s Service) GetDistanceCols(lat float64, lon float64, radius_meters int) DistanceColumns {
	var d DistanceColumns
	d.AddressCoordinatesColumnName = fmt.Sprintf(
		"%s.%s",
		table.Address.Coordinates.TableName(),
		table.Address.Coordinates.Name(),
	)
	d.DistanceColumnName = fmt.Sprintf("%s.%s", table.Address.Coordinates.TableName(), "distance")
	d.DistanceColumn = postgres.RawString(
		fmt.Sprintf(
			"ST_Distance(%s, 'SRID=4326;POINT(%f %f)'::geometry)",
			d.AddressCoordinatesColumnName,
			lon,
			lat,
		),
	).AS(d.DistanceColumnName)
	d.DistanceWhereClauseWithRadius = postgres.RawBool(
		fmt.Sprintf(
			"ST_DWithin(%s, 'POINT(%f %f)'::geometry, %d, TRUE)",
			d.AddressCoordinatesColumnName,
			lon,
			lat,
			radius_meters,
		),
	)
	return d
}
