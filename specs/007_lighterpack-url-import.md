# LighterPack URL Import - Specification

**Status**: Merged
**Author**: Claude Agent
**Date**: 2026-03-05
**GitHub Issue**: [#185](https://github.com/Angak0k/pimpmypack/issues/185)
**PR**: [#204](https://github.com/Angak0k/pimpmypack/pull/204)

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
- [x] Endpoint is protected (JWT required)
- [x] Accepts JSON body with `url` field (required)
- [x] Returns pack_id and success message on success
- [x] Returns 400 for invalid/missing URL
- [x] Returns 502 for unreachable URL

**Priority**: High

#### FR2: Validate URL Format

**Description**: Only accept valid LighterPack sharing URLs to prevent SSRF.

**Acceptance Criteria**:
- [x] URL must match pattern: `https://lighterpack.com/r/<alphanumeric>`
- [x] Reject URLs with other domains
- [x] Reject URLs with extra path segments or query parameters
- [x] Return 400 with clear error message for invalid URLs

**Priority**: High

#### FR3: Extract Pack Name and Description

**Description**: Use the LighterPack list name as the pack name and the list description as the pack description.

**Acceptance Criteria**:
- [x] Extract list name from `<h1 class="lpListName">` element
- [x] Use extracted name as pack name (instead of generic "LighterPack Import")
- [x] Extract list description from `<span class="lpListDescription">` element
- [x] Use extracted description as pack description
- [x] Handle missing list name gracefully (fallback to "LighterPack Import")
- [x] Handle missing list description gracefully (fallback to pack name)

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
- [x] All fields extracted correctly from sample pages
- [x] Weight correctly converted from milligrams to grams
- [x] Price correctly parsed with currency symbol detection
- [x] Item URL extracted when present (some items have URLs, some don't)
- [x] Worn and Consumable flags detected by `lpActive` class presence
- [x] Quantity defaults to 1 when not explicitly shown
- [x] Category correctly associated from parent category element

**Priority**: High

#### FR5: Currency Detection

**Description**: Detect currency from the price symbol and store appropriately.

**Acceptance Criteria**:
- [x] `$` prefix -> `USD`
- [x] `€` prefix -> `EUR`
- [x] `£` prefix -> `GBP`
- [x] Unknown currency symbol -> default to `EUR` (project default)

**Priority**: Medium

#### FR6: Reuse Existing Import Logic

**Description**: Use `insertLighterPack()` for database operations.

**Acceptance Criteria**:
- [x] Existing deduplication logic applies (same behavior as CSV import)
- [x] Inventory items created or reused based on matching criteria
- [x] PackContent records created with correct worn/consumable/quantity

**Priority**: High

---

## Design

### Architecture

```
HTTP Request (JSON with URL)
    -> ImportFromLighterPackURL handler (handlers.go)
        -> validateLighterPackURL()      // URL validation (SSRF prevention)
        -> fetchLighterPackPage()        // HTTP GET with context, timeout, redirect checks
        -> parseLighterPackHTML()         // golang.org/x/net/html DOM parsing
        -> insertLighterPack()           // Existing DB logic (reused, in repository.go)
    <- JSON Response (pack_id, message)
```

### HTML Parsing Strategy

Uses `golang.org/x/net/html` (already in `go.mod` as indirect dependency) to parse the DOM tree. Walks the tree to extract structured data.

**Parsing flow**:
1. Parse full HTML into node tree with `html.Parse()`
2. Find `h1.lpListName` -> pack name
3. Find `span.lpListDescription` -> pack description
4. Find all `li.lpCategory` elements
5. For each category, extract `h2.lpCategoryName`
6. For each `li.lpItem` within the category, extract all fields via dedicated extraction functions

**Helper functions** (in `lighterpack.go`):
- `hasClass(node, className)` - check CSS class on element
- `getAttr(node, attrName)` - get attribute value
- `textContent(node)` - recursively collect text content
- `findNodesByClass(node, tag, className)` - find descendants by tag and class

**Item extraction functions** (split out to reduce cognitive complexity):
- `extractItemName(node, item)` - name + URL from `span.lpName`
- `extractItemDescription(node, item)` - from `span.lpDescription`
- `extractItemWeight(node, item)` - from `input.lpMG` value (mg to g)
- `extractItemPrice(node, item)` - from `span.lpPriceCell.lpNumber` with currency detection
- `extractItemQuantity(node, item)` - from `span.lpQtyCell`
- `hasActiveFlag(node, className)` - checks `lpActive` class on `i` elements (worn/consumable)

### URL Validation (SSRF Prevention)

```go
var lighterPackURLPattern = regexp.MustCompile(`^https://lighterpack\.com/r/[a-zA-Z0-9]+$`)
```

Additionally:
- HTTP client with 15 second timeout
- Redirect following restricted to `lighterpack.com` only (max 3 redirects)
- Request created with `http.NewRequestWithContext` for proper context propagation
- Response body limited to 5MB via `io.LimitReader`

### Modifying `insertLighterPack`

Added `packName string` and `packDescription string` parameters to `insertLighterPack()`:
- CSV import handler passes `"LighterPack Import"` for both (backward compatible)
- URL import handler passes the extracted list name and description

Added `Currency` field to `LighterPackItem` struct. `insertLighterPack` uses `item.Currency` with fallback to `"USD"` when empty (CSV import compatibility).

### Edge Cases

1. **Empty list**: Return error "no items found in LighterPack page"
2. **LighterPack down/slow**: HTTP timeout (15s) -> return 502 Bad Gateway
3. **HTML structure changed**: Missing expected elements -> return 422 with descriptive error
4. **Price with no currency symbol**: Default to EUR
5. **Item with no URL**: URL field empty string (same as CSV import)
6. **Item with 0 weight**: Import as-is (valid for some items)
7. **HTML-encoded characters in URLs**: Decoded via `html.UnescapeString()`

---

## Testing Strategy

### Unit Tests (no DB required)

In `pkg/packs/lighterpack_test.go`:

#### URL Validation Tests (`TestValidateLighterPackURL`)
- Valid URL: `https://lighterpack.com/r/oo18ii` -> accepted
- Valid alphanumeric: `https://lighterpack.com/r/ABC123xyz` -> accepted
- Missing scheme: `lighterpack.com/r/oo18ii` -> rejected
- HTTP (not HTTPS): `http://lighterpack.com/r/oo18ii` -> rejected
- Wrong domain: `https://evil.com/r/oo18ii` -> rejected
- Extra path: `https://lighterpack.com/r/oo18ii/extra` -> rejected
- Query params: `https://lighterpack.com/r/oo18ii?foo=bar` -> rejected
- Empty ID: `https://lighterpack.com/r/` -> rejected
- Special chars in ID: `https://lighterpack.com/r/../../etc` -> rejected
- Empty string -> rejected
- User profile URL: `https://lighterpack.com/u/someone` -> rejected

#### HTML Parsing Tests (`TestParseLighterPackHTML`)
- **Full parse test**: USD fixture with 6 items across 4 categories, verifies all fields
- **EUR fixture** (`TestParseLighterPackHTML_EUR`): EUR currency detection and weight conversion
- **Empty page** (`TestParseLighterPackHTML_Empty`): Returns "no items found" error
- **Missing name** (`TestParseLighterPackHTML_MissingName`): Fallback to "LighterPack Import"

#### Price Parsing Tests (`TestParsePrice`)
- `$399.50` -> Price=39950, Currency=USD
- `€100.00` -> Price=10000, Currency=EUR
- `£50.00` -> Price=5000, Currency=GBP
- `$0.00` -> Price=0, Currency=USD
- Empty string -> Price=0, Currency=""
- No symbol `199.99` -> Price=19999, Currency=""

### Test Fixtures

- `pkg/packs/testdata/lighterpack_sample.html` - USD fixture with 4 categories, 6 items (with/without URL, worn, consumable, qty=6)
- `pkg/packs/testdata/lighterpack_eur.html` - EUR fixture with 1 category, 1 item

**Note**: Tests are in the `packs` package alongside `packs_test.go` which has a `TestMain` requiring DB. Unit tests compile and pass when DB is available (`make test`).

---

## Implementation Summary

### Phase 1: Types and Signature Updates [Done]

- [x] Added `ImportFromURLRequest` struct to `pkg/packs/types.go`
- [x] Added `Currency` field to `LighterPackItem` struct
- [x] Added `packName string` and `packDescription string` parameters to `insertLighterPack()` in `pkg/packs/repository.go`
- [x] Updated `insertLighterPack` to use `Currency` field from item (with fallback to `"USD"`)
- [x] Updated CSV import handler call in `pkg/packs/handlers.go` to pass `"LighterPack Import"` for both

### Phase 2: Create `lighterpack.go` [Done]

- [x] Created `pkg/packs/lighterpack.go` with URL import business logic:
  - `validateLighterPackURL()` with SSRF-safe regex
  - `fetchLighterPackPage(ctx, url)` with timeout, redirect protection, body size limit
  - `parseLighterPackHTML(data)` with DOM tree walking
  - `parseLighterPackItem(node, category)` with dedicated extraction functions
  - `parsePrice(s)` with currency symbol detection via switch statement
  - `extractItemURL(node)` with `html.UnescapeString` for HTML entity decoding
  - HTML helper functions: `hasClass`, `getAttr`, `textContent`, `findNodesByClass`

**Note**: `readLineFromCSV()` and `insertLighterPack()` were NOT moved from their original files (kept in `repository.go`) to minimize diff and risk. Only new URL import logic was added to `lighterpack.go`.

### Phase 3: Handler and Route [Done]

- [x] Added `ImportFromLighterPackURL` handler in `pkg/packs/handlers.go`
- [x] Added route in `main.go`: `protected.POST("/importfromlighterpackurl", packs.ImportFromLighterPackURL)`

### Phase 4: Tests [Done]

- [x] Created test fixture HTML files in `pkg/packs/testdata/`
- [x] Wrote URL validation tests (11 cases)
- [x] Wrote HTML parsing tests (4 test functions covering full parse, EUR, empty, missing name)
- [x] Wrote price parsing tests (6 cases)

### Phase 5: Verification [Done]

- [x] `go build ./...` compiles
- [x] `make test` passes (all tests including new ones)
- [x] `make lint` passes (0 issues)
- [x] `make api-test-full` passes (139/139 tests, no regression)
- [x] `make api-doc` regenerated

---

## Security Considerations

### SSRF Prevention
- Strict URL regex validation (only `https://lighterpack.com/r/<alphanumeric>`)
- HTTP client with 15s timeout
- Redirects restricted to `lighterpack.com` domain only (max 3)
- Response body limited to 5MB

### Input Sanitization
- All extracted strings trimmed
- HTML entities decoded safely via `html.UnescapeString()` and `golang.org/x/net/html` parser
- No raw HTML stored in database

### Rate Limiting
- Relies on existing API rate limiting (if any)
- Each import makes one outbound HTTP request

---

## Files Created/Modified

| File | Action | Description |
|------|--------|-------------|
| `pkg/packs/lighterpack.go` | Created | URL validation, HTTP fetch, HTML parsing, helper functions |
| `pkg/packs/lighterpack_test.go` | Created | Unit tests for URL validation, HTML parsing, price parsing |
| `pkg/packs/testdata/lighterpack_sample.html` | Created | USD test fixture (4 categories, 6 items) |
| `pkg/packs/testdata/lighterpack_eur.html` | Created | EUR test fixture (1 category, 1 item) |
| `pkg/packs/handlers.go` | Modified | Added `ImportFromLighterPackURL` handler, updated CSV import call |
| `pkg/packs/repository.go` | Modified | Added `packName`/`packDescription` params and `Currency` field usage to `insertLighterPack` |
| `pkg/packs/types.go` | Modified | Added `Currency` field to `LighterPackItem`, added `ImportFromURLRequest` struct |
| `main.go` | Modified | Added `/importfromlighterpackurl` route |
| `api-doc/docs.go` | Modified | Regenerated swagger docs |
| `api-doc/swagger.json` | Modified | Regenerated swagger docs |
| `api-doc/swagger.yaml` | Modified | Regenerated swagger docs |

---

## References

- CSV import handler: `pkg/packs/handlers.go:ImportFromLighterPack`
- Insert logic: `pkg/packs/repository.go:insertLighterPack`
- LighterPackItem type: `pkg/packs/types.go:150-165`
- Deduplication spec: `specs/003_lighterpack-import-deduplication.md`
- DB Currency enum: `pkg/database/migration/migration_scripts/000002_inventory.up.sql`
