package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	appmw "inventory/internal/middleware"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
)

type GoodsReceipt struct {
	ID         string             `json:"id"`
	Number     string             `json:"number"`
	SupplierID *string            `json:"supplier_id,omitempty"`
	Supplier   *Supplier          `json:"supplier,omitempty"`
	LocationID *string            `json:"location_id,omitempty"`
	Location   *Location          `json:"location,omitempty"`
	Status     string             `json:"status"`
	Reference  *string            `json:"reference,omitempty"`
	Notes      *string            `json:"notes,omitempty"`
	CreatedBy  *string            `json:"created_by,omitempty"`
	ApprovedBy *string            `json:"approved_by,omitempty"`
	PostedBy   *string            `json:"posted_by,omitempty"`
	ApprovedAt *time.Time         `json:"approved_at,omitempty"`
	PostedAt   *time.Time         `json:"posted_at,omitempty"`
	Lines      []GoodsReceiptLine `json:"lines,omitempty"`
	Total      decimal.Decimal    `json:"total"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

type Location struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}



type GoodsReceiptLine struct {
	ID        string          `json:"id"`
	ReceiptID string          `json:"receipt_id"`
	ItemID    string          `json:"item_id"`
	Item      *Item           `json:"item,omitempty"`
	Qty       int             `json:"qty"`
	UnitCost  decimal.Decimal `json:"unit_cost"`
	LineTotal decimal.Decimal `json:"line_total"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (h *Handler) ListReceipts(c echo.Context) error {
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
	supplierID := c.QueryParam("supplier_id")
	locationID := c.QueryParam("location_id")
	sort := c.QueryParam("sort")
	if sort == "" {
		sort = "created_at DESC"
	}

	offset := (page - 1) * pageSize

	// Build query
	query := `
		SELECT 
			gr.id, gr.number, gr.status, gr.supplier_id, gr.location_id, gr.created_by, 
			gr.approved_by, gr.posted_by, gr.approved_at, gr.posted_at, gr.reference, gr.notes,
			gr.created_at, gr.updated_at,
			s.name as supplier_name,
			l.name as location_name, l.code as location_code,
			COALESCE(SUM(grl.qty * grl.unit_cost), 0) as total
		FROM goods_receipts gr
		LEFT JOIN suppliers s ON gr.supplier_id = s.id
		LEFT JOIN locations l ON gr.location_id = l.id
		LEFT JOIN goods_receipt_lines grl ON gr.id = grl.receipt_id
		WHERE gr.tenant_id = $1`

	args := []interface{}{tenantID}
	argCount := 1

	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND (gr.number ILIKE $%d)", argCount)
		args = append(args, "%"+search+"%")
	}

	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND gr.status = $%d", argCount)
		args = append(args, status)
	}

	if supplierID != "" {
		argCount++
		query += fmt.Sprintf(" AND gr.supplier_id = $%d", argCount)
		args = append(args, supplierID)
	}

	if locationID != "" {
		argCount++
		query += fmt.Sprintf(" AND gr.location_id = $%d", argCount)
		args = append(args, locationID)
	}

	query += " GROUP BY gr.id, s.name, l.name, l.code"

	// Add sorting
	switch sort {
	case "number", "number ASC":
		query += " ORDER BY gr.number ASC"
	case "number DESC":
		query += " ORDER BY gr.number DESC"
	case "status", "status ASC":
		query += " ORDER BY gr.status ASC"
	case "status DESC":
		query += " ORDER BY gr.status DESC"
	case "created_at", "created_at ASC":
		query += " ORDER BY gr.created_at ASC"
	default:
		query += " ORDER BY gr.created_at DESC"
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

	var receipts []GoodsReceipt
	for rows.Next() {
		var gr GoodsReceipt
		var supplierName, locationName, locationCode sql.NullString
		var approvedBy, postedBy sql.NullString
		var approvedAt, postedAt sql.NullTime
		var reference, notes sql.NullString
		var total string

		err := rows.Scan(
			&gr.ID, &gr.Number, &gr.Status, &gr.SupplierID, &gr.LocationID, &gr.CreatedBy,
			&approvedBy, &postedBy, &approvedAt, &postedAt, &reference, &notes,
			&gr.CreatedAt, &gr.UpdatedAt, &supplierName, &locationName, &locationCode, &total,
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Database scan error")
		}

		if approvedBy.Valid {
			gr.ApprovedBy = &approvedBy.String
		}
		if postedBy.Valid {
			gr.PostedBy = &postedBy.String
		}
		if approvedAt.Valid {
			gr.ApprovedAt = &approvedAt.Time
		}
		if postedAt.Valid {
			gr.PostedAt = &postedAt.Time
		}
		if reference.Valid {
			gr.Reference = &reference.String
		}
		if notes.Valid {
			gr.Notes = &notes.String
		}

		// Parse total
		gr.Total, _ = decimal.NewFromString(total)

		// Add supplier info if available
		if supplierName.Valid && gr.SupplierID != nil {
			gr.Supplier = &Supplier{
				ID:   *gr.SupplierID,
				Name: supplierName.String,
			}
		}

		// Add location info if available
		if locationName.Valid && locationCode.Valid && gr.LocationID != nil {
			gr.Location = &Location{
				ID:   *gr.LocationID,
				Name: locationName.String,
				Code: locationCode.String,
			}
		}

		receipts = append(receipts, gr)
	}

	// Get total count
	countQuery := `
		SELECT COUNT(DISTINCT gr.id)
		FROM goods_receipts gr
		LEFT JOIN suppliers s ON gr.supplier_id = s.id
		LEFT JOIN locations l ON gr.location_id = l.id
		WHERE gr.tenant_id = $1`

	countArgs := []interface{}{tenantID}
	countArgCount := 1

	if search != "" {
		countArgCount++
		countQuery += fmt.Sprintf(" AND (gr.number ILIKE $%d)", countArgCount)
		countArgs = append(countArgs, "%"+search+"%")
	}

	if status != "" {
		countArgCount++
		countQuery += fmt.Sprintf(" AND gr.status = $%d", countArgCount)
		countArgs = append(countArgs, status)
	}

	if supplierID != "" {
		countArgCount++
		countQuery += fmt.Sprintf(" AND gr.supplier_id = $%d", countArgCount)
		countArgs = append(countArgs, supplierID)
	}

	if locationID != "" {
		countArgCount++
		countQuery += fmt.Sprintf(" AND gr.location_id = $%d", countArgCount)
		countArgs = append(countArgs, locationID)
	}

	var total int
	err = h.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	totalPages := (total + pageSize - 1) / pageSize

	return c.JSON(http.StatusOK, PaginatedResponse{
		Data:       receipts,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Total:      int64(total),
	})
}

