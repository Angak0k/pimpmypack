-- Refresh tokens are now stored as hex-encoded SHA-256 hashes so a database
-- leak cannot be replayed as live sessions. Hash existing plaintext tokens in
-- place: lookups hash the presented token, so current sessions keep working.
UPDATE refresh_token SET token = encode(sha256(convert_to(token, 'UTF8')), 'hex');
