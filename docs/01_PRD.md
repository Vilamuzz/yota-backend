# 01_PRD — YOTA Backend (Foundation Management & Donation API)

> This document defines **what** the YOTA Backend does, **who** it serves, and the **business rules** governing its domains.
> Complements `AGENTS.md` (constitution & iron rules) and `docs/02_TRD.md` (technical architecture & database flows).

---

## 1. Product Overview

- **Name:** YOTA Backend (`vilamuzz-yota-backend`)
- **Type:** RESTful API Backend + Background Task Scheduler + Webhook Processor
- **Purpose:** Power the digital ecosystem of the YOTA Foundation—enabling transparent public fundraising, recurring social program subscriptions, foster children sponsorships, emergency ambulance dispatch, financial ledger auditing, and community content publishing.
- **One-liner:** A high-integrity backend platform for managing charity operations, automated recurring donations via Midtrans, ambulance logistics, and transparent financial reporting.

---

## 2. User Personas & Actors

### Persona 1: Donor / Member (Authenticated User)

- **Goal:** Donate to campaigns, sponsor foster children, subscribe to recurring social programs, or request emergency ambulance services.
- **Needs:**
  - Seamless registration, OAuth login, email verification, and password recovery.
  - Instant payment checkout via Midtrans (QRIS, Virtual Account, E-Wallet).
  - Real-time transaction history, recurring invoice tracking, and transparent expense reports showing where funds go.

### Persona 2: Public Visitor (Unauthenticated User)

- **Goal:** Browse foundation programs, verify financial transparency, read foundation news, and submit/amen prayers.
- **Needs:**
  - Fast, read-only access to active donation campaigns, foster children profiles, foundation gallery, news, prayers, and public financial summaries.

### Persona 3: Foundation Administrator / Staff

- **Goal:** Manage day-to-day charity operations, approve requests, disburse funds, and maintain audit logs.
- **Needs:**
  - Full CRUD over donation programs, social programs, foster children & candidates, and ambulance fleet.
  - Record program expenses (`*_expense`) and maintain unified financial records (`finance_record`).
  - Review and approve/dispatch `ambulance_service_request` and review `foster_children_candidate` submissions.
  - Perform system backups (`backup`) and monitor audit logs (`log`).

### Persona 4: External Systems (Midtrans & Cron Scheduler)

- **Goal:** Automate payment settlement and recurring billing lifecycles without human intervention.
- **Needs:**
  - **Midtrans Webhook (`app/payment`):** Idempotent HTTP callback endpoint to settle or expire transactions across all financial pillars.
  - **Internal Scheduler (`internal/scheduler`):** Automated cron jobs to generate monthly `social_program_invoice` records for active subscriptions and handle overdue statuses.

---

## 3. Core Functional Domains (7 Pillars)

### 3.1 Identity, Access & Audit (`auth`, `account`, `log`, `backup`)

- **Authentication (`auth`):**
  - Standard email/password registration & login issuing JWT tokens.
  - Social OAuth login via `goth`.
  - Email verification (`email_verification.go`) and secure password reset flows (`password_reset.go`).
- **Account Management (`account`):**
  - Role-based user profiles (Admin, Staff, Donor/User).
  - Profile updates, avatar uploads (via `media`), and role assignments.
- **System Audit & Backup (`log`, `backup`):**
  - Activity logging (`log`) tracking critical state changes and admin actions.
  - On-demand or scheduled database/system backup management (`backup`).

---

### 3.2 Pillar 1: Crowdfunding & Donation Programs (`donation_program*`)

- **Campaign Management (`donation_program`):**
  - Create and manage fundraising campaigns with target amounts, deadlines, slugs, cover media, and HTML-sanitized descriptions.
  - Track collected funds automatically as transactions settle.
- **Donations (`donation_program_transaction`):**
  - Donors initiate one-time donations -> system creates a pending transaction and requests a Midtrans Snap token/redirect URL.
  - Upon Midtrans settlement webhook, transaction status updates to `paid`/`settled` and increments campaign progress.
- **Disbursement & Transparency (`donation_program_expense`):**
  - Admins record how campaign funds are spent (amount, description, receipt/proof media) and sync outflows with the master `finance_record`.

---

### 3.3 Pillar 2: Recurring Social Programs (`social_program*`)

- **Program Catalog (`social_program`):**
  - Ongoing social initiatives designed for recurring/subscription-based community support.
- **Subscriptions (`social_program_subscription`):**
  - Users subscribe to a social program with a recurring commitment cycle.
  - Users can view, pause, or cancel their active subscriptions.
