package setup

import (
	"context"
	"database/sql"
	"os"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/ayaanqui/go-migration-tool/migration_tool"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
	"github.com/pricetra/api/graph"
	gresolver "github.com/pricetra/api/graph/resolver"
	"github.com/pricetra/api/services"
	"github.com/pricetra/api/types"
	"google.golang.org/api/option"
	"googlemaps.github.io/maps"
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
		GoogleMapsApiKey: os.Getenv("GOOGLE_MAPS_API_KEY"),
		ExpoPushNotificationClientKey: os.Getenv("EXPO_PUSH_NOTIFICATION_CLIENT_KEY"),
		GoogleCloudVisionApiKey: os.Getenv("GOOGLE_CLOUD_VISION_API_KEY"),
		OpenAiApiKey: os.Getenv("OPENAI_API_SECRET"),
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

	// Setup Google maps client
	maps_client, err := maps.NewClient(maps.WithAPIKey(server.Tokens.GoogleMapsApiKey))
	if err != nil {
		panic(err)
	}

	vision_client, err := vision.NewImageAnnotatorClient(context.Background(), option.WithAPIKey(server.Tokens.GoogleCloudVisionApiKey))
	if err != nil {
		panic(err)
	}

	service := services.Service{
		DB: server.DB,
		StructValidator: server.StructValidator,
		Tokens: server.Tokens,
		Cloudinary: cloudinary,
		GoogleMapsClient: maps_client,
		ExpoPushClient: expo.NewPushClient(&expo.ClientConfig{
			AccessToken: server.Tokens.ExpoPushNotificationClientKey,
		}),
		GoogleVisionApiClient: vision_client,
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
