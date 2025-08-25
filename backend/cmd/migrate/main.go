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
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL CHECK (role IN ('ADMIN', 'MANAGER', 'CLERK')),
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
			unit_cost NUMERIC(10,2) NOT NULL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Transfers table
		`CREATE TABLE IF NOT EXISTS transfers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			number VARCHAR(255) UNIQUE NOT NULL,
			from_location_id UUID NOT NULL REFERENCES locations(id),
			to_location_id UUID NOT NULL REFERENCES locations(id),
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
			item_id UUID NOT NULL REFERENCES items(id),
			qty INTEGER NOT NULL CHECK (qty > 0),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Adjustments table
		`CREATE TABLE IF NOT EXISTS adjustments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			number VARCHAR(255) UNIQUE NOT NULL,
			location_id UUID NOT NULL REFERENCES locations(id),
			reason VARCHAR(50) NOT NULL CHECK (reason IN ('COUNT', 'DAMAGE', 'CORRECTION')),
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
			item_id UUID NOT NULL REFERENCES items(id),
			qty_diff INTEGER NOT NULL,
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

	return nil
}
