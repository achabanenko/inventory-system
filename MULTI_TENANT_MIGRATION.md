# Multi-Tenant Migration Guide

This document outlines the conversion of the inventory management system from single-tenant to multi-tenant architecture.

## Overview

The system has been converted to support multiple tenants with complete data isolation. Each tenant operates as an independent instance with their own:

- Users and authentication
- Items, categories, and suppliers
- Locations and inventory levels
- Purchase orders, transfers, and adjustments
- Audit logs and stock movements

## Database Changes

### New Tables

#### `tenants`
- `id` (UUID, Primary Key)
- `name` (VARCHAR, Company/tenant name)
- `slug` (VARCHAR, Unique URL-safe identifier)
- `domain` (VARCHAR, Optional custom domain)
- `settings` (JSONB, Tenant-specific settings)
- `contact` (JSONB, Contact information)
- `is_active` (BOOLEAN, Activation status)
- `created_at`, `updated_at` (Timestamps)

### Modified Tables

All existing tables now include:
- `tenant_id` (UUID, Foreign Key to tenants.id, NOT NULL)

### Updated Constraints

All unique constraints are now scoped per tenant:
- `users`: email unique per tenant
- `items`: SKU and barcode unique per tenant
- `locations`: code unique per tenant
- `suppliers`: code unique per tenant
- `purchase_orders`: number unique per tenant
- `transfers`: number unique per tenant
- `adjustments`: number unique per tenant
- `inventory_levels`: item+location unique per tenant
- `categories`: name unique per tenant

## Backend Changes

### New Files

1. **`backend/ent/schema/tenant.go`**
   - Tenant entity schema with relationships to all other entities

2. **`backend/internal/middleware/tenant.go`**
   - Tenant context middleware for request isolation
   - Functions to extract and validate tenant ID from requests
   - Support for multiple tenant identification methods (header, JWT, subdomain)

3. **`backend/internal/services/tenant.go`**
   - Tenant management service with CRUD operations
   - Slug validation and tenant lookup functions

4. **`backend/internal/handlers/tenants.go`**
   - HTTP handlers for tenant management API endpoints
   - System admin only access control

5. **`backend/cmd/migrate-to-multitenant/main.go`**
   - Migration script to convert existing single-tenant data
   - Creates default tenant and updates existing records

6. **`backend/internal/middleware/tenant_test.go`**
   - Unit tests for tenant middleware functionality

### Modified Files

1. **All Entity Schemas (`backend/ent/schema/*.go`)**
   - Added tenant relationships to all entities
   - Updated indexes to include tenant constraints

2. **Auth Middleware (`backend/internal/middleware/auth.go`)**
   - Updated JWT claims to include tenant ID
   - Automatic tenant context setting from JWT tokens

3. **Auth Handler (`backend/internal/handlers/auth.go`)**
   - Updated login to include tenant information in JWT
   - Enhanced user lookup with tenant validation

4. **Items Handler (`backend/internal/handlers/items.go`)**
   - Added tenant filtering to all database queries
   - Tenant context validation in all endpoints

5. **Main Application (`backend/cmd/api/main.go`)**
   - Added tenant middleware to protected routes
   - New system admin routes for tenant management

## Frontend Changes

### Modified Files

1. **Auth Context (`frontend/src/contexts/AuthContext.tsx`)**
   - Added tenant information to user context
   - Updated authentication flow to handle tenant data

2. **API Client (`frontend/src/lib/api.ts`)**
   - Added tenant ID header to all requests
   - Enhanced request interceptor for tenant context

3. **Layout Component (`frontend/src/components/Layout.tsx`)**
   - Display tenant name in application header
   - Enhanced user information display

## API Changes

### New Endpoints

#### System Admin Endpoints (No tenant context required)
- `GET /api/v1/system/tenants` - List all tenants
- `POST /api/v1/system/tenants` - Create new tenant
- `GET /api/v1/system/tenants/{id}` - Get tenant details
- `PUT /api/v1/system/tenants/{id}` - Update tenant
- `DELETE /api/v1/system/tenants/{id}` - Deactivate tenant

