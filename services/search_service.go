package services

import (
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/utils"
)

type FullTextSearchComponents struct {
	RankColumn postgres.Projection // ts_rank column
	WhereClause postgres.BoolExpression // where clause with tsquery
	OrderByClause postgres.ColumnFloat // float column using the rank column
}

func (s Service) BuildFullTextSearchQueryComponents(search_vector_col postgres.ColumnString, query string) (res FullTextSearchComponents) {
	rank_col := "rank"
	search_vector_col_name := utils.BuildFullTableName(search_vector_col)
	args := postgres.RawArgs{"$query": query}

	// Rank column
	res.RankColumn = postgres.RawFloat(
		fmt.Sprintf(
			"ts_rank(%s, plainto_tsquery('english', $query::TEXT))",
			search_vector_col_name,
		),
		args,
	).AS(rank_col)

	// Where clause with tsquery
	res.WhereClause = postgres.RawBool(
		fmt.Sprintf("%s @@ plainto_tsquery('english', $query::TEXT)", search_vector_col_name),
		args,
	)

	// Order by
	res.OrderByClause = postgres.FloatColumn(rank_col)
	return res
}
