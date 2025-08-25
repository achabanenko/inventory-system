package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"inventory/internal/middleware"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
)

// ItemDTO represents the API contract for items
type ItemDTO struct {
	ID         uuid.UUID              `json:"id"`
	SKU        string                 `json:"sku"`
	Name       string                 `json:"name"`
	Barcode    *string                `json:"barcode,omitempty"`
	UOM        string                 `json:"uom"`
	CategoryID *uuid.UUID             `json:"category_id,omitempty"`
	Category   *CategoryDTO           `json:"category,omitempty"`
	Cost       decimal.Decimal        `json:"cost"`
	Price      decimal.Decimal        `json:"price"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	IsActive   bool                   `json:"is_active"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	DeletedAt  *time.Time             `json:"deleted_at,omitempty"`
}

type createOrUpdateItemRequest struct {
	SKU        string                 `json:"sku" validate:"required"`
	Name       string                 `json:"name" validate:"required"`
	Barcode    *string                `json:"barcode"`
	UOM        string                 `json:"uom" validate:"required"`
	CategoryID *uuid.UUID             `json:"category_id"`
	Cost       string                 `json:"cost" validate:"required"`  // decimal as string to avoid float issues
	Price      string                 `json:"price" validate:"required"` // decimal as string to avoid float issues
	Attributes map[string]interface{} `json:"attributes"`
	IsActive   *bool                  `json:"is_active"`
}

func (h *Handler) ListItems(c echo.Context) error {
	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantID(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant context required")
	}

	// Pagination params
	var qp PaginationParams
	if err := c.Bind(&qp); err != nil {
		// ignore bind error, use defaults
	}
	page := qp.Page
	if page < 1 {
		page = 1
	}
	pageSize := qp.PageSize
	if pageSize <= 0 {
		pageSize = h.Config.DefaultPageSize
	}
	if pageSize > h.Config.MaxPageSize {
		pageSize = h.Config.MaxPageSize
	}

	q := c.QueryParam("q")

	// Build filters with tenant isolation
	where := "WHERE i.tenant_id = $1 AND i.deleted_at IS NULL"
	var args []interface{}
	args = append(args, tenantID)

	if q != "" {
		where += " AND (i.sku ILIKE $2 OR i.name ILIKE $2 OR i.barcode ILIKE $2)"
		args = append(args, "%"+q+"%")
	}

	// Count total (need to fix this to use same table alias)
	countSQL := "SELECT COUNT(1) FROM items i " + where
	var total int64
	if err := h.DB.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrorDetail{Code: "INTERNAL_ERROR", Message: err.Error()}})
	}

	// Fetch page with category information
	offset := (page - 1) * pageSize
	listSQL := `SELECT i.id, i.sku, i.name, i.barcode, i.uom, i.category_id, i.cost, i.price, i.attributes, i.is_active, i.created_at, i.updated_at, i.deleted_at,
				c.id as cat_id, c.name as cat_name
				FROM items i 
				LEFT JOIN categories c ON i.category_id = c.id ` + where + " ORDER BY i.created_at DESC LIMIT $%d OFFSET $%d"
	// Prepare LIMIT/OFFSET placeholders depending on existing args
	limitIndex := len(args) + 1
	offsetIndex := len(args) + 2
	listSQL = fmt.Sprintf(listSQL, limitIndex, offsetIndex)
	args = append(args, pageSize, offset)

	rows, err := h.DB.Query(listSQL, args...)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrorDetail{Code: "INTERNAL_ERROR", Message: err.Error()}})
	}
	defer rows.Close()

	items := make([]ItemDTO, 0, pageSize)
	for rows.Next() {
		var dto ItemDTO
		var barcode sql.NullString
		var categoryID sql.NullString
		var catID sql.NullString
		var catName sql.NullString
		var rawAttrs []byte
		if err := rows.Scan(&dto.ID, &dto.SKU, &dto.Name, &barcode, &dto.UOM, &categoryID, &dto.Cost, &dto.Price, &rawAttrs, &dto.IsActive, &dto.CreatedAt, &dto.UpdatedAt, &dto.DeletedAt, &catID, &catName); err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrorDetail{Code: "INTERNAL_ERROR", Message: err.Error()}})
		}
		if barcode.Valid {
			s := barcode.String
			dto.Barcode = &s
		}
		if categoryID.Valid {
			if cid, err := uuid.Parse(categoryID.String); err == nil {
				dto.CategoryID = &cid
			}
		}
		if catID.Valid && catName.Valid {
			if cid, err := uuid.Parse(catID.String); err == nil {
				dto.Category = &CategoryDTO{
					ID:   cid,
					Name: catName.String,
				}
			}
		}
		if len(rawAttrs) > 0 {
			_ = json.Unmarshal(rawAttrs, &dto.Attributes)
		}
		items = append(items, dto)
	}
	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrorDetail{Code: "INTERNAL_ERROR", Message: err.Error()}})
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       items,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Total:      total,
	})
}

