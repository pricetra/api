package services

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pricetra/api/graph/gmodel"
	"github.com/pricetra/api/types"
)

func (service Service) IsAuthenticatedDirective(ctx context.Context, obj any, next graphql.Resolver, role *gmodel.UserRole) (res any, err error) {
	authorization, ok := ctx.Value(types.AuthorizationKey).(types.AuthorizationKeyType)
	if !ok {
		return nil, fmt.Errorf("authorization header type error")
	}
	user, err := service.VerifyJwt(ctx, authorization)
	if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}
	new_ctx := context.WithValue(ctx, types.AuthUserKey, user)
	return next(new_ctx)
}
