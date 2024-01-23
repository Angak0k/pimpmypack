CREATE TYPE "WeightUnit" AS ENUM ('METRIC', 'IMPERIAL');
CREATE TYPE "Currency" AS ENUM ('USD', 'GBP', 'EUR');

CREATE TABLE "inventory"
(
    "id"          SERIAL PRIMARY KEY,
    "user_id"     INT    NOT NULL,
    "item_name"   TEXT    NOT NULL,
    "category"    TEXT,
    "description" TEXT,
    "weight"      INT,
    "weight_unit" "WeightUnit"  NOT NULL DEFAULT E'METRIC',
    "url"         TEXT,
    "price"       INT,
    "currency"   "Currency"  DEFAULT E'EUR',
    "created_at"  TIMESTAMP    NOT NULL,
    "updated_at"  TIMESTAMP    NOT NULL,
    CONSTRAINT fk_account
        FOREIGN KEY(user_id)
            REFERENCES "account"(id)
            ON DELETE CASCADE
);

