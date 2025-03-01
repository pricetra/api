package gresolver

import (
	"github.com/pricetra/api/services"
	"github.com/pricetra/api/types"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{
	AppContext types.ServerBase
	Service services.Service
}
