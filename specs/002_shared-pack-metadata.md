# Shared Pack Metadata Enhancement - Specification

**Status**: Draft
**Author**: Claude Agent
**Date**: 2026-01-20
**Last Updated**: 2026-01-20

---

## 📋 Overview

### Purpose

Currently, the `/api/sharedlist/{sharing_code}` endpoint only returns the list of items contained in a shared pack (PackContents). Users accessing a shared pack have no information about the pack itself (name, description, weight, etc.), which limits the user experience.

This specification aims to enhance the shared pack API to return a structured response that includes:
1. Pack metadata (name, description, item count, creation date)
2. Pack contents (the list of items)

### Context

- Current endpoint: `GET /api/sharedlist/{sharing_code}`
- Current implementation: `SharedList` handler in `pkg/packs/packs.go:1590-1618`
- Current response: Returns `dataset.PackContents` (list of items only)
- No pack metadata is returned to the public user

### Goals

- Provide pack metadata to users accessing shared packs
- Maintain a clear separation between pack information and pack contents
- Improve user experience for shared pack viewers
- Keep the API response structure intuitive and well-documented

### Non-Goals

- Changing the sharing mechanism or sharing code format
- Adding authentication to the public endpoint
- Exposing private pack owner information (user_id should not be exposed)

---

## 🎯 Requirements

### Functional Requirements

#### FR1: Return Pack Metadata
**Description**: The shared pack endpoint should return metadata about the pack itself.

**Acceptance Criteria**:
- [x] Endpoint returns pack name
- [x] Endpoint returns pack description
- [x] Endpoint returns pack creation date
- [x] Pack owner's user_id is NOT exposed in the response

**Note**: Pack items count is NOT included as this column doesn't exist in the database schema. Clients can count items from the contents array if needed.

**Priority**: High

#### FR2: Return Pack Contents
**Description**: The shared pack endpoint should return the list of items in the pack.

**Acceptance Criteria**:
- [x] Endpoint returns complete pack contents (items with details)
- [x] Item details include: name, description, weight, quantity, unit, category, if it's worn or consumable.


**Priority**: High

#### FR3: Structured Response Format
**Description**: The response should have a clear structure separating pack metadata from contents.

**Acceptance Criteria**:
- [x] Response has a `pack` field with metadata
- [x] Response has a `contents` field with items list
- [x] JSON structure is intuitive and well-documented
- [x] Swagger documentation is updated

**Priority**: High

#### FR4: Error Handling
**Description**: Proper error handling for invalid or non-existent sharing codes.

**Acceptance Criteria**:
- [x] 404 error if sharing code doesn't exist
- [x] 404 error if pack is no longer shared (sharing_code is NULL)
- [x] Clear error messages in responses

**Priority**: Medium

### Non-Functional Requirements

#### NFR1: Performance
- Response time: < 100ms for shared pack retrieval
- Single database query optimization (JOIN pack + contents)

#### NFR2: Security
- No exposure of sensitive user data (user_id, email, etc.)
- Sharing code remains the only access method
- No rate limiting needed (public endpoint)

#### NFR3: API Design
- RESTful response structure
- Clear and consistent JSON field naming
- Complete Swagger documentation

---

## 🏗️ Design

### API Design

#### Endpoint (Unchanged)
```
GET /api/sharedlist/{sharing_code}
```

#### Response Structure (NEW)

```json
{
  "pack": {
    "id": 123,
    "pack_name": "Weekend Camping Trip",
    "pack_description": "Essential gear for a 2-day camping trip",
    "created_at": "2026-01-15T10:30:00Z"
  },
  "contents": [
    {
      "id": 1,
      "pack_id": 123,
      "inventory_id": 456,
      "item_name": "Tent",
      "item_description": "2-person ultralight tent",
      "item_weight": 2500,
      "item_weight_unit": "g",
      "item_quantity": 1,
      "item_category": "Shelter",
      "worn": false,
      "consumable": false
    },
    // ... more items
  ]
}
```

### Data Model

#### New Dataset Type: `SharedPackResponse`

```go
// pkg/dataset/dataset.go
type SharedPackResponse struct {
    Pack     SharedPackInfo       `json:"pack"`
    Contents dataset.PackContents `json:"contents"`
}

type SharedPackInfo struct {
    ID              uint      `json:"id"`
    PackName        string    `json:"pack_name"`
    PackDescription string    `json:"pack_description"`
    CreatedAt       time.Time `json:"created_at"`
    // Note: UserID and SharingCode are intentionally NOT included
    // Note: Worn and Consumable are at item level (in PackContents), not pack level
    // Note: pack_items_count is not included - column doesn't exist in DB schema
    //       Clients can count items from the contents array if needed
}
```

