package setup

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/services"
	"github.com/pricetra/api/utils"
)

func MigrateAllRemoteProductImagesToCDN(service services.Service) {
	ctx := context.Background()
	qb := table.Product.
		SELECT(table.Product.Image.AS("image"), table.Product.Code.AS("code")).
		FROM(table.Product).
		WHERE(table.Product.Image.NOT_EQ(postgres.String("")))
	var res []struct{
		Image string
		Code string
	}
	err := qb.QueryContext(ctx, service.DB, &res)
	if err != nil {
		log.Fatal(err)
		return
	}
	
	for _, ob := range res {
		_, err := service.ImageUrlUpload(ctx, ob.Image, uploader.UploadParams{
			PublicID: ob.Code,
			Tags: []string{"PRODUCT"},
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func MigrateFillWeightTypeColumn(service services.Service) {
	ctx := context.Background()
	qb := table.Product.
		SELECT(
			table.Product.ID.AS("id"),
			table.Product.Weight.AS("weight"),
		).
		WHERE(
			postgres.AND(
				table.Product.Weight.IS_NOT_NULL(),
				table.Product.Weight.NOT_EQ(postgres.String("")),
			),
		).
		ORDER_BY(table.Product.ID)
	var weights []struct {
		ID     int64 `sql:"primary_key"`
		Weight string
	}
	if err := qb.QueryContext(ctx, service.DB, &weights); err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("Found %d products with weight to migrate", len(weights))

	var stmts []string
	for _, w := range weights {
		weight_comps, err := utils.ParseWeightIntoStruct(w.Weight)
		if err != nil {
			log.Printf("Failed to parse weight for product ID %d: %v", w.ID, err)
			continue
		}
		stmts = append(
			stmts, 
			strings.ReplaceAll(strings.ReplaceAll(
				table.Product.
					UPDATE(
						table.Product.WeightType,
						table.Product.WeightValue,
					).
					SET(
						table.Product.WeightType.SET(postgres.String(weight_comps.WeightType)),
						table.Product.WeightValue.SET(postgres.Float(weight_comps.Weight)),
					).
					WHERE(table.Product.ID.EQ(postgres.Int64(w.ID))).
					DebugSql(),
				"    ",
				" ",
			), "\n", " "),
		)
	}
	file, err := os.OpenFile("database/migrations/1755381363685385_migrate_product_weights_and_units.sql", os.O_APPEND | os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	
	if _, err := file.WriteString(strings.Join(stmts, "\n")); err != nil {
		log.Fatal(err)
		return
	}
}

func StartupUtils(service services.Service) {
	MigrateFillWeightTypeColumn(service)
}
