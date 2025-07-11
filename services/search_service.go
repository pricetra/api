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
	rank_col := utils.BuildFullTableNameHyphen(search_vector_col) + "_rank"
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

func (s Service) CreateSearchHistoryEntry(ctx context.Context, search_term string, user *gmodel.User) (search_history gmodel.SearchHistory, err error) {
	var user_id *int64
	if user != nil {
		user_id = &user.ID
	}
	qb := table.SearchHistory.
		INSERT(
			table.SearchHistory.SearchTerm,
			table.SearchHistory.UserID,
		).
		MODEL(model.SearchHistory{
			SearchTerm: search_term,
			UserID: user_id,
		}).
		RETURNING(table.SearchHistory.AllColumns)
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &search_history); err != nil {
		return gmodel.SearchHistory{}, err
	}
	return search_history, nil
}
