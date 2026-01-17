# [Feature Name] - Specification

**Status**: Draft | Under Review | Approved | In Progress | Completed  
**Author**: [Your Name/Agent Name]  
**Date**: YYYY-MM-DD  
**Last Updated**: YYYY-MM-DD  

---

## 📋 Overview

### Purpose
Brief description of what this feature aims to achieve and why it's needed.

### Context
Background information, related features, or dependencies that provide context for this specification.

### Goals
- Clear, measurable goal 1
- Clear, measurable goal 2
- Clear, measurable goal 3

### Non-Goals
- What this feature will NOT do
- Out of scope items

---

## 🎯 Requirements

### Functional Requirements

#### FR1: [Requirement Title]
**Description**: Detailed description of the functional requirement.

**Acceptance Criteria**:
- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

**Priority**: High | Medium | Low

#### FR2: [Another Requirement]
...

### Non-Functional Requirements

#### NFR1: Performance
- Response time: < Xms for Y% of requests
- Throughput: Z requests/second

#### NFR2: Security
- Authentication requirements
- Authorization levels
- Data protection needs

#### NFR3: Reliability
- Uptime requirements
- Error handling expectations

---

## 🏗️ Design

### Data Model

#### New Types (pkg/dataset)
```go
// Define new structs here
type NewEntity struct {
    ID          uint      `json:"id"`
    UserID      uint      `json:"user_id"`
    FieldName   string    `json:"field_name"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type NewEntities []NewEntity

// Input/Output types
type NewEntityInput struct {
    FieldName string `json:"field_name" binding:"required"`
}
```

#### Database Schema
```sql
-- Migration file: 00000X_feature_name.up.sql
CREATE TABLE "new_entity" (
    "id"          SERIAL PRIMARY KEY,
    "user_id"     INTEGER NOT NULL,
    "field_name"  TEXT NOT NULL,
    "created_at"  TIMESTAMP NOT NULL,
    "updated_at"  TIMESTAMP NOT NULL,
    FOREIGN KEY ("user_id") REFERENCES "account"("id") ON DELETE CASCADE
);

