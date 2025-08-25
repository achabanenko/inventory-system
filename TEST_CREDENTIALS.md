# Test Credentials

## Login to the application

**Email:** `admin@example.com`  
**Password:** `admin123`

## URLs

- **Frontend:** http://localhost:3000
- **Backend API:** http://localhost:8080/api/v1

## Current Status ✅

- ✅ **Backend:** Running on port 8080 with CORS enabled
- ✅ **Frontend:** Running on port 3000  
- ✅ **CORS:** Configured for http://localhost:3000
- ✅ **Database:** Seeded with test data
- ✅ **Purchase Orders:** All functions implemented (Create/Edit/View/Delete/Search)

## Test Data

The database has been seeded with:
- ✅ Admin user (credentials above)
- ✅ Sample suppliers (Tech Solutions Inc, Office Pro Supply, Industrial Hardware Co)
- ✅ Sample items (Laptops, Mice, Paper, Pens, Monitors)
- ✅ Sample purchase order (PO-000001) with 2 line items

## To test Purchase Orders:

1. Go to http://localhost:3000
2. Login with the credentials above
3. Navigate to "Purchase Orders" in the sidebar
4. You should see the sample purchase order "PO-000001"

## Available Actions:

- **View** purchase orders list with search/filtering
- **Create** new purchase orders (button in top right)
- **Edit** draft purchase orders
- **Approve** draft purchase orders  
- **Receive** items from approved purchase orders
- **Close** completed purchase orders
- **Delete** draft purchase orders
