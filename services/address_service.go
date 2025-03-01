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
