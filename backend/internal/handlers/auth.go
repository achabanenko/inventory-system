package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"inventory/internal/middleware"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// Google OAuth structures
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

type GoogleOAuthRequest struct {
	Code        string `json:"code" validate:"required"`
	TenantSlug  string `json:"tenant_slug"` // Optional: allow OAuth without tenant for new users
	RedirectURI string `json:"redirect_uri" validate:"required"`
}

type GoogleOAuthResponse struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	ExpiresIn    int             `json:"expires_in"`
	User         UserResponse    `json:"user"`
	Tenant       *TenantResponse `json:"tenant,omitempty"` // Optional: may not have tenant yet
	IsNewUser    bool            `json:"is_new_user"`
	NeedsTenant  bool            `json:"needs_tenant"` // Flag if user needs to select/create tenant
}

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
	var oauthProvider string

	var query string
	var args []interface{}

	if req.TenantSlug != "" {
		// Specific tenant login
		query = `
			SELECT u.id, u.tenant_id, t.name, t.slug, u.password_hash, u.name, u.role, u.is_active, u.oauth_provider
			FROM users u
			INNER JOIN tenants t ON u.tenant_id = t.id
			WHERE u.email = $1 AND t.slug = $2 AND u.is_active = true AND t.is_active = true
		`
		args = []interface{}{req.Email, req.TenantSlug}
	} else {
		// Find first active tenant for this email (backward compatibility)
		query = `
			SELECT u.id, u.tenant_id, t.name, t.slug, u.password_hash, u.name, u.role, u.is_active, u.oauth_provider
			FROM users u
			INNER JOIN tenants t ON u.tenant_id = t.id
			WHERE u.email = $1 AND u.is_active = true AND t.is_active = true
			ORDER BY u.created_at ASC
			LIMIT 1
		`
		args = []interface{}{req.Email}
	}

	err := h.DB.QueryRow(query, args...).Scan(&userID, &tenantID, &tenantName, &tenantSlug, &hashedPassword, &name, &role, &isActive, &oauthProvider)

	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("User not found")
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
	}

	// Check if user is an OAuth user (no password)
	if oauthProvider != "" {
		log.Error().Str("email", req.Email).Str("oauth_provider", oauthProvider).Msg("OAuth user attempting traditional login")
		return echo.NewHTTPError(http.StatusUnauthorized, "this account uses Google OAuth. Please sign in with Google instead.")
	}

	// Check if user has a password
	if hashedPassword == "" {
		log.Error().Str("email", req.Email).Msg("User has no password set")
		return echo.NewHTTPError(http.StatusUnauthorized, "this account has no password set. Please use Google OAuth to sign in.")
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

// GoogleOAuth handles Google OAuth authentication
func (h *Handler) GoogleOAuth(c echo.Context) error {
	var req GoogleOAuthRequest
	if err := c.Bind(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind Google OAuth request")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Debug: Log the received request
	log.Info().
		Str("email", "google_oauth").
		Str("tenant_slug", req.TenantSlug).
		Bool("has_tenant_slug", req.TenantSlug != "").
		Str("tenant_slug_length", fmt.Sprintf("%d", len(req.TenantSlug))).
		Msg("Google OAuth request received")

	// Exchange authorization code for access token
	googleToken, err := h.exchangeCodeForToken(req.Code, req.RedirectURI)
	if err != nil {
		log.Error().Err(err).Msg("Failed to exchange code for token")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to authenticate with Google")
	}

	// Get user info from Google
	googleUser, err := h.getGoogleUserInfo(googleToken)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get Google user info")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user information")
	}

	var userID, name, role string
	var isActive bool
	var isNewUser bool
	var needsTenant bool
	var tenantID, tenantName, tenantSlug string
	var hashedPassword string

	if req.TenantSlug != "" {
		// Scenario 1: OAuth with specific tenant
		log.Info().Str("tenant_slug", req.TenantSlug).Msg("OAuth with specific tenant")

		// Verify tenant exists and is active
		err = h.DB.QueryRow(`
			SELECT id, name, slug FROM tenants 
			WHERE slug = $1 AND is_active = true
		`, req.TenantSlug).Scan(&tenantID, &tenantName, &tenantSlug)

		if err != nil {
			log.Error().Err(err).Str("tenant_slug", req.TenantSlug).Msg("Tenant not found or inactive")
			return echo.NewHTTPError(http.StatusNotFound, "tenant not found or inactive")
		}

		// Check if user already exists in this tenant
		err = h.DB.QueryRow(`
			SELECT id, password_hash, name, role, is_active 
			FROM users 
			WHERE email = $1 AND tenant_id = $2
		`, googleUser.Email, tenantID).Scan(&userID, &hashedPassword, &name, &role, &isActive)

		if err != nil {
			// User doesn't exist in this tenant, create new user
			isNewUser = true
			role = "CLERK" // Default role for new OAuth users
			name = googleUser.Name

			// Insert new user
			err = h.DB.QueryRow(`
				INSERT INTO users (email, name, role, tenant_id, oauth_provider, oauth_id, avatar_url, is_active)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING id
			`, googleUser.Email, googleUser.Name, role, tenantID, "google", googleUser.ID, googleUser.Picture, true).Scan(&userID)

			if err != nil {
				log.Error().Err(err).Str("email", googleUser.Email).Msg("Failed to create new user")
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user account")
			}
		} else {
			// User exists in this tenant, update OAuth info and last login
			if !isActive {
				return echo.NewHTTPError(http.StatusUnauthorized, "user account is inactive")
			}

			// Update OAuth info and last login
			_, err = h.DB.Exec(`
				UPDATE users 
				SET oauth_provider = $1, oauth_id = $2, avatar_url = $3, last_login = $4, updated_at = $4
				WHERE id = $5
			`, "google", googleUser.ID, googleUser.Picture, time.Now(), userID)

			if err != nil {
				log.Error().Err(err).Str("user_id", userID).Msg("Failed to update user OAuth info")
			}
		}

		needsTenant = false
	} else {
		// Scenario 2: OAuth without tenant (new user flow)
		log.Info().Msg("OAuth without tenant - new user flow")

		// Check if user exists in any tenant
		err = h.DB.QueryRow(`
			SELECT u.id, u.name, u.role, u.is_active, t.id, t.name, t.slug
			FROM users u
			INNER JOIN tenants t ON u.tenant_id = t.id
			WHERE u.email = $1 AND u.is_active = true AND t.is_active = true
			ORDER BY u.created_at ASC
			LIMIT 1
		`, googleUser.Email).Scan(&userID, &name, &role, &isActive, &tenantID, &tenantName, &tenantSlug)

		if err != nil {
			// User doesn't exist anywhere, create user and assign to default tenant
			isNewUser = true
			role = "ADMIN" // Promote to ADMIN for new users
			name = googleUser.Name
			needsTenant = false

			// Get the default tenant
			err = h.DB.QueryRow(`
				SELECT id, name, slug FROM tenants 
				WHERE slug = 'default' AND is_active = true
			`).Scan(&tenantID, &tenantName, &tenantSlug)

			if err != nil {
				// If default tenant doesn't exist, create it
				log.Info().Msg("Default tenant not found, creating one")
				err = h.DB.QueryRow(`
					INSERT INTO tenants (name, slug, is_active, settings, contact)
					VALUES ($1, $2, $3, $4, $5)
					RETURNING id, name, slug
				`, "Default Company", "default", true,
					`{"currency": "USD", "timezone": "UTC"}`,
					`{"email": "", "phone": ""}`).Scan(&tenantID, &tenantName, &tenantSlug)

				if err != nil {
					log.Error().Err(err).Msg("Failed to create default tenant")
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to create default tenant")
				}
			}

			// Insert new user with default tenant
			err = h.DB.QueryRow(`
				INSERT INTO users (email, name, role, tenant_id, oauth_provider, oauth_id, avatar_url, is_active)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING id
			`, googleUser.Email, googleUser.Name, role, tenantID, "google", googleUser.ID, googleUser.Picture, true).Scan(&userID)

			if err != nil {
				log.Error().Err(err).Str("email", googleUser.Email).Msg("Failed to create new user")
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user account")
			}

			log.Info().Str("user_id", userID).Str("tenant_id", tenantID).Msg("New OAuth user created and assigned to default tenant")
		} else {
			// User exists in a tenant, update OAuth info
			if !isActive {
				return echo.NewHTTPError(http.StatusUnauthorized, "user account is inactive")
			}

			// Update OAuth info and last login
			_, err = h.DB.Exec(`
				UPDATE users 
				SET oauth_provider = $1, oauth_id = $2, avatar_url = $3, last_login = $4, updated_at = $4
				WHERE id = $5
			`, "google", googleUser.ID, googleUser.Picture, time.Now(), userID)

			if err != nil {
				log.Error().Err(err).Str("user_id", userID).Msg("Failed to update user OAuth info")
			}

			needsTenant = false
		}
	}

	// Generate tokens
	accessToken, err := h.generateToken(
		userID,
		tenantID,
		googleUser.Email,
		role,
		h.Config.JWTExpiry,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate access token")
	}

	refreshToken, err := h.generateToken(
		userID,
		tenantID,
		googleUser.Email,
		role,
		h.Config.RefreshExpiry,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate refresh token")
	}

	response := GoogleOAuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(h.Config.JWTExpiry.Seconds()),
		User: UserResponse{
			ID:       userID,
			Name:     name,
			Email:    googleUser.Email,
			Role:     role,
			TenantID: tenantID,
		},
		IsNewUser:   isNewUser,
		NeedsTenant: needsTenant,
	}

	// Only include tenant info if user has a tenant
	if !needsTenant && tenantID != "" {
		response.Tenant = &TenantResponse{
			ID:   tenantID,
			Name: tenantName,
			Slug: tenantSlug,
		}
	}

	return c.JSON(http.StatusOK, response)
}

