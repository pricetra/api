package tests

import (
	"fmt"
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
}
