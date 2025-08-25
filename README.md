# Multi-Tenant Inventory Management System

A production-ready multi-tenant inventory management system built with Go (backend) and React (frontend). Each tenant has completely isolated data and can have their own users, items, locations, and all other resources.

## Features

### Core Inventory Management
- **Items & Catalog**: Create/edit items with SKU, barcode, pricing, categories
- **Stock Tracking**: Multi-location inventory with on-hand, allocated, available tracking
- **Purchase Orders**: Create POs, receive items, auto-adjust stock
- **Transfers**: Move stock between locations with approval workflow
- **Adjustments**: Cycle counts, corrections, damages with audit trail
- **Suppliers**: Manage supplier master data
- **Users & Roles**: RBAC with Admin, Manager, Clerk roles
- **Barcode Support**: Barcode scanning compatibility
- **Audit Trail**: Complete audit log of all stock-affecting actions

### Multi-Tenancy
- **Complete Data Isolation**: Each tenant has completely separate data
- **Tenant Management**: Create and manage multiple tenants/companies
- **Per-Tenant Users**: Users belong to specific tenants
- **Unique Constraints**: SKUs, barcodes, etc. are unique per tenant
- **Tenant Context**: All API requests are automatically scoped to the user's tenant

## Tech Stack

### Backend
- **Go 1.22+** - Main language
- **PostgreSQL** - Database
- **Ent ORM** - Database ORM and migrations
- **Echo** - HTTP framework
- **JWT** - Authentication
- **Zerolog** - Structured logging

### Frontend
- **React 18+** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool
- **Tailwind CSS** - Styling
- **React Router** - Client-side routing
- **React Query** - Server state management
- **React Hook Form** - Form handling
- **Zod** - Schema validation

## Quick Start

### Prerequisites
- Go 1.22+
- Node.js 18+
- PostgreSQL 15+
- Docker (optional)

### With Docker (Recommended)

1. Clone the repository:
```bash
git clone <repository-url>
cd inventory
```

2. Start all services:
```bash
make docker-up
```

3. Access the application:
- Frontend: http://localhost
- Backend API: http://localhost:8080

### Manual Setup

1. Start PostgreSQL database

2. Set up backend:
```bash
cd backend
cp .env.example .env  # Edit with your database URL
go mod tidy
make migrate
go run cmd/api/main.go
```

3. Set up frontend:
```bash
cd frontend
npm install
npm run dev
```

## Environment Variables

### Backend (.env)
```bash
DATABASE_URL=postgres://postgres:postgres@localhost:5432/inventory?sslmode=disable
JWT_SECRET=change-me-in-production
PORT=8080
ENVIRONMENT=development
LOG_LEVEL=debug
CORS_ORIGINS=http://localhost:5173,http://localhost:3000
JWT_EXPIRY_MINUTES=15
REFRESH_EXPIRY_DAYS=7
```

### Frontend (.env)
```bash
VITE_API_URL=http://localhost:8080/api/v1
```

## Multi-Tenant Migration

If you're upgrading from a single-tenant version, run the migration script:

```bash
cd backend
go run cmd/migrate-to-multitenant/main.go
```

This script will:
1. Create the `tenants` table
2. Create a default tenant for existing data
3. Add `tenant_id` columns to all existing tables
4. Update constraints to be unique per tenant
5. Add necessary indexes for performance

The default tenant will be created with:
- **Name**: "Default Company"
- **Slug**: "default"

All existing data will be assigned to this default tenant.

## API Documentation

### Authentication & Registration
- `POST /api/v1/auth/login` - Login with email/password (optionally specify tenant)
- `POST /api/v1/auth/register` - Register new user and/or tenant
- `GET /api/v1/auth/tenant-lookup` - Find tenants associated with email
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout

### Items
- `GET /api/v1/items` - List items with pagination and filters
- `POST /api/v1/items` - Create new item
- `GET /api/v1/items/{id}` - Get item by ID
- `PUT /api/v1/items/{id}` - Update item
- `DELETE /api/v1/items/{id}` - Soft delete item

