# go-backend

Production-ready REST API built with Go and Gin.

## Features

- JWT authentication with refresh token rotation
- Clean architecture: Handler → Service → Repository
- PostgreSQL with GORM ORM and soft deletes
- Structured logging via zerolog
- In-memory token-bucket rate limiter
- CORS, panic recovery, request ID middleware
- Input validation via go-playground/validator
- Swagger UI at `/swagger/index.html`
- Graceful shutdown (30s drain)
- Docker multi-stage build + dev hot-reload via Air

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Framework | [Gin](https://github.com/gin-gonic/gin) |
| ORM | [GORM](https://gorm.io) + `gorm.io/driver/postgres` |
| Database | PostgreSQL 16 |
| Auth | [golang-jwt/jwt v5](https://github.com/golang-jwt/jwt) + bcrypt |
| Config | [Viper](https://github.com/spf13/viper) |
| Logging | [zerolog](https://github.com/rs/zerolog) |
| Validation | [go-playground/validator v10](https://github.com/go-playground/validator) |
| Docs | [swaggo/gin-swagger](https://github.com/swaggo/gin-swagger) |
| Testing | [testify](https://github.com/stretchr/testify) |

## Requirements

- Go 1.25+
- Docker + Docker Compose
- PostgreSQL 16 (or use Docker)

## Quick Start

```bash
# 1. Clone and enter directory
git clone <repo-url>
cd go-backend

# 2. Configure environment
cp .env.example .env
# Edit .env: set DB_DSN and JWT_SECRET (min 32 chars)

# 3a. Run with Docker (production)
docker compose up --build

# 3b. Run with Docker (development — hot reload)
docker compose -f docker-compose.dev.yml up

# 3c. Run locally
go run ./cmd/api
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DB_DSN` | Yes | — | PostgreSQL DSN |
| `JWT_SECRET` | Yes | — | Min 32 chars |
| `APP_PORT` | No | `8080` | HTTP port |
| `APP_ENV` | No | `development` | `development` / `production` |
| `LOG_LEVEL` | No | `info` | `debug` / `info` / `warn` / `error` |
| `LOG_PRETTY` | No | `true` | Pretty-print logs (false in production) |
| `JWT_ACCESS_EXPIRY` | No | `15m` | Access token TTL |
| `JWT_REFRESH_EXPIRY` | No | `168h` | Refresh token TTL |
| `RATE_LIMIT_RPS` | No | `100` | Requests/sec per IP |
| `RATE_LIMIT_BURST` | No | `200` | Burst capacity |

## API

Base URL: `http://localhost:8080/api/v1`

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/register` | — | Register user |
| POST | `/auth/login` | — | Login |
| POST | `/auth/refresh` | — | Refresh token |
| POST | `/auth/logout` | Bearer | Logout |
| GET | `/users/me` | Bearer | Get profile |
| PUT | `/users/me` | Bearer | Update profile |
| PUT | `/users/me/password` | Bearer | Change password |
| DELETE | `/users/me` | Bearer | Delete account |
| GET | `/health` | — | Health check |
| GET | `/health/db` | — | DB health check |
| GET | `/swagger/*any` | — | Swagger UI |

Full API docs: [`docs/API.md`](docs/API.md) | Swagger: `http://localhost:8080/swagger/index.html`

## Development

```bash
make run          # run locally
make test         # run tests with race detector
make build        # compile binary to bin/api
make swagger      # regenerate Swagger docs
make lint         # run golangci-lint
make migrate-up   # apply DB migrations
make migrate-down # rollback last migration
```

## Project Structure

```
go-backend/
├── cmd/api/main.go          # Entry point, DI wiring, graceful shutdown
├── internal/
│   ├── config/              # Viper config loader
│   ├── handler/             # HTTP handlers (Gin)
│   ├── middleware/          # Auth, CORS, logger, rate limiter, recovery
│   ├── model/               # GORM models
│   ├── repository/          # Data access layer + mocks
│   ├── server/              # Router setup
│   └── service/             # Business logic
├── pkg/
│   ├── apperror/            # Typed errors with HTTP status codes
│   ├── logger/              # zerolog wrapper
│   ├── response/            # JSON envelope helpers
│   └── token/               # JWT generation and verification
├── migrations/              # SQL migration files
└── docs/                    # Swagger generated files + documentation
```

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for full technical analysis.

## Testing

```bash
go test -race -cover ./...
```

Coverage: service layer 82%+.

## License

MIT
