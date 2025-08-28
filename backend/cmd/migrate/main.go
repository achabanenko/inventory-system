package main

import (
	"context"
	"database/sql"
	"fmt"
	"inventory/internal/config"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Create database schema
	if err := createSchema(ctx, db); err != nil {
		log.Fatal("Failed to create schema:", err)
	}

	// Add OAuth fields to users table
	log.Println("Adding OAuth fields to users table...")

	_, err = db.Exec(`
		ALTER TABLE users 
		ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50),
		ADD COLUMN IF NOT EXISTS oauth_id VARCHAR(255),
		ADD COLUMN IF NOT EXISTS avatar_url TEXT;
	`)

	if err != nil {
		log.Printf("Warning: Failed to add OAuth fields: %v", err)
	} else {
		log.Println("OAuth fields added successfully")
	}

	// Make password_hash optional for OAuth users
	log.Println("Making password_hash optional...")

	_, err = db.Exec(`
		ALTER TABLE users 
		ALTER COLUMN password_hash DROP NOT NULL;
	`)

	if err != nil {
		log.Printf("Warning: Failed to make password_hash optional: %v", err)
	} else {
		log.Println("password_hash made optional successfully")
	}

	// Add tenant_id to transfers table
	log.Println("Adding tenant_id to transfers table...")
	_, err = db.Exec(`
		ALTER TABLE transfers
		ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id)
	`)
	if err != nil {
		log.Printf("Warning: Failed to add tenant_id to transfers: %v", err)
	} else {
		log.Println("tenant_id added to transfers successfully")
	}

	// Add tenant_id to transfer_lines table
	log.Println("Adding tenant_id to transfer_lines table...")
	_, err = db.Exec(`
		ALTER TABLE transfer_lines
		ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id)
	`)
	if err != nil {
		log.Printf("Warning: Failed to add tenant_id to transfer_lines: %v", err)
	} else {
		log.Println("tenant_id added to transfer_lines successfully")
	}

	// Add item_identifier to transfer_lines table for non-existent items
	log.Println("Adding item_identifier to transfer_lines table...")
	_, err = db.Exec(`
		ALTER TABLE transfer_lines
		ADD COLUMN IF NOT EXISTS item_identifier VARCHAR(255)
	`)
	if err != nil {
		log.Printf("Warning: Failed to add item_identifier to transfer_lines: %v", err)
	} else {
		log.Println("item_identifier added to transfer_lines successfully")
	}

	// Make item_id nullable for transfer_lines
	log.Println("Making item_id nullable in transfer_lines table...")
	_, err = db.Exec(`
		ALTER TABLE transfer_lines
		ALTER COLUMN item_id DROP NOT NULL
	`)
	if err != nil {
		log.Printf("Warning: Failed to make item_id nullable: %v", err)
	} else {
		log.Println("item_id made nullable in transfer_lines successfully")
	}

	// Add description field to transfer_lines
	log.Println("Adding description to transfer_lines table...")
	_, err = db.Exec(`
		ALTER TABLE transfer_lines
		ADD COLUMN IF NOT EXISTS description TEXT
	`)
	if err != nil {
		log.Printf("Warning: Failed to add description to transfer_lines: %v", err)
	} else {
		log.Println("description added to transfer_lines successfully")
	}

	fmt.Println("Database migration completed successfully!")
}

