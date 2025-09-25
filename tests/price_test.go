package tests

import (
	"testing"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/table"
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
	img := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAC0lEQVR42mP8//8/AwAI/wH+9Q4AAAAASUVORK5CYII="
	store, err = service.CreateStore(ctx, user, gmodel.CreateStore{
		Name: "Price Test Store",
		LogoBase64: &img,
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
	product, _ := service.CreateProduct(ctx, user, gmodel.CreateProduct{
		Name: "Random test product",
		Description: "Some description",
		Brand: "Pricetra",
		Code: "12345678901",
		CategoryID: category.ID,
	}, nil)
	product_2, _ := service.CreateProduct(ctx, user, gmodel.CreateProduct{
		Name: "Random test product 2",
		Description: "Some description",
		Brand: "Pricetra",
		Code: "12345678901234",
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
		if price.UnitType != "item" {
			t.Fatal("unit type should be item by default, but got", price.UnitType)
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
			if new_price.OriginalPrice != nil {
				t.Fatal("original_price should not be set if sale is false")
			}
			if new_price.ExpiresAt != nil {
				t.Fatal("expires_at should only be set when sale is true")
			}
			price = new_price

			if prices, err := service.FindPrices(ctx, product.ID, branch.ID); err != nil || len(prices) != 2 {
				t.Fatal("there should only be 2 price row")
			}
		})

		t.Run("create price with empty sale", func(t *testing.T) {
			stock, err := service.FindStock(ctx, product.ID, branch.ID, store.ID)
			if err != nil {
				t.Fatal(err)
			}

			price_input.Amount = 5.99
			price_input.OriginalPrice = nil
			price_input.UnitType = "lb"
			price_input.Sale = true
			new_price, err := service.CreatePrice(ctx, user, price_input)
			if err != nil {
				t.Fatal(err)
			}
			
			if new_price.UnitType != "lb" {
				t.Fatal("unit type should be lb, but got", new_price.UnitType)
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
			if *new_price.OriginalPrice != stock.LatestPrice.Amount {
				t.Fatal("original price is incorrect")
			}
			if prices, err := service.FindPrices(ctx, product.ID, branch.ID); err != nil || len(prices) != 3 {
				t.Fatal("there should only be 3 price row")
			}

			prices, err := service.FindPrices(ctx, product.ID, branch.ID)
			if err != nil {
				t.Fatal(err)
			}
			if len(prices) != 3 && prices[0].ID == new_price.ID {
				t.Fatal("prices returned are incorrect")
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
			price_input.UnitType = "oz"
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

			prices, err := service.FindPrices(ctx, product.ID, branch.ID)
			if err != nil || len(prices) != 5 {
				t.Fatal("there should only be 4 price row")
			}
			if prices[0].ID != new_price.ID {
				t.Fatal("latest price is not being returned first")
			}
			if prices[1].Amount != *new_price.OriginalPrice {
				t.Fatal("second latest price should be the original price")
			}
			if prices[1].UnitType != new_price.UnitType {
				t.Fatal("unit type should be same as the sale price")
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

		t.Run("create new price with sale", func(t *testing.T) {
			now := time.Now().Add(time.Hour * 24 * 7)
			original_price := 7.99
			price_input := gmodel.CreatePrice{
				ProductID: product_2.ID,
				Amount: 5.99,
				BranchID: branch.ID,
				CurrencyCode: nil,
				Sale: true,
				ExpiresAt: &now,
				OriginalPrice: &original_price,
				Condition: nil,
				UnitType: "item",
				ImageID: nil,
			}
			if _, err := service.CreatePrice(ctx, user, price_input); err != nil {
				t.Fatal(err)
			}

			var stocks []gmodel.Stock
			err = table.Stock.
				SELECT(table.Stock.ID).
				FROM(table.Stock).
				WHERE(
					table.Stock.ProductID.EQ(postgres.Int(product_2.ID)).
						AND(table.Stock.BranchID.EQ(postgres.Int(branch.ID))),
				).Query(db, &stocks)
			if err != nil {
				t.Fatal(err)
			}
			if len(stocks) != 1 {
				t.Fatal("there should be no duplicate stocks", len(stocks))
			}
		})

		t.Run("duplicate stock", func(t *testing.T) {
			_, err := service.CreateStock(ctx, user, gmodel.CreateStock{
				ProductID: product_2.ID,
				BranchID: branch.ID,
				StoreID: store.ID,
			})
			if err == nil {
				t.Fatal("duplicate stocks should not be allowed")
			}
		})
	})
}