### Database Operations

#### Option 1: Two Separate Queries (Current Approach)
1. Query pack by sharing_code to get pack info
2. Query pack_contents to get items

**Pros**: Simple, reuses existing functions
**Cons**: Two database round-trips

#### Option 2: Single JOIN Query (Recommended)
```sql
-- Get pack info + contents in one query
SELECT
    p.id, p.pack_name, p.pack_description, p.pack_weight,
    p.pack_items_count, p.created_at,
    pc.id, pc.pack_id, pc.inventory_id, pc.item_name,
    pc.item_description, pc.item_weight, pc.item_weight_unit,
    pc.item_quantity, pc.item_category
FROM packs p
LEFT JOIN pack_contents pc ON p.id = pc.pack_id
WHERE p.sharing_code = $1
```

**Pros**: Single database query, better performance
**Cons**: More complex query parsing

**Recommendation**: Start with Option 1 (two queries) for simplicity, optimize to Option 2 if needed.

### Implementation Approach

#### Step 1: Create New Dataset Types
- Add `SharedPackResponse` and `SharedPackInfo` to `pkg/dataset/dataset.go`

#### Step 2: Create Business Function
```go
// pkg/packs/packs.go
func returnSharedPack(ctx context.Context, sharingCode string) (*dataset.SharedPackResponse, error)
```

This function will:
1. Get pack info by sharing code
2. Get pack contents by pack ID
3. Combine into `SharedPackResponse`

#### Step 3: Update Handler
- Modify `SharedList` handler to use new business function
- Update response to return `SharedPackResponse`
- Update Swagger annotations

#### Step 4: Testing
- Test with valid sharing codes
- Test with invalid sharing codes
- Test with unshared packs (NULL sharing_code)
- Test response structure matches specification

---

## 📐 Technical Design

### File Changes

#### 1. `pkg/dataset/dataset.go`
- Add `SharedPackResponse` struct
- Add `SharedPackInfo` struct

#### 2. `pkg/packs/packs.go`
- Add `returnSharedPack(ctx context.Context, sharingCode string) (*dataset.SharedPackResponse, error)`
- Modify `SharedList` handler to use new function
- Update Swagger annotations

#### 3. `pkg/packs/packs_test.go`
- Add tests for `returnSharedPack`
- Update tests for `SharedList` handler
- Test error cases (invalid code, unshared pack)

### Swagger Documentation

```go
// @Summary Get shared pack with metadata
// @Description Retrieves pack metadata and contents using a sharing code
// @Tags Public
// @Accept json
// @Produce json
// @Param sharing_code path string true "Pack sharing code"
// @Success 200 {object} dataset.SharedPackResponse "Shared pack with metadata and contents"
// @Failure 404 {object} dataset.ErrorResponse "Pack not found or not shared"
// @Failure 500 {object} dataset.ErrorResponse "Internal server error"
// @Router /sharedlist/{sharing_code} [get]
```

---

## 🧪 Testing Strategy

### Unit Tests

#### Test Cases for `returnSharedPack`
1. **Success case**: Valid sharing code returns pack + contents
2. **Not found**: Invalid sharing code returns error
3. **Unshared pack**: Pack with NULL sharing_code returns error
4. **Empty pack**: Valid code with no contents returns pack with empty array

#### Test Cases for `SharedList` Handler
1. **Success response**: Returns 200 with correct structure
2. **404 response**: Invalid code returns 404
3. **Response format**: JSON matches `SharedPackResponse` structure

### Integration Tests
1. Create a pack, share it, retrieve via sharing code
2. Unshare a pack, verify 404 on retrieval
3. Verify no user_id exposed in response

### Manual Testing
1. Test with real sharing codes in development
2. Verify Swagger UI shows correct response structure
3. Test with frontend client (if applicable)

---

## 📝 Implementation Plan

### Phase 1: Data Model ✅ Completed (2026-01-20)

- [x] Add `SharedPackResponse` to `pkg/dataset/dataset.go`
- [x] Add `SharedPackInfo` to `pkg/dataset/dataset.go`

