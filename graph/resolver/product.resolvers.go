package gresolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.44

import (
	"context"
	"fmt"

	"github.com/pricetra/api/graph/gmodel"
)

// BarcodeScan is the resolver for the barcodeScan field.
func (r *queryResolver) BarcodeScan(ctx context.Context, barcode string) (*gmodel.Product, error) {
	user := r.Service.GetAuthUserFromContext(ctx)
	product, err := r.Service.FindProductWithCode(ctx, barcode)
	if err == nil {
		return &product, nil
	}

	result, err := r.Service.UPCItemDbLookupWithUpcCode(barcode)
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("no results found for this barcode")
	}
	item := result.Items[0]
	product, err = r.Service.CreateProduct(ctx, user, item.ToCreateProduct(&barcode))
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// AllProducts is the resolver for the allProducts field.
func (r *queryResolver) AllProducts(ctx context.Context) ([]*gmodel.Product, error) {
	products, err := r.Service.FindAllProducts(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*gmodel.Product, len(products))
	for i := 0; i < len(products); i++ {
		res[i] = &products[i]
	}
	return res, nil
}
