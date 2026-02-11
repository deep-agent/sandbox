package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtKeyfunc       jwt.Keyfunc
	jwtParserOptions []jwt.ParserOption
	jwtAuthRequired  bool
	jwtOnce          sync.Once
)

func initJWT() {
	jwtOnce.Do(func() {
		jwtKeyfunc, jwtParserOptions = loadJWTVerifierFromEnv(os.Getenv)
		jwtAuthRequired = isTruthy(os.Getenv("JWT_AUTH_REQUIRED"))
	})
}

func isTruthy(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes"
}

func Auth() app.HandlerFunc {
	initJWT()

	return func(ctx context.Context, c *app.RequestContext) {
		authHeader := string(c.GetHeader("Authorization"))

		if jwtKeyfunc == nil {
			if jwtAuthRequired {
				c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"code":    401,
					"message": "authentication required but not configured",
				})
				c.Abort()
				return
			}
			c.Next(ctx)
			return
		}

		if authHeader == "" {
			if jwtAuthRequired {
				c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"code":    401,
					"message": "missing authorization header",
				})
				c.Abort()
				return
			}
			c.Next(ctx)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "invalid authorization format",
			})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, jwtKeyfunc, jwtParserOptions...)

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "invalid token",
			})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("claims", claims)
		}

		c.Next(ctx)
	}
}
