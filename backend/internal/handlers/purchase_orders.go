package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	appmw "inventory/internal/middleware"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
)

type Supplier struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Item struct {
	ID   string `json:"id"`
	SKU  string `json:"sku"`
	Name string `json:"name"`
}

type PurchaseOrder struct {
	ID         string              `json:"id"`
	Number     string              `json:"number"`
	Status     string              `json:"status"`
	SupplierID string              `json:"supplier_id"`
	Supplier   *Supplier           `json:"supplier,omitempty"`
	CreatedBy  string              `json:"created_by"`
	ApprovedBy *string             `json:"approved_by,omitempty"`
	ExpectedAt *time.Time          `json:"expected_at,omitempty"`
	ApprovedAt *time.Time          `json:"approved_at,omitempty"`
	Notes      *string             `json:"notes,omitempty"`
	Lines      []PurchaseOrderLine `json:"lines,omitempty"`
	Total      decimal.Decimal     `json:"total"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}

type PurchaseOrderLine struct {
	ID          string          `json:"id"`
	ItemID      string          `json:"item_id"`
	Item        *Item           `json:"item,omitempty"`
	QtyOrdered  int             `json:"qty_ordered"`
	QtyReceived int             `json:"qty_received"`
	UnitCost    decimal.Decimal `json:"unit_cost"`
	Tax         interface{}     `json:"tax,omitempty"`
	LineTotal   decimal.Decimal `json:"line_total"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type CreatePurchaseOrderRequest struct {
	SupplierID string                           `json:"supplier_id" validate:"required"`
	ExpectedAt *string                          `json:"expected_at"`
	Notes      *string                          `json:"notes"`
	Lines      []CreatePurchaseOrderLineRequest `json:"lines" validate:"required,min=1"`
}

type CreatePurchaseOrderLineRequest struct {
	ItemID     string      `json:"item_id" validate:"required"`
	QtyOrdered int         `json:"qty_ordered" validate:"required,min=1"`
	UnitCost   string      `json:"unit_cost" validate:"required"`
	Tax        interface{} `json:"tax,omitempty"`
}

type UpdatePurchaseOrderRequest struct {
	SupplierID string                           `json:"supplier_id"`
	ExpectedAt *string                          `json:"expected_at"`
	Notes      *string                          `json:"notes"`
	Lines      []UpdatePurchaseOrderLineRequest `json:"lines"`
}

type UpdatePurchaseOrderLineRequest struct {
	ID         *string     `json:"id,omitempty"`
	ItemID     string      `json:"item_id" validate:"required"`
	QtyOrdered int         `json:"qty_ordered" validate:"required,min=1"`
	UnitCost   string      `json:"unit_cost" validate:"required"`
	Tax        interface{} `json:"tax,omitempty"`
}

func (h *Handler) ListPurchaseOrders(c echo.Context) error {
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
	supplierID := c.QueryParam("supplier_id")
	sort := c.QueryParam("sort")
	if sort == "" {
		sort = "created_at DESC"
	}

	offset := (page - 1) * pageSize

	// Build query
	query := `
		SELECT 
			po.id, po.number, po.status, po.supplier_id, po.created_by, 
			po.approved_by, po.expected_at, po.approved_at, po.notes,
			po.created_at, po.updated_at,
			s.name as supplier_name,
			COALESCE(SUM(pol.qty_ordered * pol.unit_cost), 0) as total
		FROM purchase_orders po
		LEFT JOIN suppliers s ON po.supplier_id = s.id
		LEFT JOIN purchase_order_lines pol ON po.id = pol.purchase_order_id
		WHERE 1=1`

	args := []interface{}{}
	argCount := 0

	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND (po.number ILIKE $%d)", argCount)
		args = append(args, "%"+search+"%")
	}

	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND po.status = $%d", argCount)
		args = append(args, status)
	}

	if supplierID != "" {
		argCount++
		query += fmt.Sprintf(" AND po.supplier_id = $%d", argCount)
		args = append(args, supplierID)
	}

	query += " GROUP BY po.id, s.name"

	// Add sorting
	switch sort {
	case "number", "number ASC":
		query += " ORDER BY po.number ASC"
	case "number DESC":
		query += " ORDER BY po.number DESC"
	case "status", "status ASC":
		query += " ORDER BY po.status ASC"
	case "status DESC":
		query += " ORDER BY po.status DESC"
	case "created_at", "created_at ASC":
		query += " ORDER BY po.created_at ASC"
	default:
		query += " ORDER BY po.created_at DESC"
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
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	var purchaseOrders []PurchaseOrder
	for rows.Next() {
		var po PurchaseOrder
		var supplierName sql.NullString
		var approvedBy sql.NullString
		var expectedAt sql.NullTime
		var approvedAt sql.NullTime
		var notes sql.NullString
		var total string

		err := rows.Scan(
			&po.ID, &po.Number, &po.Status, &po.SupplierID, &po.CreatedBy,
			&approvedBy, &expectedAt, &approvedAt, &notes,
			&po.CreatedAt, &po.UpdatedAt, &supplierName, &total,
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Database scan error")
		}

		if approvedBy.Valid {
			po.ApprovedBy = &approvedBy.String
		}
		if expectedAt.Valid {
			po.ExpectedAt = &expectedAt.Time
		}
		if approvedAt.Valid {
			po.ApprovedAt = &approvedAt.Time
		}
		if notes.Valid {
			po.Notes = &notes.String
		}

		// Parse total
		po.Total, _ = decimal.NewFromString(total)

		// Add supplier info if available
		if supplierName.Valid {
			po.Supplier = &Supplier{
				ID:   po.SupplierID,
				Name: supplierName.String,
			}
		}

		purchaseOrders = append(purchaseOrders, po)
	}

	// Get total count
	countQuery := `
		SELECT COUNT(DISTINCT po.id)
		FROM purchase_orders po
		LEFT JOIN suppliers s ON po.supplier_id = s.id
		WHERE 1=1`

	countArgs := []interface{}{}
	countArgCount := 0

	if search != "" {
		countArgCount++
		countQuery += fmt.Sprintf(" AND (po.number ILIKE $%d)", countArgCount)
		countArgs = append(countArgs, "%"+search+"%")
	}

	if status != "" {
		countArgCount++
		countQuery += fmt.Sprintf(" AND po.status = $%d", countArgCount)
		countArgs = append(countArgs, status)
	}

	if supplierID != "" {
		countArgCount++
		countQuery += fmt.Sprintf(" AND po.supplier_id = $%d", countArgCount)
		countArgs = append(countArgs, supplierID)
	}

	var total int
	err = h.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	totalPages := (total + pageSize - 1) / pageSize

	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       purchaseOrders,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Total:      int64(total),
	})
}

