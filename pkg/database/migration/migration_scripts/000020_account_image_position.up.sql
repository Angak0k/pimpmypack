ALTER TABLE "account"
    ADD COLUMN "image_position_x" INTEGER NOT NULL DEFAULT 50 CHECK ("image_position_x" BETWEEN 0 AND 100),
    ADD COLUMN "image_position_y" INTEGER NOT NULL DEFAULT 50 CHECK ("image_position_y" BETWEEN 0 AND 100);
