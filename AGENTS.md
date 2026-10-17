# AGENTS.md — YOTA Backend Constitution

> This file defines the non-negotiable rules, architecture boundaries, and workflow protocols for any AI agent or contributor working in `vilamuzz-yota-backend`.
> Read this file FIRST before exploring the codebase, running commands, or writing code.

---

## 1. Project Identity & Domain Overview

- **Name:** YOTA Backend (`github.com/Vilamuzz/yota-backend`)
- **One-line description:** Core REST API for managing foundation operations, donation programs, social programs (subscriptions, invoices, transactions), foster children sponsorships, ambulance services, financial records, and public content (news, prayers, gallery).
- **Architecture:** Modular Domain-Driven Layered Architecture (`handler` → `service` → `repository`) wired via a centralized Dependency Injection container (`internal/container`).
- **API Documentation:** OpenAPI / Swagger 2.0 generated via `swaggo/swag`.

---

## 2. Tech Stack (Exact Versions — `go.mod` Source of Truth)

- **Language:** Go `1.25.0`
- **HTTP Framework:** Gin (`github.com/gin-gonic/gin v1.11.0`) + `gin-contrib/cors v1.7.6`
- **Database & ORM:** PostgreSQL + GORM (`gorm.io/gorm v1.31.1`, `gorm.io/driver/postgres v1.6.0`)
- **Schema Migrations:** Atlas + GORM Provider (`ariga.io/atlas-provider-gorm v0.6.0`)
- **Cache & Rate Limiting:** Redis (`github.com/redis/go-redis/v9 v9.17.3`) + `github.com/ulule/limiter/v3 v3.11.2`
- **Object Storage (S3):** MinIO Client (`github.com/minio/minio-go/v7 v7.0.98`) + `golang.org/x/image v0.37.0`
- **Authentication & OAuth:** JWT v5 (`github.com/golang-jwt/jwt/v5 v5.3.0`), Goth OAuth (`github.com/markbates/goth v1.82.0`), Gorilla Sessions (`v1.4.0`), `golang.org/x/crypto v0.48.0`
- **Payment Gateway:** Midtrans (`github.com/midtrans/midtrans-go v1.3.8`)
- **Background Jobs:** Cron v3 (`github.com/robfig/cron/v3 v3.0.1`)
- **Security & Sanitization:** Bluemonday HTML Sanitizer (`github.com/microcosm-cc/bluemonday v1.0.27`)
- **Utilities:** UUID (`google/uuid v1.6.0`), Slug (`gosimple/slug v1.15.0`), Logrus (`sirupsen/logrus v1.9.3`), Godotenv (`joho/godotenv v1.5.1`), `golang.org/x/sync v0.20.0`
- **API Docs:** Swaggo (`swaggo/swag v1.16.6`, `swaggo/gin-swagger v1.6.1`, `swaggo/files v1.0.1`)

> **Dependency Rule:** Never run `go get` to add new packages or upgrade versions without explicit owner approval. Always use existing packages in `go.mod`.

---

## 3. Directory Structure & Sacred Boundaries

```text
vilamuzz-yota-backend/
├── app/                          # Domain modules (business logic & HTTP delivery)
│   ├── <domain_name>/            # Standard 6-file anatomy per domain module:
│   │   ├── <domain_name>.go      # Domain entity / GORM model structs & interfaces
│   │   ├── handler.go            # Gin HTTP handlers (bind request, call service, return JSON)
│   │   ├── service.go            # Business logic, orchestration, MinIO/Redis/Midtrans calls
│   │   ├── repository.go         # GORM database queries & transactions
│   │   ├── request.go            # Incoming payload structs + validator/v10 binding tags
│   │   └── response.go           # Outgoing JSON DTOs & mapper functions
│   ├── auth/                     # Auth, OAuth (Goth), email_verification.go, password_reset.go
│   ├── media/                    # Internal S3/MinIO media service (no handler.go)
│   ├── payment/                  # Midtrans webhook/notification handler (handler.go only)
│   └── middleware/               # cors.go, jwt.go, logger.go, rate_limit.go, recovery.go
├── cmd/
│   ├── server/main.go            # Main Gin API server & cron entrypoint
│   ├── migrate/atlas/main.go     # Atlas GORM schema loader entrypoint
│   └── seed/                     # Database seeder (main.go + cmd/seed/models/*.go)
├── config/                       # Env loaders: postgre, redis, s3, jwt, oauth, email, media, rate_limit
├── internal/
│   ├── app/                      # App bootstrap (app.go) & Gin route registration (routes.go)
│   ├── container/container.go    # Dependency Injection container wiring all repos/services/handlers
│   └── scheduler/scheduler.go    # robfig/cron background jobs (e.g., invoice/subscription checks)
├── migrations/                   # Versioned Atlas SQL files + atlas.sum (DO NOT EDIT MANUALLY)
├── models/postgre.go             # Central GORM model registry for PostgreSQL & Atlas
├── docs/                         # [HYBRID DOCS] Business specs, API flows, & Swagger output
└── .agents/skills/               # Modular AI Agent Skills (read on-demand per task)

```

### Sacred Architectural Rules

1. **Strict 3-Layer Flow (`handler` → `service` → `repository`):**

