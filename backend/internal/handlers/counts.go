package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type CountBatch struct {
	ID          string  `json:"id"`
	Number      string  `json:"number"`
	LocationID  string  `json:"location_id"`
	Status      string  `json:"status"`
	Notes       *string `json:"notes,omitempty"`
	CreatedBy   *string `json:"created_by,omitempty"`
	CompletedAt *string `json:"completed_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type CountLine struct {
	ID             string `json:"id"`
	BatchID        string `json:"batch_id"`
	ItemID         string `json:"item_id"`
	ItemSKU        string `json:"item_sku,omitempty"`
	ItemName       string `json:"item_name,omitempty"`
	ExpectedOnHand int    `json:"expected_on_hand"`
	CountedQty     int    `json:"counted_qty"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// Batches
func (h *Handler) ListCountBatches(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	status := c.QueryParam("status")
	locationID := c.QueryParam("location_id")

	offset := (page - 1) * pageSize

	query := `SELECT id, number, location_id, status, notes, created_by, completed_at, created_at, updated_at FROM count_batches WHERE 1=1`
	args := []interface{}{}
	n := 0
	if status != "" {
		n++
		query += fmt.Sprintf(" AND status = $%d", n)
		args = append(args, status)
	}
	if locationID != "" {
		n++
		query += fmt.Sprintf(" AND location_id = $%d", n)
		args = append(args, locationID)
	}
	query += " ORDER BY created_at DESC"
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

	res := []CountBatch{}
	for rows.Next() {
		var m CountBatch
		var notes, createdBy, completedAt sql.NullString
		if err := rows.Scan(&m.ID, &m.Number, &m.LocationID, &m.Status, &notes, &createdBy, &completedAt, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "database scan error")
		}
		if notes.Valid {
			m.Notes = &notes.String
		}
		if createdBy.Valid {
			m.CreatedBy = &createdBy.String
		}
		if completedAt.Valid {
			m.CompletedAt = &completedAt.String
		}
		res = append(res, m)
	}

	var total int
	countQ := `SELECT COUNT(*) FROM count_batches WHERE 1=1`
	countArgs := []interface{}{}
	k := 0
	if status != "" {
		k++
		countQ += fmt.Sprintf(" AND status = $%d", k)
		countArgs = append(countArgs, status)
	}
	if locationID != "" {
		k++
		countQ += fmt.Sprintf(" AND location_id = $%d", k)
		countArgs = append(countArgs, locationID)
	}
	if err := h.DB.QueryRow(countQ, countArgs...).Scan(&total); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	return c.JSON(http.StatusOK, PaginatedResponse{Data: res, Page: page, PageSize: pageSize, TotalPages: (total + pageSize - 1) / pageSize, Total: int64(total)})
}

