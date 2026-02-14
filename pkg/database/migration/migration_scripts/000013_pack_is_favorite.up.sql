ALTER TABLE "pack" ADD COLUMN "is_favorite" BOOLEAN NOT NULL DEFAULT false;
-- Enforce at most one favorite pack per user at the database level
CREATE UNIQUE INDEX idx_pack_one_favorite_per_user ON pack (user_id) WHERE is_favorite = true;
