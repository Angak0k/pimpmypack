# Design and Architecture Principles - PimpMyPack

This document outlines the main design and architecture principles to contribute in the PimpMyPack project.

> **For AI Agents**: See [claude.md](claude.md) for quick reference (automatically loaded by Claude Code). This file provides comprehensive details.
>
> **Related Documentation**:
>
> - [claude.md](claude.md) - Quick reference for Claude Code (stack, conventions, critical info)
> - [Code Patterns & Examples](agents/PATTERNS.md) - Detailed code patterns and templates
> - [Collaboration Guide](agents/COLLABORATION.md) - Spec-driven development and learnings
> - [Code Templates](agents/templates/) - Ready-to-use code templates

## 🎯 Purpose of the project

The PimpMyPack project is a backend service for managing user accounts, inventories, and packs. It provides a RESTful API built with Go and Gin, using PostgreSQL as the database. The project emphasizes clean architecture, code quality, security, and maintainability.

In a first iteration that projects handles only user created items, future versions may include editor curated public items shared between users and collaborative packs.

## 🧑‍💻 Contribution Workflow

If you are an agent willing to contribute to this project, follow this workflow:

1. **Write specs**: Create clear specifications in `specs/*.md`
2. **Discuss**: Validate specs with the project owner
3. **Design**: Design the solution, create a task plan, amend the specs file with design decisions
4. **Validate**: Get design approval before implementation
5. **All-in mode**: Confirm if you can run all tasks at once or need per-task validation
6. **Implement**: Code following guidelines, write tests, document
7. **Update status**: Systematically update task checkboxes in specs with implementation dates, files modified, and key decisions

See [COLLABORATION.md](agents/COLLABORATION.md) for detailed best practices.

## 🏗️ Architecture Overview

### Project Structure

- **Functional package organization**: By business domain (`accounts`, `inventories`, `packs`, `security`, `config`, `database`, `helper`)
- **Separation of concerns**: Clear package responsibilities
- **Repository Pattern**: Gin handlers separated from business logic

### Naming Conventions

- **Files**: `snake_case` for SQL (`000001_account.up.sql`)
- **Packages**: Plural names for domains (`accounts`, `inventories`)
- **Public functions**: Uppercase start (`GetAccounts`, `PostMyInventory`)
- **Private functions**: Lowercase start (`returnInventories`, `registerUser`)

## 📦 Code Organization

### HTTP Handlers (Gin)

- **Naming**: `GetXxx`, `PostXxx`, `PutXxx`, `DeleteXxx` (add `My` for user resources)
- **Structure**: Bind → Validate → Execute → Respond
- **Documentation**: Complete Swagger annotations required