- `handler.go` ONLY parses Gin contexts (`*gin.Context`), validates requests (`request.go`), invokes `service.go`, and returns formatted responses (`response.go`). **Never call GORM or `repository.go` from a handler.**
- `service.go` holds business rules, sanitizes HTML inputs via `bluemonday` (for news, descriptions, comments), generates slugs (`gosimple/slug`), uploads files via `app/media`, and coordinates Midtrans payments.
- `repository.go` ONLY interacts with `*gorm.DB`. **Never import `github.com/gin-gonic/gin` inside a repository.**

2. **Centralized Wiring:** Every new domain module in `app/<domain>/` MUST be instantiated in `internal/container/container.go` and mounted in `internal/app/routes.go`.
3. **Model Registration for Atlas:** Any new GORM model struct must be registered in `models/postgre.go` (and `cmd/migrate/atlas/main.go` if applicable) so Atlas can detect schema diffs.

---

## 4. Hybrid Context & Skill Routing

Before writing code, load the relevant specialized context:

| Task / Context                                                     | Read This File First                                         |
| ------------------------------------------------------------------ | ------------------------------------------------------------ |
| Standardizing JSON errors, HTTP codes, & GORM error mapping        | `.agents/skills/error-handling/SKILL.md`                     |
| Naming Go files, structs, JSON fields, routes, & DB columns        | `.agents/skills/naming-convention/SKILL.md`                  |
| Creating/updating Midtrans transactions, webhooks, or invoices     | `.agents/skills/midtrans-payment/SKILL.md` _(recommended)_   |
| Modifying GORM models, running Atlas diffs, & updating seeds       | `.agents/skills/database-migration/SKILL.md` _(recommended)_ |
| Writing or regenerating Swaggo (`@Summary`, `@Router`) annotations | `.agents/skills/swagger-docs/SKILL.md` _(recommended)_       |

---

## 5. Essential CLI Commands

- **Verify Compilation & Static Analysis:** `go vet ./... && go build ./...`
- **Run API Server Locally:** `go run ./cmd/server/main.go`
- **Run Database Seeder:** `go run ./cmd/seed/main.go`
- **Generate Atlas Migration from GORM:** `atlas migrate diff <migration_name> --env gorm` _(check `atlas.hcl` for exact env name)_
- **Re-hash Atlas Migrations (if instructed):** `atlas migrate hash`
- **Regenerate Swagger Docs:** `swag init -g cmd/server/main.go`
- **Run Local Infrastructure:** `docker compose up -d`

---

## 6. Iron Rules (Never Break These)

1. **Never manually edit existing `migrations/*.sql` or `migrations/atlas.sum`.** Editing committed SQL files breaks the `atlas.sum` cryptographic checksum. Always create a new migration via Atlas.
2. **Guarantee Financial & Payment Idempotency.**

- Midtrans webhook callbacks (`app/payment/handler.go`) can arrive multiple times. Always verify the Midtrans signature/status and ensure transaction state updates (`donation_program_transaction`, `foster_children_transaction`, `social_program_transaction`, `social_program_invoice`, `finance_record`) are **idempotent** and wrapped in a GORM database transaction (`db.Transaction(...)`).

3. **Sanitize User-Generated Content.** Any rich-text or HTML input (e.g., `news`, `news_comment`, `prayer`, program descriptions) MUST be sanitized using `bluemonday` before saving to PostgreSQL.
4. **Always Use UUID v4/v7 (`google/uuid`) for Primary/Public Keys** according to the existing domain model pattern.
5. **Keep Seeders in Sync.** Whenever you change a GORM model schema in `app/<domain>/<domain>.go`, check `cmd/seed/models/` and update the corresponding seed model so `go run ./cmd/seed/main.go` never breaks.
6. **Never Commit Secrets.** Any new environment variable required in `config/*.go` must be added with a safe placeholder to `.env.example`.
7. **Always Propagate `context.Context`.** Pass `c.Request.Context()` from Gin handlers down to services, GORM (`db.WithContext(ctx)`), Redis, and MinIO calls.

---

## 7. Agent Workflow Protocol

When assigned a task:

1. **Read & Orient:** Read this `AGENTS.md`, check relevant `.agents/skills/*/SKILL.md`, and inspect a reference module (e.g., `app/donation_program/` or `app/social_program_transaction/`) to mirror the exact coding style.
2. **Ask Before Guessing:** If business rules around donations, foster child sponsorship, social subscriptions, or ambulance dispatch states are unclear, ask the owner first.
3. **Plan:** Present a short checklist of files to create/modify (`app/<domain>/*`, `internal/container/container.go`, `internal/app/routes.go`, `models/postgre.go`, `cmd/seed/models/*`). Wait for approval on architectural changes.
4. **Implement & Verify:** Write the code, then run `go vet ./...` and `go build ./...` to confirm zero compile errors.

---

## 8. What You Must Not Touch Without Approval

- `migrations/atlas.sum` and historical `migrations/*.sql`
- `.github/workflows/dev.yaml`
- `Dockerfile` & `docker-compose.yml`
- `atlas.hcl`
- `app/middleware/jwt.go` & `app/middleware/rate_limit.go` security parameters