- **Automated Invoicing (`social_program_invoice`):**
  - Generated automatically by `internal/scheduler` (or upon initial subscription) for each billing period.
  - Tracks `unpaid`, `paid`, `expired`, or `cancelled` states.
- **Subscription Payments (`social_program_transaction`):**
  - Links a user's Midtrans payment attempt to a specific `social_program_invoice`.
- **Program Expenses (`social_program_expense`):**
  - Tracks operational spending and aid distribution for each social program.

---

### 3.4 Pillar 3: Foster Children Sponsorship (`foster_children*`)

- **Candidate Pipeline (`foster_children_candidate`):**
  - Submission and verification workflow for prospective foster children before they are officially enrolled.
- **Foster Children Directory (`foster_children`):**
  - Profiles of enrolled foster children (biodata, education level, status, achievements, photos).
- **Sponsorship & Donations (`foster_children_transaction`):**
  - Direct financial contributions toward a specific foster child or the general foster pool via Midtrans.
- **Child Care Expenses (`foster_children_expense`):**
  - Records tuition, living allowances, and medical expenses disbursed for foster children.

---

### 3.5 Pillar 4: Master Financial Ledger (`finance_record`, `payment`)

- **Unified Ledger (`finance_record`):**
  - Single source of truth for foundation-wide cash flow (inflows from settled transactions across Donation, Social, and Foster pillars; outflows from recorded expenses and general operational costs).
- **Unified Payment Webhook (`payment`):**
  - Single callback handler (`app/payment/handler.go`) receiving Midtrans HTTP notifications, verifying cryptographic signatures, identifying the transaction domain (Donation vs. Social Invoice vs. Foster Child), and atomically updating both the domain transaction and financial balances.

---

### 3.6 Pillar 5: Ambulance Fleet & Emergency Dispatch (`ambulance*`)

- **Fleet Inventory (`ambulance`):**
  - Manage ambulance vehicles, license plates, equipment specs, availability status (`available`, `on_duty`, `maintenance`), and photos.
- **Service Requests (`ambulance_service_request`):**
  - Users/patients submit ambulance booking or emergency requests (pickup/drop-off locations, patient condition, contact info, urgency).
  - Admins approve, assign an available vehicle, or reject requests.
- **Trip & Usage History (`ambulance_history`):**
  - Logs completed ambulance dispatches, mileage/routes, driver notes, and timestamps for operational transparency.

---

### 3.7 Pillar 6: Content, Community & Media (`foundation_profile`, `news*`, `prayer`, `gallery`, `media`)

- **Foundation Profile (`foundation_profile`):**
  - Singleton/organization metadata: vision, mission, legal documents, address, contact info, and leadership structure.
- **News & Comments (`news`, `news_comment`):**
  - SEO-friendly articles (auto-slugged via `gosimple/slug`, sanitized via `bluemonday`).
  - Authenticated users can post and manage comments (`news_comment`) on articles.
- **Community Prayers (`prayer`):**
  - Donors and visitors can post prayers/supplications and interact (e.g., "Aamiin" counter/support).
- **Gallery (`gallery`) & Media Service (`media`):**
  - Centralized image/document processing (`golang.org/x/image`) and S3/MinIO object storage management (`minio-go/v7`).

---

## 4. Non-Functional Requirements (NFRs)

### Financial & Data Integrity

- **ACID Transactions:** All payment settlements, invoice updates, and expense recordings must execute inside GORM database transactions (`db.Transaction`).
- **Webhook Idempotency:** Replaying the same Midtrans `settlement` or `expire` webhook payload multiple times must never double-count balances or duplicate `finance_record` entries.
- **Auditability:** Financial records and settled transactions are immutable; corrections must be handled via adjustment records or explicit status transitions, never hard-deleted.

### Security

- **HTML Sanitization:** All rich-text inputs (`news`, `news_comment`, `prayer`, program descriptions) must pass through `bluemonday` before persistence to prevent XSS.
- **Authentication & Authorization:** Protected endpoints require valid JWTs (`golang-jwt/jwt/v5`) and appropriate role permissions.
- **Rate Limiting:** Public and authentication endpoints are protected against brute-force and abuse using Redis-backed rate limiting (`ulule/limiter/v3`).
- **CORS & Recovery:** Enforced globally via `app/middleware/`.

### Performance & Scalability

- **Response Time:** < 200ms p95 for standard read/write API endpoints; < 500ms for endpoints involving S3 presigned URLs or Midtrans API calls.
- **Caching:** Frequently accessed public data and rate-limit counters stored in Redis (`go-redis/v9`).
- **Concurrency:** Safe concurrent operations using `golang.org/x/sync` (`errgroup`) where parallel I/O is needed.

