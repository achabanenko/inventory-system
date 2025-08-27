package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	appmw "inventory/internal/middleware"
)

// Adjustment represents an inventory adjustment
type Adjustment struct {
	ID         string           `json:"id"`
	Number     string           `json:"number"`
	LocationID string           `json:"location_id"`
	Location   *Location        `json:"location,omitempty"`
	TenantID   string           `json:"tenant_id"`
	Reason     string           `json:"reason"`
	Status     string           `json:"status"`
	Notes      *string          `json:"notes,omitempty"`
	CreatedBy  *string          `json:"created_by,omitempty"`
	ApprovedBy *string          `json:"approved_by,omitempty"`
	ApprovedAt *time.Time       `json:"approved_at,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
	Lines      []AdjustmentLine `json:"lines,omitempty"`
}

// AdjustmentLine represents a line item in an adjustment
type AdjustmentLine struct {
	ID             string  `json:"id"`
	AdjustmentID   string  `json:"adjustment_id"`
	ItemID         *string `json:"item_id,omitempty"`
	ItemIdentifier string  `json:"item_identifier"`
	Item           *Item   `json:"item,omitempty"`
	QtyExpected    int     `json:"qty_expected"`
	QtyActual      int     `json:"qty_actual"`
	QtyDiff        int     `json:"qty_diff"`
	Notes          *string `json:"notes,omitempty"`
}

// CreateAdjustmentRequest represents the request to create an adjustment
type CreateAdjustmentRequest struct {
	LocationID string `json:"location_id" validate:"required"`
	Reason     string `json:"reason" validate:"required"`
	Notes      string `json:"notes"`
	Lines      []struct {
		ItemID      string `json:"item_id"`
		QtyExpected int    `json:"qty_expected"`
		QtyActual   int    `json:"qty_actual"`
		Notes       string `json:"notes"`
	} `json:"lines" validate:"required,min=1"`
}

// UpdateAdjustmentRequest represents the request to update an adjustment
type UpdateAdjustmentRequest struct {
	LocationID string `json:"location_id" validate:"required"`
	Reason     string `json:"reason" validate:"required"`
	Notes      string `json:"notes"`
	Lines      []struct {
		ItemID      string `json:"item_id"`
		QtyExpected int    `json:"qty_expected"`
		QtyActual   int    `json:"qty_actual"`
		Notes       string `json:"notes"`
	} `json:"lines" validate:"required,min=1"`
}

// generateAdjustmentNumber generates a unique adjustment number
func generateAdjustmentNumber() string {
	return fmt.Sprintf("ADJ-%d", time.Now().Unix())
}

// resolveOrCreateItemForAdjustment handles item lookup for adjustments
func (h *Handler) resolveOrCreateItemForAdjustment(tx *sql.Tx, itemIdentifier, tenantID string) (*string, error) {
	if itemIdentifier == "" {
		return nil, fmt.Errorf("item identifier is required")
	}

	var foundItemID string

	// First try by SKU
	err := tx.QueryRow(`
		SELECT id FROM items WHERE sku = $1 AND tenant_id = $2 AND is_active = true
	`, itemIdentifier, tenantID).Scan(&foundItemID)

	// If not found by SKU and looks like UUID, try by ID
	if err == sql.ErrNoRows {
		_, uuidErr := uuid.Parse(itemIdentifier)
		if uuidErr == nil {
			err = tx.QueryRow(`
				SELECT id FROM items WHERE id = $1 AND tenant_id = $2 AND is_active = true
			`, itemIdentifier, tenantID).Scan(&foundItemID)
		}
	}

	if err == sql.ErrNoRows {
		// Item doesn't exist, create a minimal one
		newItemID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO items (id, sku, name, uom, tenant_id, cost, price, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, 'EA', $4, 0, 0, true, NOW(), NOW())
		`, newItemID, itemIdentifier, itemIdentifier, tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to create item: %w", err)
		}
		return &newItemID, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to query item: %w", err)
	}

	return &foundItemID, nil
}

