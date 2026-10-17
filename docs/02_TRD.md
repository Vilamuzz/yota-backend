# 02_TRD — YOTA Backend (Technical Requirements Document)

> This document defines **how** the YOTA Backend is architected, the rationale behind technical choices, data flows, and infrastructure constraints.
> Complements `AGENTS.md` (constitution & iron rules) and `docs/01_PRD.md` (product scope & business domains).

---

## Architecture

- **Modular Monolith REST API:** Built on Go `1.25.0` and Gin (`v1.11.0`), organized by domain inside `app/<domain>/`.
- **Strict 3-Layer Domain Encapsulation:** Every domain module follows `handler.go` (HTTP transport) → `service.go` (business logic & external integrations) → `repository.go` (GORM database access), backed by `<domain>.go` (entities/interfaces), `request.go` (DTO validation), and `response.go` (JSON serialization).
- **Centralized Dependency Injection:** All repositories, services, handlers, and infrastructure clients (PostgreSQL, Redis, MinIO, Midtrans, Email, OAuth) are instantiated once in `internal/container/container.go` and wired to Gin routes in `internal/app/routes.go`.
- **Declarative Schema Migrations:** GORM models registered in `models/postgre.go` are inspected by `ariga.io/atlas-provider-gorm` (`cmd/migrate/atlas/main.go`) to generate versioned SQL files and cryptographic checksums (`migrations/atlas.sum`).
- **In-Process Background Scheduler:** `internal/scheduler/scheduler.go` runs `robfig/cron/v3` alongside the API server to automate recurring social program invoicing and expiration sweeps.
- **Unified Webhook & Media Services:**
  - `app/payment/handler.go` acts as the single ingress for Midtrans HTTP payment notifications across all financial modules.
  - `app/media/` acts as an internal service (no HTTP handler) wrapping MinIO/S3 object storage and image processing (`golang.org/x/image`).

---

## Stack Decisions (and Why)

| Choice                                         | Reason                                                                                                                                                                        |
| :--------------------------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Go 1.25.0 + Gin (`v1.11.0`)**                | High throughput, low memory footprint, fast compilation, and mature middleware ecosystem (`cors`, `gin-swagger`).                                                             |
| **GORM (`v1.31.1`) + PostgreSQL**              | Relational integrity and ACID transactions required for donations, invoices, and financial ledgers (`finance_record`).                                                        |
| **Atlas + `atlas-provider-gorm`**              | Eliminates unsafe production `db.AutoMigrate()` calls by generating reviewable, versioned SQL migrations directly from GORM structs.                                          |
| **Midtrans (`midtrans-go v1.3.8`)**            | Indonesia-native payment gateway supporting QRIS, Virtual Accounts, and E-Wallets out of the box for IDR charity transactions.                                                |
| **Redis (`go-redis/v9`) + `ulule/limiter/v3`** | Distributed rate limiting for public/auth routes and low-latency caching/session state.                                                                                       |
| **MinIO SDK (`minio-go/v7`) + `x/image`**      | S3-compatible object storage keeps binary media (receipts, gallery, foster child photos, news banners) out of PostgreSQL; `x/image` handles server-side resizing/compression. |
| **Goth (`v1.82.0`) + JWT (`jwt/v5`)**          | Stateless JWT authentication for API clients paired with multi-provider OAuth2 login via Goth + Gorilla Sessions.                                                             |
| **Bluemonday (`v1.0.27`)**                     | Sanitizes rich-text HTML in news articles, comments, prayers, and campaign descriptions to prevent stored XSS.                                                                |
| **Robfig Cron (`cron/v3`)**                    | Lightweight in-process cron scheduler for recurring subscription invoices without needing a separate worker cluster.                                                          |

---

## Constraints

