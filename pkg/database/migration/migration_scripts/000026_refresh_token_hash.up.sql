-- Refresh tokens are now stored as hex-encoded SHA-256 hashes so a database
-- leak cannot be replayed as live sessions. Hash existing plaintext tokens in
-- place: lookups hash the presented token, so current sessions keep working.
-- The WHERE guard makes the migration idempotent: plaintext tokens (base64url)
-- never look like a 64-char hex digest, already-hashed ones always do.
UPDATE refresh_token
SET token = encode(sha256(convert_to(token, 'UTF8')), 'hex')
WHERE token !~ '^[0-9a-f]{64}$';
