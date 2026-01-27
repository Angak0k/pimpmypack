CREATE INDEX IF NOT EXISTS "idx_inventory_dedup"
ON "inventory" ("user_id", "item_name", "category", "description");