CREATE INDEX idx_new_entity_user_id ON new_entity(user_id);
```

```sql
-- Migration file: 00000X_feature_name.down.sql
DROP TABLE IF EXISTS "new_entity";
```

### API Design

#### Endpoints

##### 1. GET /api/v1/myentities
**Description**: Get all entities for the authenticated user

**Authentication**: Required (JWT)

**Request**: None

**Response** (200 OK):
```json
[
  {
    "id": 1,
    "user_id": 1,
    "field_name": "value",
    "created_at": "2026-01-17T10:00:00Z",
    "updated_at": "2026-01-17T10:00:00Z"
  }
]
```

**Error Responses**:
- 401: Unauthorized
- 404: No entities found
- 500: Internal server error

##### 2. POST /api/v1/myentity
**Description**: Create a new entity

**Authentication**: Required (JWT)

**Request Body**:
```json
{
  "field_name": "value"
}
```

**Response** (201 Created):
```json
{
  "message": "Entity created successfully",
  "id": 1
}
```

**Error Responses**:
- 400: Bad request (validation errors)
- 401: Unauthorized
- 500: Internal server error

##### 3. PUT /api/v1/myentity/:id
**Description**: Update an existing entity

**Authentication**: Required (JWT)

**Path Parameters**:
- `id` (integer): Entity ID

**Request Body**:
```json
{
  "field_name": "new value"
}
```

**Response** (200 OK):
```json
{
  "message": "Entity updated successfully"
}
```

**Error Responses**:
- 400: Bad request (validation errors)
- 401: Unauthorized
- 403: Forbidden (not owner)
- 404: Entity not found
- 500: Internal server error

##### 4. DELETE /api/v1/myentity/:id
**Description**: Delete an entity

**Authentication**: Required (JWT)

**Path Parameters**:
- `id` (integer): Entity ID

**Response** (200 OK):
```json
{
  "message": "Entity deleted successfully"
}
```

**Error Responses**:
- 401: Unauthorized
- 403: Forbidden (not owner)
- 404: Entity not found
- 500: Internal server error

### Architecture Components

#### Package: pkg/entities

**Sentinel Errors**:
```go
var ErrNoEntityFound = errors.New("no entity found")
var ErrEntityNotOwned = errors.New("entity not owned by user")
```

**Handler Functions**:
- `GetMyEntities(c *gin.Context)`
- `GetMyEntityByID(c *gin.Context)`
- `PostMyEntity(c *gin.Context)`
- `PutMyEntityByID(c *gin.Context)`
- `DeleteMyEntityByID(c *gin.Context)`

**Business Functions**:
- `returnMyEntities(ctx context.Context, userID uint) (*dataset.NewEntities, error)`
- `returnMyEntityByID(ctx context.Context, userID uint, entityID uint) (*dataset.NewEntity, error)`
- `createEntity(ctx context.Context, userID uint, input dataset.NewEntityInput) (uint, error)`
- `updateEntity(ctx context.Context, userID uint, entityID uint, input dataset.NewEntityInput) error`
- `deleteEntity(ctx context.Context, userID uint, entityID uint) error`

#### Routes (main.go)
```go
// In setupProtectedRoutes()
protected.GET("/myentities", entities.GetMyEntities)
protected.GET("/myentity/:id", entities.GetMyEntityByID)
protected.POST("/myentity", entities.PostMyEntity)
protected.PUT("/myentity/:id", entities.PutMyEntityByID)
protected.DELETE("/myentity/:id", entities.DeleteMyEntityByID)
```

#### Admin Routes (if applicable)
```go
// In setupPrivateRoutes()
private.GET("/entities", entities.GetEntities)
private.GET("/entities/:id", entities.GetEntityByID)
// ... other admin endpoints
```

---

## 🧪 Testing Strategy

### Unit Tests

#### Test Files
- `pkg/entities/entities_test.go`
- `pkg/entities/testdata.go`

#### Test Cases

**GetMyEntities**:
- [ ] Successfully retrieve user's entities
- [ ] Return 404 when no entities found
- [ ] Return 401 for unauthenticated request

**PostMyEntity**:
- [ ] Successfully create new entity
- [ ] Return 400 for invalid input
- [ ] Return 401 for unauthenticated request

**PutMyEntityByID**:
- [ ] Successfully update entity
- [ ] Return 403 when user doesn't own entity
- [ ] Return 404 when entity not found
- [ ] Return 400 for invalid input

**DeleteMyEntityByID**:
- [ ] Successfully delete entity
- [ ] Return 403 when user doesn't own entity
- [ ] Return 404 when entity not found

### Integration Tests
- [ ] End-to-end workflow: create → read → update → delete
- [ ] Test with multiple users to verify isolation
- [ ] Test cascade deletion (if applicable)

### Test Data

```go
// In testdata.go
var testEntities = []dataset.NewEntity{
    {
        UserID:    1,
        FieldName: "Test Entity 1",
    },
    {
        UserID:    2,
        FieldName: "Test Entity 2",
    },
}
```

---

## 📚 Documentation Updates

### Swagger Documentation
- [ ] Add Swagger annotations to all handlers
- [ ] Regenerate Swagger documentation (`make api-doc`)
- [ ] Verify documentation at `/swagger/index.html`

### README Updates
- [ ] Update feature list (if needed)
- [ ] Add any new setup instructions
- [ ] Update examples (if needed)

### Code Comments
- [ ] Document all public functions
- [ ] Document sentinel errors
- [ ] Add inline comments for complex logic

---

## 🚀 Implementation Plan

### Phase 1: Database & Data Models
- [ ] Create migration files (up and down)
- [ ] Define types in `pkg/dataset/dataset.go`
- [ ] Test migration locally

### Phase 2: Business Logic
- [ ] Create `pkg/entities` package
- [ ] Implement sentinel errors
- [ ] Implement business functions with context
- [ ] Add proper error handling and wrapping

### Phase 3: HTTP Handlers
- [ ] Implement handler functions
- [ ] Add Swagger documentation
- [ ] Add input validation
- [ ] Implement proper HTTP status codes

### Phase 4: Routing
- [ ] Add routes in `main.go`
- [ ] Test routes with curl/Postman
- [ ] Verify JWT authentication

### Phase 5: Testing
- [ ] Create test data in `testdata.go`
- [ ] Write unit tests for handlers
- [ ] Write unit tests for business functions
- [ ] Achieve >80% coverage

### Phase 6: Quality & Documentation
- [ ] Run linter (`make lint`)
- [ ] Fix all linting issues
- [ ] Generate and verify Swagger docs
- [ ] Update README if needed

### Phase 7: Final Validation
- [ ] Run full test suite (`make test`)
- [ ] Test manually in local environment
- [ ] Verify all acceptance criteria met
- [ ] Code review ready

---

## 🔍 Security Considerations

### Authentication & Authorization
- JWT token required for all endpoints
- User can only access/modify their own entities
- Admin endpoints require admin role

### Input Validation
- Validate all user inputs
- Sanitize strings to prevent SQL injection
- Use parameterized queries

### Data Protection
- No sensitive data in logs
- Appropriate error messages (no data leakage)
- HTTPS in production

---

## 📈 Performance Considerations

### Database
- Add appropriate indexes
- Use connection pooling
- Implement context timeouts

### API
- Response pagination (if list can be large)
- Caching strategy (if applicable)
- Rate limiting (if needed)

---

## 🔄 Migration & Rollback Plan

### Deployment Steps
1. Run database migration (`up`)
2. Deploy new application version
3. Verify endpoints work correctly
4. Monitor for errors

### Rollback Steps
1. Revert to previous application version
2. Run migration rollback (`down`)
3. Verify system stability

### Data Migration (if applicable)
- No existing data to migrate
- OR: Detail data migration steps

---

## 🤔 Open Questions & Decisions

### Questions
- [ ] Question 1 that needs project owner input?
- [ ] Question 2 about scope or implementation?

### Decisions Made
- **Decision 1**: Reasoning behind the decision
- **Decision 2**: Alternative considered and why rejected

---

## 📎 References

- Related issues/tickets: #XXX
- Related specifications: [Link to related spec]
- External documentation: [Link]
- Design discussions: [Link to discussion]

---

## ✅ Sign-off

### Design Approval
- [ ] Reviewed by project owner
- [ ] Design approved
- [ ] Ready for implementation

**Approver**: [Name]  
**Date**: YYYY-MM-DD

### Implementation Approval
- [ ] All implementation tasks completed
- [ ] Tests passing
- [ ] Lint checks passing
- [ ] Documentation updated
- [ ] Ready for merge

**Approver**: [Name]  
**Date**: YYYY-MM-DD
