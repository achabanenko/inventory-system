package handlers

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	appmw "inventory/internal/middleware"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

type Transfer struct {
	ID             string         `json:"id"`
	Number         string         `json:"number"`
	FromLocationID string         `json:"from_location_id"`
	FromLocation   *Location      `json:"from_location,omitempty"`
	ToLocationID   string         `json:"to_location_id"`
	ToLocation     *Location      `json:"to_location,omitempty"`
	Status         string         `json:"status"`
	Notes          string         `json:"notes"`
	CreatedBy      *string        `json:"created_by"`
	ApprovedBy     *string        `json:"approved_by"`
	ShippedAt      *time.Time     `json:"shipped_at"`
	ReceivedAt     *time.Time     `json:"received_at"`
	Lines          []TransferLine `json:"lines,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type TransferLine struct {
	ID             string  `json:"id"`
	ItemID         *string `json:"item_id,omitempty"`
	ItemIdentifier string  `json:"item_identifier"`
	Description    string  `json:"description"`
	Item           *Item   `json:"item,omitempty"`
	Qty            int     `json:"qty"`
}

type CreateTransferRequest struct {
	FromLocationID string `json:"from_location_id"`
	ToLocationID   string `json:"to_location_id"`
	Notes          string `json:"notes"`
	Lines          []struct {
		ItemID      string `json:"item_id"`
		Description string `json:"description"`
		Qty         int    `json:"qty"`
	} `json:"lines"`
}

type UpdateTransferRequest struct {
	Notes string `json:"notes"`
	Lines []struct {
		ItemID      string `json:"item_id"`
		Description string `json:"description"`
		Qty         int    `json:"qty"`
	} `json:"lines"`
}

func (h *Handler) ListTransfers(c echo.Context) error {
	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

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
	status := c.QueryParam("status")
	fromLocationID := c.QueryParam("from_location_id")
	toLocationID := c.QueryParam("to_location_id")
	sort := c.QueryParam("sort")
	if sort == "" {
		sort = "created_at DESC"
	}

	offset := (page - 1) * pageSize

	// Build query
	query := `
		SELECT
			t.id, t.number, t.status, t.from_location_id, t.to_location_id,
			t.notes, t.created_by, t.approved_by, t.shipped_at, t.received_at,
			t.created_at, t.updated_at,
			fl.name as from_location_name, fl.code as from_location_code,
			tl.name as to_location_name, tl.code as to_location_code
		FROM transfers t
		LEFT JOIN locations fl ON t.from_location_id = fl.id
		LEFT JOIN locations tl ON t.to_location_id = tl.id
		WHERE t.tenant_id = $1`

	args := []interface{}{tenantID}
	argCount := 1

	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND t.number ILIKE $%d", argCount)
		args = append(args, "%"+search+"%")
	}

	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND t.status = $%d", argCount)
		args = append(args, status)
	}

	if fromLocationID != "" {
		argCount++
		query += fmt.Sprintf(" AND t.from_location_id = $%d", argCount)
		args = append(args, fromLocationID)
	}

	if toLocationID != "" {
		argCount++
		query += fmt.Sprintf(" AND t.to_location_id = $%d", argCount)
		args = append(args, toLocationID)
	}

	// Add sorting
	switch sort {
	case "number", "number ASC":
		query += " ORDER BY t.number ASC"
	case "number DESC":
		query += " ORDER BY t.number DESC"
	case "status", "status ASC":
		query += " ORDER BY t.status ASC"
	case "status DESC":
		query += " ORDER BY t.status DESC"
	case "created_at", "created_at ASC":
		query += " ORDER BY t.created_at ASC"
	default:
		query += " ORDER BY t.created_at DESC"
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM (" + query + ") as subquery"
	var total int
	err := h.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to count transfers")
	}

	// Add pagination
	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, pageSize)

	argCount++
	query += fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, offset)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch transfers")
	}
	defer rows.Close()

	var transfers []Transfer
	for rows.Next() {
		var t Transfer
		var fromLocationName, fromLocationCode, toLocationName, toLocationCode sql.NullString
		var notes sql.NullString

		err := rows.Scan(
			&t.ID, &t.Number, &t.Status, &t.FromLocationID, &t.ToLocationID,
			&notes, &t.CreatedBy, &t.ApprovedBy, &t.ShippedAt, &t.ReceivedAt,
			&t.CreatedAt, &t.UpdatedAt,
			&fromLocationName, &fromLocationCode,
			&toLocationName, &toLocationCode,
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to scan transfer")
		}

		if notes.Valid {
			t.Notes = notes.String
		}

		if fromLocationName.Valid {
			t.FromLocation = &Location{
				ID:   t.FromLocationID,
				Name: fromLocationName.String,
				Code: fromLocationCode.String,
			}
		}

		if toLocationName.Valid {
			t.ToLocation = &Location{
				ID:   t.ToLocationID,
				Name: toLocationName.String,
				Code: toLocationCode.String,
			}
		}

		transfers = append(transfers, t)
	}

	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       transfers,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + pageSize - 1) / pageSize,
		Total:      int64(total),
	})
}

func (h *Handler) CreateTransfer(c echo.Context) error {
	// Get user claims for tenant ID and user ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID
	userID := claims.UserID

	var req CreateTransferRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Log the incoming request data
	log.Printf("CreateTransfer request - UserID: %s, TenantID: %s, FromLocationID: %s, ToLocationID: %s, Lines: %d",
		userID, tenantID, req.FromLocationID, req.ToLocationID, len(req.Lines))

	if req.FromLocationID == req.ToLocationID {
		return echo.NewHTTPError(http.StatusBadRequest, "From and to locations cannot be the same")
	}

	if len(req.Lines) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Transfer must have at least one item")
	}

	// Validate that locations exist
	var fromLocationExists, toLocationExists bool
	err := h.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM locations WHERE id = $1 AND tenant_id = $2 AND is_active = true)
	`, req.FromLocationID, tenantID).Scan(&fromLocationExists)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to validate from location")
	}

	err = h.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM locations WHERE id = $1 AND tenant_id = $2 AND is_active = true)
	`, req.ToLocationID, tenantID).Scan(&toLocationExists)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to validate to location")
	}

	if !fromLocationExists {
		return echo.NewHTTPError(http.StatusBadRequest, "From location does not exist or is inactive")
	}

	if !toLocationExists {
		return echo.NewHTTPError(http.StatusBadRequest, "To location does not exist or is inactive")
	}

	// Validate stock availability before creating transfer (only for existing items)
	for _, line := range req.Lines {
		// Try to resolve item (could be SKU or item ID)
		var itemID string
		err := h.DB.QueryRow(`
			SELECT id FROM items WHERE (sku = $1 OR id = $1) AND tenant_id = $2 AND is_active = true
		`, line.ItemID, tenantID).Scan(&itemID)

		// If item exists, check stock availability
		if err == nil {
			// Check if source location has enough stock
			var currentStock int
			err = h.DB.QueryRow(`
				SELECT COALESCE(qty, 0) FROM inventory
				WHERE item_id = $1 AND location_id = $2 AND tenant_id = $3
			`, itemID, req.FromLocationID, tenantID).Scan(&currentStock)
			if err != nil && err != sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to check inventory")
			}

			if currentStock < line.Qty {
				return echo.NewHTTPError(http.StatusBadRequest,
					fmt.Sprintf("Insufficient stock for item '%s'. Available: %d, Requested: %d",
						line.ItemID, currentStock, line.Qty))
			}
		}
		// If item doesn't exist, skip stock validation (allow any item code)
	}

	// Start transaction
	tx, err := h.DB.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	defer tx.Rollback()

	// Generate transfer number
	number := fmt.Sprintf("TRF-%d", time.Now().Unix())

	// Create transfer
	transferID := uuid.New().String()
	_, err = tx.Exec(`
		INSERT INTO transfers (id, number, from_location_id, to_location_id, tenant_id, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`, transferID, number, req.FromLocationID, req.ToLocationID, tenantID, "DRAFT", req.Notes)
	if err != nil {
		log.Printf("Failed to create transfer: %v. Parameters: transferID=%s, number=%s, fromLocationID=%s, toLocationID=%s, tenantID=%s",
			err, transferID, number, req.FromLocationID, req.ToLocationID, tenantID)

		// Log additional details
		if pqErr, ok := err.(*pq.Error); ok {
			log.Printf("PostgreSQL Error - Code: %s, Message: %s, Detail: %s, Constraint: %s",
				pqErr.Code, pqErr.Message, pqErr.Detail, pqErr.Constraint)
			if pqErr.Code == "23505" {
				return echo.NewHTTPError(http.StatusConflict, "Transfer number already exists")
			}
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create transfer: %v", err))
	}

	// Verify transfer was created successfully
	var createdTransferID string
	err = tx.QueryRow(`SELECT id FROM transfers WHERE id = $1`, transferID).Scan(&createdTransferID)
	if err != nil {
		log.Printf("Transfer verification failed: %v. TransferID: %s", err, transferID)
		return echo.NewHTTPError(http.StatusInternalServerError, "Transfer creation verification failed")
	}
	log.Printf("Transfer verified successfully: %s", createdTransferID)

	// Create transfer lines with item identifiers (no auto-creation)
	for _, line := range req.Lines {
		var itemID sql.NullString
		var itemIdentifier string = line.ItemID
		var description string = line.Description

		// Try to find existing item - first by SKU, then by ID if it looks like UUID
		var foundItemID string
		var err error

		// First try by SKU
		err = tx.QueryRow(`
			SELECT id FROM items WHERE sku = $1 AND tenant_id = $2 AND is_active = true
		`, line.ItemID, tenantID).Scan(&foundItemID)

		// If not found by SKU and looks like UUID, try by ID
		if err == sql.ErrNoRows {
			_, uuidErr := uuid.Parse(line.ItemID)
			if uuidErr == nil {
				err = tx.QueryRow(`
					SELECT id FROM items WHERE id = $1 AND tenant_id = $2 AND is_active = true
				`, line.ItemID, tenantID).Scan(&foundItemID)
			}
		}

		if err == nil {
			// Item exists, use it
			itemID.String = foundItemID
			itemID.Valid = true
			log.Printf("Item found: %s -> %s", line.ItemID, foundItemID)
		} else if err != sql.ErrNoRows {
			// Log unexpected errors (not just "not found")
			log.Printf("Item lookup error for '%s': %v", line.ItemID, err)
		}
		// If item doesn't exist, itemID remains invalid (NULL)

		lineID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO transfer_lines (id, transfer_id, item_id, item_identifier, description, tenant_id, qty, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		`, lineID, transferID, itemID, itemIdentifier, description, tenantID, line.Qty)
		if err != nil {
			log.Printf("Failed to create transfer line: %v. Parameters: lineID=%s, transferID=%s, itemID=%+v, itemIdentifier=%s, tenantID=%s, qty=%d",
				err, lineID, transferID, itemID, itemIdentifier, tenantID, line.Qty)

			// Log additional PostgreSQL error details
			if pqErr, ok := err.(*pq.Error); ok {
				log.Printf("PostgreSQL Error in transfer_lines - Code: %s, Message: %s, Detail: %s, Constraint: %s",
					pqErr.Code, pqErr.Message, pqErr.Detail, pqErr.Constraint)
			}
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create transfer line: %v", err))
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	// Return created transfer
	transfer := Transfer{
		ID:             transferID,
		Number:         number,
		FromLocationID: req.FromLocationID,
		ToLocationID:   req.ToLocationID,
		Status:         "DRAFT",
		Notes:          req.Notes,
		CreatedBy:      &userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	return c.JSON(http.StatusCreated, transfer)
}

func (h *Handler) GetTransfer(c echo.Context) error {
	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

	id := c.Param("id")

	log.Printf("GetTransfer called for ID: %s, TenantID: %s", id, tenantID)

	// Get transfer
	var t Transfer
	t.FromLocation = &Location{} // Initialize before scanning
	t.ToLocation = &Location{}   // Initialize before scanning
	var notes sql.NullString

	err := h.DB.QueryRow(`
		SELECT
			t.id, t.number, t.status, t.from_location_id, t.to_location_id,
			t.notes, t.created_by, t.approved_by, t.shipped_at, t.received_at,
			t.created_at, t.updated_at,
			fl.name as from_location_name, fl.code as from_location_code,
			tl.name as to_location_name, tl.code as to_location_code
		FROM transfers t
		LEFT JOIN locations fl ON t.from_location_id = fl.id
		LEFT JOIN locations tl ON t.to_location_id = tl.id
		WHERE t.id = $1 AND t.tenant_id = $2
	`, id, tenantID).Scan(
		&t.ID, &t.Number, &t.Status, &t.FromLocationID, &t.ToLocationID,
		&notes, &t.CreatedBy, &t.ApprovedBy, &t.ShippedAt, &t.ReceivedAt,
		&t.CreatedAt, &t.UpdatedAt,
		&t.FromLocation.Name, &t.FromLocation.Code,
		&t.ToLocation.Name, &t.ToLocation.Code,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Transfer not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch transfer")
	}

	if notes.Valid {
		t.Notes = notes.String
	}

	// Get transfer lines
	linesRows, err := h.DB.Query(`
		SELECT tl.id, tl.item_id, tl.item_identifier, COALESCE(tl.description, '') as description, tl.qty, COALESCE(i.sku, '') as sku, COALESCE(i.name, '') as name
		FROM transfer_lines tl
		LEFT JOIN items i ON tl.item_id = i.id
		WHERE tl.transfer_id = $1 AND tl.tenant_id = $2
	`, id, tenantID)
	if err != nil {
		log.Printf("Failed to execute transfer lines query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch transfer lines")
	}
	defer linesRows.Close()

	var lines []TransferLine
	for linesRows.Next() {
		var line TransferLine
		var itemID sql.NullString
		var itemIdentifier string
		var itemSKU string
		var itemName string
		err := linesRows.Scan(&line.ID, &itemID, &itemIdentifier, &line.Description, &line.Qty, &itemSKU, &itemName)
		if err != nil {
			log.Printf("Failed to scan transfer line: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to scan transfer line")
		}

		line.ItemIdentifier = itemIdentifier

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
		// If itemID is NULL, Item will remain nil (which is correct for non-existent items)

		lines = append(lines, line)
	}

	t.Lines = lines

	return c.JSON(http.StatusOK, t)
}

func (h *Handler) UpdateTransfer(c echo.Context) error {
	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

	id := c.Param("id")
	var req UpdateTransferRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Check if transfer exists and is in DRAFT status
	var status, fromLocationID string
	err := h.DB.QueryRow(`
		SELECT status, from_location_id FROM transfers WHERE id = $1 AND tenant_id = $2
	`, id, tenantID).Scan(&status, &fromLocationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Transfer not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch transfer")
	}

	if status != "DRAFT" {
		return echo.NewHTTPError(http.StatusBadRequest, "Can only update draft transfers")
	}

	// Start transaction
	tx, err := h.DB.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	defer tx.Rollback()

	// Update transfer
	_, err = tx.Exec(`
		UPDATE transfers SET notes = $1, updated_at = NOW()
		WHERE id = $2 AND tenant_id = $3
	`, req.Notes, id, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update transfer")
	}

	// No stock validation - allow any item codes and descriptions

	// Delete existing lines
	_, err = tx.Exec(`DELETE FROM transfer_lines WHERE transfer_id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete existing lines")
	}

	// Create new lines with item identifiers (no auto-creation)
	for _, line := range req.Lines {
		var itemID sql.NullString
		var itemIdentifier string = line.ItemID
		var description string = line.Description

		// Try to find existing item - first by SKU, then by ID if it looks like UUID
		var foundItemID string
		var err error

		// First try by SKU
		err = tx.QueryRow(`
			SELECT id FROM items WHERE sku = $1 AND tenant_id = $2 AND is_active = true
		`, line.ItemID, tenantID).Scan(&foundItemID)

		// If not found by SKU and looks like UUID, try by ID
		if err == sql.ErrNoRows {
			_, uuidErr := uuid.Parse(line.ItemID)
			if uuidErr == nil {
				err = tx.QueryRow(`
					SELECT id FROM items WHERE id = $1 AND tenant_id = $2 AND is_active = true
				`, line.ItemID, tenantID).Scan(&foundItemID)
			}
		}

		if err == nil {
			// Item exists, use it
			itemID.String = foundItemID
			itemID.Valid = true
		} else if err != sql.ErrNoRows {
			// Log unexpected errors (not just "not found")
			log.Printf("Item lookup error for '%s' in UpdateTransfer: %v", line.ItemID, err)
		}
		// If item doesn't exist, itemID remains invalid (NULL)

		lineID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO transfer_lines (id, transfer_id, item_id, item_identifier, description, tenant_id, qty, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		`, lineID, id, itemID, itemIdentifier, description, tenantID, line.Qty)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create transfer line")
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Transfer updated successfully"})
}

func (h *Handler) DeleteTransfer(c echo.Context) error {
	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

	id := c.Param("id")

	// Check if transfer exists and is in DRAFT status
	var status string
	err := h.DB.QueryRow(`
		SELECT status FROM transfers WHERE id = $1 AND tenant_id = $2
	`, id, tenantID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Transfer not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch transfer")
	}

	if status != "DRAFT" {
		return echo.NewHTTPError(http.StatusBadRequest, "Can only delete draft transfers")
	}

	// Delete transfer (lines will be deleted automatically due to CASCADE)
	_, err = h.DB.Exec(`DELETE FROM transfers WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete transfer")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Transfer deleted successfully"})
}

func (h *Handler) ApproveTransfer(c echo.Context) error {
	// Get user claims for tenant ID and user ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID
	userID := claims.UserID

	id := c.Param("id")

	// Check if transfer exists and is in DRAFT status
	var status string
	err := h.DB.QueryRow(`
		SELECT status FROM transfers WHERE id = $1 AND tenant_id = $2
	`, id, tenantID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Transfer not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch transfer")
	}

	if status != "DRAFT" {
		return echo.NewHTTPError(http.StatusBadRequest, "Can only approve draft transfers")
	}

	// Update transfer status to IN_TRANSIT
	_, err = h.DB.Exec(`
		UPDATE transfers SET status = 'IN_TRANSIT', approved_by = $1, updated_at = NOW()
		WHERE id = $2 AND tenant_id = $3
	`, userID, id, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to approve transfer")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Transfer approved successfully"})
}

func (h *Handler) ShipTransfer(c echo.Context) error {
	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

	id := c.Param("id")

	// Check if transfer exists and is in IN_TRANSIT status
	var status string
	err := h.DB.QueryRow(`
		SELECT status FROM transfers WHERE id = $1 AND tenant_id = $2
	`, id, tenantID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Transfer not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch transfer")
	}

	if status != "IN_TRANSIT" {
		return echo.NewHTTPError(http.StatusBadRequest, "Can only ship in-transit transfers")
	}

	// Update transfer status to RECEIVED and set shipped timestamp
	_, err = h.DB.Exec(`
		UPDATE transfers SET status = 'RECEIVED', shipped_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`, id, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to ship transfer")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Transfer shipped successfully"})
}

func (h *Handler) ReceiveTransfer(c echo.Context) error {
	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

	id := c.Param("id")

	// Check if transfer exists and is in RECEIVED status
	var status string
	var transfer Transfer
	err := h.DB.QueryRow(`
		SELECT id, from_location_id, to_location_id FROM transfers WHERE id = $1 AND tenant_id = $2
	`, id, tenantID).Scan(&transfer.ID, &transfer.FromLocationID, &transfer.ToLocationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Transfer not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch transfer")
	}

	// Get current status
	err = h.DB.QueryRow(`SELECT status FROM transfers WHERE id = $1 AND tenant_id = $2`, id, tenantID).Scan(&status)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch transfer status")
	}

	if status != "RECEIVED" {
		return echo.NewHTTPError(http.StatusBadRequest, "Can only receive received transfers")
	}

	// Start transaction to update stock levels
	tx, err := h.DB.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	defer tx.Rollback()

	// Get transfer lines
	lines, err := tx.Query(`
		SELECT item_id, qty FROM transfer_lines WHERE transfer_id = $1 AND tenant_id = $2
	`, id, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch transfer lines")
	}
	defer lines.Close()

	// Update inventory for each line
	for lines.Next() {
		var itemID string
		var qty int
		err := lines.Scan(&itemID, &qty)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to scan transfer line")
		}

		// Reduce stock at source location
		_, err = tx.Exec(`
			INSERT INTO inventory (item_id, location_id, tenant_id, qty, updated_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (item_id, location_id, tenant_id)
			DO UPDATE SET qty = inventory.qty - EXCLUDED.qty, updated_at = NOW()
		`, itemID, transfer.FromLocationID, tenantID, qty)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update source inventory")
		}

		// Increase stock at destination location
		_, err = tx.Exec(`
			INSERT INTO inventory (item_id, location_id, tenant_id, qty, updated_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (item_id, location_id, tenant_id)
			DO UPDATE SET qty = inventory.qty + EXCLUDED.qty, updated_at = NOW()
		`, itemID, transfer.ToLocationID, tenantID, qty)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update destination inventory")
		}
	}

	// Update transfer status to COMPLETED and set received timestamp
	_, err = tx.Exec(`
		UPDATE transfers SET status = 'COMPLETED', received_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`, id, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to complete transfer")
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Transfer received successfully"})
}
