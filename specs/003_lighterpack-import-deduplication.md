# LighterPack Import - Inventory Item Deduplication - Specification

**Status**: Draft
**Author**: Claude Agent
**Date**: 2026-01-20

---

## 📋 Overview

### Purpose

Improve the LighterPack import feature to avoid creating duplicate items in the user's inventory. Currently, importing multiple packs with common items results in duplicate inventory entries, one for each import.

### Problem

In `pkg/packs/packs.go:insertLighterPack()` (lines 1537-1577):
- Each item from the CSV is systematically inserted into the inventory
- No check is performed to see if the item already exists
- Importing two packs with common items creates duplicates in the inventory

**Example scenario**:
1. User imports "Summer Hiking Pack" with item "Water Bottle"
2. User imports "Winter Camping Pack" with the same "Water Bottle"
3. Result: Two "Water Bottle" entries in inventory instead of one shared item

### Goals

- Detect existing inventory items before creating new ones during import
- Reuse existing items when available
- Maintain data integrity and user experience
- Keep import process efficient

### Non-Goals

- Merging existing duplicate items (retroactive deduplication)
- Advanced fuzzy matching or similarity detection
- Manual deduplication UI
- Cross-user item matching

---

## 🎯 Requirements

### Functional Requirements

#### FR1: Check for Existing Items Before Insert
**Description**: Before creating a new inventory item during import, check if an identical item already exists for the user.

**Acceptance Criteria**:
- [ ] Query inventory by UserID + ItemName + Category + Description before insert
- [ ] If item exists, use its ID instead of creating new one
- [ ] If item doesn't exist, create it as before

**Priority**: High

#### FR2: Link Pack to Existing Item
**Description**: When an existing item is found, link the imported pack to that item.

**Acceptance Criteria**:
- [ ] PackContent created with existing item ID
- [ ] Quantity, Worn, and Consumable attributes from CSV applied to PackContent
- [ ] No duplicate items created in inventory

**Priority**: High

#### FR3: Matching Criteria
**Description**: Define clear criteria for determining if two items are identical.

**Acceptance Criteria**:
- [ ] Match on: UserID (exact match)
- [ ] Match on: ItemName (exact match, case-sensitive)
- [ ] Match on: Category (exact match, case-sensitive)
- [ ] Match on: Description (exact match, case-sensitive)
- [ ] Note: Weight, Price are NOT used for matching

**Rationale**:
- Weight/Price may vary over time (updates, different vendors)
- Description may be refined without changing the item identity
- ItemName + Category provide sufficient specificity

**Priority**: High

---

## 🏗️ Design

### Matching Logic

**Criteria for identifying duplicate**:
```
Item A == Item B if:
  - A.UserID == B.UserID AND
  - A.ItemName == B.ItemName AND
  - A.Category == B.Category
  - A.Description == B.Description
```

**Case sensitivity**: Exact match (case-sensitive) to avoid false positives.

**Why not fuzzy matching?**:
- Simpler implementation
- Predictable behavior
- Users can standardize naming conventions
- Future enhancement if needed

### Database Query

New function in `pkg/inventories/inventories.go`:

```go
// FindInventoryItemByAttributes finds an existing inventory item for a user
// by exact match on item_name, category, and description.
// Returns the item if found, nil otherwise.
func FindInventoryItemByAttributes(ctx context.Context, userID uint, itemName, category, description string) (*dataset.Inventory, error) {
    query := `
        SELECT id, user_id, item_name, category, description, weight, url, price, currency, created_at, updated_at
        FROM inventory
        WHERE user_id = $1 AND item_name = $2 AND category = $3 AND description = $4
        LIMIT 1
    `

    var item dataset.Inventory
    err := database.DB().QueryRowContext(ctx, query, userID, itemName, category, description).Scan(
        &item.ID,
        &item.UserID,
        &item.ItemName,
        &item.Category,
        &item.Description,
        &item.Weight,
        &item.URL,
        &item.Price,
        &item.Currency,
        &item.CreatedAt,
        &item.UpdatedAt,
    )

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil  // Not an error, just not found
        }
        return nil, fmt.Errorf("failed to query inventory: %w", err)
    }

    return &item, nil
}
```

