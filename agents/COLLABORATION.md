# Collaboration Guide - PimpMyPack

This document captures learnings and best practices from collaborative development with AI agents on the PimpMyPack project.

> **See also**: [agents.md](../agents.md) for core principles and architecture overview.

## 🎓 Spec-Driven Development

### Key Principle

**The specification file is a living document that evolves throughout the implementation.**

### Best Practices

#### 1. Comprehensive Planning

Include all phases from design to deployment in the spec:

- Design decisions and architecture
- Acceptance criteria (functional requirements)
- Implementation phases with tasks
- Testing strategy
- Documentation requirements
- Deployment considerations

**Important**: Do NOT include time estimates or "Estimated Effort" in implementation plans. Focus on what needs to be done, not how long it might take.

#### 2. Checkbox Tracking

Use checkboxes to track completion status at multiple levels:

- **Acceptance criteria**: Functional requirements validation
- **Test cases**: Testing strategy progress
- **Implementation phases**: Development task completion

Example structure:

```markdown
## Implementation Plan

### Phase 1: Database Schema ✅ Completed (2025-01-15)

- [x] Create migration files
- [x] Test migration up/down
- [x] Update dataset types

**Status**: Complete. Files created at `pkg/database/migrations/000009_pack_sharing.up.sql`

### Phase 2: Business Functions ⏳ In Progress

- [x] Implement `sharePack` function
- [ ] Implement `unsharePack` function
- [ ] Add helper functions

**Status**: In progress. Function `sharePack` added to `pkg/packs/pack.go:245-267`
```

#### 3. Real-Time Updates

Update checkboxes and status immediately as work is completed:

