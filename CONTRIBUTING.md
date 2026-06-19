# Contributing

## Workflow

1. Fork → feature branch (`feat/my-feature`) → PR to `main`
2. One feature/fix per PR
3. All tests must pass, coverage must stay ≥ 80%

## Commit Format

```
feat: add password reset endpoint
fix: prevent email enumeration on login
refactor: extract token generation to pkg/token
docs: update API.md with refresh flow
test: add handler integration tests for auth
```

Types: `feat` `fix` `refactor` `docs` `test` `chore` `perf` `ci`

## Development Setup

```bash
cp .env.example .env   # fill DB_DSN and JWT_SECRET
make run               # start server
make test              # run tests
make swagger           # regenerate Swagger docs after handler changes
```

## Code Standards

- No direct DB calls in handlers — go through service → repository
- All errors use `pkg/apperror` typed errors
- All responses use `pkg/response` helpers
- New endpoints need Swagger annotations and at least one test
- Run `go vet ./...` before pushing