### Modified Import Logic

Update `insertLighterPack()` in `pkg/packs/packs.go`:

```go
func insertLighterPack(lp *dataset.LighterPack, userID uint) error {
    if lp == nil {
        return errors.New("payload is empty")
    }

    // Create new pack
    var newPack dataset.Pack
    newPack.UserID = userID
    newPack.PackName = "LighterPack Import"
    newPack.PackDescription = "LighterPack Import"
    err := insertPack(&newPack)
    if err != nil {
        return err
    }

    ctx := context.Background()

    // Insert content in new pack with insertPackContent
    for _, item := range *lp {
        var itemID uint

        // Check if item already exists in inventory
        existingItem, err := inventories.FindInventoryItemByAttributes(
            ctx,
            userID,
            item.ItemName,
            item.Category,
            item.Desc,
        )
        if err != nil {
            return fmt.Errorf("failed to check for existing item: %w", err)
        }

        if existingItem != nil {
            // Item exists, reuse it
            itemID = existingItem.ID
        } else {
            // Item doesn't exist, create it
            var i dataset.Inventory
            i.UserID = userID
            i.ItemName = item.ItemName
            i.Category = item.Category
            i.Description = item.Desc
            i.Weight = item.Weight
            i.URL = item.URL
            i.Price = item.Price
            i.Currency = "USD"
            err := inventories.InsertInventory(ctx, &i)
            if err != nil {
                return err
            }
            itemID = i.ID
        }

        // Create PackContent with the item (existing or new)
        var pc dataset.PackContent
        pc.PackID = newPack.ID
        pc.ItemID = itemID
        pc.Quantity = item.Qty
        pc.Worn = item.Worn
        pc.Consumable = item.Consumable
        err = insertPackContent(&pc)
        if err != nil {
            return err
        }
    }

    return nil
}
```

### Edge Cases

1. **Empty item name**: Skip item or return error
2. **Empty category**: Use empty string as valid category value
3. **Empty description**: Use empty string as valid description value (will match other items with empty description)
4. **Concurrent imports**: Database handles uniqueness, but deduplication is best-effort
5. **Case variations** ("Water Bottle" vs "water bottle"): Treated as different items (exact match)

---

## 🧪 Testing Strategy

### Unit Tests

#### Test: Existing Item Reused
```go
// Setup: Create item "Water Bottle" in inventory
// Action: Import pack with "Water Bottle"
// Assert: No new inventory item created, PackContent links to existing item
```

#### Test: New Item Created
```go
// Setup: Empty inventory
// Action: Import pack with "Tent"
// Assert: New inventory item created, PackContent links to new item
```

#### Test: Multiple Packs Same Item
```go
// Setup: Empty inventory
// Action: Import pack A with "Headlamp", then import pack B with "Headlamp"
// Assert: Only one "Headlamp" in inventory, two PackContents link to it
```

#### Test: Case Sensitivity
```go
// Setup: Item "Water Bottle" exists
// Action: Import pack with "water bottle"
// Assert: New item created (different case)
```

#### Test: Different Categories
```go
// Setup: Item "Cord" in category "Shelter" exists
// Action: Import pack with "Cord" in category "Electronics"
// Assert: New item created (different category)
```

#### Test: Different Descriptions
```go
// Setup: Item "Tent" with description "Shelter" exists
// Action: Import pack with "Tent" with description "Tent"
// Assert: New item created (different description)
```

#### Test: Same Name/Category, Different Description
```go
// Setup: Item "Water Bottle" / "Water" / "1L bottle" exists
// Action: Import pack with "Water Bottle" / "Water" / "500ml bottle"
// Assert: New item created (different description)
```

### Integration Tests

- Import real CSV file (GR54.csv) twice
- Verify inventory count equals unique items, not total items
- Verify both packs link to shared items correctly

---

## 📝 Implementation Plan

### Phase 1: Create Database Index
- [ ] Create migration file for index on (user_id, item_name, category, description)
- [ ] Test migration up/down
- [ ] Apply migration to database

### Phase 2: Add Lookup Function
- [ ] Create `FindInventoryItemByAttributes` in `pkg/inventories/inventories.go`
- [ ] Export the function (capitalize first letter)
- [ ] Add context parameter for proper request context propagation
- [ ] Add parameters: userID, itemName, category, description

