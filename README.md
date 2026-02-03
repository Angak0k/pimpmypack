# ⛰️ pimpmypack

> PimpMyPack is a set of backend APIs dedicated to CRUD operations on hiking equipment inventories and packing lists.  
> It should be used in conjunction with any frontend candidates.  
> It could replace [Lighterpack](https://lighterpack.com/) if this project dies (because it's not maintained anymore)  

## PimpMyPack API

The server is based on [Gin Framework](https://github.com/gin-gonic/gin) and provides endpoints to manage Accounts, Inventories & Packs

A dedicated API documentation is available [here](https://pmp-dev.alki.earth/swagger/index.html).

## For Developers

- **Backend Setup**: See [Setup for local development](#setup-for-local-development)
- **API Documentation**: [Swagger/OpenAPI](https://pmp-dev.alki.earth/swagger/index.html)
- **Frontend Integration**: [Frontend Integration Guide](docs/frontend-integration.md)

## Authentication

PimpMyPack uses JWT-based authentication with refresh tokens for secure API access.

### Token Types

- **Access Token**: Short-lived token (default: 15 minutes) used for API requests
- **Refresh Token**: Long-lived token (default: 1 day, or 30 days with "remember me") used to obtain new access tokens

### Authentication Flow

1. **Login**: POST to `/api/login` with username and password

   ```json
   {
     "username": "your_username",
     "password": "your_password",
     "remember_me": false
   }
   ```

2. **Response**: Receive both tokens

   ```json
   {
     "token": "...",              // Backward compatibility (same as access_token)
     "access_token": "...",       // Use for API requests
     "refresh_token": "...",      // Use to refresh access token
     "access_expires_in": 900,    // Access token lifetime in seconds
     "refresh_expires_in": 86400  // Refresh token lifetime in seconds
   }
   ```

3. **API Requests**: Include access token in Authorization header

   ```http
   Authorization: Bearer <access_token>
   ```

4. **Token Refresh**: POST to `/api/refresh` when access token expires

   ```json
   {
     "refresh_token": "..."
   }
   ```

5. **Logout**: POST to `/api/logout` to revoke refresh token

   ```json
   {
     "refresh_token": "..."
   }
   ```

### Configuration

Token lifetimes can be configured via environment variables in `.env`:

- `ACCESS_TOKEN_MINUTES`: Access token lifetime (default: 15 minutes)
- `REFRESH_TOKEN_DAYS`: Refresh token lifetime (default: 1 day)
- `REFRESH_TOKEN_REMEMBER_ME_DAYS`: Refresh token lifetime with "remember me" (default: 30 days)
- `REFRESH_TOKEN_CLEANUP_INTERVAL_HOURS`: Cleanup interval for expired tokens (default: 24 hours)

## Setup for local development

### 1. clone this repo

```shell
git clone git@github.com:Angak0k/pimpmypack.git
```

### 3. Start a local postgres database

The app need a local DB.

You need to use docker to start a postgres database:

```shell
docker run --name pmp_db \
    -d -p 5432:5432 \
    -e POSTGRES_PASSWORD=pmp1234 \
    -e POSTGRES_USER=pmp_user \
    -e POSTGRES_DB=pmp_db postgres:17
```

**Note:** PostgreSQL 17 is required for this project.

### 4. Configure the environment

Pimpmypack app read its conf from the environment and/or `.env` file.

The simplest way is to:

- copy the `.env.sample` file to `.env`
- customize the values in the `.env` file to match your setup

### 5. Start the API server

```shell
go build . && ./pimpmypack
```

## Run tests

```shell
go test ./...
```

or with verbose mode

```shell
go test -v ./...
```
