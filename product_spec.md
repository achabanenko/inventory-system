Inventory Management Application — Product Spec & System Prompt

1) High-Level Overview

Build a production-ready inventory management system for small/medium businesses. The system supports multi-location stock, purchasing, stock adjustments, transfers, barcode scanning, and audit trails. Backend is Go (Golang) with PostgreSQL and Ent ORM. Frontend is React with TailwindCSS.

2) Core Use Cases
	•	Items & Catalog: Create/edit items with SKU, barcode, unit of measure, pricing, cost, categories, variants, attributes.
	•	Stock Tracking: Maintain on-hand, allocated, available per location. Batch/lot and expiry optional.
	•	Purchasing: Create Purchase Orders (PO), receive items (partial/complete), auto-adjust stock, update cost (moving average).
	•	Transfers: Move stock between locations with in-transit state, approvals, and receipts.
	•	Adjustments: Cycle counts, corrections, damages, write-offs, and reasons.
	•	Suppliers: Manage supplier master data and price lists.
	•	Audit & Reporting: Audit log of all stock-affecting actions; basic reports: stock on hand, movement history, slow movers.
	•	Users & Roles: Admin, Manager, Clerk roles with RBAC.
	•	Barcode Scanning: Support scanning item/barcode fields; keyboard wedge compatible.

3) Non-Functional Requirements
	•	Security: JWT auth (access + refresh tokens), password hashing (bcrypt/argon2). OWASP best practices.
	•	Performance: API P95 < 200ms for typical calls with 100k items. Pagination for list endpoints.
	•	Reliability: ACID transactions for stock updates. Idempotent receiving & adjustments. Soft delete with deleted_at.
	•	Observability: Structured logging, request tracing (middleware), metrics (Prometheus format), health/readiness endpoints.
	•	Testing: Unit + integration tests. Seed data for dev.
	•	Internationalization: English default; design for i18n.
	•	CI/CD: Docker, migrations, linting, tests on PR.

4) Data Model (PostgreSQL via Ent)

Item
	•	id (uuid)
	•	sku (string, unique)
	•	name (string)
	•	barcode (string, unique, nullable)
	•	uom (string)
	•	category_id (fk -> Category)
	•	attributes (jsonb)
	•	cost (numeric)
	•	price (numeric)
	•	is_active (bool)
	•	created_at, updated_at, deleted_at

Category: id, name, parent_id (nullable)

Location: id, code (unique), name, address jsonb, is_active

Supplier: id, code (unique), name, contact jsonb, is_active

InventoryLevel (per-item per-location): id, item_id, location_id, on_hand (int), allocated (int), reorder_point (int), reorder_qty (int)

StockMovement: id, item_id, location_id, qty (int, +receive/-issue), reason (enum: PO_RECEIPT, ADJUSTMENT, TRANSFER_OUT, TRANSFER_IN, COUNT), reference (string), ref_id (uuid), user_id, occurred_at, meta jsonb

PurchaseOrder: id, number (unique), supplier_id, status (enum: DRAFT, APPROVED, PARTIAL, RECEIVED, CLOSED, CANCELED), expected_at, created_by, approved_by, notes

PurchaseOrderLine: id, po_id, item_id, qty_ordered, qty_received, unit_cost, tax jsonb

Transfer: id, number, from_location_id, to_location_id, status (DRAFT, IN_TRANSIT, RECEIVED, CANCELED), created_by, approved_by

TransferLine: id, transfer_id, item_id, qty

Adjustment: id, number, location_id, reason (enum: COUNT, DAMAGE, CORRECTION), notes, created_by, approved_by

AdjustmentLine: id, adjustment_id, item_id, qty_diff

User: id, email (unique), password_hash, name, role (ADMIN, MANAGER, CLERK), is_active, last_login

AuditLog: id, user_id, action (string), entity (string), entity_id (uuid), before jsonb, after jsonb, at

Notes
	•	Maintain stock integrity: available = on_hand - allocated.
	•	Use DB constraints and transactions for stock-changing operations.

5) API Design (REST)

Base URL: /api/v1 | JSON only | Use cursor or page-based pagination (page, page_size, sort), filter via query params.

Auth
	•	POST /auth/login → {access_token, refresh_token}
	•	POST /auth/refresh → new access_token
	•	POST /auth/logout

