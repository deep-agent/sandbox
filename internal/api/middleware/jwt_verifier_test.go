package middleware

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestLoadJWTVerifierFromEnv_HS256(t *testing.T) {
	secret := "my-secret"

	getenv := func(k string) string {
		if k == "JWT_SECRET" {
			return secret
		}
		return ""
	}

	keyfunc, opts := loadJWTVerifierFromEnv(getenv)
	if keyfunc == nil {
		t.Fatalf("expected keyfunc")
	}

	claims := jwt.MapClaims{
		"sub": "user1",
		"exp": jwt.NewNumericDate(time.Now().Add(time.Minute)),
	}

	tokenStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	fmt.Println(tokenStr)

	token, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, keyfunc, opts...)
	if err != nil || !token.Valid {
		t.Fatalf("parse token: valid=%v err=%v", token != nil && token.Valid, err)
	}
}

func TestLoadJWTVerifierFromEnv_NoSecret(t *testing.T) {
	getenv := func(k string) string {
		return ""
	}

	keyfunc, opts := loadJWTVerifierFromEnv(getenv)
	if keyfunc != nil {
		t.Fatalf("expected nil keyfunc when no secret")
	}
	if opts != nil {
		t.Fatalf("expected nil opts when no secret")
	}
}

func TestLoadJWTVerifierFromEnv_WrongMethod(t *testing.T) {
	secret := "my-secret"

	getenv := func(k string) string {
		if k == "JWT_SECRET" {
			return secret
		}
		return ""
	}

	keyfunc, _ := loadJWTVerifierFromEnv(getenv)
	if keyfunc == nil {
		t.Fatalf("expected keyfunc")
	}

	fakeToken := &jwt.Token{
		Method: jwt.SigningMethodRS256,
	}
	_, err := keyfunc(fakeToken)
	if err == nil {
		t.Fatalf("expected error for wrong signing method")
	}
}
