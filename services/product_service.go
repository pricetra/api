package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
	"github.com/pricetra/api/utils"
)

func (s Service) CreateProduct(ctx context.Context, user gmodel.User, input gmodel.CreateProduct, source *model.ProductSourceType) (product gmodel.Product, err error) {
	if err := s.StructValidator.StructCtx(ctx, input); err != nil {
		return product, err
	}
	if s.BarcodeExists(ctx, input.Code) {
		return product, fmt.Errorf("barcode already exists in the database. please use the update method")
	}

	var source_val model.ProductSourceType
	if source == nil {
		source_val = model.ProductSourceType_Pricetra
	} else {
		source_val = *source
	}

	// product.image should always be pointed to the CDN with public_id == product.code 
	image := fmt.Sprintf("%s/%s", CLOUDINARY_UPLOAD_BASE, input.Code)
	input.Image = &image

	qb := table.Product.
		INSERT(
			table.Product.Name,
			table.Product.Image,
			table.Product.Description,
			table.Product.URL,
			table.Product.Brand,
			table.Product.Code,
			table.Product.Color,
			table.Product.Model,
			table.Product.CategoryID,
			table.Product.Weight,
			table.Product.LowestRecordedPrice,
			table.Product.HighestRecordedPrice,
			table.Product.Source,
			table.Product.CreatedByID,
			table.Product.UpdatedByID,
		).
		MODEL(struct{
			gmodel.CreateProduct
			Source model.ProductSourceType
			CreatedByID *int64
			UpdatedByID *int64
		}{
			CreateProduct: input,
			Source: source_val,
			CreatedByID: &user.ID,
			UpdatedByID: &user.ID,
		}).
		RETURNING(table.Product.AllColumns)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &product)

	// Add category
	category, _ := s.FindCategoryById(ctx, input.CategoryID)
	product.Category = &category
	return product, err
}

func (s Service) FindProductById(ctx context.Context, id int64) (product gmodel.Product, err error) {
	qb := table.Product.
		SELECT(
			table.Product.AllColumns,
			table.Category.AllColumns,
		).
		FROM(
			table.Product.
				INNER_JOIN(table.Category, table.Category.ID.EQ(table.Product.CategoryID)),
		).
		WHERE(table.Product.ID.EQ(postgres.Int(id)))
	
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &product)
	return product, err
}

func (s Service) FindProductWithCode(ctx context.Context, barcode string) (product gmodel.Product, err error) {
	qb := table.Product.
		SELECT(
			table.Product.AllColumns,
			table.Category.AllColumns,
		).
		FROM(
			table.Product.
				INNER_JOIN(table.Category, table.Category.ID.EQ(table.Product.CategoryID)),
		).
		WHERE(table.Product.Code.EQ(postgres.String(barcode)))
	
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &product)
	return product, err
}

func (s Service) BarcodeExists(ctx context.Context, barcode string) bool {
	qb := table.Product.
		SELECT(table.Product.Code.AS("code")).
		FROM(table.Product).
		WHERE(table.Product.Code.EQ(postgres.String(barcode))).
		LIMIT(1)
	var product struct{
		Code string
	}
	err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &product)
	return err == nil
}

func (s Service) FindAllProducts(ctx context.Context) (products []gmodel.Product, err error) {
	created_user_table, updated_user_table, cols := s.CreatedAndUpdatedUserTable()
	cols = append(cols, table.Category.AllColumns)
	qb := table.Product.
		SELECT(table.Product.AllColumns, cols...).
		FROM(
			table.Product.
				INNER_JOIN(table.Category, table.Category.ID.EQ(table.Product.CategoryID)).
				LEFT_JOIN(created_user_table, created_user_table.ID.EQ(table.Product.CreatedByID)).
				LEFT_JOIN(updated_user_table, updated_user_table.ID.EQ(table.Product.UpdatedByID)),
		).
		ORDER_BY(table.Product.CreatedAt.DESC())
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &products)
	return products, err
}

