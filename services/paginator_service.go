package services

import (
	"context"
	"fmt"
	"math"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/graph/gmodel"
)

type SqlPaginator struct {
	gmodel.Paginator
	Offset int
}

func (s Service) Paginate(
	ctx context.Context,
	paginator_input gmodel.PaginatorInput,
	sql_table postgres.ReadableTable,
	id_column postgres.Column,
	where_clause postgres.BoolExpression,
	group_clauses ...postgres.GroupByClause,
) (SqlPaginator, error) {
	if err := s.StructValidator.StructCtx(ctx, paginator_input); err != nil {
		return SqlPaginator{}, err
	}

	db := s.DbOrTxQueryable()
	total_qb := sql_table.
		SELECT(postgres.COUNT(id_column).AS("total")).
		FROM(sql_table).
		WHERE(where_clause)
	if len(group_clauses) > 0 {
		total_qb = total_qb.GROUP_BY(group_clauses...)
	}

	var p_total struct{ Total int }
	if err := total_qb.QueryContext(ctx, db, &p_total); err != nil {
		return SqlPaginator{}, err
	}

	offset := (paginator_input.Page - 1) * paginator_input.Limit
	num_pages := int(math.Ceil(float64(p_total.Total) / float64(paginator_input.Limit)))

	if paginator_input.Page > num_pages {
		return SqlPaginator{}, fmt.Errorf("page cannot be exceed total pages amount")
	}

	paginator := SqlPaginator{
		Paginator: gmodel.Paginator{
			Page: paginator_input.Page,
			Total: p_total.Total,
			Limit: paginator_input.Limit,
			NumPages: num_pages,
		},
		Offset: offset,
	}
	if paginator_input.Page > 1 {
		val := paginator_input.Page - 1
		paginator.Prev = &val
	}
	if paginator_input.Page < num_pages {
		val := paginator_input.Page + 1
		paginator.Next = &val
	}
	return paginator, nil
}
