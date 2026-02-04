# PimpMyPack - Context for Claude Code

> 📚 **Full Documentation**: See [agents.md](agents.md) for complete architecture guidelines, patterns, and collaboration workflow.

## 🚀 Quick Reference

**Purpose**: Backend API for managing user packs and gear inventories (hiking, camping, etc.)

**Stack**:
- **Go 1.21+** (Gin Web Framework)
- **PostgreSQL** with **direct SQL** (`database/sql`, **NO GORM**)
- **JWT authentication** (golang-jwt/jwt v5)
- **bcrypt** for password hashing
- **Testing**: stdlib testing, httptest, testify/assert

**⚠️ CRITICAL**: Always use `QueryContext`, `ExecContext`, `QueryRowContext` - **NEVER use GORM or any ORM**.

## 📁 Project Structure

```
pimpmypack/
├── main.go                 # Application entry point, routing
├── pkg/
│   ├── accounts/           # User accounts, login, registration
│   ├── inventories/        # User gear inventories
│   ├── packs/              # Pack management (user packs with items)
│   ├── security/           # JWT, auth middleware, password hashing
│   ├── config/             # Environment configuration loading
│   ├── database/           # DB connection singleton + migrations
│   │   └── migration/      # SQL migration files
│   └── helper/             # Utilities (email, validation, etc.)
├── specs/                  # Technical specifications (*.md)
├── agents/                 # Agent definitions for Claude Code
├── docs/                   # Documentation (patterns, templates, collaboration)
└── .env.sample             # Environment variables template
```

## 🔑 Key Conventions

### Database Access (CRITICAL)

**Always use direct SQL queries** - No ORM, no GORM:

```go
// ✅ CORRECT - Direct SQL with context
err := database.DB().QueryRowContext(ctx,
    `SELECT id, username, email FROM account WHERE id = $1`,
    userID,
).Scan(&user.ID, &user.Username, &user.Email)

// ✅ CORRECT - ExecContext for INSERT/UPDATE/DELETE
_, err := database.DB().ExecContext(ctx,
    `INSERT INTO account (username, email) VALUES ($1, $2)`,
    username, email,
)

// ❌ WRONG - Never use GORM or any ORM
// db.Where("id = ?", userID).First(&user)
```

**Required patterns**:
- Always pass `context.Context` as first parameter
- Always use `QueryContext`, `ExecContext`, `QueryRowContext` (never non-context versions)
- Always `defer rows.Close()` after `Query` or `QueryContext`
- Always distinguish `sql.ErrNoRows` from other errors

### Naming Conventions

**Packages**:
- Use **plural names** for business domains: `accounts`, `inventories`, `packs`
- Use **singular** for utilities: `config`, `helper`, `database`, `security`

**Functions**:
- **Public handlers** (Gin): `GetXxx`, `PostXxx`, `PutXxx`, `DeleteXxx`
  - Add `My` prefix for user-scoped resources: `GetMyInventory`, `PostMyPack`
- **Private business functions**: lowercase start (`registerUser`, `returnInventories`)
- **Read operations**: prefix with `return` (`returnInventories`, `returnPackByID`)

**Files**:
- Migration files: `snake_case` with sequence number (`000001_account.up.sql`)
- Go files: `snake_case` (`accounts.go`, `security_test.go`)

### HTTP Handlers Structure

All Gin handlers follow this pattern:

```go
func PostMyResource(c *gin.Context) {
    var input ResourceInput

    // 1. Bind JSON
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 2. Validate (if needed)
    if !isValid(input) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "validation error"})
        return
    }

    // 3. Execute business logic
    result, err := createResource(c.Request.Context(), input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 4. Respond
    c.JSON(http.StatusOK, gin.H{"data": result})
}
```

### Security & Authentication

**Routes**:
- `/api` - Public endpoints (no auth)
- `/api/v1` - Protected endpoints (JWT required via `JwtAuthProcessor()`)
- `/api/admin` - Admin-only endpoints (JWT + admin role via `JwtAuthAdminProcessor()`)

**Middleware**:
- `security.JwtAuthProcessor()` - Standard JWT validation
- `security.JwtAuthAdminProcessor()` - JWT validation + admin role check

**Token extraction**:
- Header: `Authorization: Bearer <token>`
- Query param: `?token=<token>` (legacy, avoid in new code)

### Testing

**Organization**:
- Test files: `*_test.go` in same package as tested code
- Use `httptest` for HTTP handlers
- Use `testify/assert` for assertions

**Example**:
```go
func TestGetMyInventory_Success(t *testing.T) {
    router := setupTestRouter(t)
    user := createTestUser(t)
    token := generateTestToken(user.ID)

    req := httptest.NewRequest("GET", "/api/v1/my/inventory", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)
}
```

## 📋 Before You Start Coding

**Always do these steps first**:

