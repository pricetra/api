package services

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/cloudinary/cloudinary-go/v2/api/admin"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

const CLOUDINARY_UPLOAD_BASE string = "https://res.cloudinary.com/pricetra-cdn/image/upload"

func (s Service) ImageUrlUpload(
	ctx context.Context,
	image_url string,
	params uploader.UploadParams,
) (*uploader.UploadResult, error) {
	return s.Cloudinary.Upload.Upload(ctx, image_url, params)
}

func (s Service) GraphImageUpload(
	ctx context.Context,
	image graphql.Upload,
	params uploader.UploadParams,
) (*uploader.UploadResult, error) {
	return s.Cloudinary.Upload.Upload(ctx, image.File, params)
}

func (s Service) DeleteImageUpload(
	ctx context.Context,
	upload_id string,
) (*uploader.DestroyResult, error) {
	invalidate := true
	return s.Cloudinary.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: upload_id,
		Invalidate: &invalidate,
	})
}

func (s Service) GetImageUpload(
	ctx context.Context,
	upload_id string,
) (*admin.AssetResult, error) {
	return s.Cloudinary.Admin.AssetByAssetID(ctx, admin.AssetByAssetIDParams{
		AssetID: upload_id,
	})
}
