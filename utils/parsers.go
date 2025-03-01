package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt"
)

// Given a JSON file, map the contents into any struct dest
func FileMapper[T any](filename string, dest T) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("%s not found", filename)
	}
	if err = json.Unmarshal(file, dest); err != nil {
		return err
	}
	return nil
}

// Given a raw jwt token and an encryption key return the mapped jwt claims or an error
func GetJwtClaims(jwt_token string, key string) (jwt.MapClaims, error) {
	token, token_err := jwt.Parse(jwt_token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		return []byte(key), nil
	})
	if token_err != nil {
		return nil, fmt.Errorf("could not parse jwt token")
	}
	
	// Get claims stored in parsed JWT token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("could not fetch jwt claims")
	}
	return claims, nil
}