- **Database Migration Discipline:** `migrations/*.sql` and `migrations/atlas.sum` are strictly managed by Atlas CLI. Manual edits to committed migration files invalidate the hash chain and fail deployments.
- **Currency & Locale:** All financial amounts (`donation_program`, `social_program`, `foster_children`, `finance_record`) are denominated in Indonesian Rupiah (IDR) as integer/exact numeric types—never floating-point (`float32`/`float64`) to prevent rounding drift.
- **Webhook Idempotency:** Midtrans guarantees at-least-once delivery. `app/payment/handler.go` and downstream transaction services must tolerate duplicate `settlement`, `pending`, or `expire` notifications without double-crediting campaigns or duplicating `finance_record` entries.
- **Stateless API Tier:** Local filesystem storage is ephemeral inside Docker containers (`Dockerfile`, `docker-compose.yml`). All persistent uploads must go through `app/media/` to S3/MinIO.
- **Single-Binary Entrypoints:** Separate `cmd/` entrypoints for server (`cmd/server`), migrations (`cmd/migrate/atlas`), and seeding (`cmd/seed`).

---

## Non-Functional Requirements

- **API Latency:** < 200ms p95 for standard CRUD/read endpoints; < 500ms for endpoints invoking Midtrans Snap API or S3/MinIO uploads.
- **Financial Consistency:** 100% ACID compliance via `db.Transaction(...)` whenever mutating transaction statuses, campaign collected totals, invoice states, and `finance_record` rows simultaneously.
- **Security:**
  - Passwords hashed via `golang.org/x/crypto` (bcrypt).
  - Protected routes gated by `app/middleware/jwt.go`.
  - Public and auth routes throttled by `app/middleware/rate_limit.go`.
  - Panic isolation via `app/middleware/recovery.go` + structured logging via `sirupsen/logrus`.
- **Zero Compile/Vet Warnings:** `go vet ./...` and `go build ./...` must pass with 0 errors before merging.

---

## Dependencies (Approved List — `go.mod`)

### Core Runtime & Framework

- `github.com/gin-gonic/gin` v1.11.0 — HTTP router & middleware engine
- `github.com/gin-contrib/cors` v1.7.6 — CORS middleware
- `github.com/joho/godotenv` v1.5.1 — `.env` loader (`config/*.go`)
- `github.com/sirupsen/logrus` v1.9.3 — structured logger

### Database, Migrations & Caching

- `gorm.io/gorm` v1.31.1 & `gorm.io/driver/postgres` v1.6.0 — ORM & PostgreSQL driver
- `ariga.io/atlas-provider-gorm` v0.6.0 — GORM-to-Atlas schema loader
- `github.com/redis/go-redis/v9` v9.17.3 — Redis client
- `github.com/ulule/limiter/v3` v3.11.2 — rate limiting middleware

### Auth, Security & Validation

- `github.com/golang-jwt/jwt/v5` v5.3.0 — JWT signing & verification
- `github.com/markbates/goth` v1.82.0 & `github.com/gorilla/sessions` v1.4.0 — OAuth2 flows
- `golang.org/x/crypto` v0.48.0 — password hashing & cryptography
- `github.com/microcosm-cc/bluemonday` v1.0.27 — HTML XSS sanitizer
- `github.com/go-playground/validator/v10` v10.27.0 _(via Gin)_ — struct tag validation

### Payments, Storage, Scheduling & Utilities

- `github.com/midtrans/midtrans-go` v1.3.8 — Midtrans payment gateway SDK
- `github.com/minio/minio-go/v7` v7.0.98 & `golang.org/x/image` v0.37.0 — S3 storage & image processing
- `github.com/robfig/cron/v3` v3.0.1 — background job scheduler
- `github.com/google/uuid` v1.6.0 — UUID generation
- `github.com/gosimple/slug` v1.15.0 — URL-safe slug generation
- `golang.org/x/sync` v0.20.0 — `errgroup` concurrency primitives
- `github.com/swaggo/swag` v1.16.6, `gin-swagger` v1.6.1, `files` v1.0.1 — OpenAPI/Swagger docs

