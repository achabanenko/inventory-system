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

	if err := db.PingContext(ctx); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Insert default tenant
	_, err = db.ExecContext(ctx, `
		INSERT INTO tenants (id, name, slug, is_active, settings, contact)
		VALUES (
			gen_random_uuid(),
			'Default Company', 
			'default', 
			true, 
			'{"currency": "USD", "timezone": "UTC"}',
			'{"email": "", "phone": ""}'
		) 
		ON CONFLICT (slug) DO NOTHING
	`)
	if err != nil {
		log.Fatal("Failed to insert default tenant:", err)
	}

	// Insert admin user
	result, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, name, role, tenant_id, is_active)
		SELECT 
			gen_random_uuid(),
			'admin@example.com',
			'$2a$10$yQChQTp1s49xoVrXMkJRFu8dZJY6mHhbaGyg85QmJW5omMmgW29pG',
			'Admin User',
			'ADMIN',
			t.id,
			true
		FROM tenants t 
		WHERE t.slug = 'default'
		AND NOT EXISTS (SELECT 1 FROM users WHERE email = 'admin@example.com')
	`)
	if err != nil {
		log.Fatal("Failed to insert admin user:", err)
	}

	affected, _ := result.RowsAffected()
	if affected > 0 {
		fmt.Println("Admin user created successfully!")
		fmt.Println("Email: admin@example.com")
		fmt.Println("Password: admin123")
	} else {
		fmt.Println("Admin user already exists or no default tenant found")
	}
}