func (h *Handler) CreateCountBatch(c echo.Context) error {
	var req struct {
		LocationID string  `json:"location_id"`
		Notes      *string `json:"notes"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.LocationID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "location_id is required")
	}

	// next number
	var maxNumber int
	_ = h.DB.QueryRow(`SELECT COALESCE(MAX(CAST(SUBSTRING(number FROM 'CB-([0-9]+)') AS INTEGER)), 0) FROM count_batches WHERE number ~ '^CB-[0-9]+$'`).Scan(&maxNumber)
	number := fmt.Sprintf("CB-%06d", maxNumber+1)

	id := uuid.New().String()
	var created CountBatch
	var notes sql.NullString
	err := h.DB.QueryRow(`
        INSERT INTO count_batches (id, number, location_id, status, notes, created_at, updated_at)
        VALUES ($1, $2, $3, 'OPEN', $4, NOW(), NOW())
        RETURNING id, number, location_id, status, notes, created_at, updated_at
    `, id, number, req.LocationID, req.Notes).Scan(&created.ID, &created.Number, &created.LocationID, &created.Status, &notes, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if notes.Valid {
		created.Notes = &notes.String
	}
	return c.JSON(http.StatusCreated, created)
}

func (h *Handler) UpdateCountBatch(c echo.Context) error {
	id := c.Param("id")
	var req struct {
		LocationID *string `json:"location_id"`
		Status     *string `json:"status"`
		Notes      *string `json:"notes"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	sets := []string{}
	args := []interface{}{}
	i := 1
	if req.LocationID != nil {
		sets = append(sets, fmt.Sprintf("location_id = $%d", i))
		args = append(args, *req.LocationID)
		i++
	}
	if req.Status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", i))
		args = append(args, *req.Status)
		i++
	}
	if req.Notes != nil {
		sets = append(sets, fmt.Sprintf("notes = $%d", i))
		args = append(args, *req.Notes)
		i++
	}
	if len(sets) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "no fields to update")
	}
	sets = append(sets, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(`UPDATE count_batches SET %s WHERE id = $%d RETURNING id, number, location_id, status, notes, created_at, updated_at`, strings.Join(sets, ", "), i)
	var out CountBatch
	var notes sql.NullString
	if err := h.DB.QueryRow(query, args...).Scan(&out.ID, &out.Number, &out.LocationID, &out.Status, &notes, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "batch not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if notes.Valid {
		out.Notes = &notes.String
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) DeleteCountBatch(c echo.Context) error {
	id := c.Param("id")
	res, err := h.DB.Exec(`DELETE FROM count_batches WHERE id = $1`, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusConflict, "cannot delete batch (in use)")
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "batch not found")
	}
	return c.NoContent(http.StatusNoContent)
}