> **Adding any new dependency via `go get` requires explicit owner approval (`AGENTS.md` §6).**

---

## Explicitly Rejected

- **No Runtime `db.AutoMigrate()` in Server Startup:** Schema mutations must never run automatically on `cmd/server/main.go` boot; all DDL changes must go through Atlas versioned SQL migrations.
- **No Direct DB Access in Handlers:** `handler.go` files must never import `gorm.io/gorm` or execute SQL queries directly.
- **No Cross-Repository Database Joins Outside Service Orchestration:** Domain modules should remain decoupled; cross-domain workflows (e.g., payment webhook updating transaction + `finance_record`) are orchestrated at the service/container layer.
- **No Local Disk File Uploads:** Uploaded images/documents must stream to MinIO/S3 via `app/media/`, never saved to local container directories.
- **No Hard-Deletes on Financial Ledgers:** Settled transactions (`*_transaction`), paid invoices (`social_program_invoice`), and `finance_record` entries must never be hard-deleted (`DELETE FROM`).
- **No External Message Brokers (Kafka/RabbitMQ) Yet:** Background tasks and periodic invoicing use in-process `robfig/cron/v3` and `golang.org/x/sync` to keep infrastructure lean.

---

## Data Flow

### 1. Standard Request-Response Flow (CRUD & Queries)

1. Client sends HTTP request → Gin router (`internal/app/routes.go`).
2. Middleware pipeline executes: `recovery` → `logger` → `cors` → `rate_limit` → `jwt` (if protected).
3. `app/<domain>/handler.go` binds & validates payload using `request.go` (`c.ShouldBindJSON` / `ShouldBind`).
4. `app/<domain>/service.go` executes business logic, sanitizes HTML (`bluemonday`), uploads files via `app/media/` (if multipart), and calls `repository.go`.
5. `app/<domain>/repository.go` executes GORM query with `c.Request.Context()` (`db.WithContext(ctx)`).
6. `handler.go` maps entity to `response.go` DTO and returns standardized JSON response.

### 2. Financial & Midtrans Payment Flow

1. **Checkout Initiation:** User submits a donation/sponsorship/invoice payment → `*_transaction/service.go` creates a `pending` transaction record in PostgreSQL → calls `midtrans-go` to generate a Snap token/redirect URL → returns payment details to client.
2. **Webhook Settlement:** User completes payment → Midtrans sends HTTP POST notification to `app/payment/handler.go` → handler verifies signature & order ID → routes notification to the target transaction service (`donation_program_transaction`, `social_program_transaction`, or `foster_children_transaction`).
3. **Atomic Ledger Update:** Inside a GORM transaction (`db.Transaction`), the service marks the transaction `paid`/`settled`, updates the parent target (campaign collected total or `social_program_invoice` status), and inserts an inflow entry into `finance_record`.

### 3. Recurring Subscription & Invoice Flow (`internal/scheduler`)

1. `robfig/cron/v3` triggers scheduled job in `internal/scheduler/scheduler.go`.
2. Scheduler queries due `social_program_subscription` records → generates `unpaid` `social_program_invoice` records for the billing cycle → notifies user / awaits payment via `social_program_transaction`.

---

## CI/CD & Deployment Pipeline (`.github/workflows/dev.yaml`)

1. **Trigger:** Push or Pull Request to target development/deployment branch.
2. **Build & Verification:** Checkout code → Setup Go `1.25.0` → Verify dependencies (`go mod download`) → Static check & build (`go vet ./... && go build ./...`).
3. **Containerization:** Build production container image via `Dockerfile` and orchestrate services (PostgreSQL, Redis, MinIO/S3, API Server) via `docker-compose.yml`.
4. **Database Readiness:** Apply pending Atlas migrations from `migrations/` before serving traffic.

---

**Last Updated:** 2026-10-17  
**Owner:** Vilamuzz  
**Status:** Active Development
