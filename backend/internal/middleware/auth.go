package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Claims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func JWT(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.Split(auth, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			tokenString := parts[1]

			token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secret), nil
			})

			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			if claims, ok := token.Claims.(*Claims); ok && token.Valid {
				c.Set("user", claims)

				// Set tenant ID in context from JWT claims
				if claims.TenantID != "" {
					if tenantID, err := uuid.Parse(claims.TenantID); err == nil {
						ctx := context.WithValue(c.Request().Context(), TenantIDKey, tenantID)
						c.SetRequest(c.Request().WithContext(ctx))
					}
				}

				return next(c)
			}

			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
		}
	}
}

func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get("user").(*Claims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "user not found in context")
			}

			for _, role := range roles {
				if user.Role == role {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
		}
	}
}

func GetUserClaims(c echo.Context) (*Claims, error) {
	user, ok := c.Get("user").(*Claims)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}
	return user, nil
}
