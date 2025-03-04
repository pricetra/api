package tests

import (
	"reflect"
	"testing"

	"github.com/pricetra/api/graph/gmodel"
)

func TestProduct(t *testing.T) {
	var err error
	var product gmodel.Product
	input := gmodel.CreateProduct{
		Name: "Lucerne Provolone Cheese",
		Image: "my_image.jpg",
		Description: "Provolone cheese slices",
		Brand: "ALDI",
		Code: "021130045198",
	}
	
	t.Run("create product", func(t *testing.T) {
		product, err = service.CreateProduct(ctx, input)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("find product", func(t *testing.T) {
		found_product, err := service.FindProductWithCode(ctx, input.Code)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(found_product, product) {
			t.Fatal("products not equal", found_product, product)
		}
	})
}
