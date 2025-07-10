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

const CATEGORY_DELIM string = " > "

func (s Service) FindCategoryById(ctx context.Context, id int64) (category gmodel.Category, err error) {
	qb := table.Category.
		SELECT(table.Category.AllColumns).
		FROM(table.Category).
		WHERE(table.Category.ID.EQ(postgres.Int(id))).
		LIMIT(1)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &category)
	return category, err
}

func (s Service) FindCategoryByExactName(ctx context.Context, name string) (category gmodel.Category, err error) {
	qb := table.Category.
		SELECT(table.Category.AllColumns).
		FROM(table.Category).
		WHERE(table.Category.Name.EQ(postgres.String(name))).
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

// This method takes a category path and goes through each id to 
// verify the hierarchial order and return the expanded_pathname.
// Ex. `[462,463,464]` -> "Food, Beverages & Tobacco > Dairy & Eggs > Milk"
func (s Service) CategoryPathToExpandedPathname(ctx context.Context, path []int, name *string) (expanded_pathname string, err error) {
	if len(path) == 0 {
		if name != nil {
			return *name, nil
		}
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
	expanded_pathname = strings.Join(names, CATEGORY_DELIM)
	if categories[len(path) - 1].ExpandedPathname != expanded_pathname {
		return "", fmt.Errorf("incorrect path")
	}
	if name != nil {
		names = append(names, *name)
		return strings.Join(names, CATEGORY_DELIM), nil
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
	// add random negative int to temporarily avoid path unique key constraint collisions
	path := utils.ToPostgresArray(append(input.ParentPath, -1 * utils.RangedRandomInt(1617, 4916170)))

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

func (s Service) FindCategories(
	ctx context.Context,
	depth *int,
	parent_id *int64,
	search *string,
) (categories []gmodel.Category, err error) {
	depth_col_name := fmt.Sprintf("%s.depth", table.Category.TableName())
	path_col_name := utils.BuildFullTableName(table.Category.Path)
	order_by := []postgres.OrderByClause{table.Category.ID.ASC()}
	cols := []postgres.Projection{
		postgres.RawInt(fmt.Sprintf("array_length(%s, 1)", path_col_name)).
			AS(depth_col_name),
	}
	where_clause := postgres.Bool(true)
	if depth != nil {
		where_clause = postgres.RawBool(
			fmt.Sprintf("array_length(%s, 1) = %d", path_col_name, *depth),
		)
	}
	if parent_id != nil {
		path_col := utils.BuildFullTableName(table.Category.Path)
		contains_clause := postgres.RawBool(
			fmt.Sprintf("$id = any(%s)", path_col), 
			map[string]any{
				"$id": *parent_id,
			},
		)
		if where_clause == nil {
			where_clause = contains_clause
		} else {
			where_clause = postgres.AND(where_clause, contains_clause)
		}
	}
	if search != nil && *search != "" {
		fts_components := s.BuildFullTextSearchQueryComponents(table.Category.SearchVector, *search)
		cols = append(cols, fts_components.RankColumn)
		where_clause = where_clause.AND(fts_components.WhereClause)
		order_by = append(order_by, fts_components.OrderByClause.DESC())
	}
	qb := table.Category.SELECT(
			table.Category.AllColumns,
			cols...,
		).
		FROM(table.Category).
		WHERE(where_clause).
		ORDER_BY(order_by...)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &categories)
	return categories, err
}

// Given a category string (Ex. "Food, Beverages & Tobacco > Cooking & Baking Ingredients > Baking Decorations")
// recursively insert each category and then return the final category
func (s Service) CategoryRecursiveInsert(ctx context.Context, category_str string) (gmodel.Category, error) {
	var err error
	parsed_category := strings.Split(category_str, CATEGORY_DELIM)
	categories := make([]gmodel.Category, len(parsed_category))
	for i, category_name := range parsed_category {
		categories[i], err = s.FindCategoryByExactName(ctx, category_name)
		if err == nil {
			continue
		}

		// Category was not found so insert new one
		input := gmodel.CreateCategory{ Name: category_name }
		if i > 0 {
			input.ParentPath = utils.PostgresArrayToIntArray(categories[i-1].Path)
		}
		categories[i], err = s.CreateCategory(ctx, input)
		if err != nil {
			// Category could already exist with a different path
			// so we will use it instead
			categories[i], err = s.FindCategoryByExactName(ctx, category_name)
			if err != nil {
				return gmodel.Category{}, err
			}
		}
	}
	return categories[len(categories) - 1], nil
}
