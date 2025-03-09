package services

import (
	"context"

	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
)

func (s Service) CreateCompany(ctx context.Context, user gmodel.User, input gmodel.CreateCompany) (company gmodel.Company, err error) {
	if err := s.StructValidator.StructCtx(ctx, input); err != nil {
		return company, err
	}

	qb := table.Company.
		INSERT(
			table.Company.AllColumns.Except(
				table.Company.ID,
				table.Company.CreatedAt,
				table.Company.UpdatedAt,
			),
		).
		MODEL(struct{
			gmodel.CreateCompany
			CreatedByID *int64
			UpdatedByID *int64
		}{
			CreateCompany: input,
			CreatedByID: &user.ID,
			UpdatedByID: &user.ID,
		}).
		RETURNING(table.Company.AllColumns)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &company)
	return company, err
}