**Implementation**: Added at [pkg/dataset/dataset.go:156-170](../pkg/dataset/dataset.go#L156-L170)

**Deviation**: Removed `pack_items_count` from `SharedPackInfo` as this column doesn't exist in the database schema (table `pack` only has: id, user_id, pack_name, pack_description, sharing_code, created_at, updated_at). Clients can count items from the contents array if needed.

### Phase 2: Business Logic ✅ Completed (2026-01-20)

- [x] Implement `returnSharedPack` function
- [x] Add helper function `returnPackInfoBySharingCode` for pack info retrieval
- [x] Handle error cases (not found, unshared)

**Implementation**:

- `returnPackInfoBySharingCode` at [pkg/packs/packs.go:1622-1647](../pkg/packs/packs.go#L1622-L1647)
- `returnSharedPack` at [pkg/packs/packs.go:1650-1678](../pkg/packs/packs.go#L1650-L1678)
- Added `fmt` import at [pkg/packs/packs.go:6](../pkg/packs/packs.go#L6)

### Phase 3: Handler Update ✅ Completed (2026-01-20)

- [x] Modify `SharedList` handler
- [x] Update Swagger annotations
- [x] Ensure proper error handling

**Implementation**: Updated handler at [pkg/packs/packs.go:1591-1606](../pkg/packs/packs.go#L1591-L1606)

- Simplified to use `returnSharedPack` function
- Updated Swagger annotations to reflect new response structure
- Proper error handling: 404 for not found, 500 for internal errors

### Phase 4: Testing ✅ Completed (2026-01-20)

- [x] Write unit tests
- [x] Test error cases
- [x] Verify response structure

**Implementation**: Tests added at [pkg/packs/packs_test.go:1235-1302](../pkg/packs/packs_test.go#L1235-L1302)

- Test successful retrieval with metadata
- Test invalid sharing code (404)
- Test NULL sharing code (404)
- All tests passing ✅

### Phase 5: Documentation ✅ Completed (2026-01-20)

- [x] Regenerate Swagger documentation
- [x] Verify Swagger generation

**Implementation**: Swagger regenerated successfully with `make api-doc`

---

## 🔄 Migration Strategy

### Backward Compatibility

**Impact**: This is a **BREAKING CHANGE** for API consumers.

**Current Response**:
```json
[
  { "id": 1, "item_name": "...", ... },
  // ... items
]
```

**New Response**:
```json
{
  "pack": { ... },
  "contents": [ ... ]
}
```

**Migration Options**:

#### Option 1: Version the API (Recommended for production)
- Create new endpoint: `/public/sharedpack/{sharing_code}` (new structure)
- Keep old endpoint: `/api/sharedlist/{sharing_code}` (old structure)
- Deprecate old endpoint with sunset date

#### Option 2: Direct Breaking Change (OK for development)
- Update existing endpoint with new structure
- Update any frontend clients immediately
- Document breaking change in release notes

**Recommendation**: Since this is a development project, use Option 2 and document the breaking change.

### Database Changes

**None required** - This is purely an API response structure change.

---

## ✅ Acceptance Criteria Summary

- [x] API returns both pack metadata and contents
- [x] Response structure matches specification
- [x] No sensitive data (user_id) exposed
- [x] Error handling works correctly (404 for invalid/unshared packs)
- [x] All tests pass
- [x] Swagger documentation updated and accurate
- [x] Code follows project patterns (context usage, error wrapping)
- [x] Linter passes without errors
---

## 📚 References

- Related spec: [001_optional-pack-sharing.md](001_optional-pack-sharing.md)
- Current implementation: `pkg/packs/packs.go:1590-1618` (SharedList handler)
- Dataset types: `pkg/dataset/dataset.go`
- Architecture guide: [../agents.md](../agents.md)
- Code patterns: [../agents/PATTERNS.md](../agents/PATTERNS.md)

---

## 🤔 Open Questions

1. **Q**: Should we include `updated_at` timestamp in the pack metadata?
   **A**: No - Not needed for public shared pack view

2. **Q**: Should we expose the sharing_code in the response?
   **A**: No - User already has it in the URL, no need to return it

3. **Q**: Should we include a count of items vs relying on array length?
   **A**: No - `pack_items_count` column doesn't exist in database schema. Clients can count items from contents array

4. **Q**: Do we want to support query parameters for filtering contents?
   **A**: No - Out of scope for this spec, not needed for initial implementation

---

## 📅 Timeline

**Target Completion**: ✅ Completed on 2026-01-20
**Dependencies**: None (builds on existing pack sharing feature)

**Final Status**: All phases completed successfully. Breaking change implemented without issues as endpoint was not in use.

---

## Notes

- This enhancement improves UX significantly for shared pack viewers
- Consider frontend impact when implementing breaking changes
