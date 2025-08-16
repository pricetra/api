package tests

import (
	"fmt"
	"testing"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
)

func TestProduct(t *testing.T) {
	var err error
	user, _, err := service.CreateInternalUser(ctx, gmodel.CreateAccountInput{
		Name: "Product test user",
		Email: "product_test@pricetra.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatal(err)
	}

	var product gmodel.Product
	image := "my_image.jpg"
	category, err := service.CategoryRecursiveInsert(ctx, "Product Test Category > Product Test Subcategory")
	if err != nil {
		t.Fatal("could not create category", err.Error())
	}
	weight := "2.5lb"
	input := gmodel.CreateProduct{
		Name: "Random test product",
		Image: &image,
		Description: "Some description",
		Brand: "Pricetra",
		Code: "ABC123BARCODETEST",
		CategoryID: category.ID,
		Weight: &weight,
	}
	
	t.Run("create product", func(t *testing.T) {
		product, err = service.CreateProduct(ctx, user, input, nil)
		if err != nil {
			t.Fatal(err)
		}

		if product.CreatedByID == nil || *product.CreatedByID != user.ID {
			t.Fatal("product createdById was not inserted")
		}
		if product.UpdatedByID == nil || *product.UpdatedByID != user.ID {
			t.Fatal("product updatedById was not inserted")
		}
		if product.Category.ExpandedPathname != category.ExpandedPathname {
			t.Fatal("category expanded pathname is incorrect")
		}
		if *product.WeightValue != 2.5 || *product.WeightType != "lb" {
			t.Fatal("weight value or type is incorrect", product.WeightValue, product.WeightType)
		}
	})

	t.Run("find product", func(t *testing.T) {
		t.Run("find product with correct barcode", func(t *testing.T) {
			found_product, err := service.FindProductWithCode(ctx, input.Code)
			if err != nil {
				t.Fatal(err)
			}

			if found_product.ID != product.ID {
				t.Fatal("product ids are not equal", found_product, product)
			}
		})

		t.Run("find product with incorrect barcode", func(t *testing.T) {
			_, err := service.FindProductWithCode(ctx, "01234567810")
			if err == nil {
				t.Fatal("product should throw an error")
			}
		})
	})

	t.Run("update product", func(t *testing.T) {
		updated_name := fmt.Sprintf("%s (updated)", product.Name)
		updated_product, _, err := service.UpdateProductById(ctx, user, product.ID, gmodel.UpdateProduct{
			Name: &updated_name,
		})
		if err != nil {
			t.Fatal(err)
		}
		if updated_product.ID != product.ID {
			t.Fatal("did not update the correct product")
		}
		if updated_product.Name == product.Name {
			t.Fatal("did not update name column")
		}
		if updated_product.Code != product.Code {
			t.Fatal("should not have updated anything but name, updated_at, and updated_by_id")
		}
	})

	t.Run("pagination", func(t *testing.T) {
		p, err := service.PaginatedProducts(ctx, gmodel.PaginatorInput{
			Limit: 3,
			Page: 1,
		}, nil);
		if err != nil {
			t.Fatal(err)
		}

		total_qb := table.Product.
			SELECT(postgres.COUNT(table.Product.ID).AS("total")).
			FROM(
				table.Product.
					INNER_JOIN(table.Category, table.Category.ID.EQ(table.Product.CategoryID)).
					INNER_JOIN(table.Stock, table.Stock.ProductID.EQ(table.Product.ID)).
					INNER_JOIN(table.Store, table.Store.ID.EQ(table.Stock.StoreID)).
					INNER_JOIN(table.Branch, table.Branch.ID.EQ(table.Stock.BranchID)).
					INNER_JOIN(table.Price, table.Price.ID.EQ(table.Stock.LatestPriceID)).
					INNER_JOIN(table.Address, table.Address.ID.EQ(table.Branch.AddressID)),
			)
		var p_total struct{ Total int }
		if err := total_qb.QueryContext(ctx, db, &p_total); err != nil {
			t.Fatal(err)
		}
		if p.Paginator.Total != p_total.Total {
			t.Fatal("total products value is incorrect", p.Paginator.Total, p_total.Total)
		}
	})
}
