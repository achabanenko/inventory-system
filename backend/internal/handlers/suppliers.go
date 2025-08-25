package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type SupplierModel struct {
	ID       string      `json:"id"`
	Code     string      `json:"code"`
	Name     string      `json:"name"`
	Contact  interface{} `json:"contact,omitempty"`
	IsActive bool        `json:"is_active"`
}

func (h *Handler) ListSuppliers(c echo.Context) error {
	// Parse query parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	search := c.QueryParam("q")
	isActiveParam := c.QueryParam("is_active")

	offset := (page - 1) * pageSize

	// Build query
	query := `
		SELECT id, code, name, contact, is_active
		FROM suppliers
		WHERE 1=1`

	args := []interface{}{}
	argCount := 0

	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND (code ILIKE $%d OR name ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+search+"%")
	}

	if isActiveParam != "" {
		isActive := isActiveParam == "true"
		argCount++
		query += fmt.Sprintf(" AND is_active = $%d", argCount)
		args = append(args, isActive)
	}

	query += " ORDER BY name ASC"

	// Add pagination
	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, pageSize)

	argCount++
	query += fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, offset)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	var suppliers []SupplierModel
	for rows.Next() {
		var supplier SupplierModel
		var contact sql.NullString

		err := rows.Scan(
			&supplier.ID, &supplier.Code, &supplier.Name, &contact, &supplier.IsActive,
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Database scan error")
		}

		if contact.Valid {
			supplier.Contact = contact.String
		}

		suppliers = append(suppliers, supplier)
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM suppliers WHERE 1=1`
	countArgs := []interface{}{}
	countArgCount := 0

	if search != "" {
		countArgCount++
		countQuery += fmt.Sprintf(" AND (code ILIKE $%d OR name ILIKE $%d)", countArgCount, countArgCount)
		countArgs = append(countArgs, "%"+search+"%")
	}

	if isActiveParam != "" {
		isActive := isActiveParam == "true"
		countArgCount++
		countQuery += fmt.Sprintf(" AND is_active = $%d", countArgCount)
		countArgs = append(countArgs, isActive)
	}

	var total int
	err = h.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	totalPages := (total + pageSize - 1) / pageSize

	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       suppliers,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Total:      int64(total),
	})
}

func (h *Handler) CreateSupplier(c echo.Context) error {
	var req struct {
		Code     string                 `json:"code" validate:"required"`
		Name     string                 `json:"name" validate:"required"`
		Contact  map[string]interface{} `json:"contact"`
		IsActive *bool                  `json:"is_active"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	if req.Code == "" || req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "code and name are required")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	var contactJSON []byte
	if req.Contact != nil {
		b, err := json.Marshal(req.Contact)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid contact")
		}
		contactJSON = b
	}

	query := `
        INSERT INTO suppliers (code, name, contact, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
        RETURNING id, code, name, contact, is_active
    `

	var (
		id         string
		code       string
		name       string
		isActiveDB bool
		contact    sql.NullString
	)

	err := h.DB.QueryRow(query, req.Code, req.Name, nullableJSON(contactJSON), isActive).Scan(&id, &code, &name, &contact, &isActiveDB)
	if err != nil {
		if isUniqueViolation(err) {
			return echo.NewHTTPError(http.StatusConflict, "supplier code already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	resp := SupplierModel{ID: id, Code: code, Name: name, IsActive: isActiveDB}
	if contact.Valid {
		resp.Contact = contact.String
	}
	return c.JSON(http.StatusCreated, resp)
}

func (h *Handler) GetSupplier(c echo.Context) error {
	id := c.Param("id")

	var s SupplierModel
	var contact sql.NullString
	err := h.DB.QueryRow(`
        SELECT id, code, name, contact, is_active
        FROM suppliers WHERE id = $1
    `, id).Scan(&s.ID, &s.Code, &s.Name, &contact, &s.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "supplier not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if contact.Valid {
		s.Contact = contact.String
	}
	return c.JSON(http.StatusOK, s)
}

func (h *Handler) UpdateSupplier(c echo.Context) error {
	id := c.Param("id")
	var req struct {
		Code     *string                `json:"code"`
		Name     *string                `json:"name"`
		Contact  map[string]interface{} `json:"contact"`
		IsActive *bool                  `json:"is_active"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Build dynamic update
	sets := []string{}
	args := []interface{}{}
	idx := 1

	if req.Code != nil {
		sets = append(sets, fmt.Sprintf("code = $%d", idx))
		args = append(args, strings.TrimSpace(*req.Code))
		idx++
	}
	if req.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", idx))
		args = append(args, strings.TrimSpace(*req.Name))
		idx++
	}
	if req.IsActive != nil {
		sets = append(sets, fmt.Sprintf("is_active = $%d", idx))
		args = append(args, *req.IsActive)
		idx++
	}
	if req.Contact != nil {
		b, err := json.Marshal(req.Contact)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid contact")
		}
		sets = append(sets, fmt.Sprintf("contact = $%d", idx))
		args = append(args, string(b))
		idx++
	}

	if len(sets) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "no fields to update")
	}
	sets = append(sets, fmt.Sprintf("updated_at = NOW()"))
	args = append(args, id)

	query := fmt.Sprintf(`UPDATE suppliers SET %s WHERE id = $%d RETURNING id, code, name, contact, is_active`, strings.Join(sets, ", "), idx)

	var out SupplierModel
	var contact sql.NullString
	if err := h.DB.QueryRow(query, args...).Scan(&out.ID, &out.Code, &out.Name, &contact, &out.IsActive); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "supplier not found")
		}
		if isUniqueViolation(err) {
			return echo.NewHTTPError(http.StatusConflict, "supplier code already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if contact.Valid {
		out.Contact = contact.String
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) DeleteSupplier(c echo.Context) error {
	id := c.Param("id")
	res, err := h.DB.Exec(`DELETE FROM suppliers WHERE id = $1`, id)
	if err != nil {
		// FK conflict or others
		return echo.NewHTTPError(http.StatusConflict, "cannot delete supplier (in use)")
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "supplier not found")
	}
	return c.NoContent(http.StatusNoContent)
}

// helpers
func isUniqueViolation(err error) bool {
	// crude detection by message text, avoids importing driver-specific codes
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key value") || strings.Contains(msg, "unique constraint")
}

func nullableJSON(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	return string(b)
}
