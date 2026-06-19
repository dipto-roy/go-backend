# Changelog

All notable changes to this project will be documented in this file.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)  
Versioning: [Semantic Versioning](https://semver.org/)

## [Unreleased]

## [1.0.0] - 2026-06-19

### Added
- JWT authentication with refresh token rotation
- User registration, login, logout, refresh endpoints
- User profile CRUD (get, update, change password, delete)
- Clean architecture: handler → service → repository
- PostgreSQL + GORM with soft deletes
- Structured JSON logging via zerolog
- Per-IP token-bucket rate limiter
- CORS, panic recovery, request ID middleware
- Input validation via go-playground/validator
- Swagger UI at `/swagger/index.html`
- `/health` and `/health/db` endpoints
- Docker multi-stage build (production)
- Docker Compose dev setup with Air hot-reload
- 82%+ service layer test coverage
