-- Restore NOT NULL constraint with DEFAULT (rollback)
UPDATE "pack" SET "sharing_code" = '' WHERE "sharing_code" IS NULL;
ALTER TABLE "pack" ALTER COLUMN "sharing_code" SET DEFAULT '';
ALTER TABLE "pack" ALTER COLUMN "sharing_code" SET NOT NULL;
