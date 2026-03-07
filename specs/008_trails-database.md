# Trails Database & Admin Management - Specification

**Status**: Draft
**Author**: Claude Agent
**Date**: 2026-03-07
**Issue**: #187

---

## Overview

### Purpose
Move trails from a hardcoded 24-item string allowlist to a database table with country/continent classification, providing admin CRUD+bulk endpoints and a V2 pack-options endpoint with grouped trail hierarchy.

### Goals
- Store trails in a database table with country, continent, distance, and URL metadata
- Provide admin CRUD and bulk endpoints for trail management
- V2 pack-options endpoint returning trails grouped by continent/country
- Maintain V1 backward compatibility (trail name strings still work)

### Non-Goals
- V2 pack create/update endpoints with trail_id (follow-up issue)
- Removing the legacy `trail` TEXT column from pack table (future migration)

---

## Requirements

### Functional Requirements

#### FR1: Trail Database Table
**Description**: Trails are stored in a dedicated table with metadata.

**Acceptance Criteria**:
- [ ] `trail` table with id, name, country, continent, distance_km, url, timestamps
- [ ] Unique index on trail name
- [ ] Seeded with existing 24 trails plus ~50-80 additional curated trails
- [ ] `pack` table gains `trail_id` FK column (nullable, ON DELETE SET NULL)
- [ ] Existing pack trail values are migrated to trail_id via name matching

**Priority**: High

#### FR2: Admin Trail CRUD
**Description**: Admin users can manage trails via REST endpoints.

**Acceptance Criteria**:
- [ ] GET /admin/trails - list all trails
- [ ] GET /admin/trails/:id - get trail by ID
- [ ] POST /admin/trails - create a trail
- [ ] PUT /admin/trails/:id - update a trail
- [ ] DELETE /admin/trails/:id - delete a trail (reject if in use by packs)
- [ ] POST /admin/trails/bulk - bulk create trails
- [ ] DELETE /admin/trails/bulk - bulk delete trails

**Priority**: High

#### FR3: V2 Pack Options
**Description**: New endpoint returns trails grouped by continent and country.

**Acceptance Criteria**:
- [ ] GET /v2/pack-options returns trails grouped: `{continent: {country: [trails]}}`
- [ ] V1 GET /v1/pack-options still returns flat trail name list (from DB)

**Priority**: High

#### FR4: V1 Backward Compatibility
**Description**: Existing V1 pack create/update still accepts trail names.

**Acceptance Criteria**:
- [ ] Pack create/update with trail name resolves to trail_id internally
- [ ] Invalid trail names are rejected with 400 error
- [ ] Pack responses still include trail name string
- [ ] Dual-write: both trail (TEXT) and trail_id (FK) are written

**Priority**: High

---

## Design

### Data Model

#### New Table: trail
```sql
CREATE TABLE "trail" (
    "id"           SERIAL PRIMARY KEY,
    "name"         TEXT NOT NULL,
    "country"      TEXT NOT NULL,
    "continent"    TEXT NOT NULL,
    "distance_km"  INT,
    "url"          TEXT,
    "created_at"   TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at"   TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_trail_name ON trail(name);
```

#### Modified Table: pack
```sql
ALTER TABLE "pack" ADD COLUMN "trail_id" INT REFERENCES trail(id) ON DELETE SET NULL;
-- Backfill existing data
UPDATE pack SET trail_id = t.id FROM trail t WHERE pack.trail = t.name;
```

### API Design

#### Admin Endpoints (all under /api/admin)

| Method | Path | Description |
|--------|------|-------------|
| GET | /trails | List all trails |
| GET | /trails/:id | Get trail by ID |
| POST | /trails | Create trail |
| PUT | /trails/:id | Update trail |
| DELETE | /trails/:id | Delete trail |
| POST | /trails/bulk | Bulk create |
| DELETE | /trails/bulk | Bulk delete |

#### V2 Endpoint

| Method | Path | Description |
|--------|------|-------------|
| GET | /v2/pack-options | Grouped trail hierarchy |

### Package Structure

New package: `pkg/trails/`
- `types.go` - Trail structs and request/response types
- `repository.go` - Database access layer
- `service.go` - Public service functions for cross-package use
- `handlers.go` - Admin CRUD + bulk handlers
- `trails_test.go` - Test suite
- `testdata.go` - Test fixtures

---

## Testing Strategy

### Unit Tests
- Full CRUD cycle for admin endpoints
- Bulk create/delete operations
- V1 pack-options returns DB-driven flat trail list
- V2 pack-options returns grouped hierarchy
- Trail validation in pack create/update
- Delete rejection when trail is in use

### API Tests
- Scenario 008: admin trail CRUD, bulk operations, V1/V2 pack-options, pack create with trail

---

## Implementation Plan

### Phase 1: Database Migrations
- [ ] 000023_trail.up.sql - trail table + seed data
- [ ] 000024_pack_trail_fk.up.sql - add trail_id FK to pack

### Phase 2: Trails Package
- [ ] types.go, repository.go, service.go, handlers.go
- [ ] trails_test.go, testdata.go

### Phase 3: Pack Integration
- [ ] Update packs to use DB trails for validation
- [ ] Add trail_id to pack queries (dual-write)
- [ ] Add V2 pack-options handler

### Phase 4: Route Registration
- [ ] Admin trail routes in main.go
- [ ] V2 pack-options route

### Phase 5: Testing & Verification
- [ ] Unit tests pass
- [ ] API test scenario 008
- [ ] Linter clean
