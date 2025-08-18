package tests

import (
	"testing"

	"github.com/pricetra/api/graph/gmodel"
)

func TestBranch(t *testing.T) {
	var err error
	user, _, err := service.CreateInternalUser(ctx, gmodel.CreateAccountInput{
		Name: "Branch test user",
		Email: "branch_test@pricetra.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatal(err)
	}

	img := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAC0lEQVR42mP8//8/AwAI/wH+9Q4AAAAASUVORK5CYII="
	store, err := service.CreateStore(ctx, user, gmodel.CreateStore{
		Name: "Target",
		LogoBase64: &img,
		Website: "https://www.target.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	var branch gmodel.Branch

	t.Run("create branch", func(t *testing.T) {
		lat, lon := 41.8523563, -88.3454819
		branch, err = service.CreateBranch(ctx, user, gmodel.CreateBranch{
			Name: "Target Batavia",
			StoreID: store.ID,
			Address: &gmodel.CreateAddress{
				Latitude: lat,
				Longitude: lon,
				MapsLink: "https://www.google.com/maps/place/Target/@41.8523563,-88.3454819,17z/data=!3m1!4b1!4m15!1m8!3m7!1s0x880ee33e1a7537bb:0x56345c2beb5690cd!2s115+N+Randall+Rd,+Batavia,+IL+60510!3b1!8m2!3d41.8523563!4d-88.342907!16s%2Fg%2F11bw3yctkq!3m5!1s0x880ee33e154c66dd:0x13887485f034e6e8!8m2!3d41.8523563!4d-88.342907!16s%2Fg%2F1tfh57d3?entry=ttu&g_ep=EgoyMDI1MDMxMC4wIKXMDSoJLDEwMjExNDU1SAFQAw%3D%3D",
				FullAddress: "115 N Randall Rd, Batavia, IL 60510, USA",
				City: "Batavia",
				AdministrativeDivision: "Illinois",
				CountryCode: "US",
				ZipCode: 60510,
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		if !service.AddressExists(ctx, lat, lon) {
			t.Fatal("address could not be created")
		}

		if branch.ID == 0 {
			t.Fatal("branch was not created")
		}
		if branch.Address == nil {
			t.Fatal("address must be defined")
		}
		if branch.Address.Latitude != lat || branch.Address.Longitude != lon {
			t.Fatal("lat and lon data does not match")
		}
	})

	t.Run("find branch", func(t *testing.T) {
		found_branch, err := service.FindBranchByBranchIdAndStoreId(ctx, branch.ID, store.ID)
		if err != nil {
			t.Fatal(err)
		}

		if found_branch.ID != branch.ID || found_branch.AddressID != branch.AddressID {
			t.Fatal("valued don't match")
		}
	})

	t.Run("find branches for store", func(t *testing.T) {
		branches, err := service.FindBranchesByStoreId(ctx, store.ID, gmodel.PaginatorInput{
			Limit: 1,
			Page: 1,
		}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		if len(branches.Branches) != 1 {
			t.Fatal("should only have 1 branch", len(branches.Branches))
		}
	})
}
