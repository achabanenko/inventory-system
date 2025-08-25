package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"inventory/internal/middleware"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// CategoryDTO represents the API contract for categories
type CategoryDTO struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type createOrUpdateCategoryRequest struct {
	Name     string     `json:"name" validate:"required"`
	ParentID *uuid.UUID `json:"parent_id"`
}

func (h *Handler) ListCategories(c echo.Context) error {
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
	where := "WHERE tenant_id = $1"
	var args []interface{}
	args = append(args, tenantID)

	if q != "" {
		where += " AND name ILIKE $2"
		args = append(args, "%"+q+"%")
	}

	// Count total
	countSQL := "SELECT COUNT(1) FROM categories " + where
	var total int64
	if err := h.DB.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrorDetail{Code: "INTERNAL_ERROR", Message: err.Error()}})
	}

	// Fetch page
	offset := (page - 1) * pageSize
	listSQL := "SELECT id, name, parent_id, created_at, updated_at FROM categories " + where + " ORDER BY name ASC LIMIT $%d OFFSET $%d"
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

	categories := make([]CategoryDTO, 0, pageSize)
	for rows.Next() {
		var dto CategoryDTO
		var parentID sql.NullString
		if err := rows.Scan(&dto.ID, &dto.Name, &parentID, &dto.CreatedAt, &dto.UpdatedAt); err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrorDetail{Code: "INTERNAL_ERROR", Message: err.Error()}})
		}
		if parentID.Valid {
			if pid, err := uuid.Parse(parentID.String); err == nil {
				dto.ParentID = &pid
			}
		}
		categories = append(categories, dto)
	}
	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrorDetail{Code: "INTERNAL_ERROR", Message: err.Error()}})
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       categories,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Total:      total,
	})
}

func (h *Handler) CreateCategory(c echo.Context) error {
	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantID(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant context required")
	}

	var req createOrUpdateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "invalid request body"}})
	}
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "name is required"}})
	}

	id := uuid.New()
	now := time.Now().UTC()

	query := `
        INSERT INTO categories (id, tenant_id, name, parent_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, name, parent_id, created_at, updated_at
    `

	var (
		returned CategoryDTO
		parentID sql.NullString
	)

	err := h.DB.QueryRow(
		query,
		id,
		tenantID,
		req.Name,
		req.ParentID,
		now,
		now,
	).Scan(
		&returned.ID,
		&returned.Name,
		&parentID,
		&returned.CreatedAt,
		&returned.UpdatedAt,
	)
	if err != nil {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: ErrorDetail{Code: "CONFLICT", Message: err.Error()}})
	}

	if parentID.Valid {
		if pid, err := uuid.Parse(parentID.String); err == nil {
			returned.ParentID = &pid
		}
	}

	return c.JSON(http.StatusCreated, returned)
}

func (h *Handler) GetCategory(c echo.Context) error {
	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantID(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant context required")
	}

	idParam := c.Param("id")
	categoryID, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "invalid id"}})
	}

	query := `
        SELECT id, name, parent_id, created_at, updated_at
        FROM categories WHERE id = $1 AND tenant_id = $2
    `

	var (
		dto      CategoryDTO
		parentID sql.NullString
	)

	err = h.DB.QueryRow(query, categoryID, tenantID).Scan(
		&dto.ID, &dto.Name, &parentID, &dto.CreatedAt, &dto.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: ErrorDetail{Code: "NOT_FOUND", Message: "category not found"}})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrorDetail{Code: "INTERNAL_ERROR", Message: err.Error()}})
	}

	if parentID.Valid {
		if pid, err := uuid.Parse(parentID.String); err == nil {
			dto.ParentID = &pid
		}
	}

	return c.JSON(http.StatusOK, dto)
}

func (h *Handler) UpdateCategory(c echo.Context) error {
	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantID(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant context required")
	}

	idParam := c.Param("id")
	categoryID, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "invalid id"}})
	}

	var req createOrUpdateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "invalid request body"}})
	}
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "name is required"}})
	}

	query := `
        UPDATE categories
        SET name = $1,
            parent_id = $2,
            updated_at = $3
        WHERE id = $4 AND tenant_id = $5
        RETURNING id, name, parent_id, created_at, updated_at
    `

	var (
		dto      CategoryDTO
		parentID sql.NullString
	)

	err = h.DB.QueryRow(
		query,
		req.Name,
		req.ParentID,
		time.Now().UTC(),
		categoryID,
		tenantID,
	).Scan(
		&dto.ID, &dto.Name, &parentID, &dto.CreatedAt, &dto.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: ErrorDetail{Code: "NOT_FOUND", Message: "category not found"}})
		}
		return c.JSON(http.StatusConflict, ErrorResponse{Error: ErrorDetail{Code: "CONFLICT", Message: err.Error()}})
	}

	if parentID.Valid {
		if pid, err := uuid.Parse(parentID.String); err == nil {
			dto.ParentID = &pid
		}
	}

	return c.JSON(http.StatusOK, dto)
}

func (h *Handler) DeleteCategory(c echo.Context) error {
	// Get tenant ID from context
	tenantID, ok := middleware.GetTenantID(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant context required")
	}

	idParam := c.Param("id")
	categoryID, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: "invalid id"}})
	}

	// Check if there are items using this category
	var itemCount int64
	err = h.DB.QueryRow("SELECT COUNT(1) FROM items WHERE category_id = $1 AND tenant_id = $2 AND deleted_at IS NULL", categoryID, tenantID).Scan(&itemCount)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrorDetail{Code: "INTERNAL_ERROR", Message: err.Error()}})
	}
	if itemCount > 0 {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: ErrorDetail{Code: "CONFLICT", Message: "Cannot delete category with items assigned to it"}})
	}

	// Check if there are child categories
	var childCount int64
	err = h.DB.QueryRow("SELECT COUNT(1) FROM categories WHERE parent_id = $1 AND tenant_id = $2", categoryID, tenantID).Scan(&childCount)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrorDetail{Code: "INTERNAL_ERROR", Message: err.Error()}})
	}
	if childCount > 0 {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: ErrorDetail{Code: "CONFLICT", Message: "Cannot delete category with child categories"}})
	}

	query := `DELETE FROM categories WHERE id = $1 AND tenant_id = $2`
	result, err := h.DB.Exec(query, categoryID, tenantID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrorDetail{Code: "INTERNAL_ERROR", Message: err.Error()}})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrorDetail{Code: "INTERNAL_ERROR", Message: err.Error()}})
	}
	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: ErrorDetail{Code: "NOT_FOUND", Message: "category not found"}})
	}

	return c.NoContent(http.StatusNoContent)
}
