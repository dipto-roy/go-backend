# API Reference

Base URL: `http://localhost:8080`  
API prefix: `/api/v1`  
Content-Type: `application/json`

## Response Envelope

All responses use a consistent JSON envelope.

**Success**
```json
{
  "success": true,
  "data": { ... }
}
```

**Error**
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "human-readable message",
    "details": [ ... ]   // validation errors only
  }
}
```

**Paginated**
```json
{
  "success": true,
  "data": [ ... ],
  "meta": {
    "total": 100,
    "page": 1,
    "limit": 20,
    "pages": 5
  }
}
```

## Error Codes

| Code | HTTP | Meaning |
|------|------|---------|
| `NOT_FOUND` | 404 | Resource does not exist |
| `UNAUTHORIZED` | 401 | Missing or invalid credentials |
| `FORBIDDEN` | 403 | Authenticated but not permitted |
| `CONFLICT` | 409 | Resource already exists (e.g. email taken) |
| `BAD_REQUEST` | 400 | Malformed request |
| `VALIDATION_ERROR` | 422 | Input failed validation |
| `TOKEN_EXPIRED` | 401 | Access token expired — refresh it |
| `TOKEN_INVALID` | 401 | Token is malformed or revoked |
| `RATE_LIMITED` | 429 | Too many requests from this IP |
| `INTERNAL_ERROR` | 500 | Server error |

## Authentication

Protected routes require a Bearer token in the `Authorization` header:

```
Authorization: Bearer <access_token>
```

Access tokens expire in 15 minutes (default). Use `/auth/refresh` to obtain a new pair.

---

## Auth Endpoints

### POST /api/v1/auth/register

Create a new user account.

**Request**
```json
{
  "name": "Alice Smith",
  "email": "alice@example.com",
  "password": "supersecret"
}
```

| Field | Type | Rules |
|-------|------|-------|
| `name` | string | Required, 2–100 chars |
| `email` | string | Required, valid email |
| `password` | string | Required, min 8 chars |

**Response `201 Created`**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Alice Smith",
      "email": "alice@example.com",
      "created_at": "2026-06-19T11:00:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Errors**
| Code | Condition |
|------|-----------|
| `CONFLICT` | Email already registered |
| `VALIDATION_ERROR` | Invalid input |

---

### POST /api/v1/auth/login

Authenticate and receive a token pair.

**Request**
```json
{
  "email": "alice@example.com",
  "password": "supersecret"
}
```

**Response `200 OK`**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Alice Smith",
      "email": "alice@example.com",
      "created_at": "2026-06-19T11:00:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Errors**
| Code | Condition |
|------|-----------|
| `UNAUTHORIZED` | Wrong email or password |
| `VALIDATION_ERROR` | Invalid input |

---

### POST /api/v1/auth/refresh

Exchange a valid refresh token for a new token pair. Old refresh token is invalidated immediately (rotation).

**Request**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response `200 OK`**
```json
{
  "success": true,
  "data": {
    "user": { ... },
    "access_token": "eyJ...",
    "refresh_token": "eyJ..."
  }
}
```

**Errors**
| Code | Condition |
|------|-----------|
| `TOKEN_INVALID` | Token malformed, revoked, or expired |

---

### POST /api/v1/auth/logout

**Auth required.** Revoke the refresh token.

**Request**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response `200 OK`**
```json
{
  "success": true,
  "data": { "message": "logged out" }
}
```

---

## User Endpoints

All routes require `Authorization: Bearer <access_token>`.

### GET /api/v1/users/me

Get the authenticated user's profile.

**Response `200 OK`**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Alice Smith",
    "email": "alice@example.com",
    "created_at": "2026-06-19T11:00:00Z"
  }
}
```

---

### PUT /api/v1/users/me

Update name and/or email.

**Request**
```json
{
  "name": "Alice Jones",
  "email": "alice.jones@example.com"
}
```

| Field | Type | Rules |
|-------|------|-------|
| `name` | string | Required, 2–100 chars |
| `email` | string | Required, valid email |

**Response `200 OK`**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Alice Jones",
    "email": "alice.jones@example.com",
    "created_at": "2026-06-19T11:00:00Z"
  }
}
```

**Errors**
| Code | Condition |
|------|-----------|
| `CONFLICT` | New email already taken |
| `VALIDATION_ERROR` | Invalid input |

---

### PUT /api/v1/users/me/password

Change password.

**Request**
```json
{
  "current_password": "supersecret",
  "new_password": "evenmoresecret"
}
```

| Field | Type | Rules |
|-------|------|-------|
| `current_password` | string | Required |
| `new_password` | string | Required, min 8 chars |

**Response `200 OK`**
```json
{
  "success": true,
  "data": { "message": "password changed" }
}
```

**Errors**
| Code | Condition |
|------|-----------|
| `WRONG_PASSWORD` | Current password incorrect |
| `VALIDATION_ERROR` | Invalid input |

---

### DELETE /api/v1/users/me

Soft-delete the authenticated account. The record remains in the database with `deleted_at` set.

**Response `200 OK`**
```json
{
  "success": true,
  "data": { "message": "account deleted" }
}
```

---

## System Endpoints

### GET /health

**Response `200 OK`**
```json
{
  "success": true,
  "data": {
    "status": "ok",
    "uptime": "3h24m10s",
    "version": "1.0.0"
  }
}
```

### GET /health/db

**Response `200 OK`**
```json
{
  "success": true,
  "data": { "status": "ok", "database": "connected" }
}
```

**Response `500` (DB unreachable)**
```json
{
  "success": false,
  "error": { "code": "INTERNAL_ERROR", "message": "internal server error" }
}
```

---

## Swagger UI

Interactive API explorer available at:

```
http://localhost:8080/swagger/index.html
```

---

## Rate Limiting

- **Default:** 100 requests/second per IP, burst of 200
- **Headers:** No headers returned (standard 429 body on exceed)
- **Scope:** All endpoints including `/health`
- **Override:** Set `RATE_LIMIT_RPS` and `RATE_LIMIT_BURST` in `.env`

```json
{
  "success": false,
  "error": { "code": "RATE_LIMITED", "message": "too many requests" }
}
```
