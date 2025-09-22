# Data Model: TT Stock Backend API

## Core Entities

### User
**Purpose**: Represents system users with authentication and role-based access
**Fields**:
- `ID` (uint, primary key, auto-increment)
- `PhoneNumber` (string, unique, required) - Primary identifier for mobile auth
- `PIN` (string, hashed, required) - 4-6 digit PIN for authentication
- `Role` (enum: Admin, Owner, Staff, required) - Access control level
- `Name` (string, required) - User display name
- `IsActive` (boolean, default: true) - Account status
- `CreatedAt` (timestamp, auto-generated)
- `UpdatedAt` (timestamp, auto-updated)
- `LastLoginAt` (timestamp, nullable) - Track login activity

**Validation Rules**:
- Phone number must be valid format (10-15 digits)
- PIN must be 4-6 digits, hashed with bcrypt
- Role must be one of: Admin, Owner, Staff
- Name must be 2-100 characters

**Relationships**:
- One-to-many with StockMovement (created_by)
- One-to-many with Session (user_id)

### Product
**Purpose**: Represents individual tire or wheel items with complete specifications
**Fields**:
- `ID` (uint, primary key, auto-increment)
- `Type` (enum: Tire, Wheel, required) - Product category
- `Brand` (string, required) - Manufacturer brand
- `Model` (string, required) - Product model name
- `SKU` (string, unique, required) - Stock keeping unit
- `Description` (text, nullable) - Product description
- `ImageBase64` (text, nullable) - Product image as base64 string
- `CostPrice` (decimal, required) - Purchase cost
- `SellingPrice` (decimal, required) - Retail price
- `QuantityOnHand` (integer, default: 0) - Current stock level
- `LowStockThreshold` (integer, default: 5) - Alert threshold
- `IsActive` (boolean, default: true) - Product status
- `CreatedAt` (timestamp, auto-generated)
- `UpdatedAt` (timestamp, auto-updated)

**Validation Rules**:
- Brand and model must be 1-100 characters
- SKU must be unique and 3-50 characters
- Prices must be positive decimal values
- Quantity must be non-negative integer
- Low stock threshold must be positive integer

**Relationships**:
- One-to-many with StockMovement (product_id)
- One-to-many with ProductSpecification (product_id)

### ProductSpecification
**Purpose**: Stores detailed specifications for tires and wheels
**Fields**:
- `ID` (uint, primary key, auto-increment)
- `ProductID` (uint, foreign key, required) - References Product
- `SpecType` (enum: Tire, Wheel, required) - Specification category
- `SpecData` (jsonb, required) - Flexible specification storage

**Tire Specifications** (SpecData JSON):
```json
{
  "width": "225",
  "aspectRatio": "45",
  "diameter": "17",
  "loadIndex": "91",
  "speedRating": "W",
  "dotYear": "2023",
  "season": "All-Season",
  "runFlat": false
}
```

**Wheel Specifications** (SpecData JSON):
```json
{
  "diameter": "17",
  "width": "8.5",
  "offset": "35",
  "boltPattern": "5x114.3",
  "centerBore": "67.1",
  "color": "Black",
  "finish": "Matte",
  "weight": "22.5"
}
```

**Validation Rules**:
- SpecData must be valid JSON
- Tire specs must include width, aspectRatio, diameter
- Wheel specs must include diameter, width, offset, boltPattern
- All numeric values must be positive

**Relationships**:
- Many-to-one with Product (product_id)

### StockMovement
**Purpose**: Records all inventory changes with audit trail
**Fields**:
- `ID` (uint, primary key, auto-increment)
- `ProductID` (uint, foreign key, required) - References Product
- `UserID` (uint, foreign key, required) - References User (who made change)
- `MovementType` (enum: Incoming, Sale, Demo, Damage, Return, Adjustment, required)
- `Quantity` (integer, required) - Positive for incoming, negative for outgoing
- `Reason` (string, nullable) - Business reason for movement
- `Reference` (string, nullable) - External reference (PO number, invoice, etc.)
- `Notes` (text, nullable) - Additional details
- `CreatedAt` (timestamp, auto-generated)

