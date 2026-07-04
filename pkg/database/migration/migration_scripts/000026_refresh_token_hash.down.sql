-- Hashing is one-way; plaintext tokens cannot be restored. Drop all rows,
-- users will simply re-authenticate (same approach as migration 000025).
TRUNCATE TABLE refresh_token;
