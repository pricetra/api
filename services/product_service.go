package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
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

func (s Service) BarcodeSearch(ctx context.Context, barcode string, exact bool) (product gmodel.Product, err error) {
	where_clause := table.Product.Code.EQ(postgres.String(barcode))
	if !exact {
		where_clause = table.Product.Code.LIKE(postgres.String(fmt.Sprintf("%%%s%%", barcode)))
	}
	qb := table.Product.
		SELECT(
			table.Product.AllColumns,
			table.Category.AllColumns,
		).
		FROM(
			table.Product.
				INNER_JOIN(table.Category, table.Category.ID.EQ(table.Product.CategoryID)),
		).
		WHERE(where_clause)
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

func (s Service) ProductExists(ctx context.Context, id int64) bool {
	qb := table.Product.
		SELECT(table.Product.ID.AS("id")).
		FROM(table.Product).
		WHERE(table.Product.ID.EQ(postgres.Int(id))).
		LIMIT(1)
	var product struct{
		ID int64
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
	db := s.DbOrTxQueryable()
	created_user_table, updated_user_table, cols := s.CreatedAndUpdatedUserTable()
	tables := table.Stock.
		INNER_JOIN(table.Product, table.Product.ID.EQ(table.Stock.ProductID)).
		INNER_JOIN(table.Category, table.Category.ID.EQ(table.Product.CategoryID)).
		INNER_JOIN(table.Store, table.Store.ID.EQ(table.Stock.StoreID)).
		INNER_JOIN(table.Branch, table.Branch.ID.EQ(table.Stock.BranchID)).
		INNER_JOIN(table.Price, table.Price.ID.EQ(table.Stock.LatestPriceID)).
		INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)).
		LEFT_JOIN(created_user_table, created_user_table.ID.EQ(table.Price.CreatedByID)).
		LEFT_JOIN(updated_user_table, updated_user_table.ID.EQ(table.Price.UpdatedByID))

	where_clause := postgres.Bool(true)
	order_by := []postgres.OrderByClause{}

	if search != nil {
		if search.StoreID != nil {
			where_clause = where_clause.AND(
				table.Store.ID.EQ(postgres.Int(*search.StoreID)),
			)
		}

		if search.BranchID != nil {
			where_clause = where_clause.AND(
				table.Branch.ID.EQ(postgres.Int(*search.BranchID)),
			)
		}

		if search.Location != nil {
			l := search.Location
			distance_cols := s.GetDistanceCols(l.Latitude, l.Longitude, l.RadiusMeters)
			cols = append(cols, distance_cols.DistanceColumn)
			where_clause = where_clause.AND(distance_cols.DistanceWhereClauseWithRadius)
			order_by = append(order_by, postgres.FloatColumn(distance_cols.DistanceColumnName).ASC())
		}

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
			clause := table.Category.ExpandedPathname.LIKE(
				postgres.String(fmt.Sprintf("%%%s%%", *search.Category)),
			)
			where_clause = where_clause.AND(clause)
		}

		if search.Query != nil {
			query := strings.TrimSpace(*search.Query)
			if len(query) > 0 {
				product_ft_components := s.BuildFullTextSearchQueryComponents(table.Product.SearchVector, query)
				category_ft_components := s.BuildFullTextSearchQueryComponents(table.Category.SearchVector, query)
				cols = append(cols, product_ft_components.RankColumn, category_ft_components.RankColumn)
				where_clause = where_clause.AND(
					postgres.OR(
						product_ft_components.WhereClause,
						category_ft_components.WhereClause,
					),
				)
				order_by = append(
					order_by,
					category_ft_components.OrderByClause.DESC(),
					product_ft_components.OrderByClause.DESC(),
				)
			}
		}
	}
	order_by = append(
		order_by,
		table.Product.Views.DESC(),
		table.Price.CreatedAt.DESC(),
		table.Product.UpdatedAt.DESC(),
	)

	// get pagination data
	sql_paginator, err := s.Paginate(ctx, paginator_input, tables, table.Product.ID, where_clause)
	if err != nil {
		// Return empty result
		return gmodel.PaginatedProducts{
			Products: []*gmodel.Product{},
			Paginator: &gmodel.Paginator{},
		}, nil
	}

	cols = append(
		cols,
		table.Category.AllColumns,
		table.Stock.AllColumns,
		table.Store.AllColumns,
		table.Branch.AllColumns,
		table.Price.AllColumns,
		table.Address.AllColumns,
	)
	qb := table.Product.
		SELECT(table.Product.AllColumns, cols...).
		FROM(tables).
		WHERE(where_clause).
		ORDER_BY(order_by...).
		LIMIT(int64(sql_paginator.Limit)).
		OFFSET(int64(sql_paginator.Offset))
	if err := qb.QueryContext(ctx, db, &paginated_products.Products); err != nil {
		return gmodel.PaginatedProducts{}, err
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

type ViewerTrailFull struct {
	gmodel.ViewerTrailInput
	UserID *int64
	Platform *gmodel.AuthDeviceType
}

func (s Service) AddProductViewer(
	ctx context.Context,
	product_id int64,
	viewer_trail ViewerTrailFull,
) (viewer model.ProductView, err error) {
	if s.TX, err = s.DB.BeginTx(ctx, nil); err != nil {
		return model.ProductView{}, err
	}
	defer s.TX.Rollback()

	cols := postgres.ColumnList{
		table.ProductView.ProductID,
	}
	platform := model.AuthDeviceType_Other
	if viewer_trail.Origin != nil {
		cols = append(cols, table.ProductView.Origin)
	}
	if viewer_trail.StockID != nil {
		cols = append(cols, table.ProductView.StockID)
	}
	if viewer_trail.UserID != nil {
		cols = append(cols, table.ProductView.UserID)
	}
	if viewer_trail.Platform != nil {
		cols = append(cols, table.ProductView.Platform)
		if platform.Scan(viewer_trail.Platform.String()) != nil {
			platform = model.AuthDeviceType_Unknown
		}
	}
	create_view := table.ProductView.
		INSERT(cols).
		MODEL(model.ProductView{
			ProductID: product_id,
			StockID: viewer_trail.StockID,
			UserID: viewer_trail.UserID,
			Origin: viewer_trail.Origin,
			Platform: platform,
		}).
		RETURNING(table.ProductView.AllColumns)
	if err = create_view.QueryContext(ctx, s.TX, &viewer); err != nil {
		return model.ProductView{}, err
	}

	update_product := table.Product.
		UPDATE(table.Product.Views, table.Product.UpdatedAt).
		SET(
			table.Product.Views.SET(table.Product.Views.ADD(postgres.Int(1))),
			table.Product.UpdatedAt.SET(postgres.NOW()),
		).
		WHERE(table.Product.ID.EQ(postgres.Int(product_id)))
	if _, err = update_product.ExecContext(ctx, s.TX); err != nil {
		return model.ProductView{}, err
	}
	if err = s.TX.Commit(); err != nil {
		return model.ProductView{}, err
	}
	return viewer, nil
}

func (s Service) ExtractProductTextFromBase64Image(ctx context.Context, base64_image string) (extraction_ob gmodel.ProductExtractionResponse, err error) {
	if !utils.IsValidBase64Image(base64_image) {
		return gmodel.ProductExtractionResponse{}, fmt.Errorf("not a valid base64 encoded image")
	}

	// upload image to CDN
	upload_id := uuid.NewString()
	upload_res, err := s.ImageUrlUpload(ctx, base64_image, uploader.UploadParams{
		PublicID: upload_id,
		Tags: []string{"OCR"},
	})
	defer s.DeleteImageUpload(ctx, upload_id)
	if err != nil {
		return gmodel.ProductExtractionResponse{}, fmt.Errorf("could not upload image: %w", err)
	}
	if upload_res == nil {
		return gmodel.ProductExtractionResponse{}, fmt.Errorf("upload response was empty")
	}

	// use uploaded image to extract OCR data
	// using Google Vision
	upload_uri := fmt.Sprintf("%s/%s", CLOUDINARY_UPLOAD_BASE, upload_id)
	ocr_data, err := s.GoogleVisionOcrData(ctx, upload_uri)
	if err != nil {
		return gmodel.ProductExtractionResponse{}, fmt.Errorf("ocr error: %w", err)
	}
	
	// get prompt template
	template, err := s.GetAiTemplate(ctx, model.AiPromptType_ProductDetails)
	if err != nil {
		return gmodel.ProductExtractionResponse{}, fmt.Errorf("template error")
	}

	// replace variables with ocr data
	template.Prompt = strings.ReplaceAll(template.Prompt, template.Variable, ocr_data)

	// get gpt response
	gpt_res, err := s.GptResponse(ctx, template.Prompt, template.MaxTokens)
	if err != nil {
		return gmodel.ProductExtractionResponse{}, fmt.Errorf("could not analyze ocr data: %w", err)
	}

	extracted_fields, err := ParseRawGptResponse[gmodel.ProductExtractionFields](gpt_res)
	if err != nil {
		return gmodel.ProductExtractionResponse{}, err
	}

	extraction_ob = gmodel.ProductExtractionResponse{
		Brand: extracted_fields.Brand,
		Name: extracted_fields.ProductName,
	}
	if extracted_fields.Weight != nil {
		weight := strings.ToLower(*extracted_fields.Weight)
		extraction_ob.Weight = &weight
	}
	if len(extracted_fields.Category) > 0 {
		category, err := s.CategoryRecursiveInsert(ctx, extracted_fields.Category)
		if err == nil {
			extraction_ob.CategoryID = &category.ID
			extraction_ob.Category = &category
		}
	}
	return extraction_ob, nil
}
