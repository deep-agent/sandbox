package middleware

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func loadJWTVerifierFromEnv(getenv func(string) string) (jwt.Keyfunc, []jwt.ParserOption) {
	secret := getenv("JWT_SECRET")
	if secret == "" {
		return nil, nil
	}

	options := []jwt.ParserOption{
		jwt.WithValidMethods([]string{"HS256", "HS384", "HS512"}),
	}

	secretBytes := []byte(secret)
	keyfunc := func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return secretBytes, nil
	}

	return keyfunc, options
}
