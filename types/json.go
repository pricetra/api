package types

type EmailServer struct {
	Url string
	ApiKey string
}

type CloudinaryTokens struct {
	ApiKey string
	ApiSecret string
	CloudName string
}

type OpenFoodFactsTokens struct {
	Username string
	Password string
}

type Tokens struct {
	JwtKey string
	EmailServer EmailServer
	Cloudinary CloudinaryTokens
	UPCitemdbUserKey string
	GoogleMapsApiKey string
	ExpoPushNotificationClientKey string
	GoogleCloudVisionApiKey string
	OpenAiApiKey string
	OpenFoodFacts OpenFoodFactsTokens
}
