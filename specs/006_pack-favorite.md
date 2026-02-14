# Pack Favorite - Specification

**Status**: Draft
**Author**: Claude Agent
**Date**: 2026-02-14
**Last Updated**: 2026-02-14

---

## Overview

### Purpose
Allow users to mark one pack as their favorite. Only one pack per user can be the favorite at a time - setting a new favorite automatically unfavorites the previous one.

### Goals
- Allow users to mark a pack as their favorite
- Enforce single-favorite-per-user constraint
- Keep favorite status private (not exposed in shared pack views)

### Non-Goals
- Multiple favorites per user
- Favorite ordering/ranking

---

## Requirements

### Functional Requirements

#### FR1: Single Favorite Per User
**Description**: Only one pack per user can be marked as favorite at any time. Setting a new favorite automatically clears the previous one.

**Acceptance Criteria**:
- [ ] Only one pack per user has `is_favorite = true` (enforced by partial unique index at DB level)
- [ ] Setting a new favorite unfavorites the previous one atomically (transaction)
- [ ] New packs default to `is_favorite = false`

**Priority**: High

#### FR2: Favorite Endpoint
**Description**: Users can favorite a pack via a dedicated POST endpoint.

**Acceptance Criteria**:
- [ ] POST `/v1/mypack/:id/favorite` marks the pack as favorite
- [ ] Only pack owner can favorite their pack
- [ ] Operation is idempotent (favoriting an already-favorited pack is a no-op)
- [ ] Previous favorite is automatically cleared

**Priority**: High

#### FR3: Unfavorite Endpoint
**Description**: Users can unfavorite a pack via a dedicated DELETE endpoint.

**Acceptance Criteria**:
- [ ] DELETE `/v1/mypack/:id/favorite` removes favorite status
- [ ] Only pack owner can unfavorite their pack
- [ ] Operation is idempotent (unfavoriting a non-favorite pack is a no-op)

**Priority**: High

#### FR4: No Leaking in Shared Views
**Description**: The `is_favorite` field must not be exposed in shared/public pack views.

**Acceptance Criteria**:
- [ ] `SharedPackInfo` struct does not include `is_favorite`
- [ ] Public shared pack endpoint does not return `is_favorite`

**Priority**: High

---

## Design

### Data Model

#### Database Schema

**Migration: 000013_pack_is_favorite.up.sql**
```sql
ALTER TABLE "pack" ADD COLUMN "is_favorite" BOOLEAN NOT NULL DEFAULT false;
-- Enforce at most one favorite pack per user at the database level
CREATE UNIQUE INDEX idx_pack_one_favorite_per_user ON pack (user_id) WHERE is_favorite = true;
```

**Migration: 000013_pack_is_favorite.down.sql**
```sql
DROP INDEX IF EXISTS idx_pack_one_favorite_per_user;
ALTER TABLE "pack" DROP COLUMN IF EXISTS "is_favorite";
```

#### Modified Types
```go
type Pack struct {
    // ... existing fields ...
    SharingCode *string   `json:"sharing_code,omitempty"`
    IsFavorite  bool      `json:"is_favorite"`
    HasImage    bool      `json:"has_image"`
    // ... existing fields ...
}
```

### API Design

#### POST /api/v1/mypack/:id/favorite
**Description**: Mark a pack as favorite (idempotent, auto-unfavorites previous)

**Response** (200 OK):
```json
{"message": "Pack favorited successfully"}
```

**Error Responses**: 400, 401, 403, 404, 500

#### DELETE /api/v1/mypack/:id/favorite
**Description**: Remove favorite status from a pack (idempotent)

**Response** (200 OK):
```json
{"message": "Pack unfavorited successfully"}
```

**Error Responses**: 400, 401, 403, 404, 500

---

## Testing Strategy

### Unit Tests
- `TestFavoriteMyPack`: success, idempotent, switch favorite, forbidden, not found
- `TestUnfavoriteMyPack`: success, idempotent, forbidden, not found
- Verify new packs default to `is_favorite = false`

### API Tests
- Scenario 002: favorite, verify, idempotent, unfavorite, verify, idempotent

---

## Implementation Plan

### Phase 1: Database Migration
- [ ] Create migration files (000013)

### Phase 2: Data Model
- [ ] Add `IsFavorite` to `Pack` struct

### Phase 3: Repository
- [ ] Update SELECT queries (3 functions)
- [ ] Add `favoritePackByID` and `unfavoritePackByID`

### Phase 4: Handlers
- [ ] Add `FavoriteMyPack` and `UnfavoriteMyPack`
- [ ] Update admin PUT handler to preserve `IsFavorite`

### Phase 5: Routes
- [ ] Register routes in `main.go`

### Phase 6: Testing
- [ ] Unit tests
- [ ] API test scenario updates
