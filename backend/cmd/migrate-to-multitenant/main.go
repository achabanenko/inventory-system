package main

import (
	"database/sql"
	"fmt"
	"log"

	"inventory/internal/config"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// First, create tenants table
	if err := createTenantsTable(db); err != nil {
		log.Fatal("Failed to create tenants table:", err)
	}

	// Create a default tenant for existing data
	defaultTenantID, err := createDefaultTenant(db)
	if err != nil {
		log.Fatal("Failed to create default tenant:", err)
	}

	// Add tenant_id columns to all tables
	if err := addTenantColumns(db, defaultTenantID); err != nil {
		log.Fatal("Failed to add tenant columns:", err)
	}

	// Update constraints and indexes
	if err := updateConstraintsAndIndexes(db); err != nil {
		log.Fatal("Failed to update constraints and indexes:", err)
	}

	fmt.Println("Migration to multi-tenant completed successfully!")
	fmt.Printf("Default tenant ID: %s\n", defaultTenantID)
}

func createTenantsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS tenants (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR NOT NULL,
			slug VARCHAR UNIQUE NOT NULL,
			domain VARCHAR UNIQUE,
			settings JSONB DEFAULT '{}',
			contact JSONB DEFAULT '{}',
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);
		CREATE INDEX IF NOT EXISTS idx_tenants_domain ON tenants(domain);
		CREATE INDEX IF NOT EXISTS idx_tenants_is_active ON tenants(is_active);
	`

	_, err := db.Exec(query)
	return err
}

func createDefaultTenant(db *sql.DB) (uuid.UUID, error) {
	tenantID := uuid.New()

	query := `
		INSERT INTO tenants (id, name, slug, is_active)
		VALUES ($1, 'Default Company', 'default', true)
		ON CONFLICT (slug) DO UPDATE SET updated_at = NOW()
		RETURNING id
	`

	var returnedID uuid.UUID
	err := db.QueryRow(query, tenantID).Scan(&returnedID)
	return returnedID, err
}

func addTenantColumns(db *sql.DB, defaultTenantID uuid.UUID) error {
	tables := []string{
		"users", "items", "categories", "locations", "suppliers",
		"inventory_levels", "stock_movements", "purchase_orders",
		"transfers", "adjustments", "audit_logs",
	}

	for _, table := range tables {
		// Add tenant_id column
		query := fmt.Sprintf(`
			ALTER TABLE %s 
			ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id);
		`, table)

		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to add tenant_id to %s: %w", table, err)
		}

		// Update existing records with default tenant
		updateQuery := fmt.Sprintf(`
			UPDATE %s SET tenant_id = $1 WHERE tenant_id IS NULL;
		`, table)

		if _, err := db.Exec(updateQuery, defaultTenantID); err != nil {
			return fmt.Errorf("failed to update tenant_id in %s: %w", table, err)
		}

		// Make tenant_id NOT NULL
		alterQuery := fmt.Sprintf(`
			ALTER TABLE %s ALTER COLUMN tenant_id SET NOT NULL;
		`, table)

		if _, err := db.Exec(alterQuery); err != nil {
			return fmt.Errorf("failed to set tenant_id NOT NULL in %s: %w", table, err)
		}

		fmt.Printf("Updated table: %s\n", table)
	}

	return nil
}

func updateConstraintsAndIndexes(db *sql.DB) error {
	queries := []string{
		// Drop old unique constraints
		"ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;",
		"ALTER TABLE items DROP CONSTRAINT IF EXISTS items_sku_key;",
		"ALTER TABLE items DROP CONSTRAINT IF EXISTS items_barcode_key;",
		"ALTER TABLE locations DROP CONSTRAINT IF EXISTS locations_code_key;",
		"ALTER TABLE suppliers DROP CONSTRAINT IF EXISTS suppliers_code_key;",
		"ALTER TABLE purchase_orders DROP CONSTRAINT IF EXISTS purchase_orders_number_key;",
		"ALTER TABLE transfers DROP CONSTRAINT IF EXISTS transfers_number_key;",
		"ALTER TABLE adjustments DROP CONSTRAINT IF EXISTS adjustments_number_key;",

		// Add new unique constraints with tenant_id
		"ALTER TABLE users ADD CONSTRAINT users_tenant_email_unique UNIQUE (tenant_id, email);",
		"ALTER TABLE items ADD CONSTRAINT items_tenant_sku_unique UNIQUE (tenant_id, sku);",
		"ALTER TABLE items ADD CONSTRAINT items_tenant_barcode_unique UNIQUE (tenant_id, barcode);",
		"ALTER TABLE locations ADD CONSTRAINT locations_tenant_code_unique UNIQUE (tenant_id, code);",
		"ALTER TABLE suppliers ADD CONSTRAINT suppliers_tenant_code_unique UNIQUE (tenant_id, code);",
		"ALTER TABLE purchase_orders ADD CONSTRAINT purchase_orders_tenant_number_unique UNIQUE (tenant_id, number);",
		"ALTER TABLE transfers ADD CONSTRAINT transfers_tenant_number_unique UNIQUE (tenant_id, number);",
		"ALTER TABLE adjustments ADD CONSTRAINT adjustments_tenant_number_unique UNIQUE (tenant_id, number);",
		"ALTER TABLE inventory_levels ADD CONSTRAINT inventory_levels_tenant_item_location_unique UNIQUE (tenant_id, item_id, location_id);",
		"ALTER TABLE categories ADD CONSTRAINT categories_tenant_name_unique UNIQUE (tenant_id, name);",

		// Add indexes for performance
		"CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);",
		"CREATE INDEX IF NOT EXISTS idx_items_tenant_id ON items(tenant_id);",
		"CREATE INDEX IF NOT EXISTS idx_categories_tenant_id ON categories(tenant_id);",
		"CREATE INDEX IF NOT EXISTS idx_locations_tenant_id ON locations(tenant_id);",
		"CREATE INDEX IF NOT EXISTS idx_suppliers_tenant_id ON suppliers(tenant_id);",
		"CREATE INDEX IF NOT EXISTS idx_inventory_levels_tenant_id ON inventory_levels(tenant_id);",
		"CREATE INDEX IF NOT EXISTS idx_stock_movements_tenant_id ON stock_movements(tenant_id);",
		"CREATE INDEX IF NOT EXISTS idx_purchase_orders_tenant_id ON purchase_orders(tenant_id);",
		"CREATE INDEX IF NOT EXISTS idx_transfers_tenant_id ON transfers(tenant_id);",
		"CREATE INDEX IF NOT EXISTS idx_adjustments_tenant_id ON adjustments(tenant_id);",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_id ON audit_logs(tenant_id);",
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			// Log error but continue with other queries
			fmt.Printf("Warning: %s failed: %v\n", query, err)
		}
	}

	return nil
}
