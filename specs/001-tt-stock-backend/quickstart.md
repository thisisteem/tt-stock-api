# Quickstart Guide: TT Stock Backend API

## Overview
This guide demonstrates the core functionality of the TT Stock Backend API through practical examples. The API provides comprehensive inventory management for tire and wheel shops with real-time stock tracking, search capabilities, and business intelligence.

## Prerequisites
- API server running on `http://localhost:8080/v1`
- Valid JWT token for authentication
- Test data populated in the database

## Authentication Flow

### 1. User Login
```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phoneNumber": "1234567890",
    "pin": "1234"
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "phoneNumber": "1234567890",
      "name": "John Doe",
      "role": "Staff",
      "isActive": true
    },
    "expiresAt": "2024-09-22T12:00:00Z"
  }
}
```

### 2. Set Authorization Header
```bash
export TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
export AUTH_HEADER="Authorization: Bearer $TOKEN"
```

## Product Management

### 3. Create a New Tire Product
```bash
curl -X POST http://localhost:8080/v1/products \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "type": "Tire",
    "brand": "Michelin",
    "model": "Pilot Sport 4",
    "sku": "MIC-PS4-225-45-17",
    "description": "High-performance summer tire",
    "costPrice": 150.00,
    "sellingPrice": 200.00,
    "quantityOnHand": 25,
    "lowStockThreshold": 5,
    "specifications": {
      "specType": "Tire",
      "specData": {
        "width": "225",
        "aspectRatio": "45",
        "diameter": "17",
        "loadIndex": "91",
        "speedRating": "W",
        "dotYear": "2023",
        "season": "All-Season",
        "runFlat": false
      }
    }
  }'
```

### 4. Create a New Wheel Product
```bash
curl -X POST http://localhost:8080/v1/products \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "type": "Wheel",
    "brand": "Enkei",
    "model": "Racing RPF1",
    "sku": "ENK-RPF1-17-8.5-35",
    "description": "Lightweight racing wheel",
    "costPrice": 300.00,
    "sellingPrice": 400.00,
    "quantityOnHand": 10,
    "lowStockThreshold": 3,
    "specifications": {
      "specType": "Wheel",
      "specData": {
        "diameter": "17",
        "width": "8.5",
        "offset": "35",
        "boltPattern": "5x114.3",
        "centerBore": "67.1",
        "color": "Black",
        "finish": "Matte",
        "weight": "22.5"
      }
    }
  }'
```

### 5. Search Products by Tire Size
```bash
curl -X POST http://localhost:8080/v1/products/search \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "type": "Tire",
    "tireSpecs": {
      "width": "225",
      "aspectRatio": "45",
      "diameter": "17"
    },
    "stockStatus": "available"
  }'
```

### 6. Search Products by Wheel Fitment
```bash
curl -X POST http://localhost:8080/v1/products/search \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "type": "Wheel",
    "wheelSpecs": {
      "diameter": "17",
      "boltPattern": "5x114.3",
      "color": "Black"
    }
  }'
```

## Stock Management

### 7. Record Incoming Stock
```bash
curl -X POST http://localhost:8080/v1/stock/movements \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "productId": 1,
    "movementType": "Incoming",
    "quantity": 10,
    "reason": "New shipment from supplier",
    "reference": "PO-2024-001",
    "notes": "Received 10 units of Michelin Pilot Sport 4"
  }'
```

### 8. Process a Sale
```bash
curl -X POST http://localhost:8080/v1/stock/sale \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "productId": 1,
    "quantity": 2,
    "customerName": "Jane Smith",
    "reference": "INV-2024-001",
    "notes": "Customer requested installation service"
  }'
```

### 9. Record Damaged Stock
```bash
curl -X POST http://localhost:8080/v1/stock/movements \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "productId": 1,
    "movementType": "Damage",
    "quantity": -1,
    "reason": "Sidewall damage during handling",
    "notes": "Product damaged during warehouse operations"
  }'
```