### Inventory
- `GET /api/v1/inventory` - Get inventory levels
- `GET /api/v1/inventory/{item_id}/locations` - Get item by location
- `GET /api/v1/inventory/movements` - Get stock movements

### Purchase Orders
- `GET /api/v1/pos` - List purchase orders
- `POST /api/v1/pos` - Create purchase order
- `POST /api/v1/pos/{id}/approve` - Approve PO
- `POST /api/v1/pos/{id}/receive` - Receive items

### Tenants (System Admin Only)
- `GET /api/v1/system/tenants` - List all tenants
- `POST /api/v1/system/tenants` - Create new tenant
- `GET /api/v1/system/tenants/{id}` - Get tenant details
- `PUT /api/v1/system/tenants/{id}` - Update tenant
- `DELETE /api/v1/system/tenants/{id}` - Deactivate tenant
- `GET /api/v1/tenant` - Get current user's tenant info

### Tenant Context
All API requests (except system admin endpoints) are automatically scoped to the user's tenant. The tenant ID is extracted from the JWT token and all database queries include tenant filtering.

## Registration & Tenant Management

### New Customer/Tenant Registration

Users can register in multiple ways:

1. **Create New Company** (`POST /api/v1/auth/register`)
   ```json
   {
     "name": "John Doe",
     "email": "john@example.com", 
     "password": "securepassword",
     "tenant_name": "Acme Corporation",
     "tenant_slug": "acme-corp"
   }
   ```
   - Creates a new tenant and makes the user an admin
   - Tenant slug is auto-generated if not provided

2. **Join Existing Company** (`POST /api/v1/auth/register`)
   ```json
   {
     "name": "Jane Smith",
     "email": "jane@example.com",
     "password": "securepassword", 
     "tenant_slug": "acme-corp"
   }
   ```
   - Adds user to existing tenant as clerk role
   - Requires knowing the company identifier

3. **Tenant Lookup** (`GET /api/v1/auth/tenant-lookup?email=user@example.com`)
   - Returns all tenants associated with an email
   - Helps users find their company identifiers

### Login Options

- **Basic Login**: Email + password (uses first tenant found)
- **Tenant-Specific Login**: Email + password + tenant_slug
- **Multi-Tenant Users**: Users can belong to multiple tenants

## Default Login

For testing, use these credentials:
- **Email**: admin@example.com
- **Password**: admin123

## Development

### Available Make Commands

```bash
make help          # Show all available commands
make dev           # Start development environment
make test          # Run all tests
make build         # Build both backend and frontend
make lint          # Run linters
make docker-up     # Start with Docker
make docker-down   # Stop Docker services
```

### Database Migrations

Migrations are handled by Ent ORM. To create new migrations:

```bash
cd backend
go generate ./ent
```

### Testing

Run backend tests:
```bash
cd backend
go test ./...
```

Run frontend tests:
```bash
cd frontend
npm test
```

## Project Structure

```
inventory/
├── backend/                 # Go backend
│   ├── cmd/api/            # Main application
│   ├── ent/schema/         # Database schemas
│   ├── internal/           # Internal packages
│   │   ├── handlers/       # HTTP handlers
│   │   ├── middleware/     # HTTP middleware
│   │   ├── services/       # Business logic
│   │   └── config/         # Configuration
│   └── Dockerfile
├── frontend/               # React frontend
│   ├── src/
│   │   ├── components/     # React components
│   │   ├── pages/          # Page components
│   │   ├── contexts/       # React contexts
│   │   ├── lib/           # Utilities
│   │   └── api/           # API client
│   └── Dockerfile
├── docker-compose.yml      # Docker services
├── Makefile               # Development commands
└── README.md
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Run linters: `make lint`
5. Submit a pull request

## Security

- JWT tokens for authentication
- Password hashing with bcrypt
- Input validation and sanitization
- CORS protection
- SQL injection protection via ORM
- Request rate limiting (TODO)

## Production Deployment

1. Set strong JWT secrets
2. Use environment variables for configuration
3. Set up SSL/TLS certificates
4. Configure database backups
5. Set up monitoring and logging
6. Use a reverse proxy (nginx)

## License

MIT License - see LICENSE file for details.

## Support

For issues and questions, please create a GitHub issue.