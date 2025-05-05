package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
	"googlemaps.github.io/maps"
)

func (s Service) CreateBranchFromFullAddress(ctx context.Context, user gmodel.User, store_id int64, full_address string) (branch gmodel.Branch, err error) {
	store, err := s.FindStore(ctx, store_id)
	if err != nil {
		return branch, fmt.Errorf("store id is invalid")
	}

	data, err := s.GoogleMapsClient.Geocode(ctx, &maps.GeocodingRequest{
		Address: full_address,
	})
	if err != nil || len(data) == 0 {
		return gmodel.Branch{}, fmt.Errorf("could not parse raw address")
	}

	res := data[0];
	address := gmodel.CreateAddress{
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
	for _, component := range res.AddressComponents {
		switch component.Types[0] {
		case "locality":
			address.City = component.LongName
		case "administrative_area_level_1":
			address.AdministrativeDivision = component.LongName
		case "country":
			address.CountryCode = component.ShortName
		case "postal_code":
			zip_code, err := strconv.Atoi(component.LongName)
			if err != nil {
				return gmodel.Branch{}, err
			}
			address.ZipCode = zip_code
		}
	}

	return s.CreateBranch(ctx, user, gmodel.CreateBranch{
		Name: fmt.Sprintf("%s %s", store.Name, address.City),
		Address: &address,
		StoreID: store_id,
	})
}

func (s Service) CreateBranch(ctx context.Context, user gmodel.User, input gmodel.CreateBranch) (branch gmodel.Branch, err error) {
	if err := s.StructValidator.StructCtx(ctx, input); err != nil {
		return branch, err
	}
	if !s.StoreExists(ctx, input.StoreID) {
		return branch, fmt.Errorf("store id is invalid")
	}

	s.TX, err = s.DB.BeginTx(ctx, nil)
	if err != nil {
		return branch, err
	}
	defer s.TX.Rollback()

	address, err := s.CreateAddress(ctx, &user, *input.Address)
	if err != nil {
		return branch, err
	}

	qb := table.Branch.INSERT(
		table.Branch.Name,
		table.Branch.AddressID,
		table.Branch.StoreID,
		table.Branch.CreatedByID,
		table.Branch.UpdatedByID,
	).MODEL(model.Branch{
		Name: input.Name,
		AddressID: address.ID,
		StoreID: input.StoreID,
		CreatedByID: &user.ID,
		UpdatedByID: &user.ID,
	}).
	RETURNING(table.Branch.AllColumns)
	
	if err := qb.QueryContext(ctx, s.TX, &branch); err != nil {
		return branch, err
	}

	if err := s.TX.Commit(); err != nil {
		return branch, err
	}
	branch.Address = &address
	return branch, err
}

func (s Service) FindBranchById(ctx context.Context, id int64) (branch gmodel.Branch, err error) {
	created_user_table, updated_user_table, user_cols := s.CreatedAndUpdatedUserTable()
	columns := []postgres.Projection{
		table.Address.AllColumns,
		table.Country.Name,
	}
	columns = append(columns, user_cols...)
	qb := table.Branch.
		SELECT(
			table.Branch.AllColumns,
			columns...,
		).
		FROM(
			table.Branch.
				INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
				INNER_JOIN(table.Country, table.Country.Code.EQ(table.Address.CountryCode)).
				LEFT_JOIN(created_user_table, created_user_table.ID.EQ(table.Branch.CreatedByID)).
				LEFT_JOIN(updated_user_table, updated_user_table.ID.EQ(table.Branch.UpdatedByID)),
		).
		WHERE(table.Branch.ID.EQ(postgres.Int(id))).
		LIMIT(1)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &branch)
	return branch, err
}

func (s Service) FindBranchesByStoreId(ctx context.Context, store_id int64) (branches []gmodel.Branch, err error) {
	created_user_table, updated_user_table, user_cols := s.CreatedAndUpdatedUserTable()
	columns := []postgres.Projection{
		table.Address.AllColumns,
		table.Country.Name,
	}
	columns = append(columns, user_cols...)
	qb := table.Branch.
		SELECT(
			table.Branch.AllColumns,
			columns...,
		).
		FROM(
			table.Branch.
				INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
				INNER_JOIN(table.Country, table.Country.Code.EQ(table.Address.CountryCode)).
				LEFT_JOIN(created_user_table, created_user_table.ID.EQ(table.Branch.CreatedByID)).
				LEFT_JOIN(updated_user_table, updated_user_table.ID.EQ(table.Branch.UpdatedByID)),
		).
		WHERE(table.Branch.StoreID.EQ(postgres.Int(store_id))).
		ORDER_BY(table.Branch.CreatedAt.DESC())
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &branches)
	return branches, err
}

func (s Service) FindBranchByBranchIdAndStoreId(ctx context.Context, branch_id int64, store_id int64) (branch gmodel.Branch, err error) {
	created_user_table, updated_user_table, user_cols := s.CreatedAndUpdatedUserTable()
	columns := []postgres.Projection{
		table.Address.AllColumns,
		table.Country.Name,
	}
	columns = append(columns, user_cols...)
	qb := table.Branch.
		SELECT(
			table.Branch.AllColumns,
			columns...,
		).
		FROM(
			table.Branch.
				INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
				INNER_JOIN(table.Country, table.Country.Code.EQ(table.Address.CountryCode)).
				LEFT_JOIN(created_user_table, created_user_table.ID.EQ(table.Branch.CreatedByID)).
				LEFT_JOIN(updated_user_table, updated_user_table.ID.EQ(table.Branch.UpdatedByID)),
		).
		WHERE(
			table.Branch.StoreID.EQ(postgres.Int(store_id)).
			AND(table.Branch.ID.EQ(postgres.Int(branch_id))),
		).
		LIMIT(1)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &branch)
	return branch, err
}

func (s Service) FindBranchesByCoordinates(ctx context.Context, lat float64, lon float64, radius_meters int) (branches []gmodel.Branch, err error) {
	created_user_table, updated_user_table, user_cols := s.CreatedAndUpdatedUserTable()
	distance_cols := s.GetDistanceCols(lat, lon, radius_meters)
	columns := []postgres.Projection{
		table.Store.AllColumns,
		table.Address.AllColumns,
		table.Country.Name,
		distance_cols.DistanceColumn,
	}
	columns = append(columns, user_cols...)
	qb := table.Branch.
		SELECT(
			table.Branch.AllColumns,
			columns...,
		).
		FROM(
			table.Branch.
				INNER_JOIN(table.Store, table.Store.ID.EQ(table.Branch.StoreID)).
				INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
				INNER_JOIN(table.Country, table.Country.Code.EQ(table.Address.CountryCode)).
				LEFT_JOIN(created_user_table, created_user_table.ID.EQ(table.Branch.CreatedByID)).
				LEFT_JOIN(updated_user_table, updated_user_table.ID.EQ(table.Branch.UpdatedByID)),
		).
		WHERE(distance_cols.DistanceWhereClauseWithRadius).
		ORDER_BY(
			postgres.FloatColumn(distance_cols.DistanceColumnName).ASC(),
			table.Branch.ID.ASC(),
		)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &branches)
	return branches, err
}
