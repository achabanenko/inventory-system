package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type TenantService struct {
	db *sql.DB
}

type Tenant struct {
	ID       uuid.UUID              `json:"id"`
	Name     string                 `json:"name"`
	Slug     string                 `json:"slug"`
	Domain   *string                `json:"domain"`
	Settings map[string]interface{} `json:"settings"`
	Contact  map[string]interface{} `json:"contact"`
	IsActive bool                   `json:"is_active"`
}

func NewTenantService(db *sql.DB) *TenantService {
	return &TenantService{db: db}
}

// CreateTenant creates a new tenant
func (s *TenantService) CreateTenant(ctx context.Context, name, slug string) (*Tenant, error) {
	// Validate slug format (URL-safe)
	if !isValidSlug(slug) {
		return nil, fmt.Errorf("invalid slug format: must be URL-safe")
	}

	tenant := &Tenant{
		ID:       uuid.New(),
		Name:     name,
		Slug:     slug,
		IsActive: true,
	}

	query := `
		INSERT INTO tenants (id, name, slug, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, name, slug, domain, settings, contact, is_active
	`

	err := s.db.QueryRowContext(ctx, query, tenant.ID, tenant.Name, tenant.Slug, tenant.IsActive).
		Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Domain, &tenant.Settings, &tenant.Contact, &tenant.IsActive)

	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	return tenant, nil
}

// GetTenantByID retrieves a tenant by ID
func (s *TenantService) GetTenantByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	tenant := &Tenant{}

	query := `
		SELECT id, name, slug, domain, settings, contact, is_active
		FROM tenants
		WHERE id = $1 AND is_active = true
	`

	err := s.db.QueryRowContext(ctx, query, id).
		Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Domain, &tenant.Settings, &tenant.Contact, &tenant.IsActive)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// GetTenantBySlug retrieves a tenant by slug
func (s *TenantService) GetTenantBySlug(ctx context.Context, slug string) (*Tenant, error) {
	tenant := &Tenant{}

	query := `
		SELECT id, name, slug, domain, settings, contact, is_active
		FROM tenants
		WHERE slug = $1 AND is_active = true
	`

	err := s.db.QueryRowContext(ctx, query, slug).
		Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Domain, &tenant.Settings, &tenant.Contact, &tenant.IsActive)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// ListTenants returns all active tenants
func (s *TenantService) ListTenants(ctx context.Context) ([]*Tenant, error) {
	query := `
		SELECT id, name, slug, domain, settings, contact, is_active
		FROM tenants
		WHERE is_active = true
		ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*Tenant
	for rows.Next() {
		tenant := &Tenant{}
		err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Domain, &tenant.Settings, &tenant.Contact, &tenant.IsActive)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenants = append(tenants, tenant)
	}

	return tenants, nil
}

// UpdateTenant updates a tenant's information
func (s *TenantService) UpdateTenant(ctx context.Context, id uuid.UUID, name, slug string, domain *string) (*Tenant, error) {
	if slug != "" && !isValidSlug(slug) {
		return nil, fmt.Errorf("invalid slug format: must be URL-safe")
	}

	query := `
		UPDATE tenants
		SET name = COALESCE(NULLIF($2, ''), name),
		    slug = COALESCE(NULLIF($3, ''), slug),
		    domain = $4,
		    updated_at = NOW()
		WHERE id = $1 AND is_active = true
		RETURNING id, name, slug, domain, settings, contact, is_active
	`

	tenant := &Tenant{}
	err := s.db.QueryRowContext(ctx, query, id, name, slug, domain).
		Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Domain, &tenant.Settings, &tenant.Contact, &tenant.IsActive)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant, nil
}

// DeactivateTenant deactivates a tenant (soft delete)
func (s *TenantService) DeactivateTenant(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE tenants
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}

// isValidSlug checks if a slug is URL-safe
func isValidSlug(slug string) bool {
	if slug == "" {
		return false
	}

	// Check if slug contains only valid characters: lowercase letters, numbers, hyphens
	for _, char := range slug {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return false
		}
	}

	// Check if slug starts or ends with hyphen
	if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
		return false
	}

	return true
}
