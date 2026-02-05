# Code Patterns & Examples - PimpMyPack

This document contains detailed code patterns, examples, and templates used in the PimpMyPack project.

> **See also**: [agents.md](../agents.md) for core principles and architecture overview.

## Table of Contents

- [Architecture & Package Organization](#architecture--package-organization)
- [Handlers](#handlers)
- [Business Functions](#business-functions)
- [Testing](#testing)
- [Error Handling](#error-handling)
- [Database Operations](#database-operations)
- [Swagger Documentation](#swagger-documentation)

## Architecture & Package Organization

### Package Structure

The codebase follows a **domain-driven design** approach where each business domain has its own package with clear boundaries:

```markdown
pkg/
├── accounts/          # User accounts and authentication
│   ├── types.go       # Domain types (User, Account, LoginInput, etc.)
│   ├── accounts.go    # Handlers, service, and repository functions
│   └── testdata.go    # Test fixtures
├── packs/             # Pack management
│   ├── types.go       # Domain types (Pack, PackContent, etc.)
│   ├── handlers.go    # HTTP handlers
│   ├── service.go     # Business logic (public functions)
│   ├── repository.go  # Data access (private functions)
│   └── testdata.go    # Test fixtures
├── inventories/       # Inventory items
│   ├── types.go       # Domain types (Inventory, etc.)
│   ├── handlers.go    # HTTP handlers
│   ├── service.go     # Business logic (public functions)
│   ├── repository.go  # Data access (private functions)
│   └── testdata.go    # Test fixtures
├── images/            # Image management
│   ├── types.go       # Domain types and interfaces
│   ├── handlers.go    # HTTP handlers
│   ├── processor.go   # Image processing logic
│   └── storage_db.go  # Database storage implementation
├── apitypes/          # Shared HTTP response types
│   └── types.go       # OkResponse, ErrorResponse
├── config/            # Configuration
│   └── env.go         # Environment config, MailServer type
├── helper/            # Generic utilities
│   └── helper.go      # StringToUint, GenerateRandomCode, SMTPClient, etc.
└── security/          # Security utilities
    └── security.go    # Password hashing, JWT, etc.
```

### Domain Types Location

**IMPORTANT**: Types must be defined in their domain package, not in a centralized `dataset` package.

```go
// ✅ CORRECT: Domain-specific types in their own package
package accounts

type User struct {
    ID        uint      `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    // ...
}

type LoginInput struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}
```

```go
// ❌ INCORRECT: Don't use a centralized dataset package
package dataset

type User struct { /* ... */ }
type Pack struct { /* ... */ }
type Inventory struct { /* ... */ }
```

### Package File Organization

#### Option A: Single File (Small Packages)

For smaller packages, keep everything in one file:

```go
// pkg/accounts/accounts.go
package accounts

// Types are in types.go

// Handlers (public, Gin-specific)
func GetAccounts(c *gin.Context) { /* ... */ }

// Service functions (public, framework-agnostic)
func GetAllAccounts(ctx context.Context) ([]Account, error) { /* ... */ }

// Repository functions (private, database access)
func getAllAccounts(ctx context.Context, db *sql.DB) ([]Account, error) { /* ... */ }

// Helper functions (package-specific)
func FindUserIDByUsername(users []Account, username string) uint { /* ... */ }
```

#### Option B: Separate Files (Larger Packages)

For larger packages, separate concerns into files:

```go
// pkg/inventories/types.go or pkg/packs/types.go
package inventories
type Inventory struct { /* ... */ }

// pkg/inventories/handlers.go or pkg/packs/handlers.go
package inventories
func GetMyInventories(c *gin.Context) { /* ... */ }

// pkg/inventories/service.go or pkg/packs/service.go (public business functions)
package inventories
func GetAllInventoriesByUserID(ctx context.Context, userID uint) (Inventories, error) { /* ... */ }

// pkg/inventories/repository.go or pkg/packs/repository.go (private data access)
package inventories
func getAllInventoriesByUserID(ctx context.Context, db *sql.DB, userID uint) (Inventories, error) { /* ... */ }
```

**Examples**: Both `pkg/inventories/` and `pkg/packs/` follow this pattern.

### Helper Functions

Helper functions should live in the package they belong to, **not** in a centralized helper package.

```go
// ✅ CORRECT: Domain-specific helpers in domain package
// pkg/packs/packs.go
func FindPackIDByPackName(packs Packs, packname string) uint {
    for _, pack := range packs {
        if pack.PackName == packname {
            return pack.ID
        }
    }
    return 0
}

// pkg/accounts/accounts.go
func FindUserIDByUsername(users []User, username string) uint {
    for _, user := range users {
        if user.Username == username {
            return user.ID
        }
    }
    return 0
}
```

```go
// ✅ CORRECT: Generic utilities in helper package
// pkg/helper/helper.go
func StringToUint(s string) (uint, error) { /* ... */ }
func GenerateRandomCode(length int) (string, error) { /* ... */ }
func IsValidEmail(email string) bool { /* ... */ }
```

```go
// ❌ INCORRECT: Domain-specific logic in helper package
// pkg/helper/helper.go
func FindPackIDByPackName(packs dataset.Packs, packname string) uint { /* ... */ }
```

### Service Layer Pattern

Use a clean separation between handlers (HTTP), service (business logic), and repository (data access):

```go
// Handler layer (Gin-specific, public)
func GetMyInventories(c *gin.Context) {
    userID := c.GetUint("user_id")

    inventories, err := GetAllInventoriesByUserID(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, inventories)
}

// Service layer (framework-agnostic, public)
func GetAllInventoriesByUserID(ctx context.Context, userID uint) (Inventories, error) {
    db := database.DB()
    return getAllInventoriesByUserID(ctx, db, userID)
}

// Repository layer (data access, private)
func getAllInventoriesByUserID(ctx context.Context, db *sql.DB, userID uint) (Inventories, error) {
    query := `SELECT id, user_id, item_name, category, weight FROM inventory WHERE user_id = $1`
    // ... database query logic
}
```

### Avoiding Circular Dependencies

**Problem**: If package A imports package B, then package B cannot import package A.

**Solution**: Keep dependencies flowing in one direction:

```markdown
handlers → service → repository → database
   ↓
 types (same package)
```

**Example of proper dependency flow**:

```go
// ✅ CORRECT: Config defines MailServer, helper uses it
// pkg/config/env.go
package config

type MailServer struct {
    MailServer   string
    MailPort     int
    MailIdentity string
    MailUsername string
    MailPassword string
}

var MailServerConfig MailServer

// pkg/helper/helper.go
package helper

import "github.com/Angak0k/pimpmypack/pkg/config"

type SMTPClient struct {
    Server config.MailServer  // Uses type from config
}

// pkg/accounts/accounts.go
package accounts

import (
    "github.com/Angak0k/pimpmypack/pkg/config"
    "github.com/Angak0k/pimpmypack/pkg/helper"
)

func sendVerificationEmail() {
    client := helper.SMTPClient{
        Server: config.MailServerConfig,  // Uses instance from config
    }
    // ...
}
```

### Shared Types

For types shared across multiple packages:

- **HTTP responses**: Use `pkg/apitypes` (OkResponse, ErrorResponse)
- **Configuration**: Use `pkg/config` (MailServer, database config)
- **Utilities**: Use appropriate utility package (security, helper)

```go
// pkg/apitypes/types.go
package apitypes

type OkResponse struct {
    Response string `json:"message"`
}

type ErrorResponse struct {
    Error string `json:"error"`
}

// Usage in handlers
c.JSON(http.StatusOK, apitypes.OkResponse{Response: "Success"})
c.JSON(http.StatusBadRequest, apitypes.ErrorResponse{Error: "Invalid input"})
```

## Handlers

### Standard Handler Structure

```go
func HandlerName(c *gin.Context) {
    // 1. Bind input data (use domain-specific type)
    var input InputType
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, apitypes.ErrorResponse{Error: err.Error()})
        return
    }

    // 2. Call service function
    result, err := ServiceFunction(c.Request.Context(), input)

    // 3. Handle errors with appropriate HTTP status codes
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            c.JSON(http.StatusNotFound, apitypes.ErrorResponse{Error: "resource not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, apitypes.ErrorResponse{Error: "internal server error"})
        return
    }

    // 4. Return response
    c.JSON(http.StatusOK, result)
}
```

### Handler Naming Convention

```go
// RESTful naming examples:
func GetAccounts(c *gin.Context)        // GET collection
func GetMyInventory(c *gin.Context)     // GET user resource
func PostMyInventory(c *gin.Context)    // POST user resource
func PutMyPack(c *gin.Context)          // PUT user resource
func DeleteMyPack(c *gin.Context)       // DELETE user resource
```

### Pattern: Handler → Business Function

```go
// Public handler (Gin-specific)
func GetXxx(c *gin.Context) {
    result, err := returnXxx(c.Request.Context())
    if err != nil {
        // Handle error
        return
    }
    c.IndentedJSON(http.StatusOK, result)
}

// Private business function (framework-agnostic)
func returnXxx(ctx context.Context) (*Type, error) {
    // Business logic here
    return result, nil
}
```

## Business Functions

### Naming Pattern

```go
// Service layer: Public functions (Uppercase)
// Read operations: use "Get" prefix
func GetAllInventories(ctx context.Context, userID uint) (Inventories, error)
func GetPackByID(ctx context.Context, packID uint) (*Pack, error)

// Write operations: use descriptive verbs
func CreateInventoryItem(ctx context.Context, userID uint, item *Inventory) error
func UpdatePackContent(ctx context.Context, content *PackContent) error
func DeletePackByID(ctx context.Context, packID uint) error

// Repository layer: Private functions (lowercase)
// Read operations: prefix with "get" or "return"
func getAllInventories(ctx context.Context, db *sql.DB, userID uint) (Inventories, error)
func getPackByID(ctx context.Context, db *sql.DB, packID uint) (*Pack, error)

// Write operations: descriptive verbs
func createInventoryItem(ctx context.Context, db *sql.DB, item *Inventory) error
func updatePackContent(ctx context.Context, db *sql.DB, content *PackContent) error
func deletePackByID(ctx context.Context, db *sql.DB, packID uint) error
```

### Context Propagation Pattern

```go
// Service layer: Always accept context.Context as first parameter
func GetInventoryItems(ctx context.Context, userID uint) (Inventories, error) {
    db := database.DB()
    return getInventoryItems(ctx, db, userID)
}

// Repository layer: Accept context and db connection
func getInventoryItems(ctx context.Context, db *sql.DB, userID uint) (Inventories, error) {
    query := `SELECT id, item_name, category, weight FROM inventory WHERE user_id = $1`

    rows, err := db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to query inventory: %w", err)
    }
    defer rows.Close()

    var items Inventories
    for rows.Next() {
        var item Inventory
        if err := rows.Scan(&item.ID, &item.ItemName, &item.Category, &item.Weight); err != nil {
            return nil, fmt.Errorf("failed to scan item: %w", err)
        }
        items = append(items, item)
    }

    return items, nil
}
```

### Sentinel Errors

```go
// Define package-level sentinel errors
var (
    ErrNoAccountFound   = errors.New("no account found")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrPackNotFound     = errors.New("pack not found")
    ErrUnauthorized     = errors.New("unauthorized access")
)

// Use in service functions
// pkg/accounts/accounts.go
func GetAccountByID(ctx context.Context, accountID uint) (*Account, error) {
    db := database.DB()
    return getAccountByID(ctx, db, accountID)
}

func getAccountByID(ctx context.Context, db *sql.DB, accountID uint) (*Account, error) {
    query := `SELECT id, username, email, firstname, lastname FROM account WHERE id = $1`

    var account Account
    err := db.QueryRowContext(ctx, query, accountID).Scan(
        &account.ID,
        &account.Username,
        &account.Email,
        &account.Firstname,
        &account.Lastname,
    )
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrNoAccountFound
        }
        return nil, fmt.Errorf("failed to fetch account: %w", err)
    }
    return &account, nil
}

// Check in handlers
if err != nil {
    if errors.Is(err, ErrNoAccountFound) {
        c.JSON(http.StatusNotFound, apitypes.ErrorResponse{Error: "account not found"})
        return
    }
    c.JSON(http.StatusInternalServerError, apitypes.ErrorResponse{Error: "internal server error"})
    return
}
```

## Testing

### TestMain Pattern

```go
func TestMain(m *testing.M) {
    // 1. Initialize configuration
    config.EnvInit("../../.env")

    // 2. Initialize database
    database.Initialization()

    // 3. Run migrations
    database.Migrate()

    // 4. Load test datasets
    loadingDataset()

    // 5. Run tests
    ret := m.Run()
    os.Exit(ret)
}
```

### Test Structure Pattern

```go
func TestHandlerName(t *testing.T) {
    // Gin in test mode
    gin.SetMode(gin.TestMode)

    // Create a test router
    router := gin.Default()
    router.GET("/endpoint", HandlerToTest)

    t.Run("Success case", func(t *testing.T) {
        // Arrange
        req, _ := http.NewRequest(http.MethodGet, "/endpoint", nil)
        w := httptest.NewRecorder()

        // Act
        router.ServeHTTP(w, req)

        // Assert
        if w.Code != http.StatusOK {
            t.Errorf("Expected %d but got %d", http.StatusOK, w.Code)
        }

        var response dataset.Response
        if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
            t.Errorf("Failed to parse response: %v", err)
        }
    })

    t.Run("Error case", func(t *testing.T) {
        // Test error scenarios
    })
}
```

### Helper Function Pattern (Complexity Management)

```go
// Main test function
func TestPackSharing(t *testing.T) {
    gin.SetMode(gin.TestMode)
    router := setupRouter()

    t.Run("Share pack successfully", func(t *testing.T) {
        testShareSuccess(t, router, packID, userID)
    })

    t.Run("Share idempotent behavior", func(t *testing.T) {
        testShareIdempotent(t, router, packID, userID)
    })

    t.Run("Unshare pack successfully", func(t *testing.T) {
        testUnshareSuccess(t, router, packID, userID)
    })
}

