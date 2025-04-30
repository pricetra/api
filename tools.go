//go:build tools
// +build tools

package tools

import (
	_ "github.com/ayaanqui/go-migration-tool"
	_ "github.com/go-jet/jet/v2/cmd/jet"
	_ "github.com/lib/pq"

	_ "github.com/cosmtrek/air"

	_ "github.com/99designs/gqlgen"
	_ "github.com/99designs/gqlgen/graphql/introspection"

	_ "github.com/go-jet/jet/v2"
	_ "github.com/go-jet/jet/v2/cmd/jet"
)