func createSchema(ctx context.Context, db *sql.DB) error {
	// Create tables in the correct order (respecting foreign key constraints)
	queries := []string{
		// Categories table
		`CREATE TABLE IF NOT EXISTS categories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			parent_id UUID REFERENCES categories(id),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255),
			name VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL CHECK (role IN ('ADMIN', 'MANAGER', 'CLERK')),
			tenant_id UUID REFERENCES tenants(id),
			oauth_provider VARCHAR(50),
			oauth_id VARCHAR(255),
			avatar_url TEXT,
			is_active BOOLEAN DEFAULT TRUE,
			last_login TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Suppliers table
		`CREATE TABLE IF NOT EXISTS suppliers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			code VARCHAR(50) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			contact JSONB,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Locations table
		`CREATE TABLE IF NOT EXISTS locations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			code VARCHAR(50) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			address JSONB,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Items table
		`CREATE TABLE IF NOT EXISTS items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			sku VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			barcode VARCHAR(255) UNIQUE,
			uom VARCHAR(50) NOT NULL,
			category_id UUID REFERENCES categories(id),
			cost NUMERIC(10,2) DEFAULT 0,
			price NUMERIC(10,2) DEFAULT 0,
			attributes JSONB,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			deleted_at TIMESTAMP WITH TIME ZONE
		)`,

		// Inventory levels table
		`CREATE TABLE IF NOT EXISTS inventory_levels (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			item_id UUID NOT NULL REFERENCES items(id),
			location_id UUID NOT NULL REFERENCES locations(id),
			on_hand INTEGER DEFAULT 0 CHECK (on_hand >= 0),
			allocated INTEGER DEFAULT 0 CHECK (allocated >= 0),
			reorder_point INTEGER DEFAULT 0 CHECK (reorder_point >= 0),
			reorder_qty INTEGER DEFAULT 0 CHECK (reorder_qty >= 0),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(item_id, location_id)
		)`,

		// Stock movements table
		`CREATE TABLE IF NOT EXISTS stock_movements (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			item_id UUID NOT NULL REFERENCES items(id),
			location_id UUID NOT NULL REFERENCES locations(id),
			user_id UUID REFERENCES users(id),
			qty INTEGER NOT NULL,
			reason VARCHAR(50) NOT NULL CHECK (reason IN ('PO_RECEIPT', 'ADJUSTMENT', 'TRANSFER_OUT', 'TRANSFER_IN', 'COUNT')),
			reference VARCHAR(255),
			ref_id UUID,
			meta JSONB,
			occurred_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Purchase orders table
		`CREATE TABLE IF NOT EXISTS purchase_orders (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			number VARCHAR(255) UNIQUE NOT NULL,
			supplier_id UUID NOT NULL REFERENCES suppliers(id),
			status VARCHAR(50) DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'APPROVED', 'PARTIAL', 'RECEIVED', 'CLOSED', 'CANCELED')),
			expected_at TIMESTAMP WITH TIME ZONE,
			notes TEXT,
			created_by UUID REFERENCES users(id),
			approved_by UUID REFERENCES users(id),
			approved_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Purchase order lines table
		`CREATE TABLE IF NOT EXISTS purchase_order_lines (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			purchase_order_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
			item_id UUID NOT NULL REFERENCES items(id),
			qty_ordered INTEGER NOT NULL CHECK (qty_ordered > 0),
			qty_received INTEGER DEFAULT 0 CHECK (qty_received >= 0),
			unit_cost NUMERIC(10,2) NOT NULL,
			tax JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Goods receipts header
		`CREATE TABLE IF NOT EXISTS goods_receipts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			number VARCHAR(255) UNIQUE NOT NULL,
			supplier_id UUID REFERENCES suppliers(id),
			location_id UUID REFERENCES locations(id),
			status VARCHAR(50) NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT','APPROVED','POSTED','CLOSED','CANCELED')),
			reference VARCHAR(255),
			notes TEXT,
			created_by UUID REFERENCES users(id),
			approved_by UUID REFERENCES users(id),
			posted_by UUID REFERENCES users(id),
			approved_at TIMESTAMP WITH TIME ZONE,
			posted_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Goods receipts lines
		`CREATE TABLE IF NOT EXISTS goods_receipt_lines (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			receipt_id UUID NOT NULL REFERENCES goods_receipts(id) ON DELETE CASCADE,
			item_id UUID NOT NULL REFERENCES items(id),
			qty INTEGER NOT NULL CHECK (qty > 0),
			unit_cost NUMERIC(10,2) DEFAULT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Transfers table
		`CREATE TABLE IF NOT EXISTS transfers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			number VARCHAR(255) UNIQUE NOT NULL,
			from_location_id UUID NOT NULL REFERENCES locations(id),
			to_location_id UUID NOT NULL REFERENCES locations(id),
			tenant_id UUID REFERENCES tenants(id),
			status VARCHAR(50) DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'IN_TRANSIT', 'RECEIVED', 'CANCELED')),
			notes TEXT,
			created_by UUID REFERENCES users(id),
			approved_by UUID REFERENCES users(id),
			shipped_at TIMESTAMP WITH TIME ZONE,
			received_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			CHECK (from_location_id != to_location_id)
		)`,

		// Transfer lines table
		`CREATE TABLE IF NOT EXISTS transfer_lines (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			transfer_id UUID NOT NULL REFERENCES transfers(id) ON DELETE CASCADE,
			item_id UUID REFERENCES items(id),
			item_identifier VARCHAR(255),
			tenant_id UUID REFERENCES tenants(id),
			qty INTEGER NOT NULL CHECK (qty > 0),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Adjustments table
		`CREATE TABLE IF NOT EXISTS adjustments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			number VARCHAR(255) UNIQUE NOT NULL,
			location_id UUID NOT NULL REFERENCES locations(id),
			tenant_id UUID REFERENCES tenants(id),
			reason VARCHAR(50) NOT NULL CHECK (reason IN ('COUNT', 'DAMAGE', 'CORRECTION', 'EXPIRY', 'THEFT', 'OTHER')),
			status VARCHAR(50) DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'APPROVED', 'CANCELED')),
			notes TEXT,
			created_by UUID REFERENCES users(id),
			approved_by UUID REFERENCES users(id),
			approved_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Adjustment lines table
		`CREATE TABLE IF NOT EXISTS adjustment_lines (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			adjustment_id UUID NOT NULL REFERENCES adjustments(id) ON DELETE CASCADE,
			item_id UUID REFERENCES items(id),
			item_identifier VARCHAR(255),
			tenant_id UUID REFERENCES tenants(id),
			qty_expected INTEGER NOT NULL DEFAULT 0,
			qty_actual INTEGER NOT NULL DEFAULT 0,
			qty_diff INTEGER NOT NULL,
			notes TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Stock count batches
		`CREATE TABLE IF NOT EXISTS count_batches (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			number VARCHAR(255) UNIQUE NOT NULL,
			location_id UUID NOT NULL REFERENCES locations(id),
			status VARCHAR(50) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN','IN_PROGRESS','COMPLETED','CANCELED')),
			notes TEXT,
			created_by UUID REFERENCES users(id),
			completed_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Stock count lines
		`CREATE TABLE IF NOT EXISTS count_lines (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			batch_id UUID NOT NULL REFERENCES count_batches(id) ON DELETE CASCADE,
			item_id UUID NOT NULL REFERENCES items(id),
			expected_on_hand INTEGER NOT NULL DEFAULT 0,
			counted_qty INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Audit logs table
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id),
			action VARCHAR(255) NOT NULL,
			entity VARCHAR(255) NOT NULL,
			entity_id UUID NOT NULL,
			before JSONB,
			after JSONB,
			at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Create indexes for better performance
		`CREATE INDEX IF NOT EXISTS idx_items_sku ON items(sku)`,
		`CREATE INDEX IF NOT EXISTS idx_items_barcode ON items(barcode) WHERE barcode IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_items_category ON items(category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_items_active ON items(is_active) WHERE is_active = TRUE`,
		`CREATE INDEX IF NOT EXISTS idx_inventory_item_location ON inventory_levels(item_id, location_id)`,
		`CREATE INDEX IF NOT EXISTS idx_stock_movements_item ON stock_movements(item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_stock_movements_location ON stock_movements(location_id)`,
		`CREATE INDEX IF NOT EXISTS idx_stock_movements_occurred ON stock_movements(occurred_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity, entity_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_at ON audit_logs(at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active) WHERE is_active = TRUE`,
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query: %w\nQuery: %s", err, query)
		}
	}

	// Apply migrations for existing databases
	if err := migrateGoodsReceipts(ctx, db); err != nil {
		return fmt.Errorf("failed to migrate goods receipts: %w", err)
	}

	if err := migrateAdjustments(ctx, db); err != nil {
		return fmt.Errorf("failed to migrate adjustments: %w", err)
	}

	if err := migrateUserOAuth(ctx, db); err != nil {
		return fmt.Errorf("failed to migrate user OAuth fields: %w", err)
	}

	return nil
}

func migrateGoodsReceipts(ctx context.Context, db *sql.DB) error {
	// Check if new columns exist
	var hasApprovedBy bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'goods_receipts' AND column_name = 'approved_by'
		)
	`).Scan(&hasApprovedBy)
	if err != nil {
		return fmt.Errorf("failed to check for approved_by column: %w", err)
	}

	if !hasApprovedBy {
		// Add new columns
		migrations := []string{
			`ALTER TABLE goods_receipts ADD COLUMN IF NOT EXISTS approved_by UUID REFERENCES users(id)`,
			`ALTER TABLE goods_receipts ADD COLUMN IF NOT EXISTS posted_by UUID REFERENCES users(id)`,
			`ALTER TABLE goods_receipts ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP WITH TIME ZONE`,
		}

		// Update status constraint to include new statuses
		updateStatusConstraint := `
			ALTER TABLE goods_receipts DROP CONSTRAINT IF EXISTS goods_receipts_status_check;
			ALTER TABLE goods_receipts ADD CONSTRAINT goods_receipts_status_check 
				CHECK (status IN ('DRAFT','APPROVED','POSTED','CLOSED','CANCELED'))
		`
		migrations = append(migrations, updateStatusConstraint)

		for _, migration := range migrations {
			if _, err := db.ExecContext(ctx, migration); err != nil {
				return fmt.Errorf("failed to apply goods receipts migration: %w\nQuery: %s", err, migration)
			}
		}
	}

	// Migrate goods_receipt_lines to make unit_cost nullable
	if err := migrateGoodsReceiptLines(ctx, db); err != nil {
		return fmt.Errorf("failed to migrate goods receipt lines: %w", err)
	}

	return nil
}

func migrateGoodsReceiptLines(ctx context.Context, db *sql.DB) error {
	// Check if unit_cost column is already nullable
	var isNullable string
	err := db.QueryRowContext(ctx, `
		SELECT is_nullable FROM information_schema.columns 
		WHERE table_name = 'goods_receipt_lines' AND column_name = 'unit_cost'
	`).Scan(&isNullable)
	if err != nil {
		return fmt.Errorf("failed to check unit_cost column: %w", err)
	}

	if isNullable == "NO" {
		// Make unit_cost nullable
		if _, err := db.ExecContext(ctx, `ALTER TABLE goods_receipt_lines ALTER COLUMN unit_cost DROP NOT NULL`); err != nil {
			return fmt.Errorf("failed to make unit_cost nullable: %w", err)
		}
	}

	return nil
}

func migrateAdjustments(ctx context.Context, db *sql.DB) error {
	log.Println("Migrating adjustments table...")

	// Add tenant_id to adjustments table if not exists
	_, err := db.ExecContext(ctx, `
		ALTER TABLE adjustments
		ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id)
	`)
	if err != nil {
		log.Printf("Warning: Failed to add tenant_id to adjustments: %v", err)
	}

	// Update reason check constraint to include new values
	_, err = db.ExecContext(ctx, `
		ALTER TABLE adjustments
		DROP CONSTRAINT IF EXISTS adjustments_reason_check
	`)
	if err != nil {
		log.Printf("Warning: Failed to drop old reason constraint: %v", err)
	}

	_, err = db.ExecContext(ctx, `
		ALTER TABLE adjustments
		ADD CONSTRAINT adjustments_reason_check
		CHECK (reason IN ('COUNT', 'DAMAGE', 'CORRECTION', 'EXPIRY', 'THEFT', 'OTHER'))
	`)
	if err != nil {
		log.Printf("Warning: Failed to add new reason constraint: %v", err)
	}

	// Add new columns to adjustment_lines table
	alterQueries := []string{
		"ALTER TABLE adjustment_lines ADD COLUMN IF NOT EXISTS item_identifier VARCHAR(255)",
		"ALTER TABLE adjustment_lines ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id)",
		"ALTER TABLE adjustment_lines ADD COLUMN IF NOT EXISTS qty_expected INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE adjustment_lines ADD COLUMN IF NOT EXISTS qty_actual INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE adjustment_lines ADD COLUMN IF NOT EXISTS notes TEXT",
	}

	for _, query := range alterQueries {
		_, err = db.ExecContext(ctx, query)
		if err != nil {
			log.Printf("Warning: Failed to execute: %s - %v", query, err)
		}
	}

	log.Println("Adjustments migration completed")
	return nil
}

func migrateUserOAuth(ctx context.Context, db *sql.DB) error {
	log.Println("Migrating user OAuth fields...")

	alterQueries := []string{
		"ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_id VARCHAR(255)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT",
	}

	for _, query := range alterQueries {
		_, err := db.ExecContext(ctx, query)
		if err != nil {
			log.Printf("Warning: Failed to execute: %s - %v", query, err)
		}
	}

	log.Println("User OAuth migration completed")
	return nil
}
