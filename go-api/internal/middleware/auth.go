package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const UserIDKey = "userID"

func JWT(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			tokenValue, ok := strings.CutPrefix(authHeader, "Bearer ")
			if !ok || tokenValue == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header")
			}

			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (any, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			userID, ok := claims["user_id"].(float64)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}

			c.Set(UserIDKey, uint(userID))
			return next(c)
		}
	}
}
