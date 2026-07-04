-- rotated_at marks tokens revoked by rotation (superseded by a successor),
-- as opposed to tokens revoked by logout / password change / admin action.
-- It drives the reuse-detection grace window on /auth/refresh.
ALTER TABLE refresh_token ADD COLUMN rotated_at TIMESTAMPTZ;
