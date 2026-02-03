# Frontend Integration Guide

Complete guide for integrating PimpMyPack authentication in your frontend application.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Authentication Flow](#authentication-flow)
- [Token Storage Options](#token-storage-options)
- [Automatic Token Refresh](#automatic-token-refresh)
- [Error Handling](#error-handling)
- [Code Examples](#code-examples)
- [Security Considerations](#security-considerations)
- [Testing](#testing)
- [Troubleshooting](#troubleshooting)
- [API Reference](#api-reference)

## Overview

PimpMyPack uses **JWT-based authentication** with two types of tokens:

| Token Type | Purpose | Lifetime | Storage |
|------------|---------|----------|---------|
| **Access Token** | API requests | 15 minutes (default) | Memory or sessionStorage |
| **Refresh Token** | Obtain new access tokens | 1 day (default), 30 days with "remember me" | httpOnly cookie (recommended) or secure storage |

### Authentication Architecture

```
┌─────────┐      1. Login         ┌─────────┐
│ Frontend│ ──────────────────────>│ Backend │
│         │ <──────────────────────│         │
│         │  Access + Refresh Token│         │
└────┬────┘                        └────┬────┘
     │                                  │
     │ 2. API Request (Access Token)    │
     │ ──────────────────────────────────>
     │                                  │
     │ 3. Access Token Expired (401)    │
     │ <──────────────────────────────────
     │                                  │
     │ 4. Refresh (Refresh Token)       │
     │ ──────────────────────────────────>
     │ <──────────────────────────────────
     │  New Access Token                │
     │                                  │
     │ 5. Retry API Request             │
     │ ──────────────────────────────────>
     │                                  │
```

## Quick Start

### 1. Login

```typescript
const response = await fetch('https://api.pimpmypack.com/api/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    username: 'john_doe',
    password: 'secure_password',
    remember_me: false
  })
});

const data = await response.json();
// Store the short-lived access token in frontend storage.
// The refresh token should be sent as an httpOnly, secure cookie by the backend
// and MUST NOT be stored in localStorage or exposed to JavaScript.
localStorage.setItem('accessToken', data.access_token);
```

### 2. Make Authenticated Request

```typescript
const response = await fetch('https://api.pimpmypack.com/api/v1/mypacks', {
  headers: {
    'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
  }
});
```

### 3. Handle Token Expiration

```typescript
if (response.status === 401) {
  // Refresh access token
  const newAccessToken = await refreshAccessToken();
  // Retry original request with new token
}
```

## Authentication Flow

### Complete TypeScript Implementation

```typescript
interface LoginCredentials {
  username: string;
  password: string;
  remember_me?: boolean;
}

interface TokenPair {
  token: string;  // Backward compatibility
  access_token: string;
  refresh_token: string;
  access_expires_in: number;   // Seconds
  refresh_expires_in: number;  // Seconds
}

class AuthService {
  private baseURL = 'https://api.pimpmypack.com/api';
  private accessToken: string | null = null;
  private refreshToken: string | null = null;

  /**
   * Login and store tokens
   */
  async login(credentials: LoginCredentials): Promise<TokenPair> {
    const response = await fetch(`${this.baseURL}/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(credentials)
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Login failed');
    }

    const tokens: TokenPair = await response.json();

    // Store tokens
    this.accessToken = tokens.access_token;
    this.refreshToken = tokens.refresh_token;

    // Optional: persist in localStorage (see Security Considerations)
    localStorage.setItem('accessToken', tokens.access_token);
    localStorage.setItem('refreshToken', tokens.refresh_token);

    return tokens;
  }

  /**
   * Refresh access token
   */
  async refreshAccessToken(): Promise<string> {
    const response = await fetch(`${this.baseURL}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        refresh_token: this.refreshToken || localStorage.getItem('refreshToken')
      })
    });

    if (!response.ok) {
      // Refresh failed - redirect to login
      this.logout();
      window.location.href = '/login';
      throw new Error('Refresh token expired');
    }

    const data = await response.json();
    this.accessToken = data.access_token;
    localStorage.setItem('accessToken', data.access_token);

    return data.access_token;
  }

  /**
   * Logout and clear tokens
   */
  async logout(): Promise<void> {
    if (this.refreshToken) {
      // Revoke refresh token on server
      await fetch(`${this.baseURL}/logout`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: this.refreshToken })
      });
    }

    // Clear local tokens
    this.accessToken = null;
    this.refreshToken = null;
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
  }

  /**
   * Make authenticated request with automatic token refresh
   */
  async authenticatedFetch(url: string, options: RequestInit = {}): Promise<Response> {
    // Add access token to headers
    const headers = {
      ...options.headers,
      'Authorization': `Bearer ${this.accessToken || localStorage.getItem('accessToken')}`
    };

    let response = await fetch(url, { ...options, headers });

    // If 401, try refreshing token once
    if (response.status === 401) {
      await this.refreshAccessToken();

      // Retry with new token
      headers['Authorization'] = `Bearer ${this.accessToken}`;
      response = await fetch(url, { ...options, headers });
    }

    return response;
  }
}

// Usage
const auth = new AuthService();

// Login
await auth.login({
  username: 'john_doe',
  password: 'password',
  remember_me: true
});

// Make authenticated request
const response = await auth.authenticatedFetch('https://api.pimpmypack.com/api/v1/mypacks');
const packs = await response.json();
```

## Token Storage Options

### Comparison

| Storage Method | Security | XSS Protection | CSRF Protection | SSR Compatible |
|----------------|----------|----------------|-----------------|----------------|
| **Memory only** | ⭐⭐⭐⭐⭐ | ✅ Best | ✅ Yes | ❌ No |
| **sessionStorage** | ⭐⭐⭐ | ❌ Vulnerable | ✅ Yes | ❌ No |
| **localStorage** | ⭐⭐ | ❌ Vulnerable | ✅ Yes | ❌ No |
| **httpOnly Cookies** | ⭐⭐⭐⭐⭐ | ✅ Immune | ⚠️ Needs CSRF protection | ✅ Yes |

### Recommended Approach

**Access Token**: Memory or sessionStorage (short-lived, acceptable risk)
**Refresh Token**: httpOnly cookie with `Secure` and `SameSite=Strict` flags

### Example: httpOnly Cookie Setup

**Backend (Go)**:
```go
// Set refresh token as httpOnly cookie
http.SetCookie(w, &http.Cookie{
    Name:     "refresh_token",
    Value:    refreshToken,
    HttpOnly: true,
    Secure:   true,  // HTTPS only
    SameSite: http.SameSiteStrictMode,
    MaxAge:   86400, // 1 day
    Path:     "/api/auth/refresh",
})
```

**Frontend (TypeScript)**:
```typescript
// Refresh request automatically sends cookie
const response = await fetch('https://api.pimpmypack.com/api/auth/refresh', {
  method: 'POST',
  credentials: 'include'  // Include httpOnly cookies
});
```

## Automatic Token Refresh

### Axios Interceptor

```typescript
import axios, { AxiosInstance } from 'axios';

class ApiClient {
  private client: AxiosInstance;
  private refreshToken: string;
  private isRefreshing = false;
  private failedQueue: Array<{
    resolve: (token: string) => void;
    reject: (error: any) => void;
  }> = [];

  constructor() {
    this.client = axios.create({
      baseURL: 'https://api.pimpmypack.com/api',
      headers: { 'Content-Type': 'application/json' }
    });

    // Request interceptor - add access token
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('accessToken');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor - handle 401 and refresh
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        const originalRequest = error.config;

        // If 401 and not already retrying
        if (error.response?.status === 401 && !originalRequest._retry) {
          if (this.isRefreshing) {
            // Queue request while refresh is in progress
            return new Promise((resolve, reject) => {
              this.failedQueue.push({ resolve, reject });
            })
              .then(token => {
                originalRequest.headers.Authorization = `Bearer ${token}`;
                return this.client(originalRequest);
              })
              .catch(err => Promise.reject(err));
          }

          originalRequest._retry = true;
          this.isRefreshing = true;

          try {
            const newToken = await this.refreshAccessToken();

            // Retry all queued requests
            this.failedQueue.forEach(({ resolve }) => resolve(newToken));
            this.failedQueue = [];

            originalRequest.headers.Authorization = `Bearer ${newToken}`;
            return this.client(originalRequest);
          } catch (refreshError) {
            // Refresh failed - logout
            this.failedQueue.forEach(({ reject }) => reject(refreshError));
            this.failedQueue = [];
            this.logout();
            return Promise.reject(refreshError);
          } finally {
            this.isRefreshing = false;
          }
        }

        return Promise.reject(error);
      }
    );
  }

  private async refreshAccessToken(): Promise<string> {
    const response = await axios.post(
      'https://api.pimpmypack.com/api/auth/refresh',
      { refresh_token: localStorage.getItem('refreshToken') }
    );

    const newToken = response.data.access_token;
    localStorage.setItem('accessToken', newToken);
    return newToken;
  }

  private logout(): void {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    window.location.href = '/login';
  }

  // Public API methods
  async get(url: string) {
    return this.client.get(url);
  }

  async post(url: string, data: any) {
    return this.client.post(url, data);
  }
}

// Usage
const api = new ApiClient();
const response = await api.get('/v1/mypacks');
```

### Fetch API with Retry

```typescript
async function fetchWithRetry(
  url: string,
  options: RequestInit = {},
  retries = 1
): Promise<Response> {
  const token = localStorage.getItem('accessToken');

  const headers = {
    ...options.headers,
    'Authorization': `Bearer ${token}`
  };

  let response = await fetch(url, { ...options, headers });

  // Retry on 401 if we have retries left
  if (response.status === 401 && retries > 0) {
    const newToken = await refreshAccessToken();
    headers['Authorization'] = `Bearer ${newToken}`;
    response = await fetch(url, { ...options, headers });
  }

  return response;
}
```

## Error Handling

### HTTP Status Codes

| Status | Meaning | Action |
|--------|---------|--------|
| **200** | Success | Process response |
| **400** | Bad Request | Show validation errors to user |
| **401** | Unauthorized | Refresh token or redirect to login |
| **403** | Forbidden | Show "Access denied" message |
| **404** | Not Found | Show "Resource not found" |
| **429** | Too Many Requests | Wait `retry_after` seconds and retry |
| **500** | Internal Server Error | Show generic error, log for debugging |

### Rate Limit Handling

```typescript
async function handleRateLimitedRequest(
  url: string,
  options: RequestInit
): Promise<Response> {
  let response = await fetch(url, options);

  // If rate limited, wait and retry
  if (response.status === 429) {
    const data = await response.json();
    const retryAfter = data.retry_after || 60; // Default 60 seconds

    console.log(`Rate limited. Retrying after ${retryAfter} seconds`);

    // Wait
    await new Promise(resolve => setTimeout(resolve, retryAfter * 1000));

    // Retry
    response = await fetch(url, options);
  }

  return response;
}
```

### Error Message Display

```typescript
interface ApiError {
  error: string;
  retry_after?: number;
}

async function handleApiError(response: Response): Promise<never> {
  const error: ApiError = await response.json();

  switch (response.status) {
    case 400:
      throw new Error(`Invalid request: ${error.error}`);
    case 401:
      throw new Error('Session expired. Please login again.');
    case 403:
      throw new Error('You do not have permission to access this resource.');
    case 404:
      throw new Error('Resource not found.');
    case 429:
      throw new Error(`Too many requests. Please try again in ${error.retry_after} seconds.`);
    case 500:
      throw new Error('Server error. Please try again later.');
    default:
      throw new Error(error.error || 'An unexpected error occurred.');
  }
}
```

## Code Examples

### React Hook

```typescript
import { useState, useEffect, useCallback } from 'react';

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export function useAuth() {
  const [authState, setAuthState] = useState<AuthState>({
    accessToken: localStorage.getItem('accessToken'),
    refreshToken: localStorage.getItem('refreshToken'),
    isAuthenticated: !!localStorage.getItem('accessToken'),
    isLoading: false
  });

  const login = useCallback(async (username: string, password: string, rememberMe = false) => {
    setAuthState(prev => ({ ...prev, isLoading: true }));

    try {
      const response = await fetch('https://api.pimpmypack.com/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password, remember_me: rememberMe })
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error);
      }

      const data = await response.json();

      localStorage.setItem('accessToken', data.access_token);
      localStorage.setItem('refreshToken', data.refresh_token);

      setAuthState({
        accessToken: data.access_token,
        refreshToken: data.refresh_token,
        isAuthenticated: true,
        isLoading: false
      });

      return data;
    } catch (error) {
      setAuthState(prev => ({ ...prev, isLoading: false }));
      throw error;
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      if (authState.refreshToken) {
        await fetch('https://api.pimpmypack.com/api/logout', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: authState.refreshToken })
        });
      }
    } finally {
      localStorage.removeItem('accessToken');
      localStorage.removeItem('refreshToken');
      setAuthState({
        accessToken: null,
        refreshToken: null,
        isAuthenticated: false,
        isLoading: false
      });
    }
  }, [authState.refreshToken]);

  const refreshAccessToken = useCallback(async () => {
    try {
      const response = await fetch('https://api.pimpmypack.com/api/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: authState.refreshToken })
      });

      if (!response.ok) {
        throw new Error('Refresh failed');
      }

      const data = await response.json();
      localStorage.setItem('accessToken', data.access_token);

      setAuthState(prev => ({
        ...prev,
        accessToken: data.access_token
      }));

      return data.access_token;
    } catch (error) {
      await logout();
      throw error;
    }
  }, [authState.refreshToken, logout]);

  return {
    ...authState,
    login,
    logout,
    refreshAccessToken
  };
}

// Usage in component
function MyComponent() {
  const { isAuthenticated, login, logout } = useAuth();

  return (
    <div>
      {isAuthenticated ? (
        <button onClick={logout}>Logout</button>
      ) : (
        <button onClick={() => login('john_doe', 'password')}>Login</button>
      )}
    </div>
  );
}
```

### Vue Composable

```typescript
import { ref, computed } from 'vue';

export function useAuth() {
  const accessToken = ref<string | null>(localStorage.getItem('accessToken'));
  const refreshToken = ref<string | null>(localStorage.getItem('refreshToken'));
  const isLoading = ref(false);

  const isAuthenticated = computed(() => !!accessToken.value);

  async function login(username: string, password: string, rememberMe = false) {
    isLoading.value = true;

    try {
      const response = await fetch('https://api.pimpmypack.com/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password, remember_me: rememberMe })
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error);
      }

      const data = await response.json();

      accessToken.value = data.access_token;
      refreshToken.value = data.refresh_token;
      localStorage.setItem('accessToken', data.access_token);
      localStorage.setItem('refreshToken', data.refresh_token);

      return data;
    } finally {
      isLoading.value = false;
    }
  }

  async function logout() {
    try {
      if (refreshToken.value) {
        await fetch('https://api.pimpmypack.com/api/logout', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: refreshToken.value })
        });
      }
    } finally {
      accessToken.value = null;
      refreshToken.value = null;
      localStorage.removeItem('accessToken');
      localStorage.removeItem('refreshToken');
    }
  }

  async function refreshAccessToken() {
    try {
      const response = await fetch('https://api.pimpmypack.com/api/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken.value })
      });

      if (!response.ok) throw new Error('Refresh failed');

      const data = await response.json();
      accessToken.value = data.access_token;
      localStorage.setItem('accessToken', data.access_token);

      return data.access_token;
    } catch (error) {
      await logout();
      throw error;
    }
  }

  return {
    accessToken,
    isAuthenticated,
    isLoading,
    login,
    logout,
    refreshAccessToken
  };
}
```

## Security Considerations

### DO ✅

- **Use HTTPS** in production (TLS encryption)
- **Store access tokens** in memory or sessionStorage (short-lived)
- **Store refresh tokens** in httpOnly cookies with `Secure` and `SameSite` flags
- **Implement automatic token refresh** before access token expires
- **Clear tokens** immediately on logout
- **Validate token expiration** on the client side
- **Use environment variables** for API URLs
- **Implement CSRF protection** when using cookies
- **Log authentication events** for security monitoring
- **Rate limit** authentication endpoints

### DON'T ❌

- **Don't store tokens in regular cookies** (vulnerable to CSRF)
- **Don't log tokens** to console or analytics
- **Don't embed tokens in URLs** (appears in logs, browser history)
- **Don't store sensitive data** in JWT payload (it's base64, not encrypted)
- **Don't trust client-side expiration** - always validate on server
- **Don't ignore 401 responses** - always refresh or redirect
- **Don't store passwords** anywhere (hash on server only)
- **Don't use weak passwords** - enforce strong password policies
- **Don't share tokens** between users or devices
- **Don't skip token validation** on the backend

### XSS Protection

```typescript
// Sanitize user input before displaying
function sanitizeHTML(text: string): string {
  const element = document.createElement('div');
  element.textContent = text;
  return element.innerHTML;
}

// Use Content Security Policy
// Add to HTML head:
<meta http-equiv="Content-Security-Policy"
      content="default-src 'self'; script-src 'self'">
```

## Testing

### Mock Authentication Service

```typescript
class MockAuthService {
  async login(credentials: any) {
    return {
      access_token: 'mock_access_token',
      refresh_token: 'mock_refresh_token',
      access_expires_in: 900,
      refresh_expires_in: 86400
    };
  }

  async refreshAccessToken() {
    return 'new_mock_access_token';
  }

  async logout() {
    return { message: 'Logged out' };
  }
}
```

### Jest Test Example

```typescript
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { useAuth } from './useAuth';

// Mock fetch
global.fetch = jest.fn();

describe('useAuth', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
  });

  it('should login successfully', async () => {
    const mockResponse = {
      access_token: 'test_access_token',
      refresh_token: 'test_refresh_token',
      access_expires_in: 900,
      refresh_expires_in: 86400
    };

    (fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse
    });

    const { result } = renderHook(() => useAuth());

    await act(async () => {
      await result.current.login('testuser', 'password');
    });

    expect(result.current.isAuthenticated).toBe(true);
    expect(localStorage.getItem('accessToken')).toBe('test_access_token');
  });

  it('should handle login failure', async () => {
    (fetch as jest.Mock).mockResolvedValueOnce({
      ok: false,
      json: async () => ({ error: 'Invalid credentials' })
    });

    const { result } = renderHook(() => useAuth());

    await expect(
      result.current.login('testuser', 'wrongpassword')
    ).rejects.toThrow('Invalid credentials');

    expect(result.current.isAuthenticated).toBe(false);
  });
});
```

## Troubleshooting

### Common Issues

#### 1. **"Invalid refresh token" error**

**Cause**: Refresh token expired or revoked
**Solution**: Redirect user to login page

```typescript
if (error.message === 'Invalid refresh token') {
  await logout();
  router.push('/login');
}
```

#### 2. **CORS errors**

**Cause**: Backend not configured for cross-origin requests
**Solution**: Ensure backend sends proper CORS headers

```go
// Backend (Go)
c.Header("Access-Control-Allow-Origin", "https://yourfrontend.com")
c.Header("Access-Control-Allow-Credentials", "true")
c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
```

#### 3. **Token not included in requests**

**Cause**: Authorization header not set
**Solution**: Check header format

```typescript
// Correct format
headers: {
  'Authorization': `Bearer ${accessToken}`  // Note: "Bearer " with space
}
```

#### 4. **Rate limit exceeded (429)**

**Cause**: Too many refresh attempts
**Solution**: Implement exponential backoff

```typescript
async function retryWithBackoff(fn: () => Promise<any>, retries = 3) {
  for (let i = 0; i < retries; i++) {
    try {
      return await fn();
    } catch (error) {
      if (i === retries - 1) throw error;
      await new Promise(resolve => setTimeout(resolve, Math.pow(2, i) * 1000));
    }
  }
}
```

#### 5. **Tokens persist after logout**

**Cause**: Tokens not cleared from storage
**Solution**: Clear all storage on logout

```typescript
function clearAllTokens() {
  localStorage.removeItem('accessToken');
  localStorage.removeItem('refreshToken');
  sessionStorage.clear();
  // Clear any in-memory tokens
  accessToken = null;
  refreshToken = null;
}
```

### Debug Checklist

- [ ] Check browser console for errors
- [ ] Verify API endpoint URLs are correct
- [ ] Confirm access token is not expired (check JWT payload)
- [ ] Ensure `Authorization` header has correct format: `Bearer <token>`
- [ ] Check network tab for request/response details
- [ ] Verify CORS headers in response
- [ ] Confirm backend is running and accessible
- [ ] Check rate limiting isn't blocking requests
- [ ] Validate token is stored correctly (check localStorage/sessionStorage)
- [ ] Test with a fresh login (clear all storage first)

### JWT Decoder Tool

```typescript
function decodeJWT(token: string): any {
  try {
    const payload = token.split('.')[1];
    return JSON.parse(atob(payload));
  } catch (error) {
    console.error('Failed to decode JWT:', error);
    return null;
  }
}

// Usage - inspect token content
const tokenData = decodeJWT(accessToken);
console.log('Token expires at:', new Date(tokenData.exp * 1000));
console.log('User ID:', tokenData.user_id);
```

## API Reference

### Authentication Endpoints

#### POST /api/login

**Request**:
```json
{
  "username": "john_doe",
  "password": "secure_password",
  "remember_me": false
}
```

**Response** (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1...",
  "access_token": "eyJhbGciOiJIUzI1...",
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2g...",
  "access_expires_in": 900,
  "refresh_expires_in": 86400
}
```

**Errors**:
- `400` - Invalid request
- `401` - Invalid credentials or account not confirmed

---

#### POST /api/auth/refresh

**Request**:
```json
{
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2g..."
}
```

**Response** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzI1...",
  "expires_in": 900
}
```

**Errors**:
- `400` - Invalid request
- `401` - Invalid or expired refresh token
- `429` - Rate limit exceeded (max 10 requests/minute per IP)

---

#### POST /api/logout

**Request**:
```json
{
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2g..."
}
```

**Response** (200 OK):
```json
{
  "message": "Logged out successfully"
}
```

---

### Protected Endpoints

All endpoints under `/api/v1/*` require authentication.

**Headers**:
```
Authorization: Bearer <access_token>
```

**Example**:
```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1..." \
     https://api.pimpmypack.com/api/v1/mypacks
```

---

## Additional Resources

- **Full API Documentation**: [Swagger/OpenAPI](https://pmp-dev.alki.earth/swagger/index.html)
- **Backend Setup Guide**: [README.md](../README.md)
- **Security Specification**: [specs/005_authentication-refresh-token.md](../specs/005_authentication-refresh-token.md)

---

**Last Updated**: 2026-02-03
**API Version**: 1.0
**Maintainer**: PimpMyPack Team
