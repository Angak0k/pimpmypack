# Design and Architecture Principles - PimpMyPack

This document outlines the main design and architecture principles to contribute in the PimpMyPack project.

## 🎯 Purpose of the project

The PimpMyPack project is a backend service for managing user accounts, inventories, and packs. It provides a RESTful API built with Go and Gin, using PostgreSQL as the database. The project emphasizes clean architecture, code quality, security, and maintainability.

In a first iteration that projects handles only user created items, future versions may include editor curated public items shared between users and collaborative packs.

## 🧑‍💻 Contribution principles

If you are an agent willing to contribute to this project, please follow these steps:

- **write specs**: write  clear specifications in a md file in the specs/ folder
- **discuss**: discuss the specs with the project owner to validate them
- **design**: design the solution following the guidelines in this document and orchestrate a plan of work through structured tasks. Amend the specs file with your design decisions and the plan of work (with checkboxes to follow-it easily)
- **validate**: validate the design with the project owner before starting the implementation.
- **all-in mode**: ask to the project owner if you could run all tasks in a single go or if you should ask for a validation for each task.
- **implement**: implement the feature following the guidelines in this document, write tests and document the code.

## 🏗️ General Architecture

### Project Structure

- **Functional package organization**: Code is organized by business domain (`accounts`, `inventories`, `packs`, `security`, `config`, `database`, `helper`)
- **Separation of concerns**: Each package has a clear and delimited responsibility
- **Repository Pattern**: Gin handler functions are separated from business functions that interact with the database

### Naming Conventions

- **Files**: snake_case names for SQL files (`000001_account.up.sql`)
- **Packages**: Plural names for business domains (`accounts`, `inventories`, `packs`)
- **Public functions**: Start with an uppercase letter (e.g., `GetAccounts`, `PostMyInventory`)
- **Private functions**: Start with a lowercase letter (e.g., `returnInventories`, `registerUser`)

## 📦 Code Organization

### HTTP Handlers (Gin)

1. **RESTful Naming**:
   - GET: `GetXxx`, `GetMyXxx` (for user resources)
   - POST: `PostXxx`, `PostMyXxx`
   - PUT: `PutXxx`, `PutMyXxx`
   - DELETE: `DeleteXxx`, `DeleteMyXxx`

2. **Typical handler structure**:

```go
func HandlerName(c *gin.Context) {
    // 1. Bind input data
    var input dataset.Input
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 2. Call business function
    result, err := businessFunction(c.Request.Context(), input)

    // 3. Handle errors with appropriate HTTP status codes
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            c.IndentedJSON(http.StatusNotFound, gin.H{"error": "message"})
            return
        }
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 4. Return response
    c.IndentedJSON(http.StatusOK, result)
}
```

3. **Swagger Documentation**: Each public handler must be documented with Swagger annotations:

```go
// @Summary Short description
// @Description Detailed description
// @Security Bearer (if authentication required)
// @Tags Tag_name
// @Accept  json
// @Produce  json
// @Param   param_name  path/body/query  type  true  "Description"
// @Success 200 {object} dataset.Type
// @Failure 400 {object} dataset.ErrorResponse
// @Router /endpoint [method]
```

### Business Functions

1. **Naming**: Prefix with `return` for read functions (e.g., `returnInventories`, `returnPacks`)
2. **Context propagation**: All business functions accept a `context.Context` as the first parameter
3. **Error handling**: Use `fmt.Errorf` with wrapping (`%w`) to preserve the error chain
4. **Sentinel errors**: Define named errors for important business cases:

```go
var ErrNoAccountFound = errors.New("no account found")
```

### Database Management

1. **DB Singleton**: Database connection is managed via a singleton in the `database` package
2. **Context-aware**: Use `QueryContext`, `ExecContext`, `QueryRowContext` for all queries
3. **Defer Close**: Always close `rows` with `defer rows.Close()`
4. **Error handling**: Distinguish `sql.ErrNoRows` from other errors
5. **Transactions**: Pattern for complex operations requiring multiple queries
6. **Migration**: Use `golang-migrate` with versioned and embedded SQL files

### Security

1. **JWT**:
   - Generation with configurable lifetime
   - Extraction from `Authorization` header or `token` query param
   - Validation via Gin middleware

2. **Middlewares**:
   - `JwtAuthProcessor()`: Standard authentication
   - `JwtAuthAdminProcessor()`: Admin authentication only

3. **Passwords**:
   - Hashing with `bcrypt.GenerateFromPassword`
   - Validation with `bcrypt.CompareHashAndPassword`