#### Tenant Information
- `GET /api/v1/tenant` - Get current user's tenant info

### Modified Endpoints

All existing endpoints now automatically include tenant filtering:
- Users can only access data within their tenant
- All queries include `tenant_id` filtering
- All create operations automatically set tenant ID

### Authentication Changes

JWT tokens now include:
```json
{
  "user_id": "uuid",
  "tenant_id": "uuid",
  "email": "user@example.com",
  "role": "ADMIN|MANAGER|CLERK"
}
```

## Migration Process

### For New Installations

1. The system is multi-tenant by default
2. First user registration creates the first tenant
3. Subsequent users are assigned to tenants during registration

### For Existing Installations

1. **Run Migration Script**:
   ```bash
   cd backend
   go run cmd/migrate-to-multitenant/main.go
   ```

2. **Migration Steps**:
   - Creates `tenants` table
   - Creates default tenant ("Default Company", slug: "default")
   - Adds `tenant_id` columns to all tables
   - Assigns all existing data to default tenant
   - Updates constraints and indexes

3. **Post-Migration**:
   - All existing users belong to the default tenant
   - System continues to work as before for existing users
   - New tenants can be created via system admin APIs

## Security Considerations

### Data Isolation

- **Database Level**: All queries include tenant filtering
- **Application Level**: Middleware enforces tenant context
- **API Level**: All endpoints validate tenant access

### Access Control

- **Tenant Users**: Can only access their tenant's data
- **System Admins**: Can manage tenants but require special role
- **JWT Tokens**: Include tenant ID for automatic scoping

### Validation

- All tenant-scoped resources validate tenant ownership
- Cross-tenant access is prevented at middleware level
- Database constraints ensure data integrity

## Performance Considerations

### Indexes

Added indexes on `tenant_id` columns for all tables:
```sql
CREATE INDEX idx_tablename_tenant_id ON tablename(tenant_id);
```

### Query Performance

- All queries include tenant filtering in WHERE clauses
- Indexes ensure efficient tenant-based lookups
- Unique constraints scoped per tenant reduce conflicts

## Testing

### Unit Tests

- Tenant middleware functionality
- Tenant context validation
- Multi-tenant constraint testing

### Integration Tests

- Cross-tenant data isolation
- API endpoint tenant scoping
- Authentication with tenant context

## Future Enhancements

### Planned Features

1. **Subdomain-based Tenancy**
   - `tenant1.yourapp.com` routing
   - Automatic tenant detection from subdomain

2. **Custom Domains**
   - `inventory.company.com` support
   - SSL certificate management per tenant

3. **Tenant-specific Configurations**
   - Custom branding and themes
   - Feature flags per tenant
   - Configurable workflows

4. **Advanced Administration**
   - Tenant usage analytics
   - Resource quotas and limits
   - Billing integration

5. **Data Export/Import**
   - Tenant data backup/restore
   - Cross-tenant data migration
   - Bulk data operations

## Troubleshooting

### Common Issues

1. **Migration Fails**
   - Ensure database has sufficient permissions
   - Check for existing constraint conflicts
   - Verify database connection

2. **Tenant Context Missing**
   - Verify JWT token includes tenant_id
   - Check middleware order in routing
   - Validate token claims structure

3. **Cross-tenant Data Visible**
   - Verify tenant_id filtering in queries
   - Check middleware application to routes
   - Validate database constraints

### Debug Commands

```bash
# Check tenant data isolation
SELECT tenant_id, COUNT(*) FROM items GROUP BY tenant_id;

# Verify constraints
\d+ items  # PostgreSQL describe table

# Test API with tenant header
curl -H "X-Tenant-ID: your-tenant-id" http://localhost:8080/api/v1/items
```

## Support

For issues related to multi-tenant functionality:

1. Check this migration guide
2. Review tenant middleware logs
3. Validate database schema changes
4. Test with isolated tenant data
5. Create GitHub issue with reproduction steps