// ListAdjustments returns a paginated list of adjustments
func (h *Handler) ListAdjustments(c echo.Context) error {
	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

	// Parse query parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	status := c.QueryParam("status")
	reason := c.QueryParam("reason")
	search := c.QueryParam("search")

	// Build WHERE clause
	whereClause := "WHERE a.tenant_id = $1"
	args := []interface{}{tenantID}
	argCount := 1

	if status != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND a.status = $%d", argCount)
		args = append(args, status)
	}

	if reason != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND a.reason = $%d", argCount)
		args = append(args, reason)
	}

	if search != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND (a.number ILIKE $%d OR l.name ILIKE $%d OR a.notes ILIKE $%d)", argCount, argCount, argCount)
		args = append(args, "%"+search+"%")
	}

	// Get total count
	var total int64
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT a.id)
		FROM adjustments a
		LEFT JOIN locations l ON a.location_id = l.id
		%s
	`, whereClause)

	err := h.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to count adjustments")
	}

	// Get adjustments
	argCount++
	limitClause := fmt.Sprintf(" ORDER BY a.created_at DESC LIMIT $%d", argCount)
	args = append(args, limit)

	argCount++
	offsetClause := fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, offset)

	query := fmt.Sprintf(`
		SELECT a.id, a.number, a.location_id, a.reason, a.status, 
			   a.notes, a.created_by, a.approved_by, a.approved_at,
			   a.created_at, a.updated_at,
			   l.name as location_name, l.code as location_code
		FROM adjustments a
		LEFT JOIN locations l ON a.location_id = l.id
		%s%s%s
	`, whereClause, limitClause, offsetClause)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch adjustments")
	}
	defer rows.Close()

	var adjustments []Adjustment
	for rows.Next() {
		var adj Adjustment
		var notes sql.NullString
		var createdBy sql.NullString
		var approvedBy sql.NullString
		var approvedAt sql.NullTime
		var locationName sql.NullString
		var locationCode sql.NullString

		err := rows.Scan(
			&adj.ID, &adj.Number, &adj.LocationID, &adj.Reason, &adj.Status,
			&notes, &createdBy, &approvedBy, &approvedAt,
			&adj.CreatedAt, &adj.UpdatedAt,
			&locationName, &locationCode,
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to scan adjustment")
		}

		if notes.Valid {
			adj.Notes = &notes.String
		}
		if createdBy.Valid {
			adj.CreatedBy = &createdBy.String
		}
		if approvedBy.Valid {
			adj.ApprovedBy = &approvedBy.String
		}
		if approvedAt.Valid {
			adj.ApprovedAt = &approvedAt.Time
		}

		// Set location info
		if locationName.Valid || locationCode.Valid {
			adj.Location = &Location{
				ID:   adj.LocationID,
				Name: locationName.String,
				Code: locationCode.String,
			}
		}

		adjustments = append(adjustments, adj)
	}

	totalPages := int(total / int64(limit))
	if total%int64(limit) > 0 {
		totalPages++
	}

	response := PaginatedResponse{
		Data:       adjustments,
		Total:      total,
		Page:       page,
		PageSize:   limit,
		TotalPages: totalPages,
	}

	return c.JSON(http.StatusOK, response)
}

// GetAdjustment returns a single adjustment with its lines
func (h *Handler) GetAdjustment(c echo.Context) error {
	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

	id := c.Param("id")

	log.Printf("GetAdjustment called for ID: %s, TenantID: %s", id, tenantID)

	// Get adjustment
	var adj Adjustment
	adj.Location = &Location{} // Initialize before scanning
	var notes sql.NullString
	var createdBy sql.NullString
	var approvedBy sql.NullString
	var approvedAt sql.NullTime

	err := h.DB.QueryRow(`
		SELECT 
			a.id, a.number, a.location_id, a.reason, a.status,
			a.notes, a.created_by, a.approved_by, a.approved_at,
			a.created_at, a.updated_at,
			l.name as location_name, l.code as location_code
		FROM adjustments a
		LEFT JOIN locations l ON a.location_id = l.id
		WHERE a.id = $1 AND a.tenant_id = $2
	`, id, tenantID).Scan(
		&adj.ID, &adj.Number, &adj.LocationID, &adj.Reason, &adj.Status,
		&notes, &createdBy, &approvedBy, &approvedAt,
		&adj.CreatedAt, &adj.UpdatedAt,
		&adj.Location.Name, &adj.Location.Code,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Adjustment not found")
		}
		log.Printf("Failed to query adjustment: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch adjustment")
	}

	adj.TenantID = tenantID
	if notes.Valid {
		adj.Notes = &notes.String
	}
	if createdBy.Valid {
		adj.CreatedBy = &createdBy.String
	}
	if approvedBy.Valid {
		adj.ApprovedBy = &approvedBy.String
	}
	if approvedAt.Valid {
		adj.ApprovedAt = &approvedAt.Time
	}

	// Get adjustment lines
	linesRows, err := h.DB.Query(`
		SELECT al.id, al.item_id, al.item_identifier, COALESCE(al.notes, '') as notes, 
			   al.qty_expected, al.qty_actual, al.qty_diff,
			   COALESCE(i.sku, '') as sku, COALESCE(i.name, '') as name
		FROM adjustment_lines al
		LEFT JOIN items i ON al.item_id = i.id
		WHERE al.adjustment_id = $1 AND al.tenant_id = $2
	`, id, tenantID)
	if err != nil {
		log.Printf("Failed to execute adjustment lines query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch adjustment lines")
	}
	defer linesRows.Close()

	var lines []AdjustmentLine
	for linesRows.Next() {
		var line AdjustmentLine
		var itemID sql.NullString
		var itemIdentifier string
		var notes string
		var itemSKU string
		var itemName string

		err := linesRows.Scan(&line.ID, &itemID, &itemIdentifier, &notes,
			&line.QtyExpected, &line.QtyActual, &line.QtyDiff, &itemSKU, &itemName)
		if err != nil {
			log.Printf("Failed to scan adjustment line: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to scan adjustment line")
		}

		line.ItemIdentifier = itemIdentifier
		if notes != "" {
			line.Notes = &notes
		}

		if itemID.Valid {
			line.ItemID = &itemID.String
			// Only set Item if we have valid item data
			if itemSKU != "" || itemName != "" {
				item := Item{}
				item.SKU = itemSKU
				item.Name = itemName
				line.Item = &item
			}
		}

		lines = append(lines, line)
	}

	adj.Lines = lines

	return c.JSON(http.StatusOK, adj)
}

// CreateAdjustment creates a new adjustment
func (h *Handler) CreateAdjustment(c echo.Context) error {
	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID
	userID := claims.UserID

	var req CreateAdjustmentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Start transaction
	tx, err := h.DB.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start transaction")
	}
	defer tx.Rollback()

	// Validate location exists and belongs to tenant
	var locationExists bool
	err = tx.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM locations WHERE id = $1 AND tenant_id = $2 AND is_active = true)
	`, req.LocationID, tenantID).Scan(&locationExists)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to validate location")
	}
	if !locationExists {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid location")
	}

	// Generate adjustment ID and number
	adjustmentID := uuid.New().String()
	number := generateAdjustmentNumber()

	// Create adjustment
	_, err = tx.Exec(`
		INSERT INTO adjustments (id, number, location_id, tenant_id, reason, status, notes, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`, adjustmentID, number, req.LocationID, tenantID, req.Reason, "DRAFT", req.Notes, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create adjustment")
	}

	// Create adjustment lines
	for _, line := range req.Lines {
		lineID := uuid.New().String()
		qtyDiff := line.QtyActual - line.QtyExpected

		// Resolve or create item
		itemID, err := h.resolveOrCreateItemForAdjustment(tx, line.ItemID, tenantID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to resolve item: %v", err))
		}

		var itemIdentifier string = line.ItemID
		var notes string = line.Notes

		_, err = tx.Exec(`
			INSERT INTO adjustment_lines (id, adjustment_id, item_id, item_identifier, tenant_id, qty_expected, qty_actual, qty_diff, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		`, lineID, adjustmentID, itemID, itemIdentifier, tenantID, line.QtyExpected, line.QtyActual, qtyDiff, notes)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create adjustment line")
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	// Fetch the created adjustment
	adjustment := Adjustment{
		ID:         adjustmentID,
		Number:     number,
		LocationID: req.LocationID,
		TenantID:   tenantID,
		Reason:     req.Reason,
		Status:     "DRAFT",
		CreatedBy:  &userID,
	}

	return c.JSON(http.StatusCreated, adjustment)
}

