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

type Tokens struct {
	JwtKey string
	EmailServer EmailServer
	Cloudinary CloudinaryTokens
	UPCitemdbUserKey string
	GoogleMapsApiKey string
}
