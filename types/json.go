package types

type SendGridTemplates struct {
	EmailVerification string `json:"emailVerification"`
}

type SendGridTokens struct {
	ApiKey string `json:"apiKey"`
	RecoveryCode string `json:"recoveryCode,omitempty"`
	ApiKeyId string `json:"apiKeyId,omitempty"`
	Templates SendGridTemplates `json:"templates,omitempty"`
}

type CloudinaryTokens struct {
	ApiKey string `json:"apiKey"`
	ApiSecret string `json:"apiSecret"`
	CloudName string `json:"cloudName"`
}

type Tokens struct {
	JwtKey string `json:"jwtKey"`
	SendGrid SendGridTokens `json:"sendGrid,omitempty"`
	Cloudinary CloudinaryTokens `json:"cloudinary"`
	// To add other tokens create a struct and add them here,
	// make sure to also update ./tokens.json
}