- Mark checkboxes when tasks finish (don't batch updates)
- Add implementation dates
- Include file references with line numbers
- Note any blockers or issues

#### 4. Document Decisions

Record implementation details inline:

- **Files modified**: Include paths and line ranges
- **Functions added/changed**: Name and purpose
- **Architecture decisions**: Why this approach was chosen
- **Deviations from plan**: What changed and why

Example:

```markdown
### Phase 3: Handlers ✅ Completed (2025-01-15)

- [x] Create `ShareMyPack` handler
- [x] Create `UnshareMyPack` handler
- [x] Add Swagger documentation

**Implementation Details**:

- `ShareMyPack` handler at `pkg/packs/handlers.go:156-178`
- `UnshareMyPack` handler at `pkg/packs/handlers.go:180-202`
- Both handlers use idempotent design (no errors for already-shared/unshared)
- Swagger annotations complete with all status codes

**Deviations**:

- Did not create `ErrAlreadyShared` sentinel error - using idempotent design instead
```

#### 5. Traceability

Link code references to spec sections (bidirectional traceability):

- From spec to code: "Implemented in `file.go:123-145`"
- From code comments: "See spec section X.Y for design rationale"

### Example from Optional Pack Sharing Feature

Real learnings from implementation:

- Phase 7 distinguished between "basic unit tests complete" and "integration tests pending"
- Documentation section tracked both code changes AND generated Swagger files
- Each phase included implementation date and specific file references
- Deviations were clearly documented (e.g., "no new sentinel errors - using idempotent design")

## 🧪 Test-Driven Quality

### Cognitive Complexity Management

#### 1. Linter as Guardian

Respect complexity limits (cognitive complexity > 20 is a signal):

- When linter flags complexity, refactor immediately
- Don't disable linter warnings without good reason
- Use complexity as a code quality metric

#### 2. Helper Function Extraction

Break complex test functions into smaller, focused helpers:

```go
// Main test function - orchestrates test cases
func TestPackSharing(t *testing.T) {
    gin.SetMode(gin.TestMode)
    router := setupRouter()

    t.Run("Share pack successfully", func(t *testing.T) {
        testShareSuccess(t, router, packID, userID)
    })

    t.Run("Share idempotent behavior", func(t *testing.T) {
        testShareIdempotent(t, router, packID, userID)
    })
}

// Helper functions - focused and reusable
func testShareSuccess(t *testing.T, router *gin.Engine, packID, userID uint) {
    // Focused test logic
}

func testShareIdempotent(t *testing.T, router *gin.Engine, packID, userID uint) {
    // Idempotency verification
}
```

#### 3. Naming Clarity

Helper functions should clearly indicate what they test:

- `testShareSuccess` - tests successful sharing
- `testShareIdempotent` - tests idempotent behavior
- `testUnauthorizedAccess` - tests authorization failures
- `assertSharingCodeExists` - assertion helper

#### 4. Code Reuse

Helper functions reduce duplication AND complexity:

- Reusable across multiple test cases
- Easier to understand and maintain
- Lower cognitive complexity per function

### Testing Strategy

#### 1. Write Tests Alongside Implementation

Not as an afterthought:

- Write test structure when designing feature
- Implement tests as you implement features
- Update tests immediately when behavior changes

#### 2. Test Data Management

Use consistent patterns for test data:

- **Pointer types**: Use `*string` in test data when production uses `*string`
- **Test helpers**: Create utilities like `ComparePtrString` for common patterns
- **Random data**: Use `random.UniqueId()` to avoid conflicts between tests

Example:

```go
// Helper for pointer string comparison
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

#### 3. Idempotency Testing

Explicitly test idempotent behavior:

- Call operation twice, verify same result
- Test both success path (share → share) and cleanup path (unshare → unshare)
- Document idempotency in tests and code comments

```go
func testShareIdempotent(t *testing.T, router *gin.Engine, packID uint) {
    // Share first time
    code1 := shareAndGetCode(t, router, packID)

    // Share second time - should return same code
    code2 := shareAndGetCode(t, router, packID)

    if code1 != code2 {
        t.Errorf("Idempotency violation: got different codes %s and %s", code1, code2)
    }
}
```

## 📚 Documentation Completeness

### Multi-Level Documentation

Document at all levels:

#### 1. Swagger Annotations

Complete API documentation with all status codes:

```go
// @Summary Share a pack
// @Description Makes a pack shareable by generating a unique sharing code
// @Security Bearer
// @Tags Packs
// @Accept json
// @Produce json
// @Param id path int true "Pack ID"
// @Success 200 {object} dataset.Pack "Pack with sharing code"
// @Failure 400 {object} dataset.ErrorResponse "Bad request"
// @Failure 401 {object} dataset.ErrorResponse "Unauthorized"
// @Failure 403 {object} dataset.ErrorResponse "Forbidden (not owner)"
// @Failure 404 {object} dataset.ErrorResponse "Pack not found"
// @Failure 500 {object} dataset.ErrorResponse "Internal server error"
// @Router /api/v1/packs/{id}/share [post]
```

#### 2. Code Comments

Document the "why" and behavior (especially for idempotency):

```go
// sharePack makes a pack shareable by generating a unique sharing code.
// This operation is idempotent - calling it multiple times on the same pack
// will return the existing sharing code without generating a new one.
func sharePack(ctx context.Context, packID uint) (*string, error) {
    // Implementation
}
```

#### 3. Spec File

High-level design, decisions, and implementation tracking:

- Architecture decisions and rationale
- Trade-offs considered
- Implementation approach
- Progress tracking

#### 4. Migration Files

Include comments explaining database changes:

```sql
-- Remove NOT NULL constraint from sharing_code
-- to allow packs to be unshared (set to NULL)
ALTER TABLE packs ALTER COLUMN sharing_code DROP NOT NULL;

-- Clean up existing data: convert empty strings to NULL
UPDATE packs SET sharing_code = NULL WHERE sharing_code = '';
```

### Consistency Checks

Before committing, verify:

- **Swagger ↔ Implementation**: Annotations match actual HTTP status codes
- **Migration numbering**: Sequential (000009 follows 000008)
- **Spec ↔ Code**: Spec reflects actual implementation, not aspirational state
- **Tests ↔ Features**: All acceptance criteria have corresponding tests

## 🔄 Iterative Refinement

### Feedback Loop

Establish multiple feedback mechanisms:

#### 1. User Validation

Check spec completeness before diving into implementation:

- Share spec outline for early feedback
- Validate design approach before coding
- Confirm acceptance criteria match user expectations

#### 2. Linter Feedback

Run linter and fix issues immediately:

- Don't batch linter fixes for "later"
- Treat linter warnings as code quality signals
- Refactor when complexity limits are exceeded

#### 3. Test Feedback

Run tests frequently to catch regressions early:

- Run tests after each significant change
- Don't accumulate failing tests
- Fix test failures before adding new features

#### 4. Spec Feedback

User can point out incomplete spec sections:

- Spec serves as a communication tool
- Incomplete sections indicate more discussion needed
- Update spec based on implementation learnings

### Adaptive Planning

Plans evolve based on learnings:

#### Example Evolution

**Initial Plan**:

```markdown
### Phase 7: Testing

- [ ] Write unit tests
```

**Evolved Plan**:

```markdown
### Phase 7: Testing ⏳ In Progress

- [x] Write basic unit tests (handler level)
- [ ] Write integration tests (end-to-end)
- [ ] Test idempotent behavior
- [ ] Test unauthorized access

**Status**: Basic tests complete. Integration tests pending.
```

#### Document the Evolution

When implementation differs from initial design:

- Note what changed
- Explain why it changed
- Update acceptance criteria if needed

Example:

```markdown
**Deviations from Design**:

- Originally planned `ErrAlreadyShared` sentinel error
- Changed to idempotent design after considering user experience
- Idempotency simplifies error handling and improves retry safety
```

## 💬 Communication Patterns

### Effective Collaboration

#### 1. Explicit Status

Mark phases clearly:

- ✅ **Completed**: Done with date
- ⏳ **In Progress**: Currently working
- ⏸️ **Blocked**: Waiting on something
- ⏭️ **Pending**: Not started

#### 2. Implementation Dates

Track when work was actually done:

- Not just planned dates
- Actual completion dates
- Helps identify velocity and bottlenecks

Example:

```markdown
### Phase 2: Business Functions ✅ Completed (2025-01-15)

**Implementation Date**: January 15, 2025
```

#### 3. File References

Always include file paths and line numbers:

- Exact location: `pkg/packs/handlers.go:156-178`
- Function name: `ShareMyPack`
- Purpose: "Handler for POST /api/v1/packs/:id/share"

#### 4. Deviations

Document when implementation differs from design:

```markdown
**Design Decision**: Using idempotent operations

**Rationale**: Simplifies error handling, improves retry safety, better UX

**Impact**: No `ErrAlreadyShared` error needed
```

### Progress Visibility

Enable stakeholders to quickly understand status:

#### At-a-Glance Status

```markdown
## Implementation Progress

- ✅ Database schema (2025-01-14)
- ✅ Business functions (2025-01-15)
- ✅ Handlers (2025-01-15)
- ✅ Routes (2025-01-15)
- ⏳ Testing (in progress)
- ⏭️ Documentation (pending)
```

#### Drill-Down Details

For each phase, provide:

- Checkbox list of specific tasks
- Implementation details (files, functions)
- Status notes (blockers, pending items)
- Deviations or decisions

## 🌿 Branch Workflow

### Before Starting Any Implementation

**ALWAYS follow this workflow before beginning work on a new feature:**

1. **Checkout main branch**:

   ```bash
   git checkout main
   ```

2. **Pull latest changes**:

   ```bash
   git pull origin main
   ```

3. **Create new feature branch from main**:

   ```bash
   git checkout -b feat/feature-name
   ```

### Branch Naming Conventions

Use descriptive branch names with appropriate prefixes:

- `feat/` - New features (e.g., `feat/pack-image-upload`)
- `fix/` - Bug fixes (e.g., `fix/login-401-status`)
- `docs/` - Documentation updates (e.g., `docs/api-endpoints`)
- `refactor/` - Code refactoring (e.g., `refactor/pack-sharing`)
- `test/` - Test additions or fixes (e.g., `test/inventory-dedup`)
- `chore/` - Maintenance tasks (e.g., `chore/deps-update`)

### Rationale

- Ensures you're working from the latest code
- Prevents merge conflicts and integration issues
- Maintains clean git history
- Follows standard Git workflow best practices

## 🚫 Git and Remote Operations Policy

### CRITICAL: Never Push Without Explicit Approval

**This is a hard rule that must NEVER be violated.**

#### Forbidden Operations

The following operations are FORBIDDEN unless the user explicitly requests them:

- `git push` (pushing commits to remote)
- `git push origin <branch>` (pushing specific branches)
- `gh pr create` (creating pull requests)
- Any operation that sends local commits to a remote repository

#### Allowed Operations

The following operations are ALLOWED without asking:

- `git add` (staging files)
- `git commit` (creating local commits)
- `git status` (checking repository status)
- `git diff` (viewing changes)
- `git log` (viewing commit history)
- `git branch` (managing local branches)

#### Workflow Guidelines

**Correct workflow**:

1. Make changes to code
2. Stage changes with `git add`
3. Create local commit with `git commit`
4. **STOP and WAIT for user approval**
5. User explicitly says "push" or "create PR"
6. Only then execute `git push` or `gh pr create`

**What NOT to do**:

1. Never assume "ready to merge" means "push now"
2. Never automatically push after committing
3. Never open PRs without explicit instruction
4. Never batch "commit + push" operations without asking

#### Why This Matters

- User autonomy: The user controls when code goes to remote
- Review opportunity: User may want to review commits before pushing
- Workflow respect: User has their own process and timing
- Trust: Violating this rule breaks user trust

#### Exception Cases

The ONLY time pushing without asking is acceptable:

- User explicitly says: "push to remote"
- User explicitly says: "create a PR"
- User explicitly says: "open a pull request"
- User uses command: `/push` or similar explicit command

#### What to Do After Committing

After creating local commits, respond with:

"I've created a local commit with the changes. The commit is ready to be pushed to remote when you're ready. Would you like me to push it now?"

Or simply:

"Changes committed locally. Let me know when you'd like to push to remote."

#### If This Rule Is Broken

If you accidentally push or create a PR without approval:

1. Immediately acknowledge the mistake
2. Apologize sincerely
3. Explain what happened
4. Ask if the user wants you to close the PR or take other action
5. Reference this section to show you understand the policy

## 🏛️ Technical Decisions

### Design Patterns Validated

Patterns that proved valuable:

#### 1. Idempotency

Design share/unshare as idempotent operations:

**Benefits**:

- Simplifies error handling (no "already shared" error needed)
- Better user experience (safe to retry)
- Reduces cognitive load

**Documentation**:

- Document in code comments
- Document in Swagger annotations
- Test explicitly

**Example**:

```go
// sharePack is idempotent - calling it multiple times returns the same code
func sharePack(ctx context.Context, packID uint) (*string, error) {
    // Check if already shared
    if existingCode := getExistingCode(ctx, packID); existingCode != nil {
        return existingCode, nil
    }

    // Generate new code
    return generateAndSaveCode(ctx, packID)
}
```

#### 2. Pointer Types for Optional Fields

Use `*string` for nullable fields like `SharingCode`:

**Benefits**:

- Semantically correct (NULL vs empty string)
- Matches PostgreSQL NULL behavior
- Clear intent (presence vs absence)

**Cost**:

- Requires helper functions for comparisons
- Slightly more complex test data setup

**Example**:

```go
type Pack struct {
    ID          uint      `json:"id"`
    Name        string    `json:"name"`
    SharingCode *string   `json:"sharing_code,omitempty"` // NULL when not shared
}
```

#### 3. Context Propagation

All business functions accept `context.Context` as first parameter:

**Benefits**:

- Enables timeout handling
- Supports cancellation
- Carries request-scoped values
- Standard Go practice

**Pattern**:

```go
func businessFunction(ctx context.Context, params ...) error {
    // Use ctx in all DB operations
    rows, err := database.DB.QueryContext(ctx, query, params...)
    // ...
}
```

#### 4. Ownership Checks

Consistent security pattern across handlers:

**Pattern**:

```go
func HandlerForUserResource(c *gin.Context) {
    // 1. Extract resource ID from path
    resourceID := extractID(c)

    // 2. Get user ID from JWT
    userID := getUserIDFromContext(c)

    // 3. Verify ownership before any operation
    if !isOwner(userID, resourceID) {
        c.IndentedJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
        return
    }

    // 4. Proceed with operation
    // ...
}
```

### Migration Best Practices

Patterns that ensure safe schema evolution:

#### Up Migration

```sql
-- Remove constraints first
ALTER TABLE packs ALTER COLUMN sharing_code DROP NOT NULL;

-- Clean existing data
UPDATE packs SET sharing_code = NULL WHERE sharing_code = '';

-- Add new constraints if needed
-- (in this case, nullable is desired, so no new constraint)
```

#### Down Migration

```sql
-- Restore data (set sensible defaults for NULL values)
UPDATE packs SET sharing_code = '' WHERE sharing_code IS NULL;

-- Restore constraints
ALTER TABLE packs ALTER COLUMN sharing_code SET NOT NULL;
```

#### Key Points

- Order matters: constraints before data
- Clean data as part of migration
- Make migrations reversible
- Test both up and down paths

## ✅ Quality Gates

### Pre-Merge Checklist

Based on learnings from our collaborative process:

#### Documentation

- [ ] All phase checkboxes reviewed and accurate
- [ ] Implementation dates recorded
- [ ] File references with line numbers included
- [ ] Deviations from design documented
- [ ] Spec reflects actual implementation state

#### Code Quality

- [ ] Linter passes (`make lint`)
- [ ] Cognitive complexity within limits
- [ ] Error handling follows patterns
- [ ] Security checks implemented (ownership, validation)

#### Testing

- [ ] New feature tests written
- [ ] All tests pass (`make test`)
- [ ] Idempotent behavior tested
- [ ] Edge cases covered
- [ ] Test helpers created for complexity management

#### Documentation

- [ ] Swagger documentation complete
- [ ] All HTTP status codes documented
- [ ] Code comments explain "why"
- [ ] Public functions have doc comments

#### Database

- [ ] Migration files created (up and down)
- [ ] Migration numbering sequential
- [ ] Migrations tested (up and rollback)
- [ ] Data cleanup included in migrations

#### Integration

- [ ] Routes added to main.go
- [ ] Middleware applied correctly (auth, etc.)
- [ ] Handlers follow standard pattern
- [ ] Dataset types created/updated

### Key Insight

**The spec file itself serves as a comprehensive pre-merge checklist when properly maintained.**

By keeping the spec updated throughout development, you create a natural quality gate that ensures all requirements are met before considering the work complete.