// SelectTenantForOAuthUser allows OAuth users to select or create a tenant
func (h *Handler) SelectTenantForOAuthUser(c echo.Context) error {
	// Get user from JWT context
	user := c.Get("user").(*middleware.Claims)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not authenticated")
	}

	var req struct {
		Action       string `json:"action" validate:"required,oneof=select create"`
		TenantSlug   string `json:"tenant_slug,omitempty"`
		TenantName   string `json:"tenant_name,omitempty"`
		TenantDomain string `json:"tenant_domain,omitempty"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Action == "select" && req.TenantSlug == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "tenant_slug required for selection")
	}

	if req.Action == "create" && (req.TenantName == "" || req.TenantSlug == "") {
		return echo.NewHTTPError(http.StatusBadRequest, "tenant_name and tenant_slug required for creation")
	}

	var tenantID, tenantName, tenantSlug string

	if req.Action == "select" {
		// User wants to join existing tenant
		err := h.DB.QueryRow(`
			SELECT id, name, slug FROM tenants 
			WHERE slug = $1 AND is_active = true
		`, req.TenantSlug).Scan(&tenantID, &tenantName, &tenantSlug)

		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "tenant not found or inactive")
		}

		// Check if user already exists in this tenant
		var existingUserID string
		err = h.DB.QueryRow(`
			SELECT id FROM users WHERE email = $1 AND tenant_id = $2
		`, user.Email, tenantID).Scan(&existingUserID)

		if err == nil {
			return echo.NewHTTPError(http.StatusConflict, "user already exists in this tenant")
		}
	} else {
		// User wants to create new tenant
		// Check if tenant slug is available
		var existingTenantID string
		err := h.DB.QueryRow(`
			SELECT id FROM tenants WHERE slug = $1
		`, req.TenantSlug).Scan(&existingTenantID)

		if err == nil {
			return echo.NewHTTPError(http.StatusConflict, "tenant slug already exists")
		}

		// Create new tenant
		err = h.DB.QueryRow(`
			INSERT INTO tenants (name, slug, domain, is_active, settings, contact)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, name, slug
		`, req.TenantName, req.TenantSlug, req.TenantDomain, true,
			`{"currency": "USD", "timezone": "UTC"}`,
			`{"email": "", "phone": ""}`).Scan(&tenantID, &tenantName, &tenantSlug)

		if err != nil {
			log.Error().Err(err).Msg("Failed to create tenant")
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create tenant")
		}
	}

	// Update user with tenant_id and promote to ADMIN role
	_, err := h.DB.Exec(`
		UPDATE users 
		SET tenant_id = $1, role = $2, updated_at = $3
		WHERE id = $4
	`, tenantID, "ADMIN", time.Now(), user.UserID)

	if err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Msg("Failed to assign user to tenant")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to assign user to tenant")
	}

	// Generate new tokens with updated tenant info
	accessToken, err := h.generateToken(
		user.UserID,
		tenantID,
		user.Email,
		"ADMIN", // New role
		h.Config.JWTExpiry,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate access token")
	}

	refreshToken, err := h.generateToken(
		user.UserID,
		tenantID,
		user.Email,
		"ADMIN", // New role
		h.Config.RefreshExpiry,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate refresh token")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    int(h.Config.JWTExpiry.Seconds()),
		"user": UserResponse{
			ID:       user.UserID,
			Name:     "", // Will be updated from database
			Email:    user.Email,
			Role:     "ADMIN",
			TenantID: tenantID,
		},
		"tenant": TenantResponse{
			ID:   tenantID,
			Name: tenantName,
			Slug: tenantSlug,
		},
		"message": "Tenant assigned successfully",
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

// exchangeCodeForToken exchanges authorization code for Google access token
func (h *Handler) exchangeCodeForToken(code, redirectURI string) (string, error) {
	url := "https://oauth2.googleapis.com/token"

	// Google expects form-encoded data, not JSON
	data := make(map[string][]string)
	data["client_id"] = []string{h.Config.GoogleClientID}
	data["client_secret"] = []string{h.Config.GoogleClientSecret}
	data["code"] = []string{code}
	data["grant_type"] = []string{"authorization_code"}
	data["redirect_uri"] = []string{redirectURI}

	// Log the request for debugging (remove in production)
	log.Info().
		Str("client_id", h.Config.GoogleClientID).
		Str("redirect_uri", redirectURI).
		Str("code_length", fmt.Sprintf("%d", len(code))).
		Msg("Exchanging Google OAuth code for token")

	// Convert map to form-encoded string
	formData := make([]string, 0, len(data))
	for key, values := range data {
		for _, value := range values {
			formData = append(formData, fmt.Sprintf("%s=%s", key, value))
		}
	}
	formString := strings.Join(formData, "&")

	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(formString))
	if err != nil {
		return "", fmt.Errorf("failed to make request to Google: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	log.Info().
		Int("status_code", resp.StatusCode).
		Str("response_body", string(body)).
		Msg("Google OAuth response")

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Google OAuth error: %d - %s", resp.StatusCode, string(body))
	}

	var tokenResp map[string]interface{}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse response JSON: %w", err)
	}

	accessToken, ok := tokenResp["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("no access token in response: %+v", tokenResp)
	}

	return accessToken, nil
}

// getGoogleUserInfo retrieves user information from Google
func (h *Handler) getGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	url := "https://www.googleapis.com/oauth2/v2/userinfo"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
