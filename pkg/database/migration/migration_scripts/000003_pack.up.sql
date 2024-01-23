CREATE TABLE "pack"
(
    "id"                SERIAL PRIMARY KEY,
    "user_id"           INT    NOT NULL,
    "pack_name"         TEXT    NOT NULL,
    "pack_description"  TEXT,
    "created_at"        TIMESTAMP    NOT NULL,
    "updated_at"        TIMESTAMP    NOT NULL,
    CONSTRAINT fk_account
        FOREIGN KEY(user_id)
            REFERENCES "account"(id)
            ON DELETE CASCADE
);

CREATE TABLE "pack_content"
(
    "id"                SERIAL PRIMARY KEY,
    "pack_id"           INT    NOT NULL,
    "item_id" INT NOT NULL,
    "quantity" INT,
    "worn" BOOLEAN,
    "consumable" BOOLEAN,
    "created_at" TIMESTAMP NOT NULL,
    "updated_at" TIMESTAMP NOT NULL,
    UNIQUE ("pack_id", "item_id"),
    CONSTRAINT fk_pack
        FOREIGN KEY ("pack_id")
            REFERENCES "pack"(id)
            ON DELETE CASCADE,
    CONSTRAINT fk_item
        FOREIGN KEY ("item_id")
            REFERENCES "inventory"(id)
            ON DELETE CASCADE
);