// Helper functions reduce complexity
func testShareSuccess(t *testing.T, router *gin.Engine, packID, userID uint) {
    // Arrange
    req := createShareRequest(packID)
    w := httptest.NewRecorder()

    // Act
    router.ServeHTTP(w, req)

    // Assert
    assertStatusCode(t, w, http.StatusOK)
    assertSharingCodeExists(t, w)
}

func testShareIdempotent(t *testing.T, router *gin.Engine, packID, userID uint) {
    // Share twice, should return same result
    code1 := shareAndGetCode(t, router, packID)
    code2 := shareAndGetCode(t, router, packID)

    if code1 != code2 {
        t.Errorf("Idempotency violation: got different codes %s and %s", code1, code2)
    }
}
```

### Test Data Helpers

```go
// testdata.go
var (
    TestAccountID uint
    TestInventoryID uint
    TestPackID uint
)

func loadingDataset() {
    // Create test account
    TestAccountID = createTestAccount()

    // Create test inventory
    TestInventoryID = createTestInventory(TestAccountID)

    // Create test pack
    TestPackID = createTestPack(TestAccountID)
}

// Helper for pointer string comparison (when using *string for nullable fields)
func ComparePtrString(t *testing.T, expected, actual *string, fieldName string) {
    if expected == nil && actual == nil {
        return
    }
    if expected == nil || actual == nil {
        t.Errorf("%s mismatch: expected %v, got %v", fieldName, expected, actual)
        return
    }
    if *expected != *actual {
        t.Errorf("%s mismatch: expected %s, got %s", fieldName, *expected, *actual)
    }
}
```

## Error Handling

### Error Wrapping Pattern

```go
// Always wrap errors with context
func processData(ctx context.Context, id uint) error {
    data, err := fetchData(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to fetch data for id %d: %w", id, err)
    }

    if err := validateData(data); err != nil {
        return fmt.Errorf("data validation failed: %w", err)
    }

    return nil
}
```

### Handler Error Response Pattern

```go
func HandlerWithErrors(c *gin.Context) {
    result, err := businessFunction(c.Request.Context())

    if err != nil {
        // Check for specific errors first (most specific to least specific)
        if errors.Is(err, ErrNotFound) {
            c.IndentedJSON(http.StatusNotFound, gin.H{"error": "resource not found"})
            return
        }
        if errors.Is(err, ErrUnauthorized) {
            c.IndentedJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
            return
        }
        if errors.Is(err, sql.ErrNoRows) {
            c.IndentedJSON(http.StatusNotFound, gin.H{"error": "no data found"})
            return
        }

        // Generic error (don't expose internal details)
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
        return
    }

    c.IndentedJSON(http.StatusOK, result)
}
```

## Database Operations

### Query Pattern with Context

```go
func returnItems(ctx context.Context, userID uint) (*dataset.Items, error) {
    query := `
        SELECT id, name, description, created_at, updated_at
        FROM items
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

    rows, err := database.DB.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to query items: %w", err)
    }
    defer rows.Close()

    var items dataset.Items
    for rows.Next() {
        var item dataset.Item
        if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan item: %w", err)
        }
        items = append(items, item)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("rows iteration error: %w", err)
    }

    return &items, nil
}
```

### Single Row Query Pattern

```go
func returnItemByID(ctx context.Context, itemID uint) (*dataset.Item, error) {
    query := `
        SELECT id, name, description, created_at, updated_at
        FROM items
        WHERE id = $1
    `

    var item dataset.Item
    err := database.DB.QueryRowContext(ctx, query, itemID).Scan(
        &item.ID,
        &item.Name,
        &item.Description,
        &item.CreatedAt,
        &item.UpdatedAt,
    )

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrItemNotFound
        }
        return nil, fmt.Errorf("failed to fetch item: %w", err)
    }

    return &item, nil
}
```

### Exec Pattern (INSERT/UPDATE/DELETE)

```go
func createItem(ctx context.Context, item dataset.ItemInput, userID uint) error {
    query := `
        INSERT INTO items (user_id, name, description, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `

    now := time.Now().Truncate(time.Second)
    _, err := database.DB.ExecContext(ctx, query, userID, item.Name, item.Description, now, now)
    if err != nil {
        return fmt.Errorf("failed to create item: %w", err)
    }

    return nil
}

func updateItem(ctx context.Context, itemID uint, updates dataset.ItemInput) error {
    query := `
        UPDATE items
        SET name = $1, description = $2, updated_at = $3
        WHERE id = $4
    `

    result, err := database.DB.ExecContext(ctx, query, updates.Name, updates.Description, time.Now().Truncate(time.Second), itemID)
    if err != nil {
        return fmt.Errorf("failed to update item: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rows == 0 {
        return ErrItemNotFound
    }

    return nil
}
```

### Transaction Pattern

```go
func complexOperation(ctx context.Context, data dataset.ComplexInput) error {
    // Begin transaction
    tx, err := database.DB.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }

    // Ensure rollback on error
    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()

    // First operation
    _, err = tx.ExecContext(ctx, query1, params1...)
    if err != nil {
        return fmt.Errorf("first operation failed: %w", err)
    }

    // Second operation
    _, err = tx.ExecContext(ctx, query2, params2...)
    if err != nil {
        return fmt.Errorf("second operation failed: %w", err)
    }

    // Commit transaction
    if err = tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}
```

### Nullable Fields Pattern

```go
// Use pointer types for nullable fields
type Pack struct {
    ID          uint      `json:"id"`
    Name        string    `json:"name"`
    SharingCode *string   `json:"sharing_code,omitempty"` // nullable
    CreatedAt   time.Time `json:"created_at"`
}

// Query with nullable fields
func returnPack(ctx context.Context, packID uint) (*dataset.Pack, error) {
    query := `SELECT id, name, sharing_code, created_at FROM packs WHERE id = $1`

    var pack dataset.Pack
    var sharingCode sql.NullString

    err := database.DB.QueryRowContext(ctx, query, packID).Scan(
        &pack.ID,
        &pack.Name,
        &sharingCode,
        &pack.CreatedAt,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to fetch pack: %w", err)
    }

    // Convert sql.NullString to *string
    if sharingCode.Valid {
        pack.SharingCode = &sharingCode.String
    }

    return &pack, nil
}
```

## Swagger Documentation

### Complete Handler Documentation

```go
// @Summary Get user inventories
// @Description Retrieves all inventory items for the authenticated user
// @Security Bearer
// @Tags Inventories
// @Accept json
// @Produce json
// @Success 200 {object} inventories.Inventories "List of inventories"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 500 {object} apitypes.ErrorResponse "Internal server error"
// @Router /api/v1/inventories [get]
func GetMyInventories(c *gin.Context) {
    // Implementation
}

// @Summary Create inventory item
// @Description Creates a new inventory item for the authenticated user
// @Security Bearer
// @Tags Inventories
// @Accept json
// @Produce json
// @Param inventory body inventories.Inventory true "Inventory data"
// @Success 201 {object} inventories.Inventory "Created inventory"
// @Failure 400 {object} apitypes.ErrorResponse "Bad request"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Failure 500 {object} apitypes.ErrorResponse "Internal server error"
// @Router /api/v1/inventories [post]
func PostMyInventory(c *gin.Context) {
    // Implementation
}
```

### Public vs Protected Routes

```go
// Public route (no authentication)
// @Summary User login
// @Description Authenticate user and receive JWT token
// @Tags Public
// @Accept json
// @Produce json
// @Param credentials body accounts.LoginInput true "Login credentials"
// @Success 200 {object} accounts.Token "JWT token"
// @Failure 400 {object} apitypes.ErrorResponse "Bad request"
// @Failure 401 {object} apitypes.ErrorResponse "Invalid credentials"
// @Router /api/login [post]
func Login(c *gin.Context) {
    // Implementation
}

// Protected route (requires JWT)
// @Summary Get user profile
// @Description Get authenticated user's profile
// @Security Bearer
// @Tags Internal
// @Produce json
// @Success 200 {object} accounts.Account "User profile"
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
// @Router /api/v1/profile [get]
func GetMyProfile(c *gin.Context) {
    // Implementation
}
```

### Parameter Types

```go
// Path parameter
// @Param id path int true "Pack ID"

// Query parameter
// @Param search query string false "Search term"

// Body parameter (use domain-specific type with full package path)
// @Param pack body packs.Pack true "Pack data"

// Header parameter (usually for custom headers, JWT is handled by @Security)
// @Param X-Custom-Header header string false "Custom header"
```

### Type References in Swagger

When referencing types in Swagger comments:

- **Same package**: Use just the type name (swaggo can resolve it automatically)
- **Different package**: Use the full package path (e.g., `apitypes.ErrorResponse`)

```go
// In pkg/packs/packs.go - types from same package, no prefix needed
// @Success 200 {object} Pack "Single pack"
// @Success 200 {object} Packs "List of packs"

// Types from different packages need full path
// @Success 200 {object} apitypes.OkResponse "Success response"
// @Failure 400 {object} apitypes.ErrorResponse "Error response"
```

```go
// In pkg/accounts/accounts.go - same package types
// @Success 200 {object} Account "User account"
// @Success 200 {object} Token "Authentication token"

// Cross-package references
// @Failure 401 {object} apitypes.ErrorResponse "Unauthorized"
```

## Recurring Patterns Summary

### Context Pattern

```go
// Always pass context as first parameter
func businessFunction(ctx context.Context, params ...) error {
    // Use ctx for DB and HTTP operations
}
```

### Repository Pattern

```go
// Public handler
func GetXxx(c *gin.Context) { }

// Private business function
func returnXxx(ctx context.Context) (*Type, error) { }
```

### Error Handling Pattern

```go
if err != nil {
    if errors.Is(err, SpecificError) {
        // Specific handling
        return
    }
    // Generic handling
    return
}
```

### Handler Pattern

```go
func Handler(c *gin.Context) {
    // Bind → Validate → Execute → Respond
}
```