func (s Service) PaginatedProducts(ctx context.Context, paginator_input gmodel.PaginatorInput, search *gmodel.ProductSearch) (paginated_products gmodel.PaginatedProducts, err error) {
	created_user_table, updated_user_table, cols := s.CreatedAndUpdatedUserTable()
	tables := table.Product.
		INNER_JOIN(table.Category, table.Category.ID.EQ(table.Product.CategoryID)).
		LEFT_JOIN(created_user_table, created_user_table.ID.EQ(table.Product.CreatedByID)).
		LEFT_JOIN(updated_user_table, updated_user_table.ID.EQ(table.Product.UpdatedByID))

	where_clause := postgres.Bool(true)
	order_by := []postgres.OrderByClause{}

	if search != nil {
		if search.CategoryID != nil {
			clause := postgres.RawBool(
				fmt.Sprintf("$id = any(%s)", utils.BuildFullTableName(table.Category.Path)), 
				map[string]any{
					"$id": *search.CategoryID,
				},
			)
			where_clause = where_clause.AND(clause)
		}

		if search.Category != nil {
			col := utils.BuildFullTableName(table.Category.ExpandedPathname)
			clause := postgres.RawBool(col + " like $category", map[string]any{
				"$category": fmt.Sprintf("%%%s%%", *search.Category),
			})
			where_clause = where_clause.AND(clause)
		}

		if search.Query != nil {
			query := strings.TrimSpace(*search.Query)
			if query != "" {
				rank_col := "rank"
				search_vector_col_name := utils.BuildFullTableName(table.Product.SearchVector)
				args := postgres.RawArgs{"$query": query}
	
				// Rank column
				cols = append(cols, postgres.RawFloat(
					fmt.Sprintf(
						"ts_rank(%s, plainto_tsquery('english', $query::TEXT))",
						search_vector_col_name,
					),
					args,
				).AS(rank_col))
	
				// Where clause with tsquery
				clause := postgres.RawBool(
					fmt.Sprintf("%s @@ plainto_tsquery('english', $query::TEXT)", search_vector_col_name),
					args,
				)
				where_clause = where_clause.AND(clause)
				// Order by
				order_by = append(order_by, postgres.FloatColumn(rank_col).DESC())
			}
		}
	}
	order_by = append(order_by, table.Product.UpdatedAt.DESC())

	// get pagination data
	sql_paginator, err := s.Paginate(ctx, paginator_input, tables, table.Product.ID, where_clause)
	if err != nil {
		// Return empty result
		return gmodel.PaginatedProducts{
			Products: []*gmodel.Product{},
			Paginator: &gmodel.Paginator{},
		}, nil
	}

	cols = append(cols, table.Category.AllColumns)
	qb := table.Product.
		SELECT(table.Product.AllColumns, cols...).
		FROM(tables).
		WHERE(where_clause).
		ORDER_BY(order_by...).
		LIMIT(int64(sql_paginator.Limit)).
		OFFSET(int64(sql_paginator.Offset))
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &paginated_products.Products)
	if err != nil {
		return paginated_products, err
	}

	paginated_products.Paginator = &sql_paginator.Paginator
	return paginated_products, nil
}

func (s Service) UpdateProductById(ctx context.Context, user gmodel.User, id int64, input gmodel.UpdateProduct) (updated_product gmodel.Product, old_product gmodel.Product, err error) {
	if err := s.StructValidator.StructCtx(ctx, input); err != nil {
		return gmodel.Product{}, gmodel.Product{}, err
	}
	
	product, err := s.FindProductById(ctx, id)
	if err != nil {
		return gmodel.Product{}, gmodel.Product{}, fmt.Errorf("product with id does not exist")
	}

	cols := postgres.ColumnList{}
	code := product.Code
	if input.Name != nil && *input.Name != product.Name {
		cols = append(cols, table.Product.Name)
	}
	if input.Description != nil && *input.Description != product.Description {
		cols = append(cols, table.Product.Description)
	}
	if input.URL != nil {
		cols = append(cols, table.Product.URL)
	}
	if input.Brand != nil && *input.Brand != product.Brand {
		cols = append(cols, table.Product.Brand)
	}
	if input.Code != nil && *input.Code != product.Code {
		if s.BarcodeExists(ctx, *input.Code) {
			return gmodel.Product{}, gmodel.Product{}, fmt.Errorf("new barcode is already in use")
		}
		code = *input.Code
		cols = append(cols, table.Product.Code)
	}
	if input.Color != nil {
		cols = append(cols, table.Product.Color)
	}
	if input.Model != nil {
		cols = append(cols, table.Product.Model)
	}
	if input.CategoryID != nil {
		cols = append(cols, table.Product.CategoryID)
	}
	if input.Weight != nil {
		cols = append(cols, table.Product.Weight)
	}
	if input.LowestRecordedPrice != nil {
		cols = append(cols, table.Product.LowestRecordedPrice)
	}
	if input.HighestRecordedPrice != nil {
		cols = append(cols, table.Product.HighestRecordedPrice)
	}
	if input.ImageFile != nil {
		image := fmt.Sprintf("%s/%s", CLOUDINARY_UPLOAD_BASE, code)
		input.Image = &image
		cols = append(cols, table.Product.Image)
	}

	if len(cols) == 0 {
		return product, product, nil
	}
	cols = append(cols, table.Product.UpdatedByID, table.Product.UpdatedAt)
	qb := table.Product.
		UPDATE(cols).
		MODEL(struct{
			gmodel.UpdateProduct
			UpdatedByID *int64
			UpdatedAt time.Time
		}{
			UpdateProduct: input,
			UpdatedByID: &user.ID,
			UpdatedAt: time.Now(),
		}).
		WHERE(table.Product.ID.EQ(postgres.Int(id))).
		RETURNING(table.Product.AllColumns)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &updated_product)
	if err != nil {
		return gmodel.Product{}, gmodel.Product{}, err
	}

	// Add category
	category, _ := s.FindCategoryById(ctx, updated_product.CategoryID)
	updated_product.Category = &category
	return updated_product, product, nil
}

func (s Service) FindAllBrands(ctx context.Context) (brands []gmodel.Brand, err error) {
	qb := table.Product.
		SELECT(
			table.Product.Brand,
			postgres.COUNT(table.Product.ID).AS("brand.products"),
		).FROM(table.Product).
		GROUP_BY(table.Product.Brand).
		ORDER_BY(
			postgres.IntegerColumn("brand.products").DESC(),
			table.Product.Brand.ASC(),
		)
	db := s.DbOrTxQueryable()
	if err := qb.QueryContext(ctx, db, &brands); err != nil {
		return nil, err
	}
	return brands, nil
}
