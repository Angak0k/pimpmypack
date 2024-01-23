CREATE TYPE "UserRole" AS ENUM ('admin', 'standard');
CREATE TYPE "UserStatus" AS ENUM ('active', 'pending', 'inactive');

CREATE TABLE "account"
(
    "id"          SERIAL PRIMARY KEY,
    "username"    TEXT    NOT NULL  UNIQUE,
    "email"       TEXT    NOT NULL,
    "firstname"   TEXT    NOT NULL,
    "lastname"    TEXT    NOT NULL,
    "role"        "UserRole"   NOT NULL DEFAULT E'standard',
    "status"      "UserStatus"    NOT NULL DEFAULT E'pending',
    "created_at"  TIMESTAMP    NOT NULL,
    "updated_at"  TIMESTAMP    NOT NULL
);

