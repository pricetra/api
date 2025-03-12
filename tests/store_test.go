package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/pricetra/api/graph/gmodel"
)

func TestStore(t *testing.T) {
	var err error
	user, _, err := service.CreateInternalUser(ctx, gmodel.CreateAccountInput{
		Name: "Store test user",
		Email: "store_test@pricetra.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatal(err)
	}
	var store gmodel.Store

	t.Run("create store", func(t *testing.T) {
		store, err = service.CreateStore(ctx, user, gmodel.CreateStore{
			Name: "Walmart",
			Logo: uuid.NewString(),
			Website: "https://www.walmart.com",
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("find store", func(t *testing.T) {
		found_store, err := service.FindStore(ctx, store.ID)
		if err != nil {
			t.Fatal("could not find store", err)
		}
		if store == found_store {
			t.Fatal("store data does not match", store, found_store)
		}

		if !service.StoreExists(ctx, found_store.ID) {
			t.Fatal("store should exist")
		}

		if service.StoreExists(ctx, 123232) {
			t.Fatal("store should not exist")
		}
	})
}