// UpdateAdjustment updates an existing adjustment
func (h *Handler) UpdateAdjustment(c echo.Context) error {
	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

	id := c.Param("id")

	var req UpdateAdjustmentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Start transaction
	tx, err := h.DB.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start transaction")
	}
	defer tx.Rollback()

	// Check if adjustment exists and is modifiable
	var status string
	err = tx.QueryRow(`
		SELECT status FROM adjustments WHERE id = $1 AND tenant_id = $2
	`, id, tenantID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Adjustment not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch adjustment")
	}

	if status != "DRAFT" {
		return echo.NewHTTPError(http.StatusBadRequest, "Cannot modify non-draft adjustment")
	}

	// Validate location exists and belongs to tenant
	var locationExists bool
	err = tx.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM locations WHERE id = $1 AND tenant_id = $2 AND is_active = true)
	`, req.LocationID, tenantID).Scan(&locationExists)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to validate location")
	}
	if !locationExists {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid location")
	}

	// Update adjustment
	_, err = tx.Exec(`
		UPDATE adjustments 
		SET location_id = $1, reason = $2, notes = $3, updated_at = NOW()
		WHERE id = $4 AND tenant_id = $5
	`, req.LocationID, req.Reason, req.Notes, id, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update adjustment")
	}

	// Delete existing lines
	_, err = tx.Exec(`
		DELETE FROM adjustment_lines WHERE adjustment_id = $1 AND tenant_id = $2
	`, id, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete existing lines")
	}

	// Create new lines
	for _, line := range req.Lines {
		lineID := uuid.New().String()
		qtyDiff := line.QtyActual - line.QtyExpected

		// Resolve or create item
		itemID, err := h.resolveOrCreateItemForAdjustment(tx, line.ItemID, tenantID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to resolve item: %v", err))
		}

		var itemIdentifier string = line.ItemID
		var notes string = line.Notes

		_, err = tx.Exec(`
			INSERT INTO adjustment_lines (id, adjustment_id, item_id, item_identifier, tenant_id, qty_expected, qty_actual, qty_diff, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		`, lineID, id, itemID, itemIdentifier, tenantID, line.QtyExpected, line.QtyActual, qtyDiff, notes)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create adjustment line")
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Adjustment updated successfully"})
}

// DeleteAdjustment deletes an adjustment
func (h *Handler) DeleteAdjustment(c echo.Context) error {
	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

	id := c.Param("id")

	// Check if adjustment exists and is deletable
	var status string
	err := h.DB.QueryRow(`
		SELECT status FROM adjustments WHERE id = $1 AND tenant_id = $2
	`, id, tenantID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Adjustment not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch adjustment")
	}

	if status == "APPROVED" {
		return echo.NewHTTPError(http.StatusBadRequest, "Cannot delete approved adjustment")
	}

	// Delete adjustment (lines will be deleted by cascade)
	_, err = h.DB.Exec(`
		DELETE FROM adjustments WHERE id = $1 AND tenant_id = $2
	`, id, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete adjustment")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Adjustment deleted successfully"})
}

// ApproveAdjustment approves an adjustment and applies inventory changes
func (h *Handler) ApproveAdjustment(c echo.Context) error {
	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID
	userID := claims.UserID

	id := c.Param("id")

	// Start transaction
	tx, err := h.DB.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start transaction")
	}
	defer tx.Rollback()

	// Check if adjustment exists and can be approved
	var status, locationID string
	err = tx.QueryRow(`
		SELECT status, location_id FROM adjustments WHERE id = $1 AND tenant_id = $2
	`, id, tenantID).Scan(&status, &locationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Adjustment not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch adjustment")
	}

	if status != "DRAFT" {
		return echo.NewHTTPError(http.StatusBadRequest, "Can only approve draft adjustments")
	}

	// Get adjustment lines
	linesRows, err := tx.Query(`
		SELECT item_id, qty_diff FROM adjustment_lines 
		WHERE adjustment_id = $1 AND tenant_id = $2 AND item_id IS NOT NULL
	`, id, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch adjustment lines")
	}
	defer linesRows.Close()

	// Apply inventory changes
	for linesRows.Next() {
		var itemID string
		var qtyDiff int

		err := linesRows.Scan(&itemID, &qtyDiff)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to scan adjustment line")
		}

		if qtyDiff != 0 {
			// Update inventory levels
			_, err = tx.Exec(`
				INSERT INTO inventory_levels (item_id, location_id, on_hand, allocated, reorder_point, reorder_qty, created_at, updated_at)
				VALUES ($1, $2, $3, 0, 0, 0, NOW(), NOW())
				ON CONFLICT (item_id, location_id)
				DO UPDATE SET 
					on_hand = inventory_levels.on_hand + $3,
					updated_at = NOW()
			`, itemID, locationID, qtyDiff)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update inventory")
			}

			// Create stock movement record
			_, err = tx.Exec(`
				INSERT INTO stock_movements (id, item_id, location_id, user_id, qty, reason, reference, ref_id, occurred_at, created_at)
				VALUES ($1, $2, $3, $4, $5, 'ADJUSTMENT', 'Adjustment', $6, NOW(), NOW())
			`, uuid.New().String(), itemID, locationID, userID, qtyDiff, id)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create stock movement")
			}
		}
	}

	// Update adjustment status
	_, err = tx.Exec(`
		UPDATE adjustments 
		SET status = 'APPROVED', approved_by = $1, approved_at = NOW(), updated_at = NOW()
		WHERE id = $2 AND tenant_id = $3
	`, userID, id, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to approve adjustment")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Adjustment approved successfully"})
}
