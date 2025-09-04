package services

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pricetra/api/graph/gmodel"
)

func (s Service) IsAuthenticatedDirective(ctx context.Context, obj any, next graphql.Resolver, role *gmodel.UserRole) (res any, err error) {
	user := s.GetAuthUserFromContext(ctx)
	if user == (gmodel.User{}) {
		return nil, fmt.Errorf("unauthorized")
	}

	if role != nil && !s.IsRoleAuthorized(*role, user.Role) {
		return nil, fmt.Errorf("insufficient permissions")
	}
	return next(ctx)
}
