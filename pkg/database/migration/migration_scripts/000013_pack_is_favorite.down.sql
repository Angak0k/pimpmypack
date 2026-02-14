DROP INDEX IF EXISTS idx_pack_one_favorite_per_user;
ALTER TABLE "pack" DROP COLUMN IF EXISTS "is_favorite";
