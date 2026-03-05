# LighterPack URL Import - Specification

**Status**: Draft
**Author**: Claude Agent
**Date**: 2026-03-05
**GitHub Issue**: [#185](https://github.com/Angak0k/pimpmypack/issues/185)

---

## Overview

### Purpose

Allow users to import a pack from a LighterPack sharing URL (e.g., `https://lighterpack.com/r/oo18ii`) instead of exporting a CSV file and uploading it manually.

### Problem

The current LighterPack import requires users to:
1. Open their LighterPack list
2. Export as CSV
3. Upload the CSV file via `/api/v1/importfromlighterpack`

This is a multi-step process. LighterPack sharing URLs are commonly shared in hiking communities. Importing directly from a URL simplifies the process to a single API call.

### Goals

- Import a pack from a LighterPack sharing URL in a single request
- Extract all available item data (name, description, category, weight, price, URL, quantity, worn/consumable flags)
- Reuse existing `insertLighterPack()` logic (deduplication, inventory creation, pack content linking)
- Use the list name from LighterPack as the pack name

### Non-Goals

- Supporting LighterPack user profile URLs (only sharing URLs `/r/<id>`)
- Handling authentication-protected LighterPack lists
- Periodic sync with LighterPack (this is a one-time import)
- Importing LighterPack images

---

## Requirements

### Functional Requirements

#### FR1: Accept LighterPack Sharing URL

**Description**: New endpoint accepts a JSON body with a LighterPack sharing URL.

**Endpoint**: `POST /api/v1/importfromlighterpackurl`

**Request**:
```json
{
  "url": "https://lighterpack.com/r/oo18ii"
}
```

**Response** (same format as CSV import):
```json
{
  "message": "LighterPack URL imported successfully",
  "pack_id": 42
}
```

**Acceptance Criteria**:
- [ ] Endpoint is protected (JWT required)
- [ ] Accepts JSON body with `url` field (required)
- [ ] Returns pack_id and success message on success
- [ ] Returns 400 for invalid/missing URL
- [ ] Returns appropriate error for unreachable URL

**Priority**: High

#### FR2: Validate URL Format

**Description**: Only accept valid LighterPack sharing URLs to prevent SSRF.

**Acceptance Criteria**:
- [ ] URL must match pattern: `https://lighterpack.com/r/<alphanumeric>`
- [ ] Reject URLs with other domains
- [ ] Reject URLs with extra path segments or query parameters
- [ ] Return 400 with clear error message for invalid URLs

**Priority**: High

#### FR3: Extract Pack Name and Description

**Description**: Use the LighterPack list name as the pack name and the list description as the pack description.

**Acceptance Criteria**:
- [ ] Extract list name from `<h1 class="lpListName">` element
- [ ] Use extracted name as pack name (instead of generic "LighterPack Import")
- [ ] Extract list description from `<span class="lpListDescription">` element
- [ ] Use extracted description as pack description
- [ ] Handle missing list name gracefully (fallback to "LighterPack Import")
- [ ] Handle missing list description gracefully (fallback to pack name)

**Priority**: Medium

#### FR4: Extract Items with Full Details

**Description**: Parse each item from the HTML with all available attributes.

**Fields to extract per item**:

| Field | HTML Source | Conversion |
|-------|-----------|------------|
| ItemName | `span.lpName` text content | Trim whitespace |
| Category | Parent `h2.lpCategoryName` text | Trim whitespace |
| Desc | `span.lpDescription` text | Trim whitespace |
| Weight | `input.lpMG` value attribute | Divide by 1000 (milligrams to grams) |
| URL | `a.lpHref` href inside `span.lpName` | HTML-unescape (`&#x2F;` -> `/`) |
| Price | `span.lpPriceCell.lpNumber` text | Strip currency symbol, parse float * 100 (cents) |
| Currency | Currency symbol prefix in price | `$` -> USD, `€` -> EUR, `£` -> GBP |
| Qty | `span.lpQtyCell` text content | Parse as int, default 1 |
| Worn | `i.lpWorn` has class `lpActive` | Boolean |
| Consumable | `i.lpConsumable` has class `lpActive` | Boolean |

**Acceptance Criteria**:
- [ ] All fields extracted correctly from sample pages
- [ ] Weight correctly converted from milligrams to grams
- [ ] Price correctly parsed with currency symbol detection
- [ ] Item URL extracted when present (some items have URLs, some don't)
- [ ] Worn and Consumable flags detected by `lpActive` class presence
- [ ] Quantity defaults to 1 when not explicitly shown
- [ ] Category correctly associated from parent category element

**Priority**: High

#### FR5: Currency Detection

**Description**: Detect currency from the price symbol and store appropriately.

**Acceptance Criteria**:
- [ ] `$` prefix -> `USD`
- [ ] `€` prefix -> `EUR`
- [ ] `£` prefix -> `GBP`
- [ ] Unknown currency symbol -> default to `EUR` (project default)

**Priority**: Medium

#### FR6: Reuse Existing Import Logic

**Description**: Use `insertLighterPack()` for database operations.

**Acceptance Criteria**:
- [ ] Existing deduplication logic applies (same behavior as CSV import)
- [ ] Inventory items created or reused based on matching criteria
- [ ] PackContent records created with correct worn/consumable/quantity

**Priority**: High

---

## Design

### Architecture

```
HTTP Request (JSON with URL)
    -> ImportFromLighterPackURL handler
        -> validateLighterPackURL()      // URL validation (SSRF prevention)
        -> fetchLighterPackPage()        // HTTP GET with timeout
        -> parseLighterPackHTML()         // golang.org/x/net/html parsing
        -> insertLighterPack()           // Existing DB logic (reused)
    <- JSON Response (pack_id, message)
```

### HTML Parsing Strategy

Use `golang.org/x/net/html` (already in `go.mod` as indirect dependency) to parse the DOM tree. Walk the tree to extract structured data.

**Parsing flow**:
1. Parse full HTML into node tree with `html.Parse()`
2. Find `h1.lpListName` -> pack name
3. Find `span.lpListDescription` -> pack description
4. Find all `li.lpCategory` elements
5. For each category, extract `h2.lpCategoryName`
6. For each `li.lpItem` within the category, extract all fields

**Helper functions**:
- `hasClass(node, className)` - check CSS class on element
- `getAttr(node, attrName)` - get attribute value
- `textContent(node)` - recursively collect text content
- `findNodesByClass(node, className)` - find descendants by class

### URL Validation (SSRF Prevention)

```go
var lighterPackURLPattern = regexp.MustCompile(`^https://lighterpack\.com/r/[a-zA-Z0-9]+$`)
```

Additionally:
- Use HTTP client with timeout (15 seconds)
- Disable redirect following (or validate redirect target)

### Modifying `insertLighterPack`

Add `packName string` and `packDescription string` parameters to `insertLighterPack()`:
- CSV import handler passes `"LighterPack Import"` for both (backward compatible)
- URL import handler passes the extracted list name and description

**Note**: `LighterPackItem` struct doesn't currently have a `Currency` field. We need to either:
- (a) Add a `Currency` field to `LighterPackItem` and update `insertLighterPack` to use it
- (b) Pass currency separately to `insertLighterPack`

Option (a) is cleaner since currency is a per-item attribute (all items from one import share the same currency, but conceptually it belongs to the item).

### Edge Cases

1. **Empty list**: Return error "no items found in LighterPack page"
2. **LighterPack down/slow**: HTTP timeout (15s) -> return 502 Bad Gateway
3. **HTML structure changed**: Missing expected elements -> return 422 with descriptive error
4. **Price with no currency symbol**: Default to EUR
5. **Item with no URL**: URL field empty string (same as CSV import)
6. **Item with 0 weight**: Import as-is (valid for some items)
7. **HTML-encoded characters in URLs**: Decode `&#x2F;` -> `/`, etc.

---

## Testing Strategy

### Unit Tests (no DB required)

In `pkg/packs/lighterpack_url_test.go`:

#### URL Validation Tests
- Valid URL: `https://lighterpack.com/r/oo18ii` -> accepted
- Missing scheme: `lighterpack.com/r/oo18ii` -> rejected
- HTTP (not HTTPS): `http://lighterpack.com/r/oo18ii` -> rejected
- Wrong domain: `https://evil.com/r/oo18ii` -> rejected
- Extra path: `https://lighterpack.com/r/oo18ii/extra` -> rejected
- Query params: `https://lighterpack.com/r/oo18ii?foo=bar` -> rejected
- Empty ID: `https://lighterpack.com/r/` -> rejected
- Special chars in ID: `https://lighterpack.com/r/../../etc` -> rejected

#### HTML Parsing Tests
- **Full parse test**: Use saved HTML fixture, verify all items extracted correctly
- **Pack name extraction**: Verify list name extracted from h1
- **Pack description extraction**: Verify list description extracted from span.lpListDescription
- **Weight conversion**: `lpMG=33000` -> Weight=33 grams
- **Price parsing**: `$399.50` -> Price=39950, Currency=USD
- **EUR price**: `€100.00` -> Price=10000, Currency=EUR
- **GBP price**: `£50.00` -> Price=5000, Currency=GBP
- **Worn flag**: Item with `lpWorn lpActive` -> Worn=true
- **Consumable flag**: Item with `lpConsumable lpActive` -> Consumable=true
- **Quantity**: `<span class="lpQtyCell" qty8>8</span>` -> Qty=8
- **Item URL**: `<a class="lpHref" href="https:&#x2F;&#x2F;example.com">` -> URL=`https://example.com`
- **Item without URL**: Name directly in span -> URL=""
- **Empty page**: Returns error

### Test Fixtures

Save a representative HTML snippet in `pkg/packs/testdata/lighterpack_sample.html` containing:
- A list name and description
- 2-3 categories
- Items with various attributes (with/without URL, worn, consumable, different quantities)
- Mix of USD and EUR items (separate fixtures)

### Integration Tests

If feasible with test infrastructure:
- Full handler test with mocked HTTP server serving HTML fixture
- Verify pack and inventory items created in DB

---

## Implementation Plan

### Phase 1: Types and Signature Updates

- [ ] Add `ImportFromURLRequest` struct to `pkg/packs/types.go`
- [ ] Add `Currency` field to `LighterPackItem` struct
- [ ] Add `packName string` and `packDescription string` parameters to `insertLighterPack()` in `pkg/packs/repository.go`
- [ ] Update `insertLighterPack` to use `Currency` field from item (with fallback to `"USD"` for CSV import compatibility)
- [ ] Update CSV import handler call in `pkg/packs/handlers.go` to pass `"LighterPack Import"` for both name and description

### Phase 2: Create `lighterpack.go` (business logic only)

- [ ] Create `pkg/packs/lighterpack.go` with all LighterPack business logic:
  - Move `readLineFromCSV()` from `repository.go`
  - Move `insertLighterPack()` from `repository.go`
  - Add `validateLighterPackURL()`
  - Add `fetchLighterPackPage()` with timeout and SSRF protection
  - Add `parseLighterPackHTML()` with DOM tree walking
  - Add HTML helper functions (hasClass, getAttr, textContent, findNodesByClass)
  - Add currency detection from price symbol
  - Add item URL extraction (with HTML unescaping)
- [ ] Move existing CSV import tests from `packs_test.go` to `pkg/packs/lighterpack_test.go`

### Phase 3: Handler and Route

- [ ] Add `ImportFromLighterPackURL` handler in `pkg/packs/handlers.go`
- [ ] Add route in `main.go`: `protected.POST("/importfromlighterpackurl", packs.ImportFromLighterPackURL)`

### Phase 4: Tests

- [ ] Create test fixture HTML file in `pkg/packs/testdata/`
- [ ] Write URL validation tests
- [ ] Write HTML parsing tests
- [ ] Verify all edge cases covered

### Phase 5: Verification

- [ ] `go build ./...` compiles
- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] `make api-test` passes (no regression)
- [ ] Manual test with real LighterPack URLs

---

## Security Considerations

### SSRF Prevention
- Strict URL regex validation (only `https://lighterpack.com/r/<alphanumeric>`)
- HTTP client with timeout
- No redirect following to external domains

### Input Sanitization
- All extracted strings trimmed
- HTML entities decoded safely via `golang.org/x/net/html` parser
- No raw HTML stored in database

### Rate Limiting
- Relies on existing API rate limiting (if any)
- Each import makes one outbound HTTP request

---

## Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| `pkg/packs/lighterpack.go` | Create | LighterPack business logic: CSV parsing, URL validation, HTTP fetch, HTML parsing, DB insert |
| `pkg/packs/lighterpack_test.go` | Create | All LighterPack tests (moved CSV tests + new URL/HTML tests) |
| `pkg/packs/testdata/lighterpack_sample.html` | Create | Test fixture |
| `pkg/packs/handlers.go` | Modify | Add `ImportFromLighterPackURL` handler |
| `pkg/packs/repository.go` | Modify | Remove `readLineFromCSV`, `insertLighterPack` (moved to lighterpack.go) |
| `pkg/packs/packs_test.go` | Modify | Remove CSV import tests (moved to lighterpack_test.go) |
| `pkg/packs/types.go` | Modify | Add request struct, Currency field |
| `main.go` | Modify | Add route |

---

## References

- Current CSV import handler: `pkg/packs/handlers.go:1185-1258`
- Current insert logic: `pkg/packs/repository.go:729-794`
- LighterPackItem type: `pkg/packs/types.go:150-164`
- Deduplication spec: `specs/003_lighterpack-import-deduplication.md`
- DB Currency enum: `pkg/database/migration/migration_scripts/000002_inventory.up.sql`