func (h *Handler) CreateItem(c echo.Context) error {
	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantID(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant context required")
	}

	var req createOrUpdateItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "invalid request body"}})
	}
	if req.SKU == "" || req.Name == "" || req.UOM == "" || req.Cost == "" || req.Price == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "missing required fields"}})
	}

	cost, err := decimal.NewFromString(req.Cost)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "invalid cost"}})
	}
	price, err := decimal.NewFromString(req.Price)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "invalid price"}})
	}

	id := uuid.New()
	now := time.Now().UTC()

	var attrsJSON []byte
	if req.Attributes != nil {
		b, err := json.Marshal(req.Attributes)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "invalid attributes"}})
		}
		attrsJSON = b
	} else {
		attrsJSON = []byte("{}")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	query := `
        INSERT INTO items (id, tenant_id, sku, name, barcode, uom, category_id, cost, price, attributes, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
        RETURNING id, sku, name, barcode, uom, category_id, cost, price, attributes, is_active, created_at, updated_at, deleted_at
    `

	var (
		barcode  sql.NullString
		returned ItemDTO
		rawAttrs []byte
	)

	if req.Barcode != nil {
		barcode = sql.NullString{String: *req.Barcode, Valid: true}
	}

	err = h.DB.QueryRow(
		query,
		id,
		tenantID,
		req.SKU,
		req.Name,
		barcode,
		req.UOM,
		req.CategoryID,
		cost.String(),
		price.String(),
		attrsJSON,
		isActive,
		now,
		now,
	).Scan(
		&returned.ID,
		&returned.SKU,
		&returned.Name,
		&barcode,
		&returned.UOM,
		&returned.CategoryID,
		&returned.Cost,
		&returned.Price,
		&rawAttrs,
		&returned.IsActive,
		&returned.CreatedAt,
		&returned.UpdatedAt,
		&returned.DeletedAt,
	)
	if err != nil {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: ErrorDetail{Code: "CONFLICT", Message: err.Error()}})
	}

	if barcode.Valid {
		s := barcode.String
		returned.Barcode = &s
	}
	if len(rawAttrs) > 0 {
		_ = json.Unmarshal(rawAttrs, &returned.Attributes)
	}

	return c.JSON(http.StatusCreated, returned)
}

func (h *Handler) GetItem(c echo.Context) error {
	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantID(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant context required")
	}

	idParam := c.Param("id")
	itemID, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "invalid id"}})
	}

	query := `
        SELECT i.id, i.sku, i.name, i.barcode, i.uom, i.category_id, i.cost, i.price, i.attributes, i.is_active, i.created_at, i.updated_at, i.deleted_at,
               c.id as cat_id, c.name as cat_name
        FROM items i
        LEFT JOIN categories c ON i.category_id = c.id 
        WHERE i.id = $1 AND i.tenant_id = $2 AND i.deleted_at IS NULL
    `

	var (
		dto        ItemDTO
		barcode    sql.NullString
		categoryID sql.NullString
		catID      sql.NullString
		catName    sql.NullString
		rawAttr    []byte
	)

	err = h.DB.QueryRow(query, itemID, tenantID).Scan(
		&dto.ID, &dto.SKU, &dto.Name, &barcode, &dto.UOM, &categoryID, &dto.Cost, &dto.Price, &rawAttr, &dto.IsActive, &dto.CreatedAt, &dto.UpdatedAt, &dto.DeletedAt, &catID, &catName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: ErrorDetail{Code: "NOT_FOUND", Message: "item not found"}})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrorDetail{Code: "INTERNAL_ERROR", Message: err.Error()}})
	}
	if barcode.Valid {
		s := barcode.String
		dto.Barcode = &s
	}
	if categoryID.Valid {
		if cid, err := uuid.Parse(categoryID.String); err == nil {
			dto.CategoryID = &cid
		}
	}
	if catID.Valid && catName.Valid {
		if cid, err := uuid.Parse(catID.String); err == nil {
			dto.Category = &CategoryDTO{
				ID:   cid,
				Name: catName.String,
			}
		}
	}
	if len(rawAttr) > 0 {
		_ = json.Unmarshal(rawAttr, &dto.Attributes)
	}
	return c.JSON(http.StatusOK, dto)
}

