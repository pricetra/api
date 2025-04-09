package tests

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/pricetra/api/graph/gmodel"
	"github.com/pricetra/api/utils"
)

func TestCategory(t *testing.T) {
	var err error
	var root_category_1 gmodel.Category

	t.Run("create root categories", func(t *testing.T) {
		root_category_1, err = service.CreateCategory(ctx, gmodel.CreateCategory{
			Name: "Food",
		})
		if err != nil {
			t.Fatal(err)
		}

		if root_category_1.Name != root_category_1.ExpandedPathname {
			t.Fatal("name and expanded pathname should be the same for root categories", root_category_1.Name, root_category_1.ExpandedPathname)
		}
		path := utils.ToPostgresArray([]int{int(root_category_1.ID)})
		if root_category_1.Path != path {
			t.Fatal("paths don't match", root_category_1.Path, path)
		}
	})

	t.Run("create subcategories", func(t *testing.T) {
		parent_path := utils.PostgresArrayToIntArray(root_category_1.Path)
		subcategory, err := service.CreateCategory(ctx, gmodel.CreateCategory{
			Name: "ABC123",
			ParentPath: parent_path,
		})
		if err != nil {
			t.Fatal(err)
		}

		if subcategory.ExpandedPathname != "Food > ABC123" {
			t.Fatal("pathname is incorrect", subcategory.ExpandedPathname)
		}
		path := utils.ToPostgresArray(append(parent_path, int(subcategory.ID)))
		fmt.Println(subcategory.Path, subcategory.ExpandedPathname)
		if subcategory.Path != path {
			t.Fatal("paths don't match", subcategory.Path, path)
		}

		t.Run("incorrect path", func(t *testing.T) {
			parent_path := append(utils.PostgresArrayToIntArray(subcategory.Path), 12)
			_, err := service.CreateCategory(ctx, gmodel.CreateCategory{
				Name: "12332",
				ParentPath: parent_path,
			})
			if err == nil {
				t.Fatal("sub category has invalid path")
			}
		})
	})

	t.Run("recursive category insertion test", func(t *testing.T) {
		t.Run("confirm existing category", func(t *testing.T) {
			cat1, err := service.CategoryRecursiveInsert(ctx, "Food, Beverages & Tobacco > Meat & Seafood > Fish & Seafood")
			if err != nil {
				t.Fatal(err)
			}
			check_category, err := service.FindCategoryById(ctx, 485)
			if err != nil {
				t.Fatal("check category was not found", err)
			}

			if !reflect.DeepEqual(cat1, check_category) {
				t.Fatal("categories don't match", cat1, check_category)
			}
		})

		t.Run("sub category", func(t *testing.T) {
			cat1, err := service.CategoryRecursiveInsert(ctx, "Food, Beverages & Tobacco > Meat & Seafood > Fish & Seafood > Calamari")
			if err != nil {
				t.Fatal(err)
			}
			check_category, err := service.FindCategoryById(ctx, 485)
			if err != nil {
				t.Fatal("check category was not found", err)
			}

			cat1_path := utils.PostgresArrayToIntArray(cat1.Path)
			check_category_path := append(
				utils.PostgresArrayToIntArray(check_category.Path),
				int(cat1.ID),
			)
			if !reflect.DeepEqual(cat1_path, check_category_path) {
				t.Fatal("categories don't match", cat1, check_category)
			}
		})
	})
}