4. **Routes**:
   - `/api`: Public routes (register, login, etc.)
   - `/api/v1`: Protected routes (authentication required)
   - `/api/admin`: Admin routes (admin authentication required)

## 🧪 Tests

### Test Organization

1. **Test files**: `*_test.go` in the same package as the tested code
2. **Testdata files**: `testdata.go` for reusable test data
3. **TestMain**: Global initialization (config, DB, migrations, datasets)

### Test Structure

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

func TestFunctionName(t *testing.T) {
    // Gin in test mode
    gin.SetMode(gin.TestMode)
    
    // Create a test router
    router := gin.Default()
    router.GET("/endpoint", HandlerToTest)
    
    t.Run("Test Case Description", func(t *testing.T) {
        // Arrange
        req, _ := http.NewRequest(http.MethodGet, "/endpoint", nil)
        w := httptest.NewRecorder()
        
        // Act
        router.ServeHTTP(w, req)
        
        // Assert
        if w.Code != http.StatusOK {
            t.Errorf("Expected %d but got %d", http.StatusOK, w.Code)
        }
    })
}
```

### Test Best Practices

1. **Isolation**: Each test must be independent
2. **Test data**: Use random data (`random.UniqueId()`) to avoid conflicts
3. **Gin test mode**: Always use `gin.SetMode(gin.TestMode)`
4. **httptest**: Use `httptest.NewRecorder()` to capture responses
5. **Coverage**: Run tests with coverage (`go test -coverprofile=coverage.out`)
6. **Race detector**: Use `-race` to detect data races
7. **Sequential execution**: Use `-p=1` for sequential execution if necessary
8. **Test database**: Use a dedicated Docker container for tests

## ⚙️ Configuration

### Environment

1. **Environment variables**: Configuration via `.env` with fallback to system variables
2. **Global variables**: Exposed in the `config` package for simplified access
3. **Initialization**: `config.EnvInit(".env")` at application startup
4. **Typed types**: Convert environment variables to appropriate types (int, bool, etc.)

### Execution Modes

- `LOCAL`: Local development (Swagger enabled, bind to localhost)
- `DEV`: Development (Swagger enabled)
- Others: Production (Swagger disabled, Gin release mode)

## 🔧 Helpers and Utilities

### Helper Package

1. **Conversion functions**: `StringToUint`, `ConvertWeightUnit`
2. **Search functions**: `FindUserIDByUsername`, `FindPackIDByPackName`
3. **Validation**: `IsValidEmail` with regex
4. **Generation**: `GenerateRandomCode` for tokens/codes
5. **Email**: Email sending functions (SMTP)

## 📄 Dataset (Data Models)

### Organization

1. **Dedicated package**: `pkg/dataset` contains all data types
2. **Separation of concerns**:
   - Base types: `Account`, `Inventory`, `Pack`, `PackContent`
   - Collections: `Accounts`, `Inventories`, `Packs`, `PackContents`
   - Input types: `RegisterInput`, `LoginInput`, etc.
   - Response types: `OkResponse`, `ErrorResponse`
   - Composite types: `PackContentWithItem` (joins)

3. **JSON annotations**: All exported fields have JSON tags

### Timestamps

- **Systematic fields**: `created_at` and `updated_at` on all entities
- **Type**: `time.Time` with truncation to the second
- **Management**: `time.Now().Truncate(time.Second)`

## 🚀 Build and Deployment

### Makefile

1. **Main targets**:
   - `start-db`: Starts PostgreSQL in Docker
   - `stop-db`: Stops the container
   - `clean-db`: Removes the container
   - `test`: Runs tests (with start-db and clean-db)
   - `api-doc`: Generates Swagger documentation
   - `build`: Builds after tests
   - `lint`: Code analysis with golangci-lint

2. **Test DB management**: The Makefile automatically manages the test DB lifecycle

### Dockerfile

1. **Base image**: Alpine (lightweight and secure)
2. **Update**: `apk update && apk upgrade --no-cache`
3. **Binary**: Copy of compiled binary only (multi-stage build recommended)
4. **Entrypoint**: Direct command without shell

### CI/CD

1. **GitHub Actions**: Workflows for CI, documentation and release
2. **Tests**: PostgreSQL as Docker service for tests
3. **Checks**: 
   - Verification of `go mod tidy`
   - Tests with coverage
   - Lint with golangci-lint

## 🛡️ Code Quality

### Linting (golangci-lint)

1. **Strict configuration**: More than 50 linters enabled
2. **Key linters**:
   - Security: `gosec`, `gosimple`
   - Bugs: `errcheck`, `govet`, `staticcheck`
   - Style: `revive`, `gocritic`
   - Performance: `perfsprint`
   - Tests: `testifylint`, `testableexamples`

3. **Targeted disables**: Use `//nolint:lintername` only when justified

