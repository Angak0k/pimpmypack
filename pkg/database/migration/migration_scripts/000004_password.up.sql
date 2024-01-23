CREATE TABLE "password"
(
    "id"            SERIAL PRIMARY KEY,
    "user_id"       INT     NOT NULL,
    "password"      TEXT    NOT NULL,
    "last_password" TEXT,
    "updated_at"    TIMESTAMP    NOT NULL,
    CONSTRAINT fk_account
        FOREIGN KEY(user_id)
            REFERENCES "account"(id)
            ON DELETE CASCADE
);