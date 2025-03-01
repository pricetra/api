package gresolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.44

import (
	"context"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph"
	"github.com/pricetra/api/graph/gmodel"
)

// GetAllCountries is the resolver for the getAllCountries field.
func (r *queryResolver) GetAllCountries(ctx context.Context) ([]*gmodel.Country, error) {
	cities_col := fmt.Sprintf("%s.%s", table.AdministrativeDivision.TableName(), table.AdministrativeDivision.Cities.Name())
	qb := table.Country.
		SELECT(
			table.Country.AllColumns,
			table.AdministrativeDivision.AdministrativeDivision,
			postgres.
				Raw(fmt.Sprintf("array_to_json(%s)::TEXT", cities_col)).
				AS(cities_col),
			table.Currency.AllColumns,
		).
		FROM(
			table.Country.
				INNER_JOIN(table.AdministrativeDivision, table.AdministrativeDivision.CountryCode.EQ(table.Country.Code)).
				INNER_JOIN(table.Currency, table.Currency.CurrencyCode.EQ(table.Country.Currency)),
		).
		ORDER_BY(
			table.Country.Name.ASC(),
			table.AdministrativeDivision.AdministrativeDivision.ASC(),
		)
	var countries []*gmodel.Country
	if err := qb.QueryContext(ctx, r.AppContext.DB, &countries); err != nil {
		return nil, fmt.Errorf("could not query country data")
	}
	return countries, nil
}

// Query returns graph.QueryResolver implementation.
func (r *Resolver) Query() graph.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
