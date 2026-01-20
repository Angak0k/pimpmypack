-- Remove NOT NULL constraint and DEFAULT value from sharing_code
ALTER TABLE "pack" ALTER COLUMN "sharing_code" DROP NOT NULL;
ALTER TABLE "pack" ALTER COLUMN "sharing_code" DROP DEFAULT;

-- Set empty strings to NULL (cleanup existing data)
UPDATE "pack" SET "sharing_code" = NULL WHERE "sharing_code" = '';