Items
	•	GET /items (filter: q, category_id, barcode, sku, is_active)
	•	POST /items
	•	GET /items/{id}
	•	PUT /items/{id}
	•	DELETE /items/{id} (soft delete)

Locations
	•	GET /locations | POST /locations | GET/PUT/DELETE /locations/{id}

Suppliers
	•	GET /suppliers | POST /suppliers | GET/PUT/DELETE /suppliers/{id}

Inventory
	•	GET /inventory (item_id?, location_id?) → returns levels with available, on_hand, allocated
	•	GET /inventory/{item_id}/locations → per-location breakdown
	•	GET /inventory/movements (filters: item_id, location_id, reason, from, to)

Purchase Orders
	•	GET /pos | POST /pos (create DRAFT)
	•	GET /pos/{id} | PUT /pos/{id} (update header/lines in DRAFT)
	•	POST /pos/{id}/approve
	•	POST /pos/{id}/receive body: [{line_id, qty_received, occurred_at}] (idempotent)
	•	POST /pos/{id}/close

Transfers
	•	GET /transfers | POST /transfers
	•	POST /transfers/{id}/approve
	•	POST /transfers/{id}/ship → status IN_TRANSIT, movements (from_location, -qty)
	•	POST /transfers/{id}/receive → status RECEIVED, movements (to_location, +qty)

Adjustments
	•	GET /adjustments | POST /adjustments
	•	POST /adjustments/{id}/approve → apply qty_diff movements

Users & Audit
	•	GET /users | POST /users | GET/PUT /users/{id} | POST /users/{id}/disable
	•	GET /audit (filters: entity, entity_id, user_id, from, to)

Conventions
	•	All mutating endpoints return the updated entity with version and updated_at for optimistic concurrency.
	•	Errors: {error: {code, message, details?}}. Use codes like VALIDATION_ERROR, NOT_FOUND, CONFLICT, FORBIDDEN.

6) Stock Update Rules
	•	PO Receive: For each line: increase on_hand at target location by qty_received; create StockMovement with reason PO_RECEIPT and ref to PO.
	•	Transfer Ship: Decrease on_hand at source; create movement TRANSFER_OUT.
	•	Transfer Receive: Increase on_hand at destination; create movement TRANSFER_IN.
	•	Adjustment Approve: Apply qty_diff (+/-) to on_hand; create movement ADJUSTMENT.
	•	Wrap each operation in a DB transaction. Use idempotency key per request to avoid double-apply.

7) Security & RBAC
	•	Admin: full access.
	•	Manager: manage items, suppliers, POs, transfers, adjustments, view users.
	•	Clerk: read items/inventory, create draft POs, create transfers/adjustments (needs approval), receive POs.
	•	Use middleware to enforce role scopes per route.

8) Backend Architecture (Go)
	•	Modules: /cmd/api, /internal/{auth,db,items,inventory,po,transfer,adjustment,user,audit,http}
	•	Framework: net/http or fiber/echo; middleware: logging, recovery, auth (JWT), request ID, CORS, gzip.
	•	Ent ORM for schema/migrations.
	•	Config: env vars (DB_URL, JWT_SECRET, PORT, CORS_ORIGINS).
	•	Validation: use struct tags; return field-level errors.
	•	Services: business logic separate from handlers; keep handlers thin.

9) Frontend (React + Tailwind)
	•	Routing: /login, /dashboard, /items, /items/:id, /inventory, /pos, /pos/:id, /transfers, /adjustments, /suppliers, /settings/users.
	•	State: React Query for server state; Zod for validation; forms with React Hook Form.
	•	Components: DataTable with column filters, pagination, bulk actions; Modal forms; Barcode input.
	•	UX: Optimistic updates where safe; skeleton loaders; toasts; confirm dialogs for destructive actions.
	•	Auth: Token storage in httpOnly cookies; refresh flow; protected routes.

10) Reporting
	•	Stock On Hand by location (CSV export, printable).
	•	Movements over time with filters.
	•	Slow Movers based on last N days movements.

11) DevOps
	•	Docker: multi-stage builds (Go, Node), Nginx or Vite preview for static.
	•	Migrations: ent/migrate; auto-run on start (configurable).
	•	Makefile: make dev, make test, make lint, make migrate.
