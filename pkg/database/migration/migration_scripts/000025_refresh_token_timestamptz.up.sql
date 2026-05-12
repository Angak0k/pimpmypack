-- Purge existing rows: TIMESTAMP (no tz) stored wall-clock values whose
-- timezone is ambiguous, so converting them reliably is not possible.
-- Refresh tokens are short-lived; users will simply re-authenticate.
TRUNCATE TABLE refresh_token;

ALTER TABLE refresh_token
    ALTER COLUMN expires_at TYPE TIMESTAMPTZ USING expires_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN last_used_at TYPE TIMESTAMPTZ USING last_used_at AT TIME ZONE 'UTC';