1. ✅ **Read existing code** in the target package to understand patterns
2. ✅ **Check [agents.md](agents.md)** for detailed architecture guidelines
3. ✅ **Check [docs/PATTERNS.md](docs/PATTERNS.md)** for code examples
4. ✅ **Write specs** in `specs/*.md` before implementation (for non-trivial features)
5. ✅ **Use SQL direct queries** - verify examples in `pkg/accounts/accounts.go`

## ⚠️ Critical Workflow: Before Pushing to GitHub

**MANDATORY steps before every push**:

1. ✅ **Run full test suite**: `make test`

   - All tests must pass
   - No race conditions detected (`-race` flag enabled)

2. ✅ **ALWAYS run linter**: `make lint`

   - **CRITICAL**: Fix ALL linter issues before pushing
   - This catches common mistakes (wrong assertions, code quality issues)
   - CI will reject PRs with linter violations

3. ✅ **Run API tests** (non-regression): `make api-test`

   - Verifies all API endpoints still work correctly
   - Tests authentication, CRUD operations, file uploads
   - **ESPECIALLY important** after changes to handlers, middleware, or database schema

4. ✅ **Verify build**: `go build ./...`

   - All packages must compile without errors

**Why this matters**: The linter catches subtle bugs that tests might miss (e.g., using `assert.Equal` for pointer comparison instead of `assert.Same`). API tests catch regressions in endpoint behavior. Always run both locally before pushing to avoid CI failures and wasted review cycles.

## 🤖 Available Agents

The project includes specialized agents for common tasks. Invoke them explicitly or Claude will suggest them based on context.

### api-test-runner

**Purpose**: Run automated API test scenarios for non-regression testing

**When to use**:

- After modifying API handlers or middleware
- After database schema changes or migrations
- After authentication/security updates
- Before pushing code that affects API behavior
- When explicitly requested: "run api tests" or "check for regressions"

**Quick commands**:

```bash
# Build the test CLI
make build-apitest

# Run all test scenarios (35 tests across 4 scenarios)
make api-test

# Run specific scenario
./bin/apitest run 001    # Authentication & registration
./bin/apitest run 002    # Pack CRUD operations
./bin/apitest run 003    # Inventory management
./bin/apitest run 004    # CSV import (LighterPack)

# Run with verbose output for debugging
./bin/apitest run -v 001
```

**What it tests**:

- User registration and email confirmation
- Login and token refresh flows
- Pack creation, update, deletion
- Inventory CRUD operations
- File upload (CSV import)
- Authorization and access control

**Note**: Server must be running on `localhost:8080` with `STAGE=LOCAL` in `.env`

## ❌ Common Mistakes to Avoid

**Critical Errors**:
- ❌ **NO GORM**: Project uses `database/sql` directly, never use any ORM
- ❌ **NO auto-migrations**: Always write migration files in `pkg/database/migration/`
- ❌ **NO `go.mod` changes**: Unless explicitly requested (dependencies are stable)
- ❌ **NO breaking changes**: Always maintain backward compatibility

**Code Quality**:
- ❌ Don't use non-context DB methods (`Query`, `Exec`, `QueryRow`)
- ❌ Don't forget `defer rows.Close()`
- ❌ Don't ignore `sql.ErrNoRows` distinction
- ❌ Don't create new packages without discussing architecture

**Security**:
- ❌ Don't log sensitive data (passwords, tokens)
- ❌ Don't store passwords in plain text (always bcrypt)
- ❌ Don't skip JWT validation on protected routes

## 🔍 Quick Code Reference

**Database connection**:
```go
database.DB() // Returns *sql.DB singleton
```

**Generate JWT token**:
```go
token, err := security.GenerateToken(userID)
```

**Hash password**:
```go
hashedPassword, err := security.HashPassword(password)
```

**Verify password**:
```go
err := security.VerifyPassword(password, hashedPassword)
```

**Load config**:
```go
config.APISecret          // JWT secret
config.TokenLifespan      // Token duration (hours)
config.DBHost, config.DBPort, etc.
```

## 📖 Additional Resources

- **[agents.md](agents.md)** - Complete architecture & contribution guidelines
- **[docs/PATTERNS.md](docs/PATTERNS.md)** - Detailed code patterns & templates
- **[docs/COLLABORATION.md](docs/COLLABORATION.md)** - Spec-driven development workflow
- **[specs/](specs/)** - Technical specifications for features
- **[.env.sample](.env.sample)** - Environment variables documentation

## 💡 Example: Read User from Database

```go
func getUserByID(ctx context.Context, userID uint) (*User, error) {
    var user User

    err := database.DB().QueryRowContext(ctx,
        `SELECT id, username, email, firstname, lastname, role, status,
                created_at, updated_at
         FROM account
         WHERE id = $1`,
        userID,
    ).Scan(
        &user.ID,
        &user.Username,
        &user.Email,
        &user.Firstname,
        &user.Lastname,
        &user.Role,
        &user.Status,
        &user.CreatedAt,
        &user.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user not found")
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    return &user, nil
}
```

---

**Last Updated**: 2026-02-02
**Project Version**: See [go.mod](go.mod) for Go version and dependencies
