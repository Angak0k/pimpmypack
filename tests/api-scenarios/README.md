# API Test Scenarios

This directory contains YAML-based test scenarios for automated API testing using the `api-test-runner` agent.

## Quick Start

1. **Start the server** in LOCAL mode:
   ```bash
   # Make sure STAGE=LOCAL in .env
   go run main.go
   ```

2. **Run test scenarios** via Claude Code:
   ```
   "Run the API test scenarios"
   ```

## Scenario Format

Test scenarios are defined in YAML files with the following structure:

```yaml
name: "Test Scenario Name"
base_url: "http://localhost:8080/api"
scenarios:
  - name: "Step description"
    request:
      method: POST                    # GET, POST, PUT, DELETE
      endpoint: "/endpoint"            # Relative to base_url
      headers:                         # Optional
        Authorization: "Bearer {{access_token}}"
        Content-Type: "application/json"
      body:                            # Optional (for POST/PUT)
        field: "value"
        dynamic: "{{variable}}"
    assertions:                        # Validation rules
      - type: status_code
        expected: 200
      - type: json_path
        path: "$.field"
        equals: "expected_value"
      - type: json_path
        path: "$.other_field"
        exists: true
    store:                             # Save values for later use
      variable_name: "{{response.field}}"
```

## Variable Substitution

Variables are referenced using `{{variable_name}}` syntax and are substituted at runtime.

### Built-in Variables
- `{{timestamp}}` - Current Unix timestamp (for unique usernames/emails)
- `{{random}}` - Random string

### Custom Variables
Store response values for use in subsequent requests:

```yaml
store:
  access_token: "{{response.access_token}}"
  user_id: "{{response.id}}"
```

Use stored variables:

```yaml
request:
  endpoint: "/users/{{user_id}}"
  headers:
    Authorization: "Bearer {{access_token}}"
```

## Assertion Types

### status_code
Validates HTTP response status:

```yaml
assertions:
  - type: status_code
    expected: 200
```

### json_path
Extracts and validates JSON fields using JSONPath syntax:

**Check existence:**
```yaml
  - type: json_path
    path: "$.access_token"
    exists: true
```

**Exact match:**
```yaml
  - type: json_path
    path: "$.username"
    equals: "testuser"
```

**Contains substring:**
```yaml
  - type: json_path
    path: "$.message"
    contains: "success"
```

## Example Scenarios

### Simple Login Test

```yaml
name: "User Login"
base_url: "http://localhost:8080/api"
scenarios:
  - name: "Login with valid credentials"
    request:
      method: POST
      endpoint: "/login"
      body:
        username: "testuser"
        password: "password123"
    assertions:
      - type: status_code
        expected: 200
      - type: json_path
        path: "$.access_token"
        exists: true
    store:
      access_token: "{{response.access_token}}"
```

### Multi-Step Flow with State

```yaml
name: "User Registration and Pack Creation"
base_url: "http://localhost:8080/api"
scenarios:
  - name: "Register new user"
    request:
      method: POST
      endpoint: "/register"
      body:
        username: "user_{{timestamp}}"
        email: "user_{{timestamp}}@test.com"
        password: "TestPass123!"
        firstname: "Test"
        lastname: "User"
    assertions:
      - type: status_code
        expected: 200
    store:
      username: "{{request.body.username}}"
      email: "{{request.body.email}}"

  - name: "Confirm email (LOCAL mode)"
    request:
      method: GET
      endpoint: "/confirmemail?username={{username}}&email={{email}}"
    assertions:
      - type: status_code
        expected: 200

  - name: "Login"
    request:
      method: POST
      endpoint: "/login"
      body:
        username: "{{username}}"
        password: "TestPass123!"
    assertions:
      - type: status_code
        expected: 200
    store:
      access_token: "{{response.access_token}}"

  - name: "Create pack"
    request:
      method: POST
      endpoint: "/v1/mypack"
      headers:
        Authorization: "Bearer {{access_token}}"
      body:
        pack_name: "Test Pack"
        pack_description: "Test description"
    assertions:
      - type: status_code
        expected: 201
      - type: json_path
        path: "$.id"
        exists: true
    store:
      pack_id: "{{response.id}}"
```

## Best Practices

### 1. Use Unique Test Data
Always use timestamps or random strings for test usernames/emails:

```yaml
body:
  username: "test_api_{{timestamp}}"
  email: "test_{{timestamp}}@example.com"
```

### 2. Include Cleanup Steps
Delete test data at the end of scenarios:

```yaml
  - name: "Cleanup: Delete test pack"
    request:
      method: DELETE
      endpoint: "/v1/mypack/{{pack_id}}"
      headers:
        Authorization: "Bearer {{access_token}}"
    assertions:
      - type: status_code
        expected: 200
```

### 3. Test Error Cases
Don't just test happy paths:

```yaml
  - name: "Login with invalid password"
    request:
      method: POST
      endpoint: "/login"
      body:
        username: "{{username}}"
        password: "wrongpassword"
    assertions:
      - type: status_code
        expected: 401
```

### 4. Document Your Scenarios
Use descriptive step names that explain what's being tested:

```yaml
  - name: "Verify authenticated user can access own account info"
```

## Running Tests

### Run All Scenarios
```
"Run all API test scenarios in tests/api-scenarios/"
```

### Run Specific Scenario
```
"Run the test scenario 001-user-registration-auth.yaml"
```

### Run and Save Report
```
"Run API tests and save the report to test-results.txt"
```

## Troubleshooting

**Server not responding**:
- Ensure server is running: `go run main.go`
- Check it's on port 8080: `curl http://localhost:8080/api/register`

**Email confirmation failing**:
- Verify `STAGE=LOCAL` in `.env`
- Check server logs for "LOCAL MODE" messages

**Assertion failures**:
- Check actual response with curl directly
- Verify JSON path syntax is correct
- Use `jq` to explore response structure: `echo '$response' | jq .`

## Test Scenarios

| File | Description |
|------|-------------|
| `001-user-registration-auth.yaml` | User registration, email confirmation, login, token refresh |
| `002-pack-crud.yaml` | Pack creation, retrieval, update, share, delete |
| `003-inventory-management.yaml` | Inventory item CRUD operations |
| `004-import-lighterpack.yaml` | CSV file upload and import testing |

## Contributing

When adding new test scenarios:

1. Follow the naming convention: `NNN-description.yaml`
2. Include comprehensive assertions
3. Add cleanup steps
4. Test locally before committing
5. Update this README with scenario description
