package main

import (
	"context"
	"database/sql"
	"fmt"
	"inventory/internal/config"
	"inventory/internal/handlers"
	"inventory/internal/middleware"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	setupLogger(cfg)

	db, err := setupDatabase(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	setupMiddleware(e, cfg)

	h := handlers.New(db, cfg)
	setupRoutes(e, h)

	startServer(e, cfg)
}

func setupLogger(cfg *config.Config) {
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.DebugLevel
	}

	zerolog.SetGlobalLevel(level)

	if cfg.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}

func setupDatabase(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().Msg("Connected to database")
	return db, nil
}

func setupMiddleware(e *echo.Echo, cfg *config.Config) {
	e.Use(middleware.Logger())
	e.Use(middleware.RequestID())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.Gzip())

	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: cfg.CORSOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))
}

func setupRoutes(e *echo.Echo, h *handlers.Handler) {
	api := e.Group("/api/v1")

	api.GET("/healthz", h.Health)
	api.GET("/readyz", h.Ready)

	auth := api.Group("/auth")
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/logout", h.Logout)
	auth.POST("/register", h.RegisterUser)
	auth.GET("/tenant-lookup", h.TenantLookup)

	log.Info().Msg("Routes configured: /api/v1/auth/login")

	// Current tenant info (requires JWT but not tenant context since it returns tenant info)
	me := api.Group("/me")
	me.Use(middleware.JWT(h.Config.JWTSecret))
	me.GET("/tenant", h.GetCurrentTenant)

	// Protected routes - each with explicit middleware
	items := api.Group("/items")
	items.Use(middleware.JWT(h.Config.JWTSecret))
	items.Use(middleware.RequireTenant())
	items.GET("", h.ListItems)
	items.POST("", h.CreateItem)
	items.GET("/:id", h.GetItem)
	items.PUT("/:id", h.UpdateItem)
	items.DELETE("/:id", h.DeleteItem)

	locations := api.Group("/locations")
	locations.Use(middleware.JWT(h.Config.JWTSecret))
	locations.Use(middleware.RequireTenant())
	locations.GET("", h.ListLocations)
	locations.POST("", h.CreateLocation)
	locations.GET("/:id", h.GetLocation)
	locations.PUT("/:id", h.UpdateLocation)
	locations.DELETE("/:id", h.DeleteLocation)

	suppliers := api.Group("/suppliers")
	suppliers.Use(middleware.JWT(h.Config.JWTSecret))
	suppliers.Use(middleware.RequireTenant())
	suppliers.GET("", h.ListSuppliers)
	suppliers.POST("", h.CreateSupplier)
	suppliers.GET("/:id", h.GetSupplier)
	suppliers.PUT("/:id", h.UpdateSupplier)
	suppliers.DELETE("/:id", h.DeleteSupplier)

	categories := api.Group("/categories")
	categories.Use(middleware.JWT(h.Config.JWTSecret))
	categories.Use(middleware.RequireTenant())
	categories.GET("", h.ListCategories)
	categories.POST("", h.CreateCategory)
	categories.GET("/:id", h.GetCategory)
	categories.PUT("/:id", h.UpdateCategory)
	categories.DELETE("/:id", h.DeleteCategory)

	inventory := api.Group("/inventory")
	inventory.Use(middleware.JWT(h.Config.JWTSecret))
	inventory.Use(middleware.RequireTenant())
	inventory.GET("", h.GetInventory)
	inventory.GET("/:item_id/locations", h.GetItemLocations)
	inventory.GET("/movements", h.GetMovements)

	purchaseOrders := api.Group("/purchase-orders")
	purchaseOrders.Use(middleware.JWT(h.Config.JWTSecret))
	purchaseOrders.Use(middleware.RequireTenant())
	purchaseOrders.GET("", h.ListPurchaseOrders)
	purchaseOrders.POST("", h.CreatePurchaseOrder)
	purchaseOrders.GET("/:id", h.GetPurchaseOrder)
	purchaseOrders.PUT("/:id", h.UpdatePurchaseOrder)
	purchaseOrders.DELETE("/:id", h.DeletePurchaseOrder)
	purchaseOrders.POST("/:id/approve", h.ApprovePurchaseOrder)
	purchaseOrders.POST("/:id/receive", h.ReceivePurchaseOrder)
	purchaseOrders.POST("/:id/close", h.ClosePurchaseOrder)

	transfers := api.Group("/transfers")
	transfers.Use(middleware.JWT(h.Config.JWTSecret))
	transfers.Use(middleware.RequireTenant())
	transfers.GET("", h.ListTransfers)
	transfers.POST("", h.CreateTransfer)
	transfers.GET("/:id", h.GetTransfer)
	transfers.POST("/:id/approve", h.ApproveTransfer)
	transfers.POST("/:id/ship", h.ShipTransfer)
	transfers.POST("/:id/receive", h.ReceiveTransfer)

	adjustments := api.Group("/adjustments")
	adjustments.Use(middleware.JWT(h.Config.JWTSecret))
	adjustments.Use(middleware.RequireTenant())
	adjustments.GET("", h.ListAdjustments)
	adjustments.POST("", h.CreateAdjustment)
	adjustments.GET("/:id", h.GetAdjustment)
	adjustments.POST("/:id/approve", h.ApproveAdjustment)

	// Goods Receipts
	receipts := api.Group("/receipts")
	receipts.Use(middleware.JWT(h.Config.JWTSecret))
	receipts.Use(middleware.RequireTenant())
	receipts.GET("", h.ListReceipts)
	receipts.POST("", h.CreateReceipt)
	receipts.GET("/:id", h.GetReceipt)
	receipts.PUT("/:id", h.UpdateReceipt)
	receipts.DELETE("/:id", h.DeleteReceipt)
	receipts.POST("/:id/approve", h.ApproveReceipt)
	receipts.POST("/:id/post", h.PostReceipt)
	receipts.POST("/:id/close", h.CloseReceipt)
	receipts.GET("/:id/lines", h.ListReceiptLines)
	receipts.POST("/:id/lines", h.AddReceiptLine)
	receipts.PUT("/:id/lines/:line_id", h.UpdateReceiptLine)
	receipts.DELETE("/:id/lines/:line_id", h.DeleteReceiptLine)
	receipts.POST("/from-po", h.CreateReceiptFromPO)

	// Stock counting batches and lines
	counts := api.Group("/counts")
	counts.Use(middleware.JWT(h.Config.JWTSecret))
	counts.Use(middleware.RequireTenant())
	counts.GET("", h.ListCountBatches)
	counts.POST("", h.CreateCountBatch)
	counts.PUT("/:id", h.UpdateCountBatch)
	counts.DELETE("/:id", h.DeleteCountBatch)
	counts.GET("/:batch_id/lines", h.ListCountLines)
	counts.POST("/:batch_id/lines", h.AddCountLine)
	counts.PUT("/:batch_id/lines/:line_id", h.UpdateCountLine)
	counts.DELETE("/:batch_id/lines/:line_id", h.DeleteCountLine)

	users := api.Group("/users")
	users.Use(middleware.JWT(h.Config.JWTSecret))
	users.Use(middleware.RequireTenant())
	users.Use(middleware.RequireRole("ADMIN"))
	users.GET("", h.ListUsers)
	users.POST("", h.CreateUser)
	users.GET("/:id", h.GetUser)
	users.PUT("/:id", h.UpdateUser)
	users.POST("/:id/disable", h.DisableUser)

	audit := api.Group("/audit")
	audit.Use(middleware.JWT(h.Config.JWTSecret))
	audit.Use(middleware.RequireTenant())
	audit.GET("", h.GetAuditLogs)

	// System admin routes (no tenant context required)
	systemAdmin := api.Group("/system")
	systemAdmin.Use(middleware.JWT(h.Config.JWTSecret))
	systemAdmin.Use(middleware.RequireRole("SYSTEM_ADMIN"))

	systemAdmin.GET("/tenants", h.ListTenants)
	systemAdmin.POST("/tenants", h.CreateTenant)
	systemAdmin.GET("/tenants/:id", h.GetTenant)
	systemAdmin.PUT("/tenants/:id", h.UpdateTenant)
	systemAdmin.DELETE("/tenants/:id", h.DeactivateTenant)

}

func startServer(e *echo.Echo, cfg *config.Config) {
	go func() {
		log.Info().Str("port", cfg.Port).Msg("Starting server")
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to shutdown server")
	}

	log.Info().Msg("Server shutdown gracefully")
}
