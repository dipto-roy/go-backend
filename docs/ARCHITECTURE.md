# Architecture

## Overview

go-backend is a production-grade REST API following clean architecture principles. Dependencies flow inward: HTTP handlers depend on service interfaces, services depend on repository interfaces. No layer knows about the layer above it.

```
HTTP Request
     │
     ▼
 Middleware Stack
(recovery → requestID → logger → cors → ratelimit)
     │
     ▼
  Handler
(bind & validate input → call service → write response)
     │
     ▼
  Service
(business rules, password hashing, token generation)
     │
     ▼
 Repository
(interface → PostgreSQL implementation)
     │
     ▼
 PostgreSQL
```

## Directory Structure

```
go-backend/
├── cmd/
│   └── api/
│       └── main.go              # Entry point: load config, wire deps, start server
├── internal/
│   ├── config/
│   │   └── config.go            # Viper: reads .env + env vars, validates at startup
│   ├── handler/
│   │   ├── handler.go           # Base Handler struct, DI container, validation helper
│   │   ├── auth.go              # register, login, refresh, logout
│   │   ├── user.go              # getMe, updateMe, changePassword, deleteMe
│   │   └── health.go            # /health, /health/db
│   ├── middleware/
│   │   ├── auth.go              # JWT Bearer validation, injects user ID into context
│   │   ├── cors.go              # Configurable CORS with origin allowlist
│   │   ├── logger.go            # Structured request logging (zerolog)
│   │   ├── ratelimit.go         # Per-IP token bucket rate limiter (in-memory)
│   │   ├── recovery.go          # Panic → 500 with stack trace logging
│   │   └── request_id.go        # X-Request-ID injection and propagation
│   ├── model/
│   │   └── user.go              # User and RefreshToken GORM models
│   ├── repository/
│   │   ├── repository.go        # UserRepository and RefreshTokenRepository interfaces
│   │   ├── user_postgres.go     # PostgreSQL UserRepository implementation
│   │   ├── refresh_token_postgres.go  # PostgreSQL RefreshTokenRepository
│   │   └── mock/
│   │       └── user_mock.go     # testify/mock implementations for testing
│   ├── server/
│   │   └── server.go            # Gin engine setup, all route registration
│   └── service/
│       ├── auth.go              # AuthService: register, login, refresh, logout
│       └── user.go              # UserService: profile CRUD, password change
├── pkg/
│   ├── apperror/
│   │   └── errors.go            # Typed errors (AppError) with HTTP status codes
│   ├── logger/
│   │   └── logger.go            # Global zerolog wrapper (Init, Get, Info, Error…)
│   ├── response/
│   │   └── response.go          # Consistent JSON envelope: {success, data, error, meta}
│   └── token/
│       └── jwt.go               # JWT HS256: Generate (pair), Verify (returns Claims)
├── migrations/
│   ├── 000001_create_users.up.sql
│   └── 000001_create_users.down.sql
└── docs/                        # Swagger-generated OpenAPI files
```

## Key Design Decisions

### 1. Interface-Driven Repositories

Repository interfaces are defined in `internal/repository/repository.go` and consumed by services. Implementations (PostgreSQL) are injected at startup in `main.go`. This enables:
- Mock substitution in unit tests without a real database
- Future swap to a different storage engine without touching service code

### 2. Typed Error System (`pkg/apperror`)

All internal errors are wrapped in `AppError{Code, Message, HTTPStatus}`. The `response.Err()` helper inspects the error type and writes the correct HTTP status automatically. Handlers never hard-code status codes.

```
Service returns apperror.ErrNotFound
  → response.Err() maps to 404
  → {"success":false,"error":{"code":"NOT_FOUND","message":"resource not found"}}
```

### 3. Refresh Token Rotation

Refresh tokens are stored in the `refresh_tokens` table (not only in the JWT). On every `/auth/refresh`:
1. JWT is verified
2. Token record is looked up and checked for expiry
3. Old token is deleted
4. New token pair is issued

