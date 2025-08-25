package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	TenantSlug  string `json:"tenant_slug,omitempty" validate:"max=50"`
	TenantName  string `json:"tenant_name,omitempty" validate:"max=100"`
	InviteToken string `json:"invite_token,omitempty"`
}

type RegisterResponse struct {
	User        UserResponse   `json:"user"`
	Tenant      TenantResponse `json:"tenant"`
	AccessToken string         `json:"access_token"`
	ExpiresIn   int            `json:"expires_in"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
}

type TenantResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// RegisterUser handles both new tenant creation and joining existing tenants
func (h *Handler) RegisterUser(c echo.Context) error {
	log.Info().Msg("Registration endpoint called")

	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		log.Error().Err(err).Msg("Failed to bind registration request")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	log.Info().
		Str("email", req.Email).
		Str("tenant_name", req.TenantName).
		Str("tenant_slug", req.TenantSlug).
		Bool("has_invite", req.InviteToken != "").
		Msg("Registration request received")

	// Validate request
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Name, email, and password are required")
	}

	// Sanitize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Determine registration type
	if req.InviteToken != "" {
		return h.registerWithInvite(c, req)
	} else if req.TenantName != "" {
		// Create new tenant (with optional custom slug)
		return h.registerNewTenant(c, req)
	} else if req.TenantSlug != "" {
		// Join existing tenant by slug
		return h.registerWithTenantSlug(c, req)
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, "Either tenant_name (to create new company) or tenant_slug (to join existing company) is required")
	}
}

// registerNewTenant creates a new tenant and makes the user an admin
func (h *Handler) registerNewTenant(c echo.Context, req RegisterRequest) error {
	if req.TenantName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant name is required for new tenant registration")
	}

	// Generate tenant slug from name if not provided
	if req.TenantSlug == "" {
		req.TenantSlug = generateSlugFromName(req.TenantName)
	}

	// Check if email already exists globally
	var existingUserCount int
	err := h.DB.QueryRow(`
		SELECT COUNT(*) FROM users WHERE email = $1
	`, req.Email).Scan(&existingUserCount)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to check existing users")
	}
	if existingUserCount > 0 {
		return echo.NewHTTPError(http.StatusConflict, "Email already registered. Try joining an existing tenant instead.")
	}

	// Start transaction
	tx, err := h.DB.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start transaction")
	}
	defer tx.Rollback()

	// Create tenant within transaction
	tenantID := uuid.New()

	log.Info().
		Str("tenant_name", req.TenantName).
		Str("tenant_slug", req.TenantSlug).
		Str("tenant_id", tenantID.String()).
		Msg("Creating tenant")

	_, err = tx.Exec(`
		INSERT INTO tenants (id, name, slug, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, true, NOW(), NOW())
	`, tenantID, req.TenantName, req.TenantSlug)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_name", req.TenantName).
			Str("tenant_slug", req.TenantSlug).
			Msg("Failed to create tenant")
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return echo.NewHTTPError(http.StatusConflict, "Tenant identifier already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create tenant: %v", err))
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to secure password")
	}

	// Create user as tenant admin
	userID := uuid.New()
	_, err = tx.Exec(`
		INSERT INTO users (id, tenant_id, email, password_hash, name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, true, NOW(), NOW())
	`, userID, tenantID, req.Email, string(hashedPassword), req.Name, "ADMIN")

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to complete registration")
	}

	// Generate JWT token
	accessToken, err := h.generateToken(
		userID.String(),
		tenantID.String(),
		req.Email,
		"ADMIN",
		h.Config.JWTExpiry,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate access token")
	}

	// Log successful registration
	log.Info().
		Str("user_email", req.Email).
		Str("tenant_slug", req.TenantSlug).
		Str("user_id", userID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("New tenant and admin user created")

	return c.JSON(http.StatusCreated, RegisterResponse{
		User: UserResponse{
			ID:       userID.String(),
			Name:     req.Name,
			Email:    req.Email,
			Role:     "ADMIN",
			TenantID: tenantID.String(),
		},
		Tenant: TenantResponse{
			ID:   tenantID.String(),
			Name: req.TenantName,
			Slug: req.TenantSlug,
		},
		AccessToken: accessToken,
		ExpiresIn:   int(h.Config.JWTExpiry.Seconds()),
	})
}

// registerWithTenantSlug allows users to join an existing tenant
func (h *Handler) registerWithTenantSlug(c echo.Context, req RegisterRequest) error {
	// Find tenant by slug
	var tenantID, tenantName, tenantSlug string
	var isActive bool
	err := h.DB.QueryRow(`
		SELECT id, name, slug, is_active
		FROM tenants
		WHERE slug = $1 AND is_active = true
	`, req.TenantSlug).Scan(&tenantID, &tenantName, &tenantSlug, &isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Company not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to lookup company")
	}

	// Check if email already exists in this tenant
	var existingUserCount int
	err = h.DB.QueryRow(`
		SELECT COUNT(*) FROM users WHERE email = $1 AND tenant_id = $2
	`, req.Email, tenantID).Scan(&existingUserCount)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to check existing users")
	}
	if existingUserCount > 0 {
		return echo.NewHTTPError(http.StatusConflict, "Email already registered in this company")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to secure password")
	}

	// Create user as clerk (default role for self-registration)
	userID := uuid.New()
	_, err = h.DB.Exec(`
		INSERT INTO users (id, tenant_id, email, password_hash, name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, true, NOW(), NOW())
	`, userID, tenantID, req.Email, string(hashedPassword), req.Name, "CLERK")

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
	}

	// Generate JWT token
	accessToken, err := h.generateToken(
		userID.String(),
		tenantID,
		req.Email,
		"CLERK",
		h.Config.JWTExpiry,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate access token")
	}

	// Log successful registration
	log.Info().
		Str("user_email", req.Email).
		Str("tenant_slug", tenantSlug).
		Str("user_id", userID.String()).
		Str("tenant_id", tenantID).
		Msg("User registered to existing tenant")

	return c.JSON(http.StatusCreated, RegisterResponse{
		User: UserResponse{
			ID:       userID.String(),
			Name:     req.Name,
			Email:    req.Email,
			Role:     "CLERK",
			TenantID: tenantID,
		},
		Tenant: TenantResponse{
			ID:   tenantID,
			Name: tenantName,
			Slug: tenantSlug,
		},
		AccessToken: accessToken,
		ExpiresIn:   int(h.Config.JWTExpiry.Seconds()),
	})
}

// registerWithInvite handles invitation-based registration
func (h *Handler) registerWithInvite(c echo.Context, req RegisterRequest) error {
	// TODO: Implement invitation system
	// This would validate the invite token and extract tenant/role information
	return echo.NewHTTPError(http.StatusNotImplemented, "Invitation-based registration not yet implemented")
}

// TenantLookup allows users to find their tenant by email
func (h *Handler) TenantLookup(c echo.Context) error {
	log.Info().Msg("Tenant lookup endpoint called")

	email := c.QueryParam("email")
	if email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Email parameter required")
	}

	email = strings.ToLower(strings.TrimSpace(email))

	// Find all tenants for this email
	rows, err := h.DB.Query(`
		SELECT t.id, t.name, t.slug, u.role
		FROM users u
		INNER JOIN tenants t ON u.tenant_id = t.id
		WHERE u.email = $1 AND u.is_active = true AND t.is_active = true
		ORDER BY u.created_at ASC
	`, email)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to lookup tenants")
	}
	defer rows.Close()

	type TenantLookupResult struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
		Role string `json:"role"`
	}

	var tenants []TenantLookupResult
	for rows.Next() {
		var tenant TenantLookupResult
		if err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Role); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to scan tenant data")
		}
		tenants = append(tenants, tenant)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tenants": tenants,
	})
}

// generateSlugFromName creates a URL-safe slug from a tenant name
func generateSlugFromName(name string) string {
	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters (keep only letters, numbers, and hyphens)
	var result strings.Builder
	for _, char := range slug {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}

	slug = result.String()

	// Remove leading/trailing hyphens and multiple consecutive hyphens
	slug = strings.Trim(slug, "-")
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Limit length
	if len(slug) > 50 {
		slug = slug[:50]
		slug = strings.TrimRight(slug, "-")
	}

	return slug
}
