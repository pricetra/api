package setup

import (
	"database/sql"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/ayaanqui/go-migration-tool/migration_tool"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/pricetra/api/graph"
	gresolver "github.com/pricetra/api/graph/resolver"
	"github.com/pricetra/api/services"
	"github.com/pricetra/api/types"
)

const GRAPH_ENDPOINT string = "/graphql"

func NewServer(db_conn *sql.DB, router *chi.Mux) *types.ServerBase {
	server := types.ServerBase{
		DB: db_conn,
		Router: router,
		StructValidator: validator.New(),
		MigrationDirectory: "./database/migrations",
	}

	// Run DB migrations
	db_migration := migration_tool.New(server.DB, &migration_tool.Config{
		Directory: server.MigrationDirectory,
		TableName: "migration",
	})
	db_migration.RunMigration()

	server.Tokens = &types.Tokens{
		JwtKey: os.Getenv("JWT_KEY"),
		EmailServer: types.EmailServer{
			Url: os.Getenv("EMAIL_SERVER_URL"),
			ApiKey: os.Getenv("EMAIL_SERVER_API_KEY"),
		},
		Cloudinary: types.CloudinaryTokens{
			CloudName: os.Getenv("CLOUDINARY_CLOUD_NAME"),
			ApiKey: os.Getenv("CLOUDINARY_API_KEY"),
			ApiSecret: os.Getenv("CLOUDINARY_API_SECRET"),
		},
		UPCitemdbUserKey: os.Getenv("UPCITEMDB_USER_KEY"),
	}

	// Setup Cloudinary CDN
	cloudinary, err := cloudinary.NewFromParams(
		server.Tokens.Cloudinary.CloudName,
		server.Tokens.Cloudinary.ApiKey,
		server.Tokens.Cloudinary.ApiSecret,
	)
	if err != nil {
		panic(err)
	}

	service := services.Service{
		DB: server.DB,
		StructValidator: server.StructValidator,
		Tokens: server.Tokens,
		Cloudinary: cloudinary,
	}

	// Startup utils...
	StartupUtils(service)

	// TODO: only show in dev environment
	server.Router.Handle("/playground", playground.Handler("GraphQL Playground", GRAPH_ENDPOINT))

	server.Router.Group(func(chi_router chi.Router) {
		c := graph.Config{}
		c.Resolvers = &gresolver.Resolver{
			AppContext: server,
			Service: service,
		}
		c.Directives.IsAuthenticated = service.IsAuthenticatedDirective
		graphql_handler := handler.NewDefaultServer(graph.NewExecutableSchema(c))

		chi_router.Use(service.AuthorizationMiddleware)
		chi_router.Handle(GRAPH_ENDPOINT, graphql_handler)
	})
	return &server
}
