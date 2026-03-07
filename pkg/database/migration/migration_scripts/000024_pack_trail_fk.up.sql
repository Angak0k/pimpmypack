ALTER TABLE "pack" ADD COLUMN "trail_id" INT REFERENCES trail(id) ON DELETE SET NULL;

-- Backfill trail_id from existing trail text values
UPDATE pack SET trail_id = t.id FROM trail t WHERE pack.trail = t.name;