See [PATTERNS.md#handlers](agents/PATTERNS.md#handlers) for detailed examples.

### Business Functions

- **Naming**: Prefix `return` for read functions (`returnInventories`)
- **Context**: Always accept `context.Context` as first parameter
- **Errors**: Use `fmt.Errorf` with wrapping (`%w`), define sentinel errors

See [PATTERNS.md#business-functions](agents/PATTERNS.md#business-functions) for patterns.

### Database Management

- **Singleton**: DB connection via `database` package
- **Context-aware**: Use `QueryContext`, `ExecContext`, `QueryRowContext`
- **Cleanup**: Always `defer rows.Close()`
- **Errors**: Distinguish `sql.ErrNoRows` from other errors

### Security

- **JWT**: Generation, extraction (header/query), validation via middleware
- **Middlewares**: `JwtAuthProcessor()` (standard), `JwtAuthAdminProcessor()` (admin)
- **Passwords**: bcrypt for hashing and validation
- **Routes**: `/api` (public), `/api/v1` (protected), `/api/admin` (admin)

## 🧪 Testing

### Organization

- Test files: `*_test.go` in same package
- Test data: `testdata.go` for reusable data
- `TestMain`: Global setup (config, DB, migrations, datasets)

### Best Practices

- **Isolation**: Independent tests
- **Random data**: Use `random.UniqueId()` to avoid conflicts
- **Gin test mode**: `gin.SetMode(gin.TestMode)`
- **Coverage**: Run with `-coverprofile=coverage.out`
- **Complexity**: Extract helpers when cognitive complexity > 20

See [PATTERNS.md#testing](agents/PATTERNS.md#testing) for examples.

## ⚙️ Configuration

- **Environment**: `.env` with system variable fallback
- **Global access**: Via `config` package
- **Modes**:
  - `LOCAL`/`DEV`: Swagger enabled
  - Production: Swagger disabled, Gin release mode

## 📄 Dataset (Data Models)

All types in `pkg/dataset`:

- Base types: `Account`, `Inventory`, `Pack`, `PackContent`
- Collections: `Accounts`, `Inventories`, `Packs`
- Input/Response types: `RegisterInput`, `OkResponse`, `ErrorResponse`
- Composite types: `PackContentWithItem` (joins)
- **Timestamps**: `created_at`, `updated_at` (truncated to second)

## 🚀 Build and Deployment

### Makefile Targets

- `start-db`: Start PostgreSQL in Docker
- `stop-db`: Stop container
- `clean-db`: Remove container
- `test`: Run tests (manages DB lifecycle)
- `api-doc`: Generate Swagger docs
- `build`: Build after tests
- `lint`: golangci-lint analysis

### Docker

- Base: Alpine (lightweight)
- Multi-stage build recommended
- Direct entrypoint (no shell)

## 🛡️ Code Quality

### Linting

- **50+ linters** enabled via golangci-lint
- Key categories: Security, Bugs, Style, Performance, Tests
- Targeted `//nolint:lintername` only when justified

### Error Handling

- Wrap errors: `fmt.Errorf("context: %w", err)`
- Compare: `errors.Is()` for sentinel errors
- HTTP Status mapping:
  - 400: Bad Request
  - 401: Unauthorized
  - 403: Forbidden
  - 404: Not Found
  - 500: Internal Server Error

## 🔐 Security Best Practices

1. **Never hardcode secrets** - use environment variables
2. **Never return passwords** in API responses
3. **JWT**: Strong secret, limited lifetime
4. **SQL Injection**: Always use parameterized queries (`$1`, `$2`)
5. **CORS**: Appropriate production config
6. **Rate Limiting**: Needed for public endpoints

## 📐 Design Principles

### SOLID

- **Single Responsibility**: Each package/function has one job
- **Open/Closed**: Extensibility via interfaces
- **Dependency Inversion**: High-level independent of details

### DRY & Clean Code

- Centralize repetitive operations (helpers, middleware, testdata)
- Explicit names, short functions, early returns
- Comment the "why", not the "what"

## 📊 Database Migrations

- **Format**: `NNNNNN_description.up.sql` / `.down.sql`
- **Principles**: Idempotent, rollback support, sequential versioning
- **Embedded**: Via `//go:embed` in binary
- **Atomic**: Each migration is a coherent unit

## 📝 New Feature Checklist

- [ ] Create types in `pkg/dataset`
- [ ] Create SQL migration (up and down)
- [ ] Implement business functions with context
- [ ] Define sentinel errors if necessary
- [ ] Create Gin handlers with Swagger docs
- [ ] Add routes in `main.go` (public/protected/admin)
- [ ] Create unit tests with testdata
- [ ] Run `make lint`
- [ ] Run `make test` with coverage
- [ ] Update documentation if necessary

## 📚 Documentation

### Required Sections

- README: Project description, local setup, commands, API docs link
- Swagger: Complete annotations, organized by tags
- Code: Public functions commented, sentinel errors documented

### Generation

- Swagger: `swag init --tags !Internal`
- Access: `/swagger/*any` in DEV/LOCAL only

## 🎯 Important Points

1. **Context timeout**: Implement for DB queries
2. **Connection pooling**: Properly configure SQL pool
3. **Graceful shutdown**: Handle application shutdown
4. **Observability**: Structured logs and metrics
5. **Input validation**: Validate all user data
6. **Integration tests**: End-to-end test coverage
