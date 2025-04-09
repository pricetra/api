package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
	"github.com/pricetra/api/utils"
)

func (s Service) FindCategoryById(ctx context.Context, id int64) (category gmodel.Category, err error) {
	qb := table.Category.
		SELECT(table.Category.AllColumns).
		FROM(table.Category).
		WHERE(table.Category.ID.EQ(postgres.Int(id))).
		LIMIT(1)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &category)
	return category, err
}

func (s Service) CategoryExists(ctx context.Context, id int64) bool {
	qb := table.Category.
		SELECT(table.Category.ID).
		FROM(table.Category).
		WHERE(table.Category.ID.EQ(postgres.Int(id))).
		LIMIT(1)
	var dest struct{ID int64}
	return qb.QueryContext(ctx, s.DbOrTxQueryable(), &dest) == nil
}

func (s Service) CategoryPathToExpandedPathname(ctx context.Context, path []int, name *string) (expanded_pathname string, err error) {
	const DELIM string = " > "
	if len(path) == 0 {
		return expanded_pathname, nil
	}
	categories := make([]gmodel.Category, len(path))
	names := make([]string, len(path))
	for i, id := range path {
		categories[i], err = s.FindCategoryById(ctx, int64(id))
		if err != nil {
			return expanded_pathname, err
		}
		names[i] = categories[i].Name
	}
	expanded_pathname = strings.Join(names, DELIM)
	if categories[len(path) - 1].ExpandedPathname != expanded_pathname {
		return "", fmt.Errorf("incorrect path")
	}
	if name != nil {
		names = append(names, *name)
		return strings.Join(names, DELIM), nil
	}
	return expanded_pathname, nil
}

func (s Service) CreateCategory(ctx context.Context, input gmodel.CreateCategory) (category gmodel.Category, err error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return gmodel.Category{}, err
	}
	defer tx.Rollback()

	expanded_pathname, err := s.CategoryPathToExpandedPathname(ctx, input.ParentPath, &input.Name)
	if err != nil {
		return gmodel.Category{}, err
	}
	path := utils.ToPostgresArray(input.ParentPath)

	qb := table.Category.
		INSERT(
			table.Category.Name,
			table.Category.Path,
			table.Category.ExpandedPathname,
			table.Category.CategoryAlias,
		).
		MODEL(
			model.Category{
				Name: input.Name,
				Path: path,
				ExpandedPathname: expanded_pathname,
				CategoryAlias: &input.Name,
			},
		).RETURNING(table.Category.AllColumns)
	if err = qb.QueryContext(ctx, tx, &category); err != nil {
		return gmodel.Category{}, err
	}

	// Update category path with id
	input.ParentPath = append(input.ParentPath, int(category.ID))
	update_qb := table.Category.
		UPDATE(table.Category.Path).
		SET(utils.ToPostgresArray(input.ParentPath)).
		WHERE(table.Category.ID.EQ(postgres.Int(category.ID))).
		RETURNING(table.Category.AllColumns)
	if err := update_qb.QueryContext(ctx, tx, &category); err != nil {
		return gmodel.Category{}, err
	}

	if err := tx.Commit(); err != nil {
		return gmodel.Category{}, err
	}
	return category, nil
}

func (s Service) FindCategories(ctx context.Context, depth *int) (categories []gmodel.Category, err error) {
	depth_col_name := fmt.Sprintf("%s.depth", table.Category.TableName())
	path_len_exp := fmt.Sprintf("array_length(%s, 1)", utils.BuildFullTableName(table.Category.Path))
	var where_clause postgres.BoolExpression = nil
	if depth != nil {
		where_clause = postgres.RawBool(fmt.Sprintf("%s > %d", path_len_exp, *depth))
	}
	qb := table.Category.SELECT(
			table.Category.AllColumns,
			postgres.RawInt(path_len_exp).AS(depth_col_name),
		).
		FROM(table.Category).
		WHERE(where_clause).
		ORDER_BY(table.Category.ID.ASC())
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &categories)
	return categories, err
}