---

## 5. Key User Stories & Acceptance Criteria

### US-1: Donor Completes a Campaign Donation

**As a** donor  
**I want to** donate to an active `donation_program` and pay via Midtrans  
**So that** my contribution is recorded and the campaign's collected amount increases automatically.

**Acceptance Criteria:**

- `POST` to donation transaction endpoint validates amount and campaign active status.
- System creates a `pending` transaction and returns a Midtrans payment URL/token.
- When Midtrans sends a `settlement` webhook to `app/payment`, the system verifies the signature, marks the transaction `paid`, updates the campaign's collected total, and logs the inflow in `finance_record`.

### US-2: Subscriber Receives & Pays Recurring Social Program Invoice

**As a** social program subscriber  
**I want** the system to generate my periodic invoice and allow me to pay it  
**So that** my recurring charity commitment stays active without manual re-registration.

**Acceptance Criteria:**

- Background scheduler (`internal/scheduler`) generates a `social_program_invoice` when a subscription's billing cycle is due.
- Subscriber initiates payment via `social_program_transaction`.
- Settled payment marks the invoice `paid` and extends the next billing date on `social_program_subscription`.

### US-3: Citizen Requests an Ambulance

**As a** community member in need of medical transport  
**I want to** submit an `ambulance_service_request`  
**So that** the foundation can dispatch an available ambulance and track the trip.

**Acceptance Criteria:**

- User submits pickup details, patient info, and requested time.
- Admin reviews the request, assigns an available `ambulance`, and transitions the ambulance status.
- Upon completion, an `ambulance_history` record is created and the ambulance returns to `available` status.

### US-4: Admin Publishes Transparent Program Expenses

**As a** foundation admin  
**I want to** record expenses for donation programs, social programs, and foster children  
**So that** donors can see proof of how funds are utilized.

**Acceptance Criteria:**

- Admin submits expense details with optional media proof (uploaded to MinIO/S3 via `app/media`).
- Expense is saved in the respective `*_expense` table and reflected in `finance_record`.
- Public users can query program expenses via read-only endpoints.

---

## 6. Success Metrics

| Metric                        | Target                     | Measurement                                                                      |
| :---------------------------- | :------------------------- | :------------------------------------------------------------------------------- |
| **Payment Webhook Accuracy**  | 100% (0 balance drift)     | Reconciliation between Midtrans dashboard, `*_transaction`, and `finance_record` |
| **API Availability / Uptime** | ≥ 99.9%                    | Zero unhandled panics (`app/middleware/recovery.go` active)                      |
| **Build & Static Analysis**   | 0 errors / warnings        | `go vet ./... && go build ./...` in CI (`.github/workflows/dev.yaml`)            |
| **Migration Integrity**       | 100% reproducible          | `migrations/atlas.sum` verified by Atlas CLI on fresh PostgreSQL instance        |
| **Swagger Coverage**          | 100% public & admin routes | `swag init -g cmd/server/main.go` compiles cleanly                               |

---

## 7. Out of Scope (Current Version)

- ❌ Native mobile app UI or web frontend code (handled in separate frontend repositories).
- ❌ Multi-currency settlement (all transactions are denominated in IDR via Midtrans).
- ❌ Direct bank account auto-disbursement/payouts (expenses are disbursed offline and recorded in `*_expense` with proof).
- ❌ Real-time GPS telemetry streaming for ambulances (dispatch and history are status/milestone-based).

---

## 8. Domain Glossary

| Term                          | Domain Module               | Definition                                                                                              |
| :---------------------------- | :-------------------------- | :------------------------------------------------------------------------------------------------------ |
| **Donation Program**          | `donation_program*`         | Time-bound or target-bound crowdfunding campaign (one-time donations).                                  |
| **Social Program**            | `social_program*`           | Ongoing charity program supporting recurring user subscriptions and periodic invoices.                  |
| **Foster Children Candidate** | `foster_children_candidate` | A prospective child submitted for review before becoming an active `foster_children` beneficiary.       |
| **Finance Record**            | `finance_record`            | Master ledger aggregating inflows (settled transactions) and outflows (expenses) across the foundation. |
| **Media Service**             | `media`                     | Internal service wrapping MinIO/S3 uploads; has no HTTP handler of its own.                             |

---

**Last Updated:** 2026-10-17  
**Owner:** Vilamuzz  
**Status:** Active Development