type CreateGoodsReceiptRequest struct {
	SupplierID      *string `json:"supplier_id"`
	LocationID      *string `json:"location_id"`
	Reference       *string `json:"reference"`
	Notes           *string `json:"notes"`
	PurchaseOrderID *string `json:"purchase_order_id"`
	Lines           []struct {
		ItemID   string `json:"item_id"`
		Qty      int    `json:"qty"`
		UnitCost string `json:"unit_cost"`
	} `json:"lines"`
}

func (h *Handler) CreateReceipt(c echo.Context) error {
	var req CreateGoodsReceiptRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Get user ID from context (set by auth middleware)
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	userID := claims.UserID
	tenantID := claims.TenantID

	// Generate receipt number
	var maxNumber int
	err := h.DB.QueryRow(`
		SELECT COALESCE(MAX(CAST(SUBSTRING(number FROM 'GR-([0-9]+)') AS INTEGER)), 0)
		FROM goods_receipts 
		WHERE number ~ '^GR-[0-9]+$'
	`).Scan(&maxNumber)
	if err != nil && err != sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	grNumber := fmt.Sprintf("GR-%06d", maxNumber+1)

	// Start transaction for receipt and lines creation
	tx, err := h.DB.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	defer tx.Rollback()

	// Create receipt
	grID := uuid.New().String()
	_, err = tx.Exec(`
		INSERT INTO goods_receipts (id, number, status, supplier_id, location_id, reference, notes, tenant_id, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	`, grID, grNumber, "DRAFT", req.SupplierID, req.LocationID, req.Reference, req.Notes, claims.TenantID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create receipt")
	}

	// Create receipt lines if provided
	var total decimal.Decimal
	if req.Lines != nil && len(req.Lines) > 0 {
		for _, line := range req.Lines {
			// Resolve or create item
			var unitCostDecimal *decimal.Decimal
			if line.UnitCost != "" {
				cost, err := decimal.NewFromString(line.UnitCost)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, "Invalid unit cost format")
				}
				unitCostDecimal = &cost
			}

			resolvedItemID, resErr := h.resolveOrCreateItem(tx, line.ItemID, unitCostDecimal, tenantID)
			if resErr != nil {
				return echo.NewHTTPError(http.StatusBadRequest, resErr.Error())
			}

			// Create receipt line
			lineID := uuid.New().String()
			var unitCostValue interface{}
			if unitCostDecimal != nil {
				unitCostValue = unitCostDecimal.StringFixed(2)
			} else {
				unitCostValue = nil
			}
			_, err = tx.Exec(`
				INSERT INTO goods_receipt_lines (id, receipt_id, item_id, qty, unit_cost, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			`, lineID, grID, resolvedItemID, line.Qty, unitCostValue)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create receipt line")
			}

			// Add to total (only if unit cost is provided)
			if unitCostDecimal != nil {
				lineTotal := unitCostDecimal.Mul(decimal.NewFromInt(int64(line.Qty)))
				total = total.Add(lineTotal)
			}
		}

		// Update receipt total
		_, err = tx.Exec(`UPDATE goods_receipts SET total = $1 WHERE id = $2`, total, grID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update receipt total")
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	// Return created receipt
	gr := GoodsReceipt{
		ID:         grID,
		Number:     grNumber,
		Status:     "DRAFT",
		SupplierID: req.SupplierID,
		LocationID: req.LocationID,
		Reference:  req.Reference,
		Notes:      req.Notes,
		CreatedBy:  &userID,
		Total:      total,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return c.JSON(http.StatusCreated, gr)
}

func (h *Handler) UpdateReceipt(c echo.Context) error {
	id := c.Param("id")

	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

	var req struct {
		SupplierID *string `json:"supplier_id"`
		LocationID *string `json:"location_id"`
		Status     *string `json:"status"`
		Reference  *string `json:"reference"`
		Notes      *string `json:"notes"`
		Lines      []struct {
			ItemID   string `json:"item_id"`
			Qty      int    `json:"qty"`
			UnitCost string `json:"unit_cost"`
		} `json:"lines"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	sets := []string{}
	args := []interface{}{}
	i := 1
	if req.SupplierID != nil {
		sets = append(sets, fmt.Sprintf("supplier_id = $%d", i))
		args = append(args, *req.SupplierID)
		i++
	}
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
	if req.Reference != nil {
		sets = append(sets, fmt.Sprintf("reference = $%d", i))
		args = append(args, *req.Reference)
		i++
	}
	if req.Notes != nil {
		sets = append(sets, fmt.Sprintf("notes = $%d", i))
		args = append(args, *req.Notes)
		i++
	}
	// If no header fields to update but lines are provided, we still need to update the receipt
	// to refresh the updated_at timestamp and potentially update the total
	if len(sets) == 0 && req.Lines == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "no fields to update")
	}

	// Declare variables outside conditional blocks
	var out GoodsReceipt
	var supplierID, locationID, reference, notes sql.NullString

	// Always update the updated_at timestamp if we're making any changes
	if len(sets) > 0 {
		sets = append(sets, "updated_at = NOW()")
		args = append(args, id)

		query := fmt.Sprintf(`UPDATE goods_receipts SET %s WHERE id = $%d AND tenant_id = $%d RETURNING id, number, supplier_id, location_id, status, reference, notes, created_at, updated_at`, strings.Join(sets, ", "), i, i+1)
		if err := h.DB.QueryRow(query, append(args, tenantID)...).Scan(&out.ID, &out.Number, &supplierID, &locationID, &out.Status, &reference, &notes, &out.CreatedAt, &out.UpdatedAt); err != nil {
			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusNotFound, "receipt not found")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "database error")
		}
	} else {
		// If only lines are being updated, we need to get the current receipt data
		// and update the updated_at timestamp
		_, err := h.DB.Exec(`UPDATE goods_receipts SET updated_at = NOW() WHERE id = $1 AND tenant_id = $2`, id, tenantID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update receipt timestamp")
		}

		// Get current receipt data for response
		if err := h.DB.QueryRow(`SELECT id, number, supplier_id, location_id, status, reference, notes, created_at, updated_at FROM goods_receipts WHERE id = $1 AND tenant_id = $2`, id, tenantID).Scan(&out.ID, &out.Number, &supplierID, &locationID, &out.Status, &reference, &notes, &out.CreatedAt, &out.UpdatedAt); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get receipt data")
		}
	}

	// Set optional fields if they exist
	if supplierID.Valid {
		out.SupplierID = &supplierID.String
	}
	if locationID.Valid {
		out.LocationID = &locationID.String
	}
	if reference.Valid {
		out.Reference = &reference.String
	}
	if notes.Valid {
		out.Notes = &notes.String
	}

	// Handle lines update if provided
	if req.Lines != nil {
		// Start transaction for lines update
		tx, err := h.DB.Begin()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
		}
		defer tx.Rollback()

		// Delete existing lines
		_, err = tx.Exec(`DELETE FROM goods_receipt_lines WHERE receipt_id = $1`, id)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete existing lines")
		}

		// Create new lines
		var total decimal.Decimal
		for _, line := range req.Lines {
			// Resolve or create item
			var unitCostDecimal *decimal.Decimal
			if line.UnitCost != "" {
				cost, err := decimal.NewFromString(line.UnitCost)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, "Invalid unit cost format")
				}
				unitCostDecimal = &cost
			}

			resolvedItemID, resErr := h.resolveOrCreateItem(tx, line.ItemID, unitCostDecimal, tenantID)
			if resErr != nil {
				return echo.NewHTTPError(http.StatusBadRequest, resErr.Error())
			}

			// Create receipt line
			lineID := uuid.New().String()
			var unitCostValue interface{}
			if unitCostDecimal != nil {
				unitCostValue = unitCostDecimal.StringFixed(2)
			} else {
				unitCostValue = nil
			}
			_, err = tx.Exec(`
				INSERT INTO goods_receipt_lines (id, receipt_id, item_id, qty, unit_cost, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			`, lineID, id, resolvedItemID, line.Qty, unitCostValue)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create receipt line")
			}

			// Add to total (only if unit cost is provided)
			if unitCostDecimal != nil {
				lineTotal := unitCostDecimal.Mul(decimal.NewFromInt(int64(line.Qty)))
				total = total.Add(lineTotal)
			}
		}

		// Update receipt total
		_, err = tx.Exec(`UPDATE goods_receipts SET total = $1 WHERE id = $2`, total, id)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update receipt total")
		}

		// Commit transaction
		if err = tx.Commit(); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
		}

		// Update the returned receipt with new total
		out.Total = total
	}

	return c.JSON(http.StatusOK, out)
}

func (h *Handler) DeleteReceipt(c echo.Context) error {
	id := c.Param("id")

	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

	res, err := h.DB.Exec(`DELETE FROM goods_receipts WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusConflict, "cannot delete receipt")
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "receipt not found")
	}
	return c.NoContent(http.StatusNoContent)
}

