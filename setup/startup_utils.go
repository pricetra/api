package setup

import (
	"context"
	"log"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/services"
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

func StartupUtils(service services.Service) {
}
