package main

import (
	"context"
	"database/sql"
	"fmt"
	"inventory/internal/config"
	"log"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
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

	if err := db.PingContext(ctx); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	if err := seedData(ctx, db); err != nil {
		log.Fatal("Failed to seed data:", err)
	}

	fmt.Println("Database seeded successfully!")
}

func seedData(ctx context.Context, db *sql.DB) error {
	// Hash password for admin user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Seed data in order of dependencies
	queries := []string{
		// Insert default admin user
		fmt.Sprintf(`
			INSERT INTO users (email, password_hash, name, role, is_active) 
			VALUES ('admin@example.com', '%s', 'Admin User', 'ADMIN', true)
			ON CONFLICT (email) DO NOTHING
		`, string(hashedPassword)),

		// Insert sample categories
		`INSERT INTO categories (name) VALUES 
			('Electronics'),
			('Office Supplies'),
			('Hardware'),
			('Consumables')`,

		// Insert sample locations
		`INSERT INTO locations (code, name, address, is_active) VALUES 
			('WH01', 'Main Warehouse', '{"street": "123 Main St", "city": "Anytown", "zip": "12345"}', true),
			('WH02', 'Secondary Warehouse', '{"street": "456 Oak Ave", "city": "Somewhere", "zip": "67890"}', true),
			('STORE', 'Retail Store', '{"street": "789 Commerce Blvd", "city": "Downtown", "zip": "54321"}', true)
			ON CONFLICT (code) DO NOTHING`,

		// Insert sample suppliers
		`INSERT INTO suppliers (code, name, contact, is_active) VALUES 
			('SUP001', 'Tech Solutions Inc', '{"email": "orders@techsolutions.com", "phone": "555-0123"}', true),
			('SUP002', 'Office Pro Supply', '{"email": "sales@officepro.com", "phone": "555-0456"}', true),
			('SUP003', 'Industrial Hardware Co', '{"email": "info@industrialhardware.com", "phone": "555-0789"}', true)
			ON CONFLICT (code) DO NOTHING`,

		// Insert sample items
		`INSERT INTO items (sku, name, barcode, uom, cost, price, is_active) VALUES 
			('LAPTOP-001', 'Business Laptop', '1234567890123', 'each', 800.00, 1200.00, true),
			('MOUSE-001', 'Wireless Mouse', '2345678901234', 'each', 15.00, 25.00, true),
			('PAPER-001', 'Copy Paper A4', '3456789012345', 'ream', 3.50, 6.00, true),
			('PEN-001', 'Blue Ballpoint Pen', '4567890123456', 'each', 0.25, 0.75, true),
			('MONITOR-001', '24" LCD Monitor', '5678901234567', 'each', 150.00, 250.00, true)
			ON CONFLICT (sku) DO NOTHING`,
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to execute seed query: %w\nQuery: %s", err, query)
		}
	}

	// Insert inventory levels for sample items
	inventoryQuery := `
		INSERT INTO inventory_levels (item_id, location_id, on_hand, allocated, reorder_point, reorder_qty)
		SELECT i.id, l.id, 
			CASE 
				WHEN i.sku LIKE 'LAPTOP%' THEN 10
				WHEN i.sku LIKE 'MOUSE%' THEN 50
				WHEN i.sku LIKE 'PAPER%' THEN 100
				WHEN i.sku LIKE 'PEN%' THEN 500
				WHEN i.sku LIKE 'MONITOR%' THEN 25
				ELSE 0
			END as on_hand,
			0 as allocated,
			CASE 
				WHEN i.sku LIKE 'LAPTOP%' THEN 5
				WHEN i.sku LIKE 'MOUSE%' THEN 20
				WHEN i.sku LIKE 'PAPER%' THEN 50
				WHEN i.sku LIKE 'PEN%' THEN 200
				WHEN i.sku LIKE 'MONITOR%' THEN 10
				ELSE 5
			END as reorder_point,
			CASE 
				WHEN i.sku LIKE 'LAPTOP%' THEN 10
				WHEN i.sku LIKE 'MOUSE%' THEN 50
				WHEN i.sku LIKE 'PAPER%' THEN 100
				WHEN i.sku LIKE 'PEN%' THEN 500
				WHEN i.sku LIKE 'MONITOR%' THEN 25
				ELSE 10
			END as reorder_qty
		FROM items i
		CROSS JOIN locations l
		WHERE i.is_active = true AND l.is_active = true
		ON CONFLICT (item_id, location_id) DO NOTHING
	`

	if _, err := db.ExecContext(ctx, inventoryQuery); err != nil {
		return fmt.Errorf("failed to seed inventory levels: %w", err)
	}

	return nil
}
