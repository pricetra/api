package gresolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.44

import (
	"context"

	"github.com/pricetra/api/graph"
	"github.com/pricetra/api/graph/gmodel"
)

// CreateBranch is the resolver for the createBranch field.
func (r *mutationResolver) CreateBranch(ctx context.Context, input gmodel.CreateBranch) (*gmodel.Branch, error) {
	user := r.Service.GetAuthUserFromContext(ctx)
	branch, err := r.Service.CreateBranch(ctx, user, input)
	if err != nil {
		return nil, err
	}
	return &branch, nil
}

// AllBranches is the resolver for the allBranches field.
func (r *queryResolver) AllBranches(ctx context.Context, storeID int64) ([]*gmodel.Branch, error) {
	branches, err := r.Service.FindBranchesByStoreId(ctx, storeID)
	if err != nil {
		return nil, err
	}

	res := make([]*gmodel.Branch, len(branches))
	for i := range branches {
		res[i] = &branches[i]
	}
	return res, nil
}

// FindBranch is the resolver for the findBranch field.
func (r *queryResolver) FindBranch(ctx context.Context, storeID int64, id int64) (*gmodel.Branch, error) {
	branch, err := r.Service.FindBranchByBranchIdAndStoreId(ctx, id, storeID)
	if err != nil {
		return nil, err
	}
	return &branch, nil
}

// FindBranchesByDistance is the resolver for the findBranchesByDistance field.
func (r *queryResolver) FindBranchesByDistance(ctx context.Context, lat float64, lon float64, radiusMeters int) ([]*gmodel.Branch, error) {
	branches, err := r.Service.FindBranchesByCoordinates(ctx, lat, lon, radiusMeters)
	if err != nil {
		return nil, err
	}

	res := make([]*gmodel.Branch, len(branches))
	for i := range branches {
		res[i] = &branches[i]
	}
	return res, nil
}

// Mutation returns graph.MutationResolver implementation.
func (r *Resolver) Mutation() graph.MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
