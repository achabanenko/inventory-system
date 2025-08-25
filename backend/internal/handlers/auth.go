package handlers

import (
	"net/http"
	"time"

	"inventory/internal/middleware"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	TenantSlug string `json:"tenant_slug,omitempty"` // Optional: specific tenant to login to
}

type LoginResponse struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresIn    int            `json:"expires_in"`
	User         UserResponse   `json:"user"`
	Tenant       TenantResponse `json:"tenant"`
}

func (h *Handler) Login(c echo.Context) error {
	log.Info().
		Str("content_type", c.Request().Header.Get("Content-Type")).
		Str("method", c.Request().Method).
		Msg("Login handler called")

	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind request")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	log.Info().
		Str("email", req.Email).
		Bool("has_password", req.Password != "").
		Msg("Login attempt")

	// Query user from database with optional tenant filtering
	var userID, tenantID, tenantName, tenantSlug, hashedPassword, name, role string
	var isActive bool

	var query string
	var args []interface{}

	if req.TenantSlug != "" {
		// Specific tenant login
		query = `
			SELECT u.id, u.tenant_id, t.name, t.slug, u.password_hash, u.name, u.role, u.is_active 
			FROM users u
			INNER JOIN tenants t ON u.tenant_id = t.id
			WHERE u.email = $1 AND t.slug = $2 AND u.is_active = true AND t.is_active = true
		`
		args = []interface{}{req.Email, req.TenantSlug}
	} else {
		// Find first active tenant for this email (backward compatibility)
		query = `
			SELECT u.id, u.tenant_id, t.name, t.slug, u.password_hash, u.name, u.role, u.is_active 
			FROM users u
			INNER JOIN tenants t ON u.tenant_id = t.id
			WHERE u.email = $1 AND u.is_active = true AND t.is_active = true
			ORDER BY u.created_at ASC
			LIMIT 1
		`
		args = []interface{}{req.Email}
	}

	err := h.DB.QueryRow(query, args...).Scan(&userID, &tenantID, &tenantName, &tenantSlug, &hashedPassword, &name, &role, &isActive)

	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("User not found")
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("Invalid password")
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
	}

	// Generate tokens
	accessToken, err := h.generateToken(
		userID,
		tenantID,
		req.Email,
		role,
		h.Config.JWTExpiry,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate access token")
	}

	refreshToken, err := h.generateToken(
		userID,
		tenantID,
		req.Email,
		role,
		h.Config.RefreshExpiry,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate refresh token")
	}

	return c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(h.Config.JWTExpiry.Seconds()),
		User: UserResponse{
			ID:       userID,
			Name:     name,
			Email:    req.Email,
			Role:     role,
			TenantID: tenantID,
		},
		Tenant: TenantResponse{
			ID:   tenantID,
			Name: tenantName,
			Slug: tenantSlug,
		},
	})
}

func (h *Handler) Refresh(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	claims, err := h.validateToken(req.RefreshToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid refresh token")
	}

	accessToken, err := h.generateToken(
		claims.UserID,
		claims.TenantID,
		claims.Email,
		claims.Role,
		h.Config.JWTExpiry,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate access token")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"access_token": accessToken,
		"expires_in":   int(h.Config.JWTExpiry.Seconds()),
	})
}

func (h *Handler) Logout(c echo.Context) error {
	// TODO: Implement token blacklisting
	return c.JSON(http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}

func (h *Handler) generateToken(userID, tenantID, email, role string, expiry time.Duration) (string, error) {
	claims := &middleware.Claims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.Config.JWTSecret))
}

func (h *Handler) validateToken(tokenString string) (*middleware.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &middleware.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.Config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*middleware.Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