// Create receipt from Purchase Order remaining quantities
func (h *Handler) CreateReceiptFromPO(c echo.Context) error {
	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

	var req struct {
		PurchaseOrderNumber string  `json:"purchase_order_number"`
		PurchaseOrderID     string  `json:"purchase_order_id"`
		LocationID          string  `json:"location_id"`
		Reference           *string `json:"reference"`
		Notes               *string `json:"notes"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if (strings.TrimSpace(req.PurchaseOrderNumber) == "" && strings.TrimSpace(req.PurchaseOrderID) == "") || strings.TrimSpace(req.LocationID) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "purchase_order_number or purchase_order_id and location_id are required")
	}

	// Resolve PO id by number if provided
	poID := strings.TrimSpace(req.PurchaseOrderID)
	if poID == "" {
		if err := h.DB.QueryRow(`SELECT id FROM purchase_orders WHERE number = $1 AND tenant_id = $2`, strings.TrimSpace(req.PurchaseOrderNumber), tenantID).Scan(&poID); err != nil {
			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusNotFound, "purchase order not found")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "database error")
		}
	}

	// Load PO header
	var supplierID sql.NullString
	if err := h.DB.QueryRow(`SELECT supplier_id FROM purchase_orders WHERE id = $1 AND tenant_id = $2`, poID, tenantID).Scan(&supplierID); err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "purchase order not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	// Load remaining lines
	rows, err := h.DB.Query(`
        SELECT item_id, GREATEST(qty_ordered - qty_received, 0) AS remaining, unit_cost
        FROM purchase_order_lines
        WHERE purchase_order_id = $1`, poID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	defer rows.Close()
	type pol struct {
		itemID    string
		remaining int
		unitCost  string
	}
	var pols []pol
	for rows.Next() {
		var r pol
		if err := rows.Scan(&r.itemID, &r.remaining, &r.unitCost); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "database scan error")
		}
		if r.remaining > 0 {
			pols = append(pols, r)
		}
	}
	if len(pols) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "no remaining quantities on purchase order")
	}

	// Create receipt header
	var maxNumber int
	_ = h.DB.QueryRow(`SELECT COALESCE(MAX(CAST(SUBSTRING(number FROM 'GR-([0-9]+)') AS INTEGER)), 0) FROM goods_receipts WHERE number ~ '^GR-[0-9]+$' AND tenant_id = $1`, tenantID).Scan(&maxNumber)
	number := fmt.Sprintf("GR-%06d", maxNumber+1)
	id := uuid.New().String()
	var out GoodsReceipt
	var supplierOut, locationOut, reference, notes sql.NullString
	if err := h.DB.QueryRow(`
        INSERT INTO goods_receipts (id, number, supplier_id, location_id, status, reference, notes, tenant_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, 'DRAFT', $5, $6, $7, NOW(), NOW())
        RETURNING id, number, supplier_id, location_id, status, reference, notes, created_at, updated_at
    `, id, number, supplierID, req.LocationID, req.Reference, req.Notes, tenantID).Scan(&out.ID, &out.Number, &supplierOut, &locationOut, &out.Status, &reference, &notes, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if supplierOut.Valid {
		out.SupplierID = &supplierOut.String
	}
	if locationOut.Valid {
		out.LocationID = &locationOut.String
	}
	if reference.Valid {
		out.Reference = &reference.String
	}
	if notes.Valid {
		out.Notes = &notes.String
	}

	// Insert lines
	for _, r := range pols {
		if _, err := h.DB.Exec(`
            INSERT INTO goods_receipt_lines (id, receipt_id, item_id, qty, unit_cost, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5::numeric, NOW(), NOW())
        `, uuid.New().String(), id, r.itemID, r.remaining, r.unitCost); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "database error")
		}
	}

	return c.JSON(http.StatusCreated, out)
}

func (h *Handler) ListReceiptLines(c echo.Context) error {
	receiptID := c.Param("id")

	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

	// Verify receipt belongs to tenant
	var receiptExists bool
	err := h.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM goods_receipts WHERE id = $1 AND tenant_id = $2)`, receiptID, tenantID).Scan(&receiptExists)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	if !receiptExists {
		return echo.NewHTTPError(http.StatusNotFound, "Receipt not found")
	}

	rows, err := h.DB.Query(`
		SELECT 
			grl.id, grl.receipt_id, grl.item_id, grl.qty, grl.unit_cost, 
			grl.created_at, grl.updated_at,
			i.sku, i.name
		FROM goods_receipt_lines grl
		LEFT JOIN items i ON grl.item_id = i.id
		WHERE grl.receipt_id = $1 
		ORDER BY grl.created_at ASC
	`, receiptID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	defer rows.Close()
	res := []GoodsReceiptLine{}
	for rows.Next() {
		var m GoodsReceiptLine
		var sku, name sql.NullString
		if err := rows.Scan(&m.ID, &m.ReceiptID, &m.ItemID, &m.Qty, &m.UnitCost, &m.CreatedAt, &m.UpdatedAt, &sku, &name); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "database scan error")
		}
		// Add item info if available
		if sku.Valid || name.Valid {
			m.Item = &Item{
				ID:   m.ItemID,
				SKU:  sku.String,
				Name: name.String,
			}
		}
		res = append(res, m)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": res})
}