This enables revocation and prevents replay attacks.

### 4. Email Enumeration Prevention

`AuthService.Login` returns `ErrUnauthorized` (not `ErrNotFound`) when the email doesn't exist. This prevents attackers from probing which emails are registered.

### 5. Rate Limiter

Uses a per-IP token bucket implemented without external dependencies. Goroutine-safe via `sync.RWMutex`. A background goroutine cleans up idle buckets every 5 minutes. For multi-instance deployments, replace with a Redis-backed limiter.

### 6. Graceful Shutdown

`main.go` catches `SIGINT`/`SIGTERM`, calls `http.Server.Shutdown(ctx)` with a 30-second timeout. In-flight requests complete; new connections are rejected.

## Data Models

### `users`

| Column | Type | Notes |
|--------|------|-------|
| `id` | UUID | PK, auto-generated |
| `email` | VARCHAR(255) | Unique index |
| `password_hash` | VARCHAR(255) | bcrypt, cost=10 |
| `name` | VARCHAR(100) | |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |
| `deleted_at` | TIMESTAMPTZ | Soft delete (GORM) |

### `refresh_tokens`

| Column | Type | Notes |
|--------|------|-------|
| `id` | UUID | PK |
| `user_id` | UUID | FK → users.id, CASCADE |
| `token` | VARCHAR(500) | Unique, JWT string |
| `expires_at` | TIMESTAMPTZ | Compared on refresh |
| `created_at` | TIMESTAMPTZ | |
| `deleted_at` | TIMESTAMPTZ | Soft delete |

## Authentication Flow

```
Register:
  POST /auth/register → hash password → insert user → generate JWT pair
                      → store refresh token → return {user, access_token, refresh_token}

Login:
  POST /auth/login → find user by email → compare bcrypt → generate JWT pair
                   → store refresh token → return {user, access_token, refresh_token}

Authenticated request:
  Header: Authorization: Bearer <access_token>
  middleware/auth.go → verify JWT → inject user_id into gin.Context → handler reads it

Token refresh:
  POST /auth/refresh → verify JWT → lookup stored refresh token → check expiry
                     → delete old → generate new pair → store new refresh token

Logout:
  POST /auth/logout → delete refresh token record (access token expires naturally)
```

## Configuration

Config loads at startup via Viper. Priority order: environment variables > `.env` file > defaults. Startup fails fast if `JWT_SECRET` (< 32 chars) or `DB_DSN` is missing.

## Middleware Execution Order

```
1. Recovery       — catches panics, returns 500
2. RequestID      — attaches X-Request-ID to every request/response
3. Logger         — logs method, path, status, latency, request ID
4. CORS           — sets Access-Control-* headers, handles OPTIONS preflight
5. RateLimit      — per-IP token bucket, 429 if exhausted
6. Auth (route)   — validates Bearer JWT on protected routes only
```

## Testing Strategy

| Type | Location | Approach |
|------|----------|----------|
| Unit — service | `internal/service/*_test.go` | Mock repositories via testify/mock |
| Integration — handler | `internal/handler/*_test.go` | httptest.NewRecorder, mock services |
| Coverage target | | 80%+ on service layer |

Run:
```bash
go test -race -cover ./internal/service/... ./internal/handler/...
```

## Production Checklist

- [ ] Set `APP_ENV=production` (disables gin debug mode, forces JSON logs)
- [ ] Set `LOG_PRETTY=false`
- [ ] Rotate `JWT_SECRET` on deploy; existing tokens invalidated automatically
- [ ] Use Redis-backed rate limiter for multi-instance deployments
- [ ] Set `CORS_ORIGINS` to your frontend domain
- [ ] Run `migrate-up` before deploying new binary
- [ ] Configure TLS termination at reverse proxy (nginx/caddy)
- [ ] Set `ReadTimeout`, `WriteTimeout` appropriate for your workload