func (h *Handler) CreatePurchaseOrder(c echo.Context) error {
	var req CreatePurchaseOrderRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Get user ID from context (set by auth middleware)
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	userID := claims.UserID

	// Generate PO number
	var maxNumber int
	err := h.DB.QueryRow(`
		SELECT COALESCE(MAX(CAST(SUBSTRING(number FROM 'PO-([0-9]+)') AS INTEGER)), 0)
		FROM purchase_orders 
		WHERE number ~ '^PO-[0-9]+$'
	`).Scan(&maxNumber)
	if err != nil && err != sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	poNumber := fmt.Sprintf("PO-%06d", maxNumber+1)

	// Start transaction
	tx, err := h.DB.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	defer tx.Rollback()

	// Parse expected_at date if provided
	var expectedAt *time.Time
	if req.ExpectedAt != nil && *req.ExpectedAt != "" {
		if parsedTime, err := time.Parse("2006-01-02", *req.ExpectedAt); err == nil {
			expectedAt = &parsedTime
		} else if parsedTime, err := time.Parse(time.RFC3339, *req.ExpectedAt); err == nil {
			expectedAt = &parsedTime
		}
	}

	// Create purchase order
	poID := uuid.New().String()
	_, err = tx.Exec(`
		INSERT INTO purchase_orders (id, number, status, supplier_id, tenant_id, created_by, expected_at, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`, poID, poNumber, "DRAFT", req.SupplierID, claims.TenantID, userID, expectedAt, req.Notes)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create purchase order")
	}

	// Create purchase order lines
	var lines []PurchaseOrderLine
	for _, lineReq := range req.Lines {
		lineID := uuid.New().String()
		// Convert string unit cost to decimal for resolveOrCreateItem
		unitCostDecimal, err := decimal.NewFromString(lineReq.UnitCost)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid unit cost: %s", lineReq.UnitCost))
		}
		// Resolve or create item by provided identifier (UUID or SKU)
		resolvedItemID, resErr := h.resolveOrCreateItem(tx, lineReq.ItemID, &unitCostDecimal, claims.TenantID)
		if resErr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, resErr.Error())
		}
		// Ensure proper types for DB: numeric and jsonb
		unitCostStr := unitCostDecimal.StringFixed(2)
		var taxJSON *string
		if lineReq.Tax != nil {
			if b, marshalErr := json.Marshal(lineReq.Tax); marshalErr == nil {
				s := string(b)
				taxJSON = &s
			}
		}

		_, err = tx.Exec(`
            INSERT INTO purchase_order_lines (id, purchase_order_id, item_id, qty_ordered, qty_received, unit_cost, tax, created_at, updated_at)
            VALUES ($1, $2, $3, $4, 0, $5::numeric, COALESCE($6::jsonb, '{}'::jsonb), NOW(), NOW())
        `, lineID, poID, resolvedItemID, lineReq.QtyOrdered, unitCostStr, taxJSON)
		if err != nil {
			c.Logger().Errorf("failed to create purchase order line: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create purchase order line")
		}

		// Calculate line total
		lineTotal := unitCostDecimal.Mul(decimal.NewFromInt(int64(lineReq.QtyOrdered)))

		lines = append(lines, PurchaseOrderLine{
			ID:          lineID,
			ItemID:      resolvedItemID,
			QtyOrdered:  lineReq.QtyOrdered,
			QtyReceived: 0,
			UnitCost:    unitCostDecimal,
			Tax:         lineReq.Tax,
			LineTotal:   lineTotal,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	// Calculate total
	var total decimal.Decimal
	for _, line := range lines {
		total = total.Add(line.LineTotal)
	}

	// Return created purchase order
	po := PurchaseOrder{
		ID:         poID,
		Number:     poNumber,
		Status:     "DRAFT",
		SupplierID: req.SupplierID,
		CreatedBy:  userID,
		ExpectedAt: expectedAt,
		Notes:      req.Notes,
		Lines:      lines,
		Total:      total,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return c.JSON(http.StatusCreated, po)
}

func (h *Handler) GetPurchaseOrder(c echo.Context) error {
	id := c.Param("id")

	// Get purchase order
	var po PurchaseOrder
	var supplierName sql.NullString
	var approvedBy sql.NullString
	var expectedAt sql.NullTime
	var approvedAt sql.NullTime
	var notes sql.NullString

	err := h.DB.QueryRow(`
		SELECT 
			po.id, po.number, po.status, po.supplier_id, po.created_by,
			po.approved_by, po.expected_at, po.approved_at, po.notes,
			po.created_at, po.updated_at,
			s.name as supplier_name
		FROM purchase_orders po
		LEFT JOIN suppliers s ON po.supplier_id = s.id
		WHERE po.id = $1
	`, id).Scan(
		&po.ID, &po.Number, &po.Status, &po.SupplierID, &po.CreatedBy,
		&approvedBy, &expectedAt, &approvedAt, &notes,
		&po.CreatedAt, &po.UpdatedAt, &supplierName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Purchase order not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	if approvedBy.Valid {
		po.ApprovedBy = &approvedBy.String
	}
	if expectedAt.Valid {
		po.ExpectedAt = &expectedAt.Time
	}
	if approvedAt.Valid {
		po.ApprovedAt = &approvedAt.Time
	}
	if notes.Valid {
		po.Notes = &notes.String
	}

	// Add supplier info if available
	if supplierName.Valid {
		po.Supplier = &Supplier{
			ID:   po.SupplierID,
			Name: supplierName.String,
		}
	}

	// Get purchase order lines
	rows, err := h.DB.Query(`
		SELECT 
			pol.id, pol.item_id, pol.qty_ordered, pol.qty_received, 
			pol.unit_cost, pol.tax, pol.created_at, pol.updated_at,
			i.sku, i.name as item_name
		FROM purchase_order_lines pol
		LEFT JOIN items i ON pol.item_id = i.id
		WHERE pol.purchase_order_id = $1
		ORDER BY pol.created_at
	`, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	var lines []PurchaseOrderLine
	var total decimal.Decimal

	for rows.Next() {
		var line PurchaseOrderLine
		var unitCostStr string
		var itemSKU, itemName sql.NullString

		err := rows.Scan(
			&line.ID, &line.ItemID, &line.QtyOrdered, &line.QtyReceived,
			&unitCostStr, &line.Tax, &line.CreatedAt, &line.UpdatedAt,
			&itemSKU, &itemName,
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Database scan error")
		}

		// Parse unit cost
		line.UnitCost, _ = decimal.NewFromString(unitCostStr)
		line.LineTotal = line.UnitCost.Mul(decimal.NewFromInt(int64(line.QtyOrdered)))
		total = total.Add(line.LineTotal)

		// Add item info if available
		if itemSKU.Valid && itemName.Valid {
			line.Item = &Item{
				ID:   line.ItemID,
				SKU:  itemSKU.String,
				Name: itemName.String,
			}
		}

		lines = append(lines, line)
	}

	po.Lines = lines
	po.Total = total

	return c.JSON(http.StatusOK, po)
}

func (h *Handler) UpdatePurchaseOrder(c echo.Context) error {
	id := c.Param("id")

	var req UpdatePurchaseOrderRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	// Check if purchase order exists and is in DRAFT status
	var currentStatus string
	err := h.DB.QueryRow("SELECT status FROM purchase_orders WHERE id = $1", id).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Purchase order not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	if currentStatus != "DRAFT" {
		return echo.NewHTTPError(http.StatusBadRequest, "Can only update purchase orders in DRAFT status")
	}

	// Start transaction
	tx, err := h.DB.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	defer tx.Rollback()

	// Parse expected_at date if provided
	var expectedAt *time.Time
	if req.ExpectedAt != nil && *req.ExpectedAt != "" {
		if parsedTime, err := time.Parse("2006-01-02", *req.ExpectedAt); err == nil {
			expectedAt = &parsedTime
		} else if parsedTime, err := time.Parse(time.RFC3339, *req.ExpectedAt); err == nil {
			expectedAt = &parsedTime
		}
	}

	// Update purchase order
	_, err = tx.Exec(`
		UPDATE purchase_orders 
		SET supplier_id = $1, expected_at = $2, notes = $3, updated_at = NOW()
		WHERE id = $4
	`, req.SupplierID, expectedAt, req.Notes, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update purchase order")
	}

	// Delete existing lines
	_, err = tx.Exec("DELETE FROM purchase_order_lines WHERE purchase_order_id = $1", id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete existing lines")
	}

	// Create new lines
	var lines []PurchaseOrderLine
	for _, lineReq := range req.Lines {
		lineID := uuid.New().String()
		// Convert string unit cost to decimal for resolveOrCreateItem
		unitCostDecimal, err := decimal.NewFromString(lineReq.UnitCost)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid unit cost: %s", lineReq.UnitCost))
		}
		resolvedItemID, resErr := h.resolveOrCreateItem(tx, lineReq.ItemID, &unitCostDecimal, claims.TenantID)
		if resErr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, resErr.Error())
		}
		unitCostStr := unitCostDecimal.StringFixed(2)
		var taxJSON *string
		if lineReq.Tax != nil {
			if b, marshalErr := json.Marshal(lineReq.Tax); marshalErr == nil {
				s := string(b)
				taxJSON = &s
			}
		}
		_, err = tx.Exec(`
            INSERT INTO purchase_order_lines (id, purchase_order_id, item_id, qty_ordered, qty_received, unit_cost, tax, created_at, updated_at)
            VALUES ($1, $2, $3, $4, 0, $5::numeric, COALESCE($6::jsonb, '{}'::jsonb), NOW(), NOW())
        `, lineID, id, resolvedItemID, lineReq.QtyOrdered, unitCostStr, taxJSON)
		if err != nil {
			c.Logger().Errorf("failed to create purchase order line (update): %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create purchase order line")
		}

		// Calculate line total
		lineTotal := unitCostDecimal.Mul(decimal.NewFromInt(int64(lineReq.QtyOrdered)))

		lines = append(lines, PurchaseOrderLine{
			ID:          lineID,
			ItemID:      resolvedItemID,
			QtyOrdered:  lineReq.QtyOrdered,
			QtyReceived: 0,
			UnitCost:    unitCostDecimal,
			Tax:         lineReq.Tax,
			LineTotal:   lineTotal,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	// Get updated purchase order
	return h.GetPurchaseOrder(c)
}

// resolveOrCreateItem accepts a provided identifier which can be an Item UUID or a SKU.
// If it's a UUID, it must exist. If it's not a UUID, it's treated as a SKU; if missing, a minimal item is created.
func (h *Handler) resolveOrCreateItem(tx *sql.Tx, provided string, unitCost *decimal.Decimal, tenantID string) (string, error) {
	if provided == "" {
		return "", fmt.Errorf("item identifier is required")
	}
	if _, err := uuid.Parse(provided); err == nil {
		// Provided is UUID; ensure it exists and belongs to the tenant
		var id string
		err := tx.QueryRow(`SELECT id FROM items WHERE id = $1 AND tenant_id = $2 AND (deleted_at IS NULL OR deleted_at > NOW())`, provided, tenantID).Scan(&id)
		if err != nil {
			if err == sql.ErrNoRows {
				// Log more details for debugging
				fmt.Printf("DEBUG: Item lookup failed - provided: %s, tenantID: %s\n", provided, tenantID)
				return "", fmt.Errorf("invalid item id: %s (tenant: %s)", provided, tenantID)
			}
			return "", err
		}
		return id, nil
	}

	// Treat as SKU path
	sku := provided
	var existingID string
	err := tx.QueryRow(`SELECT id FROM items WHERE sku = $1 AND tenant_id = $2 AND (deleted_at IS NULL OR deleted_at > NOW())`, sku, tenantID).Scan(&existingID)
	if err == nil {
		return existingID, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}

	// Create minimal item
	newID := uuid.New().String()
	var costStr, priceStr string
	if unitCost != nil {
		costStr = unitCost.StringFixed(2)
		priceStr = unitCost.StringFixed(2)
	} else {
		costStr = "0.00"
		priceStr = "0.00"
	}
	_, err = tx.Exec(`
        INSERT INTO items (id, sku, name, uom, cost, price, tenant_id, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5::numeric, $6::numeric, $7, TRUE, NOW(), NOW())
    `, newID, sku, sku, "each", costStr, priceStr, tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to create item for sku %s", sku)
	}
	return newID, nil
}

func (h *Handler) ApprovePurchaseOrder(c echo.Context) error {
	id := c.Param("id")
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	userID := claims.UserID

	// Check if purchase order exists and is in DRAFT status
	var currentStatus string
	err := h.DB.QueryRow("SELECT status FROM purchase_orders WHERE id = $1", id).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Purchase order not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	if currentStatus != "DRAFT" {
		return echo.NewHTTPError(http.StatusBadRequest, "Can only approve purchase orders in DRAFT status")
	}

	// Update status to APPROVED
	_, err = h.DB.Exec(`
		UPDATE purchase_orders 
		SET status = 'APPROVED', approved_by = $1, approved_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`, userID, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to approve purchase order")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Purchase order approved successfully",
	})
}

type ReceiveItemsRequest struct {
	Lines []ReceiveLineRequest `json:"lines" validate:"required"`
}

type ReceiveLineRequest struct {
	LineID      string `json:"line_id" validate:"required"`
	QtyReceived int    `json:"qty_received" validate:"required,min=0"`
}

func (h *Handler) ReceivePurchaseOrder(c echo.Context) error {
	id := c.Param("id")

	var req ReceiveItemsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Check if purchase order exists and is in APPROVED status
	var currentStatus string
	err := h.DB.QueryRow("SELECT status FROM purchase_orders WHERE id = $1", id).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Purchase order not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	if currentStatus != "APPROVED" && currentStatus != "PARTIAL" {
		return echo.NewHTTPError(http.StatusBadRequest, "Can only receive items for approved purchase orders")
	}

	// Start transaction
	tx, err := h.DB.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	defer tx.Rollback()

	// Update received quantities
	for _, lineReq := range req.Lines {
		// Get current line info
		var qtyOrdered, currentQtyReceived int
		var itemID string
		err := tx.QueryRow(`
			SELECT qty_ordered, qty_received, item_id 
			FROM purchase_order_lines 
			WHERE id = $1 AND purchase_order_id = $2
		`, lineReq.LineID, id).Scan(&qtyOrdered, &currentQtyReceived, &itemID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Purchase order line not found")
		}

		newQtyReceived := currentQtyReceived + lineReq.QtyReceived
		if newQtyReceived > qtyOrdered {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Cannot receive more items than ordered for line %s", lineReq.LineID))
		}

		// Update line
		_, err = tx.Exec(`
			UPDATE purchase_order_lines 
			SET qty_received = $1, updated_at = NOW()
			WHERE id = $2
		`, newQtyReceived, lineReq.LineID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update line")
		}

		// Update inventory levels (this would normally go through inventory service)
		// For now, we'll just create a stock movement record
		if lineReq.QtyReceived > 0 {
			_, err = tx.Exec(`
				INSERT INTO stock_movements (id, item_id, movement_type, quantity, reference_type, reference_id, created_at)
				VALUES ($1, $2, 'IN', $3, 'PURCHASE_ORDER', $4, NOW())
			`, uuid.New().String(), itemID, lineReq.QtyReceived, id)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create stock movement")
			}
		}
	}

	// Check if all lines are fully received
	var totalLines, fullyReceivedLines int
	err = tx.QueryRow(`
		SELECT 
			COUNT(*) as total_lines,
			COUNT(CASE WHEN qty_ordered = qty_received THEN 1 END) as fully_received_lines
		FROM purchase_order_lines 
		WHERE purchase_order_id = $1
	`, id).Scan(&totalLines, &fullyReceivedLines)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	// Update purchase order status
	var newStatus string
	if fullyReceivedLines == totalLines {
		newStatus = "RECEIVED"
	} else if fullyReceivedLines > 0 {
		newStatus = "PARTIAL"
	} else {
		newStatus = "APPROVED"
	}

	_, err = tx.Exec(`
		UPDATE purchase_orders 
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`, newStatus, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update purchase order status")
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Items received successfully",
		"status":  newStatus,
	})
}

func (h *Handler) ClosePurchaseOrder(c echo.Context) error {
	id := c.Param("id")

	// Check if purchase order exists and can be closed
	var currentStatus string
	err := h.DB.QueryRow("SELECT status FROM purchase_orders WHERE id = $1", id).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Purchase order not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	if currentStatus == "CLOSED" || currentStatus == "CANCELED" {
		return echo.NewHTTPError(http.StatusBadRequest, "Purchase order is already closed or canceled")
	}

	// Update status to CLOSED
	_, err = h.DB.Exec(`
		UPDATE purchase_orders 
		SET status = 'CLOSED', updated_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to close purchase order")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Purchase order closed successfully",
	})
}

func (h *Handler) DeletePurchaseOrder(c echo.Context) error {
	id := c.Param("id")

	// Check if purchase order exists and is in DRAFT status
	var currentStatus string
	err := h.DB.QueryRow("SELECT status FROM purchase_orders WHERE id = $1", id).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Purchase order not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	if currentStatus != "DRAFT" {
		return echo.NewHTTPError(http.StatusBadRequest, "Can only delete purchase orders in DRAFT status")
	}

	// Start transaction
	tx, err := h.DB.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	defer tx.Rollback()

	// Delete purchase order lines first
	_, err = tx.Exec("DELETE FROM purchase_order_lines WHERE purchase_order_id = $1", id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete purchase order lines")
	}

	// Delete purchase order
	_, err = tx.Exec("DELETE FROM purchase_orders WHERE id = $1", id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete purchase order")
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	return c.NoContent(http.StatusNoContent)
}