func (h *Handler) UpdateItem(c echo.Context) error {
	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantID(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant context required")
	}

	idParam := c.Param("id")
	itemID, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "invalid id"}})
	}

	var req createOrUpdateItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "invalid request body"}})
	}
	if req.SKU == "" || req.Name == "" || req.UOM == "" || req.Cost == "" || req.Price == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "missing required fields"}})
	}

	cost, err := decimal.NewFromString(req.Cost)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "invalid cost"}})
	}
	price, err := decimal.NewFromString(req.Price)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "invalid price"}})
	}

	var attrsJSON []byte
	if req.Attributes != nil {
		b, err := json.Marshal(req.Attributes)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "invalid attributes"}})
		}
		attrsJSON = b
	} else {
		attrsJSON = []byte("{}")
	}

	var barcode sql.NullString
	if req.Barcode != nil {
		barcode = sql.NullString{String: *req.Barcode, Valid: true}
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	query := `
        UPDATE items
        SET sku = $1,
            name = $2,
            barcode = $3,
            uom = $4,
            category_id = $5,
            cost = $6,
            price = $7,
            attributes = $8,
            is_active = $9,
            updated_at = $10
        WHERE id = $11 AND tenant_id = $12 AND deleted_at IS NULL
        RETURNING id, sku, name, barcode, uom, category_id, cost, price, attributes, is_active, created_at, updated_at, deleted_at
    `

	var dto ItemDTO
	var rawAttrs []byte
	err = h.DB.QueryRow(
		query,
		req.SKU,
		req.Name,
		barcode,
		req.UOM,
		req.CategoryID,
		cost.String(),
		price.String(),
		attrsJSON,
		isActive,
		time.Now().UTC(),
		itemID,
		tenantID,
	).Scan(
		&dto.ID, &dto.SKU, &dto.Name, &barcode, &dto.UOM, &dto.CategoryID, &dto.Cost, &dto.Price, &rawAttrs, &dto.IsActive, &dto.CreatedAt, &dto.UpdatedAt, &dto.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: ErrorDetail{Code: "NOT_FOUND", Message: "item not found"}})
		}
		return c.JSON(http.StatusConflict, ErrorResponse{Error: ErrorDetail{Code: "CONFLICT", Message: err.Error()}})
	}

	if barcode.Valid {
		s := barcode.String
		dto.Barcode = &s
	}
	if len(rawAttrs) > 0 {
		_ = json.Unmarshal(rawAttrs, &dto.Attributes)
	}
	return c.JSON(http.StatusOK, dto)
}

func (h *Handler) DeleteItem(c echo.Context) error {
	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantID(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant context required")
	}

	idParam := c.Param("id")
	itemID, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "invalid id"}})
	}

	query := `
        UPDATE items SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND tenant_id = $3 AND deleted_at IS NULL
        RETURNING id
    `
	var id uuid.UUID
	err = h.DB.QueryRow(query, time.Now().UTC(), itemID, tenantID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: ErrorDetail{Code: "NOT_FOUND", Message: "item not found"}})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrorDetail{Code: "INTERNAL_ERROR", Message: err.Error()}})
	}
	return c.NoContent(http.StatusNoContent)
}
