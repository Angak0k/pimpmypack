# Code Templates - PimpMyPack

This directory contains ready-to-use code templates for common patterns in the PimpMyPack project.

## Available Templates

### Go Code Templates

1. **handler.go.template** - Standard Gin HTTP handler with Swagger annotations
2. **business_function.go.template** - Business logic functions (CRUD operations)
3. **test.go.template** - Test file structure with TestMain and helpers

### SQL Templates

4. **migration.up.sql.template** - Database migration (forward)
5. **migration.down.sql.template** - Database migration (rollback)

## Usage

### 1. Copy the Template

```bash
cp templates/handler.go.template pkg/mypackage/handlers.go
```

### 2. Replace Placeholders

Each template contains placeholders in UPPERCASE that you need to replace:

#### Handler Template Placeholders

- `PACKAGE_NAME` - Package name (e.g., `packs`, `inventories`)
- `SUMMARY_TEXT` - Short Swagger summary
- `DESCRIPTION_TEXT` - Detailed Swagger description
- `TAG_NAME` - Swagger tag (e.g., `Packs`, `Inventories`)
- `PARAM_NAME` - Parameter name
- `PARAM_TYPE` - Parameter location (`path`, `query`, `body`)
- `PARAM_GO_TYPE` - Go type (e.g., `int`, `string`)
- `REQUIRED` - `true` or `false`
- `PARAM_DESCRIPTION` - Parameter description
- `RESPONSE_TYPE` - Response dataset type
- `ROUTER_PATH` - API route (e.g., `/api/v1/packs/{id}`)
- `METHOD` - HTTP method (`get`, `post`, `put`, `delete`)
- `HANDLER_NAME` - Function name (e.g., `GetMyPacks`)
- `INPUT_TYPE` - Input dataset type

#### Business Function Template Placeholders

- `PACKAGE_NAME` - Package name
- `NAME` - Resource name (e.g., `Pack`, `Inventory`)
- `TYPE` - Dataset type (e.g., `dataset.Pack`)
- `COLLECTION_TYPE` - Collection type (e.g., `dataset.Packs`)
- `ITEM_TYPE` - Individual item type
- `INPUT_TYPE` - Input type (e.g., `dataset.PackInput`)

#### Test Template Placeholders

- `PACKAGE_NAME` - Package name
- `HANDLER_NAME` - Handler function name
- `RESPONSE_TYPE` - Expected response type

#### Migration Template Placeholders

- `DESCRIPTION` - Short description of migration
- `DATE` - Creation date
- `PURPOSE_DESCRIPTION` - Detailed purpose explanation

### 3. Customize Logic

After replacing placeholders, customize the template:

- Add specific validation logic
- Adjust error handling for your use case
- Add any additional fields or operations
- Implement business logic

## Example

### Creating a New Handler

```bash
# 1. Copy template
cp templates/handler.go.template pkg/items/handlers.go

# 2. Replace placeholders (using sed, manually, or your editor)
sed -i '' 's/PACKAGE_NAME/items/g' pkg/items/handlers.go
sed -i '' 's/HANDLER_NAME/GetMyItems/g' pkg/items/handlers.go
# ... continue with other replacements

# 3. Open in editor and customize
# - Add specific business logic
# - Adjust error messages
# - Add additional validations
```

### Creating a Migration

```bash
# 1. Determine next migration number
ls pkg/database/migrations/ | tail -1  # See last number

# 2. Copy templates
cp templates/migration.up.sql.template pkg/database/migrations/000010_items.up.sql
cp templates/migration.down.sql.template pkg/database/migrations/000010_items.down.sql

# 3. Edit files and replace placeholders
# 4. Test migration up and down
```

## Best Practices

1. **Always use templates** as a starting point for consistency
2. **Don't skip placeholders** - replace all of them
3. **Customize after replacing** - templates are starting points, not final code
4. **Test immediately** - run tests and linter after creating from template
5. **Update templates** - if you find a better pattern, update the template

## Template Maintenance

When you discover a better pattern or common issue:

1. Update the relevant template
2. Document the change in git commit
3. Consider whether existing code should be refactored to match

## See Also

- [agents.md](../../agents.md) - Core principles and architecture
- [PATTERNS.md](../PATTERNS.md) - Detailed code patterns and examples
- [COLLABORATION.md](../COLLABORATION.md) - Collaboration best practices