func (h *Handler) AddReceiptLine(c echo.Context) error {
	receiptID := c.Param("id")

	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

	var req struct {
		ItemID   string `json:"item_id"`
		Qty      int    `json:"qty"`
		UnitCost string `json:"unit_cost"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.ItemID == "" || req.Qty <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "item_id and qty are required")
	}

	// Start transaction
	tx, err := h.DB.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	defer tx.Rollback()

	// Verify receipt belongs to tenant
	var receiptExists bool
	err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM goods_receipts WHERE id = $1 AND tenant_id = $2)`, receiptID, tenantID).Scan(&receiptExists)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	if !receiptExists {
		return echo.NewHTTPError(http.StatusNotFound, "Receipt not found")
	}

	// Resolve or create item (similar to purchase orders)
	var unitCostDecimal *decimal.Decimal
	if req.UnitCost != "" {
		cost, err := decimal.NewFromString(req.UnitCost)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid unit cost format")
		}
		unitCostDecimal = &cost
	}
	resolvedItemID, resErr := h.resolveOrCreateItem(tx, req.ItemID, unitCostDecimal, tenantID)
	if resErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, resErr.Error())
	}

	id := uuid.New().String()
	var unitCostValue interface{}
	if req.UnitCost != "" {
		unitCostValue = req.UnitCost
	} else {
		unitCostValue = nil
	}
	var out GoodsReceiptLine
	if err := tx.QueryRow(`
        INSERT INTO goods_receipt_lines (id, receipt_id, item_id, qty, unit_cost, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
        RETURNING id, receipt_id, item_id, qty, unit_cost, created_at, updated_at
    `, id, receiptID, resolvedItemID, req.Qty, unitCostValue).Scan(&out.ID, &out.ReceiptID, &out.ItemID, &out.Qty, &out.UnitCost, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	return c.JSON(http.StatusCreated, out)
}

