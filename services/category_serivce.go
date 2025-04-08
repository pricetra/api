package services

import (
	"context"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
	"github.com/pricetra/api/utils"
)

func (s Service) FindCategories(ctx context.Context, depth *int) (categories []gmodel.Category, err error) {
	depth_col_name := fmt.Sprintf("%s.depth", table.Category.TableName())
	path_len_exp := fmt.Sprintf("array_length(%s, 1)", utils.BuildFullTableName(table.Category.Path))
	var where_clause postgres.BoolExpression = nil
	if depth != nil {
		where_clause = postgres.RawBool(fmt.Sprintf("%s > 1", path_len_exp))
	}
	qb := table.Category.SELECT(
			table.Category.AllColumns,
			postgres.RawInt(path_len_exp).AS(depth_col_name),
		).
		FROM(table.Category).
		WHERE(where_clause).
		ORDER_BY(table.Category.ID.ASC())
	fmt.Println(qb.DebugSql())
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &categories)
	return categories, err
}
