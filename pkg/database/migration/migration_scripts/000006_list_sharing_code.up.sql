ALTER TABLE "pack"
ADD COLUMN "sharing_code" TEXT NOT NULL DEFAULT '',
ADD CONSTRAINT "sharing_code_unique" UNIQUE ("sharing_code");