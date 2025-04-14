package types

type EmailServer struct {
	Url string
	ApiKey string
}

type CloudinaryTokens struct {
	ApiKey string `json:"apiKey"`
	ApiSecret string `json:"apiSecret"`
	CloudName string `json:"cloudName"`
}

type Tokens struct {
	JwtKey string
	EmailServer EmailServer
	Cloudinary CloudinaryTokens `json:"cloudinary"`
	// To add other tokens create a struct and add them here,
	// make sure to also update ./tokens.json
}
