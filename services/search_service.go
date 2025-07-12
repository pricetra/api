package services

import (
	"context"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
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

func (s Service) PaginatedSearchHistory(ctx context.Context, user gmodel.User, paginator_input gmodel.PaginatorInput) (res gmodel.PaginatedSearch, err error) {
	where_clause := table.SearchHistory.UserID.EQ(postgres.Int(user.ID))
	sql_paginator, err := s.Paginate(ctx, paginator_input, table.SearchHistory, table.SearchHistory.ID, where_clause)
	if err != nil {
		return gmodel.PaginatedSearch{
			Searches: []*gmodel.SearchHistory{},
			Paginator: &gmodel.Paginator{},
		}, nil
	}
	qb := table.SearchHistory.
		SELECT(
			table.SearchHistory.ID,
			table.SearchHistory.SearchTerm,
			table.SearchHistory.CreatedAt,
		).
		FROM(table.SearchHistory).
		WHERE(where_clause).
		ORDER_BY(table.SearchHistory.ID.DESC()).
		LIMIT(int64(sql_paginator.Limit)).
		OFFSET(int64(sql_paginator.Offset))
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &res.Searches); err != nil {
		return gmodel.PaginatedSearch{}, err
	}
	res.Paginator = &sql_paginator.Paginator
	return res, nil
}

func (s Service) DeleteSearchHistory(ctx context.Context, user gmodel.User, search_history_id *int64) bool {
	where_clause := table.SearchHistory.UserID.EQ(postgres.Int(user.ID))
	if search_history_id != nil {
		where_clause.AND(table.SearchHistory.ID.EQ(postgres.Int(*search_history_id)))
	}
	_, err := table.SearchHistory.
		DELETE().
		WHERE(where_clause).
		ExecContext(ctx, s.DB)
	return err == nil
}
