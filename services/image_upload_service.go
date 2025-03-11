package services

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func (service Service) ImageUrlUpload(
	ctx context.Context,
	image_url string,
	params uploader.UploadParams,
) (*uploader.UploadResult, error) {
	return service.Cloudinary.Upload.Upload(ctx, image_url, params)
}

func (service Service) GraphImageUpload(
	ctx context.Context,
	image graphql.Upload,
	params uploader.UploadParams,
) (*uploader.UploadResult, error) {
	return service.Cloudinary.Upload.Upload(ctx, image.File, params)
}

func (service Service) DeleteImageUpload(
	ctx context.Context,
	upload_id string,
) (*uploader.DestroyResult, error) {
	invalidate := true
	return service.Cloudinary.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: upload_id,
		Invalidate: &invalidate,
	})
}