func (h *Handler) UpdateReceiptLine(c echo.Context) error {
	receiptID := c.Param("id")
	lineID := c.Param("line_id")

	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

	var req struct {
		Qty      *int    `json:"qty"`
		UnitCost *string `json:"unit_cost"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Verify receipt belongs to tenant
	var receiptExists bool
	err := h.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM goods_receipts WHERE id = $1 AND tenant_id = $2)`, receiptID, tenantID).Scan(&receiptExists)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	if !receiptExists {
		return echo.NewHTTPError(http.StatusNotFound, "Receipt not found")
	}

	sets := []string{}
	args := []interface{}{}
	i := 1
	if req.Qty != nil {
		sets = append(sets, fmt.Sprintf("qty = $%d", i))
		args = append(args, *req.Qty)
		i++
	}
	if req.UnitCost != nil {
		sets = append(sets, fmt.Sprintf("unit_cost = $%d", i))
		args = append(args, *req.UnitCost)
		i++
	}
	if len(sets) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "no fields to update")
	}
	sets = append(sets, "updated_at = NOW()")
	args = append(args, lineID, receiptID)
	query := fmt.Sprintf(`UPDATE goods_receipt_lines SET %s WHERE id = $%d AND receipt_id = $%d RETURNING id, receipt_id, item_id, qty, unit_cost, created_at, updated_at`, strings.Join(sets, ", "), i, i+1)
	var out GoodsReceiptLine
	if err := h.DB.QueryRow(query, args...).Scan(&out.ID, &out.ReceiptID, &out.ItemID, &out.Qty, &out.UnitCost, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "line not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) DeleteReceiptLine(c echo.Context) error {
	receiptID := c.Param("id")
	lineID := c.Param("line_id")

	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

	// Verify receipt belongs to tenant
	var receiptExists bool
	err := h.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM goods_receipts WHERE id = $1 AND tenant_id = $2)`, receiptID, tenantID).Scan(&receiptExists)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	if !receiptExists {
		return echo.NewHTTPError(http.StatusNotFound, "Receipt not found")
	}

	res, err := h.DB.Exec(`DELETE FROM goods_receipt_lines WHERE id = $1 AND receipt_id = $2`, lineID, receiptID)
	if err != nil {
		return echo.NewHTTPError(http.StatusConflict, "cannot delete line")
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "line not found")
	}
	return c.NoContent(http.StatusNoContent)
}

// GetReceipt retrieves a single receipt with all details including lines
func (h *Handler) GetReceipt(c echo.Context) error {
	id := c.Param("id")

	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	tenantID := claims.TenantID

	// Get receipt header
	var gr GoodsReceipt
	var supplierName, locationName, locationCode sql.NullString
	var approvedBy, postedBy sql.NullString
	var approvedAt, postedAt sql.NullTime
	var reference, notes sql.NullString

	err := h.DB.QueryRow(`
		SELECT 
			gr.id, gr.number, gr.status, gr.supplier_id, gr.location_id, gr.created_by,
			gr.approved_by, gr.posted_by, gr.approved_at, gr.posted_at, gr.reference, gr.notes,
			gr.created_at, gr.updated_at,
			s.name as supplier_name,
			l.name as location_name, l.code as location_code
		FROM goods_receipts gr
		LEFT JOIN suppliers s ON gr.supplier_id = s.id
		LEFT JOIN locations l ON gr.location_id = l.id
		WHERE gr.id = $1 AND gr.tenant_id = $2
	`, id, tenantID).Scan(
		&gr.ID, &gr.Number, &gr.Status, &gr.SupplierID, &gr.LocationID, &gr.CreatedBy,
		&approvedBy, &postedBy, &approvedAt, &postedAt, &reference, &notes,
		&gr.CreatedAt, &gr.UpdatedAt, &supplierName, &locationName, &locationCode,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Receipt not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	if approvedBy.Valid {
		gr.ApprovedBy = &approvedBy.String
	}
	if postedBy.Valid {
		gr.PostedBy = &postedBy.String
	}
	if approvedAt.Valid {
		gr.ApprovedAt = &approvedAt.Time
	}
	if postedAt.Valid {
		gr.PostedAt = &postedAt.Time
	}
	if reference.Valid {
		gr.Reference = &reference.String
	}
	if notes.Valid {
		gr.Notes = &notes.String
	}

	// Add supplier info if available
	if supplierName.Valid && gr.SupplierID != nil {
		gr.Supplier = &Supplier{
			ID:   *gr.SupplierID,
			Name: supplierName.String,
		}
	}

	// Add location info if available
	if locationName.Valid && locationCode.Valid && gr.LocationID != nil {
		gr.Location = &Location{
			ID:   *gr.LocationID,
			Name: locationName.String,
			Code: locationCode.String,
		}
	}

	// Get receipt lines
	rows, err := h.DB.Query(`
		SELECT 
			grl.id, grl.item_id, grl.qty, grl.unit_cost, 
			grl.created_at, grl.updated_at,
			i.sku, i.name as item_name
		FROM goods_receipt_lines grl
		LEFT JOIN items i ON grl.item_id = i.id
		WHERE grl.receipt_id = $1
		ORDER BY grl.created_at
	`, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	var lines []GoodsReceiptLine
	var total decimal.Decimal

	for rows.Next() {
		var line GoodsReceiptLine
		var unitCostStr string
		var itemSKU, itemName sql.NullString

		err := rows.Scan(
			&line.ID, &line.ItemID, &line.Qty, &unitCostStr,
			&line.CreatedAt, &line.UpdatedAt,
			&itemSKU, &itemName,
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Database scan error")
		}

		// Parse unit cost
		line.UnitCost, _ = decimal.NewFromString(unitCostStr)
		line.LineTotal = line.UnitCost.Mul(decimal.NewFromInt(int64(line.Qty)))
		total = total.Add(line.LineTotal)
		line.ReceiptID = id

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

	gr.Lines = lines
	gr.Total = total

	return c.JSON(http.StatusOK, gr)
}

// ApproveReceipt approves a receipt (changes status from DRAFT to APPROVED)
func (h *Handler) ApproveReceipt(c echo.Context) error {
	id := c.Param("id")
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	userID := claims.UserID

	// Check if receipt exists and is in DRAFT status
	var currentStatus string
	err := h.DB.QueryRow("SELECT status FROM goods_receipts WHERE id = $1 AND tenant_id = $2", id, claims.TenantID).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Receipt not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	if currentStatus != "DRAFT" {
		return echo.NewHTTPError(http.StatusBadRequest, "Can only approve receipts in DRAFT status")
	}

	// Update status to APPROVED
	_, err = h.DB.Exec(`
		UPDATE goods_receipts 
		SET status = 'APPROVED', approved_by = $1, approved_at = NOW(), updated_at = NOW()
		WHERE id = $2 AND tenant_id = $3
	`, userID, id, claims.TenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to approve receipt")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Receipt approved successfully",
	})
}

// PostReceipt posts a receipt to inventory (changes status from APPROVED to POSTED)
func (h *Handler) PostReceipt(c echo.Context) error {
	id := c.Param("id")
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	userID := claims.UserID

	// Check if receipt exists and is in APPROVED status
	var currentStatus string
	var locationID sql.NullString
	err := h.DB.QueryRow("SELECT status, location_id FROM goods_receipts WHERE id = $1 AND tenant_id = $2", id, claims.TenantID).Scan(&currentStatus, &locationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Receipt not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	if currentStatus != "APPROVED" {
		return echo.NewHTTPError(http.StatusBadRequest, "Can only post approved receipts")
	}

	if !locationID.Valid {
		return echo.NewHTTPError(http.StatusBadRequest, "Receipt must have a location to post")
	}

	// Start transaction
	tx, err := h.DB.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	defer tx.Rollback()

	// Get receipt lines
	rows, err := tx.Query(`
		SELECT item_id, qty, unit_cost
		FROM goods_receipt_lines
		WHERE receipt_id = $1
	`, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	// Create stock movements and update inventory levels
	for rows.Next() {
		var itemID string
		var qty int
		var unitCostStr string

		err := rows.Scan(&itemID, &qty, &unitCostStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Database scan error")
		}

		if qty > 0 {
			// Create stock movement record
			_, err = tx.Exec(`
				INSERT INTO stock_movements (id, item_id, location_id, movement_type, quantity, unit_cost, reference_type, reference_id, occurred_at, created_at)
				VALUES ($1, $2, $3, 'IN', $4, $5::numeric, 'GOODS_RECEIPT', $6, NOW(), NOW())
			`, uuid.New().String(), itemID, locationID.String, qty, unitCostStr, id)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create stock movement")
			}

			// Update inventory levels
			_, err = tx.Exec(`
				INSERT INTO inventory_levels (id, item_id, location_id, on_hand, allocated, available, created_at, updated_at)
				VALUES ($1, $2, $3, $4, 0, $4, NOW(), NOW())
				ON CONFLICT (item_id, location_id) 
				DO UPDATE SET 
					on_hand = inventory_levels.on_hand + $4,
					available = inventory_levels.available + $4,
					updated_at = NOW()
			`, uuid.New().String(), itemID, locationID.String, qty)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update inventory levels")
			}
		}
	}

	// Update receipt status to POSTED
	_, err = tx.Exec(`
		UPDATE goods_receipts 
		SET status = 'POSTED', posted_by = $1, posted_at = NOW(), updated_at = NOW()
		WHERE id = $2 AND tenant_id = $3
	`, userID, id, claims.TenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to post receipt")
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Receipt posted successfully",
	})
}

// CloseReceipt closes a receipt (changes status from POSTED to CLOSED)
func (h *Handler) CloseReceipt(c echo.Context) error {
	id := c.Param("id")

	// Get user claims for tenant ID
	claims, errClaims := appmw.GetUserClaims(c)
	if errClaims != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	// Check if receipt exists and can be closed
	var currentStatus string
	err := h.DB.QueryRow("SELECT status FROM goods_receipts WHERE id = $1 AND tenant_id = $2", id, claims.TenantID).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Receipt not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	if currentStatus == "CLOSED" || currentStatus == "CANCELED" {
		return echo.NewHTTPError(http.StatusBadRequest, "Receipt is already closed or canceled")
	}

	// Update status to CLOSED
	_, err = h.DB.Exec(`
		UPDATE goods_receipts 
		SET status = 'CLOSED', updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`, id, claims.TenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to close receipt")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Receipt closed successfully",
	})
}
