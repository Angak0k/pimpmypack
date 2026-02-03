# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Refresh Token Authentication** - Automatic session extension without re-login
  - New endpoint: `POST /auth/refresh` - Exchange refresh token for new access token
  - "Remember me" option for extended sessions (up to 30 days)
  - Automatic cleanup of expired tokens (configurable interval)
- **Security Enhancements**
  - Rate limiting on `/auth/refresh` endpoint (10 requests/minute per IP)
  - Structured audit logging for all authentication events (JSON format)
  - Comprehensive error sanitization (no internal errors exposed to clients)
- **Documentation**
  - Frontend Integration Guide ([docs/frontend-integration.md](docs/frontend-integration.md))
  - Complete examples for React, Vue, and vanilla JavaScript
  - Security best practices for token storage and management

### Changed
- **Access Token Lifetime**: Reduced from 60 minutes to 15 minutes for better security
- **Login Endpoint**: Now returns both `access_token` and `refresh_token`
  - Response includes `expires_in` and `refresh_expires_in` fields
  - Added optional `remember_me` field in login request
- **Session Management**: Improved user experience with automatic token refresh

### Security
- Access tokens now short-lived (15 minutes) to reduce exposure window
- Refresh tokens enable seamless session extension
- Rate limiting prevents brute force attacks on refresh endpoint
- All authentication events logged with structured audit trail
- Error messages sanitized to prevent information leakage

### Developer Notes
- **Backward Compatible**: Existing clients continue to work without changes
- **Migration**: Database migration `000006_refresh_tokens` runs automatically
- **Configuration**: New environment variables for refresh token settings
  - `REFRESH_TOKEN_DAYS` - Default refresh token lifetime (default: 1 day)
  - `REFRESH_TOKEN_REMEMBER_ME_DAYS` - Remember me lifetime (default: 30 days)
  - `REFRESH_TOKEN_CLEANUP_INTERVAL_HOURS` - Cleanup job interval (default: 24 hours)
  - `REFRESH_RATE_LIMIT_REQUESTS` - Rate limit (default: 10)
  - `REFRESH_RATE_LIMIT_WINDOW_MINUTES` - Rate limit window (default: 1 minute)
- See `.env.sample` for complete configuration options
