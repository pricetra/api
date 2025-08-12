package tests

import (
	"testing"

	"github.com/pricetra/api/graph/gmodel"
)

func TestGroceryList(t *testing.T) {
	var err error
	user, _, err := service.CreateInternalUser(ctx, gmodel.CreateAccountInput{
		Name: "Grocery list test user",
		Email: "grocery_list_user@pricetra.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatal(err)
	}
	var default_list gmodel.GroceryList
	product1, _ := service.FindProductById(ctx, 1)

	t.Run("find all grocery lists", func(t *testing.T) {
		lists, err := service.GetGroceryLists(ctx, user)
		if err != nil {
			t.Fatal(err)
		}
		if len(lists) != 1 {
			t.Fatalf("expected 1 grocery list, got %d", len(lists))
		}
	})

	t.Run("find default grocery list", func(t *testing.T) {
		default_list, err = service.GetDefaultGroceryList(ctx, user)
		if err != nil {
			t.Fatal(err)
		}
		if !default_list.Default {
			t.Fatal("expected default grocery list to be true")
		}
	})

	t.Run("create grocery list item", func(t *testing.T) {
		t.Run("grocery item with product", func(t *testing.T) {
			item, err := service.CreateGroceryListItem(ctx, user, default_list.ID, gmodel.CreateGroceryListItemInput{
				ProductID: &product1.ID,
			})
			if err != nil {
				t.Fatal(err)
			}
			if item.ProductID == nil || *item.ProductID != 1 {
				t.Fatalf("incorrect values, got %+v", item)
			}
			if item.Quantity != 1 {
				t.Fatalf("incorrect quantity, got %+v", item)
			}
			if item.Unit == nil || *item.Unit != "item" {
				t.Fatalf("incorrect unit, got %+v", item)
			}
			if item.Completed != false {
				t.Fatalf("item should not be completed, got %+v", item)
			}
		})

		t.Run("grocery item with category", func(t *testing.T) {
			category := "Apple"
			quantity := 2
			unit := "lb"
			item, err := service.CreateGroceryListItem(ctx, user, default_list.ID, gmodel.CreateGroceryListItemInput{
				Category: &category,
				Quantity: &quantity,
				Unit: &unit,
			})
			if err != nil {
				t.Fatal(err)
			}
			if item.ProductID != nil {
				t.Fatalf("product id should be nil")
			}
			if item.Quantity != quantity {
				t.Fatalf("incorrect quantity, got %+v", item)
			}
			if item.Unit == nil || *item.Unit != "lb" {
				t.Fatalf("incorrect unit, got %+v", item)
			}
		})
	})
}

