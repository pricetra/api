package tests

import (
	"reflect"
	"testing"

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
	input := gmodel.CreateProduct{
		Name: "Random test product",
		Image: &image,
		Description: "Some description",
		Brand: "Pricetra",
		Code: "ABC123BARCODETEST",
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
	})

	t.Run("find product", func(t *testing.T) {
		t.Run("find product with correct barcode", func(t *testing.T) {
			found_product, err := service.FindProductWithCode(ctx, input.Code)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(found_product, product) {
				t.Fatal("products not equal", found_product, product)
			}
		})

		t.Run("find product with incorrect barcode", func(t *testing.T) {
			_, err := service.FindProductWithCode(ctx, "01234567810")
			if err == nil {
				t.Fatal("product should throw an error")
			}
		})
	})
}
