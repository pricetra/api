package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pricetra/api/graph/gmodel"
)

func TestPrice(t *testing.T) {
	var err error
	user, _, err := service.CreateInternalUser(ctx, gmodel.CreateAccountInput{
		Name: "Price test user",
		Email: "price_test@pricetra.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatal(err)
	}
	var store gmodel.Store
	store, err = service.CreateStore(ctx, user, gmodel.CreateStore{
		Name: "Price Test Store",
		Logo: uuid.NewString(),
		Website: "https://pricetra.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	branch, err := service.CreateBranch(ctx, user, gmodel.CreateBranch{
		Name: "Target Batavia",
		StoreID: store.ID,
		Address: &gmodel.CreateAddress{
			Latitude: 41.900612,
			Longitude: -88.3436658,
			MapsLink: "https://maps.google.com",
			FullAddress: "855 S Randall Rd, St. Charles, IL 60174, USA",
			City: "St. Charles",
			AdministrativeDivision: "Illinois",
			CountryCode: "US",
			ZipCode: 60174,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	category, err := service.CategoryRecursiveInsert(ctx, "Product Test Category > Product Test Subcategory")
	if err != nil {
		t.Fatal("could not create category", err.Error())
	}
	product, err := service.CreateProduct(ctx, user, gmodel.CreateProduct{
		Name: "Random test product",
		Image: nil,
		Description: "Some description",
		Brand: "Pricetra",
		Code: "12345678901",
		CategoryID: category.ID,
	}, nil)

	t.Run("create price", func(t *testing.T) {
		if _, err = service.FindStock(ctx, product.ID, branch.ID, branch.StoreID); err == nil {
			t.Fatal("stock should not exist before creating price")
		}
		if prices, err := service.FindPrices(ctx, product.ID, branch.ID); err != nil || len(prices) != 0 {
			t.Fatal("prices should not exist before creating price")
		}

		price_input := gmodel.CreatePrice{
			ProductID: product.ID,
			Amount: 10.99,
			BranchID: branch.ID,
			CurrencyCode: nil,
			Sale: false,
			ExpiresAt: nil,
			OriginalPrice: nil,
			Condition: nil,
			UnitType: "item",
			ImageID: nil,
		}
		price, err := service.CreatePrice(ctx, user, price_input)
		if err != nil {
			t.Fatal("could not create price", err.Error())
		}
		if price.OriginalPrice != nil {
			t.Fatal("original price should be nil, but got", price.OriginalPrice)
		}
		if price.ExpiresAt != nil {
			t.Fatal("expiresAt should be nil, but got", price.ExpiresAt)
		}

		stock, err := service.FindStock(ctx, product.ID, branch.ID, branch.StoreID)
		if err != nil {
			t.Fatal("stock should be created after creating price")
		}
		if stock.LatestPriceID != price.ID {
			t.Fatal("stock latest price should be set to the created price")
		}

		if prices, err := service.FindPrices(ctx, product.ID, branch.ID); err != nil || len(prices) != 1 {
			t.Fatal("there should only be 1 price row")
		}

		t.Run("validate amount", func(t *testing.T) {
			price_input.Amount = 0.0
			if _, err = service.CreatePrice(ctx, user, price_input); err == nil {
				t.Fatal("should not be able to create price with amount 0")
			}
		})

		t.Run("original price without sale", func(t *testing.T) {
			price_input.Amount = 11.99
			org_price := 50.0
			price_input.OriginalPrice = &org_price
			new_price, err := service.CreatePrice(ctx, user, price_input)
			if err != nil {
				t.Fatal(err)
			}
			if new_price.OriginalPrice == nil {
				t.Fatal("original price should be set, but got nil")
			}
			if *new_price.OriginalPrice == org_price {
				t.Fatal("original price should not be equal to the current price since sale is not set")
			}
			if *new_price.OriginalPrice != price.Amount {
				t.Fatal("since sale is not set original price should be equal to the current price, but got", *new_price.OriginalPrice)	
			}
			price = new_price

			if prices, err := service.FindPrices(ctx, product.ID, branch.ID); err != nil || len(prices) != 2 {
				t.Fatal("there should only be 2 price row")
			}
		})

		t.Run("create price with empty sale", func(t *testing.T) {
			price_input.Amount = 5.99
			price_input.OriginalPrice = nil
			price_input.Sale = true
			new_price, err := service.CreatePrice(ctx, user, price_input)
			if err != nil {
				t.Fatal(err)
			}
			
			if new_price.Sale != true {
				t.Fatal("sale should be true, but got false")
			}
			if new_price.ExpiresAt == nil {
				t.Fatal("expiresAt should not be nil, but got nil")
			}
			if new_price.OriginalPrice == nil {
				t.Fatal("original price should not be nil, but got nil")
			}
			if *new_price.OriginalPrice != price.Amount {
				t.Fatal("original price should be equal to the current price, but got", *new_price.OriginalPrice)
			}
			if prices, err := service.FindPrices(ctx, product.ID, branch.ID); err != nil || len(prices) != 3 {
				t.Fatal("there should only be 3 price row")
			}
		})

		t.Run("create sale with old expiration date", func(t *testing.T) {
			price_input.Amount = 4.99
			price_input.OriginalPrice = nil
			price_input.Sale = true
			date := time.Now().Add(-time.Hour * 24) // 1 days ago
			price_input.ExpiresAt = &date
			if _, err = service.CreatePrice(ctx, user, price_input); err == nil {
				t.Fatal("should not be able to create price with old expiration date")
			}
		})

		t.Run("create full sale price", func(t *testing.T) {
			price_input.Amount = 4.89
			org_price := 10.99
			price_input.OriginalPrice = &org_price
			price_input.Sale = true
			exp_date := time.Now().Add(time.Hour * 24 * 3 * 7) // 3 weeks from now
			price_input.ExpiresAt = &exp_date
			new_price, err := service.CreatePrice(ctx, user, price_input)
			if err != nil {
				t.Fatal(err)
			}
			exp_date_date_only, err := time.Parse(time.DateOnly, exp_date.Format("2006-01-02"))
			if err != nil {
				t.Fatal("could not parse expiration date", err)
			}
			if new_price.ExpiresAt.Equal(exp_date_date_only) {
				t.Fatal("expiresAt should not be nil, but got nil")
			}
			if new_price.OriginalPrice == nil {
				t.Fatal("original price should not be nil, but got nil")
			}
			if *new_price.OriginalPrice != org_price {
				t.Fatal("original price should be equal to the provided original price, but got", *new_price.OriginalPrice)
			}
			if prices, err := service.FindPrices(ctx, product.ID, branch.ID); err != nil || len(prices) != 5 {
				t.Fatal("there should only be 4 price row")
			}

			stock, err := service.FindStock(ctx, product.ID, branch.ID, branch.StoreID)
			if err != nil {
				t.Fatal("stock should be created after creating price")
			}
			if stock.LatestPriceID != new_price.ID {
				t.Fatal("stock latest price should be set to the created price")
			}
		})

		t.Run("create sale with original price set to 0", func(t *testing.T) {
			price_input.Amount = 4.89
			org_price := 0.0
			price_input.OriginalPrice = &org_price
			price_input.Sale = true
			if _, err := service.CreatePrice(ctx, user, price_input); err == nil {
				t.Fatal("should not be able to create price when original price set to 0")
			}

			if prices, err := service.FindPrices(ctx, product.ID, branch.ID); err != nil || len(prices) != 5 {
				t.Fatal("there should only be 4 price row")
			}
		})
	})
}