### Phase 3: Update Import Logic
- [ ] Modify `insertLighterPack` to check for existing items
- [ ] Call `FindInventoryItemByAttributes` with all 4 criteria
- [ ] Use existing item ID when found
- [ ] Create new item only when not found
- [ ] Link PackContent to correct item ID

### Phase 4: Add Tests
- [ ] Unit tests for `FindInventoryItemByAttributes`
- [ ] Test for existing item reused
- [ ] Test for new item created
- [ ] Test for different descriptions
- [ ] Update `TestImportFromLighterPack` if exists
- [ ] Integration test: import same CSV twice

### Phase 5: Documentation
- [ ] Update Swagger docs if needed
- [ ] Add comments explaining deduplication logic

---

## 🔄 Alternative Approaches Considered

### Alternative 1: Fuzzy Matching
Use similarity algorithms (Levenshtein distance) to match items with minor name variations.

**Pros**:
- More flexible (handles typos, variations)
- Better UX for casual users

**Cons**:
- Complex implementation
- Risk of false positives (merging unrelated items)
- Performance impact on large inventories
- Difficult to explain behavior to users

**Decision**: Rejected for v1. Exact match is simpler and more predictable.

### Alternative 2: Match on Name Only
Ignore category when matching items.

**Pros**:
- Simpler query
- More aggressive deduplication

**Cons**:
- False positives (e.g., "Cord" for shelter vs "Cord" for electronics)
- Category is valuable metadata

**Decision**: Rejected. Category provides important context.

### Alternative 3: Hash-based Matching
Create hash of item attributes for matching.

**Pros**:
- Flexible criteria (can include multiple fields)
- Fast lookup with index

**Cons**:
- Requires migration to add hash column
- Changes to attributes would require rehashing
- More complex than direct query

**Decision**: Rejected. Direct query is sufficient for current scale.

---

## 🔒 Security Considerations

### User Isolation
- CRITICAL: Always filter by UserID to prevent cross-user item exposure
- The `FindInventoryItemByNameAndCategory` function MUST include UserID in WHERE clause

### SQL Injection
- Use parameterized queries (already standard in codebase)
- No string concatenation for SQL

---

## 📊 Impact Analysis

### Database Impact
- One additional SELECT query per item during import
- Query is simple and uses indexed columns (user_id, item_name, category should be indexed)
- Minimal performance impact

### Breaking Changes
- None. The API contract remains unchanged
- Existing imports continue to work
- Only the internal behavior changes (fewer duplicates created)

### User Experience
- **Positive**: Cleaner inventory, no manual deduplication needed
- **Neutral**: Transparent change, users may not notice
- **Risk**: If users relied on separate items per pack (unlikely), behavior changes

---

## ✅ Acceptance Criteria Summary

- [ ] Importing same item in multiple packs creates only one inventory entry
- [ ] Each pack correctly links to the shared inventory item
- [ ] Matching is based on UserID + ItemName + Category (exact match)
- [ ] New items are created when no match is found
- [ ] All tests pass
- [ ] No breaking changes to API
- [ ] Performance remains acceptable (< 100ms per item lookup)

---

## 📚 References

- Current implementation: `pkg/packs/packs.go:1537-1577` (insertLighterPack)
- CSV format example: `specs/GR54.csv`
- Inventory data model: `pkg/dataset/dataset.go:23-35`

---

## 📌 Open Questions

1. **Should we add an index on (user_id, item_name, category)?**
   - Current implementation may not have this index
   - Would improve lookup performance
   - Decision: Check existing indexes, add if missing

2. **Should we log when items are reused vs created?**
   - Useful for debugging and user visibility
   - Could add to response: `{"created": 5, "reused": 3}`
   - Decision: TBD, nice-to-have for v2

3. **Should we handle case-insensitive matching?**
   - Would require LOWER() in query
   - Adds complexity
   - Decision: Not for v1, exact match is clearer

---

## Notes

- This is an enhancement, not a bugfix
- Backward compatible change
- No migration needed (no schema changes)
- Can be implemented and tested independently