**Validation Rules**:
- Quantity cannot be zero
- MovementType must be valid enum value
- Reason must be 1-200 characters if provided
- Reference must be 1-100 characters if provided

**Relationships**:
- Many-to-one with Product (product_id)
- Many-to-one with User (user_id)

### Session
**Purpose**: Tracks user authentication sessions
**Fields**:
- `ID` (uint, primary key, auto-increment)
- `UserID` (uint, foreign key, required) - References User
- `Token` (string, unique, required) - JWT token string
- `ExpiresAt` (timestamp, required) - Token expiration (1 day from creation)
- `IsActive` (boolean, default: true) - Session status
- `CreatedAt` (timestamp, auto-generated)
- `LastUsedAt` (timestamp, auto-updated) - Track activity

**Validation Rules**:
- Token must be valid JWT format
- ExpiresAt must be in future
- Token must be unique across all sessions

**Relationships**:
- Many-to-one with User (user_id)

### Alert
**Purpose**: Manages low-stock notifications and system warnings
**Fields**:
- `ID` (uint, primary key, auto-increment)
- `ProductID` (uint, foreign key, required) - References Product
- `AlertType` (enum: LowStock, OutOfStock, PriceChange, required)
- `Message` (string, required) - Alert message
- `IsRead` (boolean, default: false) - Read status
- `IsActive` (boolean, default: true) - Alert status
- `CreatedAt` (timestamp, auto-generated)
- `ReadAt` (timestamp, nullable) - When alert was read

**Validation Rules**:
- Message must be 1-500 characters
- AlertType must be valid enum value

**Relationships**:
- Many-to-one with Product (product_id)

## Database Schema Design

### Indexes
**Performance Optimization**:
- `users.phone_number` (unique index) - Fast login lookup
- `products.sku` (unique index) - Fast product lookup
- `products.type` (index) - Filter by product type
- `products.brand` (index) - Filter by brand
- `stock_movements.product_id` (index) - Product movement history
- `stock_movements.created_at` (index) - Time-based queries
- `sessions.token` (unique index) - Fast token validation
- `sessions.expires_at` (index) - Token cleanup queries

### Constraints
**Data Integrity**:
- Foreign key constraints on all relationships
- Check constraints for positive quantities and prices
- Unique constraints on phone numbers and SKUs
- Not null constraints on required fields
- Default values for optional fields

### Triggers
**Business Logic**:
- Update `products.quantity_on_hand` when stock movements occur
- Generate alerts when stock falls below threshold
- Update `users.last_login_at` on successful authentication
- Clean up expired sessions automatically

## State Transitions

### Product Lifecycle
1. **Created** → Active (default state)
2. **Active** → Inactive (discontinued)
3. **Inactive** → Active (reinstated)

### User Account Lifecycle
1. **Created** → Active (default state)
2. **Active** → Inactive (suspended)
3. **Inactive** → Active (reactivated)

### Session Lifecycle
1. **Created** → Active (valid token)
2. **Active** → Expired (token expired)
3. **Active** → Inactive (logout)

### Stock Movement Flow
1. **Incoming** → Increases quantity_on_hand
2. **Sale** → Decreases quantity_on_hand (with validation)
3. **Demo** → Decreases quantity_on_hand
4. **Damage** → Decreases quantity_on_hand
5. **Return** → Increases quantity_on_hand
6. **Adjustment** → Can increase or decrease quantity_on_hand

## Data Validation Rules

### Business Rules
- Stock quantity cannot go negative (enforced at application level)
- Selling price must be greater than cost price (warning, not enforced)
- PIN must be 4-6 digits (enforced at registration)
- Phone number must be unique across all users
- SKU must be unique across all products
- JWT tokens expire after 1 day
- Low stock alerts generated when quantity <= threshold

### Data Consistency
- All timestamps in UTC
- Decimal precision: 2 places for prices
- String lengths enforced at database level
- JSON schema validation for product specifications
- Cascade delete rules for dependent records
