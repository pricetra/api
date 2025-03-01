package types

import (
	"fmt"
	"strings"
)

type AuthorizationKeyType string
const AuthorizationKey AuthorizationKeyType = "Authorization"

// Given a bearer token ("Bearer <TOKEN>") returns the token or an error if parsing was unsuccessful
func (value AuthorizationKeyType) GetToken() (string, error) {
	parsed_authorization := strings.Split(string(value), " ")
	if parsed_authorization[0] != "Bearer" || len(parsed_authorization) < 2 {
		return "", fmt.Errorf("could not parse bearer token")
	}
	token := parsed_authorization[1]
	return token, nil
}

type AuthUserKeyType string
const AuthUserKey AuthUserKeyType = "AUTH_USER"