### 10. Check Current Inventory
```bash
curl -X GET "http://localhost:8080/v1/stock/inventory?stockStatus=lowStock" \
  -H "$AUTH_HEADER"
```

### 11. View Stock Movement History
```bash
curl -X GET "http://localhost:8080/v1/stock/movements?productId=1&limit=10" \
  -H "$AUTH_HEADER"
```

## Business Intelligence

### 12. Get Stock Alerts
```bash
curl -X GET "http://localhost:8080/v1/stock/alerts?isRead=false" \
  -H "$AUTH_HEADER"
```

### 13. Mark Alert as Read
```bash
curl -X PUT http://localhost:8080/v1/stock/alerts/1/read \
  -H "$AUTH_HEADER"
```

### 14. Get Inventory Summary
```bash
curl -X GET http://localhost:8080/v1/stock/inventory \
  -H "$AUTH_HEADER"
```

## Error Handling Examples

### 15. Invalid Login Credentials
```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phoneNumber": "1234567890",
    "pin": "9999"
  }'
```

**Expected Response:**
```json
{
  "success": false,
  "message": "Invalid credentials"
}
```

### 16. Insufficient Stock for Sale
```bash
curl -X POST http://localhost:8080/v1/stock/sale \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "productId": 1,
    "quantity": 1000
  }'
```

**Expected Response:**
```json
{
  "success": false,
  "message": "Insufficient stock available",
  "errors": [
    {
      "field": "quantity",
      "message": "Requested quantity (1000) exceeds available stock (25)"
    }
  ]
}
```

### 17. Unauthorized Access
```bash
curl -X GET http://localhost:8080/v1/products \
  -H "Authorization: Bearer invalid-token"
```

**Expected Response:**
```json
{
  "success": false,
  "message": "Invalid or expired token"
}
```

## Performance Testing

### 18. Load Test - Multiple Concurrent Searches
```bash
# Run 10 concurrent product searches
for i in {1..10}; do
  curl -X POST http://localhost:8080/v1/products/search \
    -H "Content-Type: application/json" \
    -H "$AUTH_HEADER" \
    -d '{"type": "Tire", "brand": "Michelin"}' &
done
wait
```

### 19. Response Time Validation
```bash
# Measure response time for product search
time curl -X POST http://localhost:8080/v1/products/search \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{"type": "Tire"}' \
  -w "Response time: %{time_total}s\n"
```

## Data Validation Examples

### 20. Invalid Product Data
```bash
curl -X POST http://localhost:8080/v1/products \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "type": "Tire",
    "brand": "",
    "model": "Test",
    "sku": "TEST-001",
    "costPrice": -10.00,
    "sellingPrice": 100.00
  }'
```

**Expected Response:**
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": [
    {
      "field": "brand",
      "message": "Brand is required"
    },
    {
      "field": "costPrice",
      "message": "Cost price must be positive"
    }
  ]
}
```

## Success Criteria Validation

### ✅ Operational Efficiency
- Product search returns results in <200ms
- Stock movements are recorded in real-time
- Inventory levels update immediately after sales

### ✅ Accuracy
- Stock quantities never go negative
- All movements are tracked with user attribution
- Product specifications are validated before storage

### ✅ Business Intelligence
- Low stock alerts are generated automatically
- Inventory summary provides accurate totals
- Movement history is complete and auditable

### ✅ Scalability
- API handles multiple concurrent requests
- Database queries are optimized with proper indexing
- Response times remain consistent under load

### ✅ Reliability
- Authentication tokens expire after 1 day
- Error responses are consistent and informative
- System maintains data integrity during concurrent operations

## Next Steps
1. **Integration Testing**: Run comprehensive test suite
2. **Performance Testing**: Validate response time requirements
3. **Security Testing**: Verify authentication and authorization
4. **Load Testing**: Test with 1000+ concurrent users
5. **Mobile Integration**: Test with Flutter mobile application