// Lines
func (h *Handler) ListCountLines(c echo.Context) error {
	batchID := c.Param("batch_id")
	rows, err := h.DB.Query(`
        SELECT cl.id, cl.batch_id, cl.item_id, COALESCE(i.sku, ''), COALESCE(i.name, ''), cl.expected_on_hand, cl.counted_qty, cl.created_at, cl.updated_at
        FROM count_lines cl
        LEFT JOIN items i ON i.id = cl.item_id
        WHERE cl.batch_id = $1
        ORDER BY cl.created_at ASC
    `, batchID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	defer rows.Close()
	res := []CountLine{}
	for rows.Next() {
		var m CountLine
		if err := rows.Scan(&m.ID, &m.BatchID, &m.ItemID, &m.ItemSKU, &m.ItemName, &m.ExpectedOnHand, &m.CountedQty, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "database scan error")
		}
		res = append(res, m)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": res})
}

func (h *Handler) AddCountLine(c echo.Context) error {
	batchID := c.Param("batch_id")
	var req struct {
		ItemID         string `json:"item_id"`
		ExpectedOnHand int    `json:"expected_on_hand"`
		CountedQty     int    `json:"counted_qty"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.ItemID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "item_id is required")
	}

	// Ensure batch exists and get its location
	var batchLocationID string
	if err := h.DB.QueryRow(`SELECT location_id FROM count_batches WHERE id = $1`, batchID).Scan(&batchLocationID); err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "batch not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	// Resolve item id: allow UUID or SKU
	resolvedItemID := ""
	if _, err := uuid.Parse(req.ItemID); err == nil {
		// UUID provided; verify exists
		if err := h.DB.QueryRow(`SELECT id FROM items WHERE id = $1`, req.ItemID).Scan(&resolvedItemID); err != nil {
			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid item id")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "database error")
		}
	} else {
		// Treat as SKU or barcode; accept hyphen-less variants
		q := strings.TrimSpace(req.ItemID)
		err := h.DB.QueryRow(`
            SELECT id FROM items
            WHERE replace(sku, '-', '') = replace($1, '-', '')
               OR sku = $1
               OR barcode = $1
            LIMIT 1
        `, q).Scan(&resolvedItemID)
		if err == sql.ErrNoRows {
			// Fallback: try by name (case-insensitive), pick first match
			err = h.DB.QueryRow(`
                SELECT id FROM items WHERE LOWER(name) = LOWER($1) OR name ILIKE $2 LIMIT 1
            `, q, "%"+q+"%").Scan(&resolvedItemID)
		}
		if err != nil {
			if err == sql.ErrNoRows {
				// Create minimal item to allow counting to continue
				newID := uuid.New().String()
				sku := strings.ReplaceAll(q, " ", "-")
				if sku == "" {
					sku = newID
				}
				name := q
				uom := "each"
				created := false
				for attempt := 0; attempt < 3; attempt++ {
					if _, insErr := h.DB.Exec(`
                    INSERT INTO items (id, sku, name, uom, cost, price, is_active, created_at, updated_at)
                    VALUES ($1, $2, $3, $4, $5::numeric, $6::numeric, TRUE, NOW(), NOW())
                `, newID, sku, name, uom, "0.00", "0.00"); insErr == nil {
						resolvedItemID = newID
						created = true
						break
					} else if strings.Contains(insErr.Error(), "duplicate key") || strings.Contains(insErr.Error(), "unique") {
						sku = sku + "-1"
						continue
					} else {
						return echo.NewHTTPError(http.StatusInternalServerError, "database error")
					}
				}
				if !created {
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to create item")
				}
			} else {
				return echo.NewHTTPError(http.StatusInternalServerError, "database error")
			}
		}
	}

	// Auto-fill expected_on_hand from inventory_levels if not provided (>0)
	if req.ExpectedOnHand <= 0 {
		var onHand sql.NullInt64
		if err := h.DB.QueryRow(`SELECT on_hand FROM inventory_levels WHERE item_id = $1 AND location_id = $2`, resolvedItemID, batchLocationID).Scan(&onHand); err == nil && onHand.Valid {
			req.ExpectedOnHand = int(onHand.Int64)
		}
	}

	id := uuid.New().String()
	var out CountLine
	err := h.DB.QueryRow(`
        INSERT INTO count_lines (id, batch_id, item_id, expected_on_hand, counted_qty, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
        RETURNING id, batch_id, item_id, expected_on_hand, counted_qty, created_at, updated_at
    `, id, batchID, resolvedItemID, req.ExpectedOnHand, req.CountedQty).Scan(&out.ID, &out.BatchID, &out.ItemID, &out.ExpectedOnHand, &out.CountedQty, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	return c.JSON(http.StatusCreated, out)
}

func (h *Handler) UpdateCountLine(c echo.Context) error {
	batchID := c.Param("batch_id")
	lineID := c.Param("line_id")
	var req struct {
		ExpectedOnHand *int `json:"expected_on_hand"`
		CountedQty     *int `json:"counted_qty"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	sets := []string{}
	args := []interface{}{}
	i := 1
	if req.ExpectedOnHand != nil {
		sets = append(sets, fmt.Sprintf("expected_on_hand = $%d", i))
		args = append(args, *req.ExpectedOnHand)
		i++
	}
	if req.CountedQty != nil {
		sets = append(sets, fmt.Sprintf("counted_qty = $%d", i))
		args = append(args, *req.CountedQty)
		i++
	}
	if len(sets) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "no fields to update")
	}
	sets = append(sets, "updated_at = NOW()")
	args = append(args, lineID, batchID)

	query := fmt.Sprintf(`UPDATE count_lines SET %s WHERE id = $%d AND batch_id = $%d RETURNING id, batch_id, item_id, expected_on_hand, counted_qty, created_at, updated_at`, strings.Join(sets, ", "), i, i+1)
	var out CountLine
	if err := h.DB.QueryRow(query, args...).Scan(&out.ID, &out.BatchID, &out.ItemID, &out.ExpectedOnHand, &out.CountedQty, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "line not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) DeleteCountLine(c echo.Context) error {
	batchID := c.Param("batch_id")
	lineID := c.Param("line_id")
	res, err := h.DB.Exec(`DELETE FROM count_lines WHERE id = $1 AND batch_id = $2`, lineID, batchID)
	if err != nil {
		return echo.NewHTTPError(http.StatusConflict, "cannot delete line")
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "line not found")
	}
	return c.NoContent(http.StatusNoContent)
}
