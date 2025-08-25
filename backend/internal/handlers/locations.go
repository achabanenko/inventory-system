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

type LocationModel struct {
	ID       string      `json:"id"`
	Code     string      `json:"code"`
	Name     string      `json:"name"`
	Address  interface{} `json:"address,omitempty"`
	IsActive bool        `json:"is_active"`
}

func (h *Handler) ListLocations(c echo.Context) error {
	// Parse pagination & filters
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
	query := `SELECT id, code, name, address, is_active FROM locations WHERE 1=1`
	args := []interface{}{}
	n := 0
	if search != "" {
		n++
		query += fmt.Sprintf(" AND (code ILIKE $%d OR name ILIKE $%d)", n, n)
		args = append(args, "%"+search+"%")
	}
	if isActiveParam != "" {
		n++
		query += fmt.Sprintf(" AND is_active = $%d", n)
		args = append(args, isActiveParam == "true")
	}
	query += " ORDER BY code ASC"
	n++
	query += fmt.Sprintf(" LIMIT $%d", n)
	args = append(args, pageSize)
	n++
	query += fmt.Sprintf(" OFFSET $%d", n)
	args = append(args, offset)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	defer rows.Close()

	res := []LocationModel{}
	for rows.Next() {
		var m LocationModel
		var addr sql.NullString
		if err := rows.Scan(&m.ID, &m.Code, &m.Name, &addr, &m.IsActive); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "database scan error")
		}
		if addr.Valid {
			m.Address = addr.String
		}
		res = append(res, m)
	}

	// Count
	countQ := `SELECT COUNT(*) FROM locations WHERE 1=1`
	countArgs := []interface{}{}
	k := 0
	if search != "" {
		k++
		countQ += fmt.Sprintf(" AND (code ILIKE $%d OR name ILIKE $%d)", k, k)
		countArgs = append(countArgs, "%"+search+"%")
	}
	if isActiveParam != "" {
		k++
		countQ += fmt.Sprintf(" AND is_active = $%d", k)
		countArgs = append(countArgs, isActiveParam == "true")
	}
	var total int
	if err := h.DB.QueryRow(countQ, countArgs...).Scan(&total); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	return c.JSON(http.StatusOK, PaginatedResponse{Data: res, Page: page, PageSize: pageSize, TotalPages: (total + pageSize - 1) / pageSize, Total: int64(total)})
}

func (h *Handler) CreateLocation(c echo.Context) error {
	var req struct {
		Code     string                 `json:"code"`
		Name     string                 `json:"name"`
		Address  map[string]interface{} `json:"address"`
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

	var addrJSON []byte
	if req.Address != nil {
		b, err := json.Marshal(req.Address)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid address")
		}
		addrJSON = b
	}

	var m LocationModel
	var addr sql.NullString
	err := h.DB.QueryRow(`
        INSERT INTO locations (code, name, address, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
        RETURNING id, code, name, address, is_active
    `, req.Code, req.Name, nullableJSON(addrJSON), isActive).Scan(&m.ID, &m.Code, &m.Name, &addr, &m.IsActive)
	if err != nil {
		if isUniqueViolation(err) {
			return echo.NewHTTPError(http.StatusConflict, "location code already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if addr.Valid {
		m.Address = addr.String
	}
	return c.JSON(http.StatusCreated, m)
}

func (h *Handler) GetLocation(c echo.Context) error {
	id := c.Param("id")
	var m LocationModel
	var addr sql.NullString
	if err := h.DB.QueryRow(`SELECT id, code, name, address, is_active FROM locations WHERE id = $1`, id).Scan(&m.ID, &m.Code, &m.Name, &addr, &m.IsActive); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "location not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if addr.Valid {
		m.Address = addr.String
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) UpdateLocation(c echo.Context) error {
	id := c.Param("id")
	var req struct {
		Code     *string                `json:"code"`
		Name     *string                `json:"name"`
		Address  map[string]interface{} `json:"address"`
		IsActive *bool                  `json:"is_active"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	sets := []string{}
	args := []interface{}{}
	i := 1
	if req.Code != nil {
		sets = append(sets, fmt.Sprintf("code = $%d", i))
		args = append(args, strings.TrimSpace(*req.Code))
		i++
	}
	if req.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", i))
		args = append(args, strings.TrimSpace(*req.Name))
		i++
	}
	if req.IsActive != nil {
		sets = append(sets, fmt.Sprintf("is_active = $%d", i))
		args = append(args, *req.IsActive)
		i++
	}
	if req.Address != nil {
		b, err := json.Marshal(req.Address)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid address")
		}
		sets = append(sets, fmt.Sprintf("address = $%d", i))
		args = append(args, string(b))
		i++
	}
	if len(sets) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "no fields to update")
	}
	sets = append(sets, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(`UPDATE locations SET %s WHERE id = $%d RETURNING id, code, name, address, is_active`, strings.Join(sets, ", "), i)

	var m LocationModel
	var addr sql.NullString
	if err := h.DB.QueryRow(query, args...).Scan(&m.ID, &m.Code, &m.Name, &addr, &m.IsActive); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "location not found")
		}
		if isUniqueViolation(err) {
			return echo.NewHTTPError(http.StatusConflict, "location code already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if addr.Valid {
		m.Address = addr.String
	}
	return c.JSON(http.StatusOK, m)
}

func (h *Handler) DeleteLocation(c echo.Context) error {
	id := c.Param("id")
	res, err := h.DB.Exec(`DELETE FROM locations WHERE id = $1`, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusConflict, "cannot delete location (in use)")
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "location not found")
	}
	return c.NoContent(http.StatusNoContent)
}