### Error Handling

1. **Wrapping**: Always wrap errors with context: `fmt.Errorf("context: %w", err)`
2. **Errors.Is**: Use `errors.Is()` to compare with sentinel errors
3. **Logging**: Use `log.Fatalf` for critical errors at startup
4. **HTTP Status**: Properly map errors to HTTP status codes
   - 400: Bad Request (validation)
   - 401: Unauthorized (no token / invalid token)
   - 403: Forbidden (valid token but no rights)
   - 404: Not Found
   - 500: Internal Server Error

## 📚 Documentation

### README

1. **Required sections**:
   - Project description
   - Setup for local development
   - Basic commands (build, test, run)
   - Link to API documentation

### Swagger

1. **Configuration**: Annotations in main and handlers
2. **Generation**: `swag init --tags !Internal` (exclude admin routes)
3. **Access**: `/swagger/*any` in DEV/LOCAL mode only
4. **Tags**: Organize endpoints by domain (`Public`, `Internal`, etc.)

### Code Comments

1. **Public functions**: Must have a comment starting with the function name
2. **Packages**: No `doc.go` but inline documentation
3. **Errors**: Document sentinel errors
4. **Swagger**: Complete annotations for all public and protected routes

## 🔐 Security Best Practices

1. **Secrets**: Never hardcode secrets, always use environment variables
2. **Passwords**: 
   - Never return passwords in API responses
   - Use bcrypt with default cost
3. **JWT**: 
   - Strong secret via environment
   - Limited and configurable lifetime
4. **SQL Injection**: Always use parameterized queries (`$1`, `$2`, etc.)
5. **CORS**: Appropriate configuration for production environments
6. **Rate Limiting**: To implement for public endpoints (register, login)

## 📐 Design Principles

### SOLID

1. **Single Responsibility**: Each package/function has a single responsibility
2. **Open/Closed**: Extensibility via interfaces (especially for DB)
3. **Dependency Inversion**: High-level packages don't depend on details

### DRY (Don't Repeat Yourself)

1. **Helper functions**: Centralize repetitive operations
2. **Middleware**: Reuse cross-cutting logic (auth, logging)
3. **Testdata**: Centralize test data

### Clean Code

1. **Explicit names**: Variables and functions with clear names
2. **Short functions**: Limit cyclomatic complexity
3. **Early returns**: Favor early returns to reduce nesting
4. **Useful comments**: Comment the "why", not the "what"

## 🔄 Recurring Patterns

### Handler Pattern

```go
func Handler(c *gin.Context) {
    // Bind → Validate → Execute → Respond
}
```

### Repository Pattern

```go
// Public handler
func GetXxx(c *gin.Context) { }

// Private business function
func returnXxx(ctx context.Context) (*Type, error) { }
```

### Context Pattern

```go
// Always pass context as first parameter
func businessFunction(ctx context.Context, params ...) error {
    // Use ctx for DB and HTTP operations
}
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

## 📊 Database Migrations

### Naming Convention

- Format: `NNNNNN_description.up.sql` / `NNNNNN_description.down.sql`
- Example: `000001_account.up.sql`

### Principles

1. **Idempotence**: Migrations must be replayable
2. **Rollback**: Always provide a `.down.sql` migration
3. **Sequential versioning**: 6-digit numbers
4. **Embedded**: SQL files embedded in binary via `//go:embed`
5. **Atomic**: Each migration must be a coherent unit of work

## 🎯 Important Points

1. **Context timeout**: Implement timeouts for DB queries
2. **Connection pooling**: Properly configure SQL connection pool
3. **Graceful shutdown**: Properly handle application shutdown
4. **Observability**: Add structured logs and metrics
5. **Input validation**: Validate all user data
6. **Integration tests**: Complete with end-to-end tests

## 📝 New Feature Checklist

- [ ] Create types in `pkg/dataset`
- [ ] Create SQL migration (up and down)
- [ ] Implement business functions with context
- [ ] Define sentinel errors if necessary
- [ ] Create Gin handlers with Swagger documentation
- [ ] Add routes in `main.go` (public/protected/admin)
- [ ] Create unit tests with testdata
- [ ] Check lint (`make lint`)
- [ ] Test locally with coverage (`make test`)
- [ ] Update documentation if necessary
