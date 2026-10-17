# 03_APP_FLOW — YOTA Backend (API & Domain Lifecycle Flows)

> This document maps out the API module inventory, access tiers, step-by-step domain lifecycles (payments, subscriptions, ambulance dispatch, auth), and edge-case handling.
> Complements `AGENTS.md` (constitution), `docs/01_PRD.md` (product scope), and `docs/02_TRD.md` (technical architecture).
> **Source of truth:** `app/*/handler.go` (routes), `app/*/service.go` (flows), and `internal/container/container.go` (wiring + cron).

---

## Roles (`pkg/enum`)

Authorization is role-based. A JWT carries `account_id`, all active `roles`, and one `active_role` (switchable via `POST /api/auth/switch-role`). `RequireRoles(...)` checks the token's roles.

| Constant                 | Role name                     |
| :----------------------- | :---------------------------- |
| `RoleOrangTuaAsuh`       | Orang Tua Asuh (default role on registration — foster parent / donor) |
| `RoleChairman`           | Ketua Yayasan                 |
| `RoleSocialManager`      | Koordinator Sosial            |
| `RoleFinance`            | Bendahara                     |
| `RoleAmbulanceManager`   | Penanggung Jawab Ambulans     |
| `RoleAmbulanceDriver`    | Supir Ambulans                |
| `RolePublicationManager` | Penanggung Jawab Publikasi    |
| `RoleSuperadmin`         | Superadmin                    |

---

## Module & Endpoint Inventory

All handlers are registered in `Container.RegisterHandlers` (`internal/container/container.go`) onto the `/api` group created in `internal/app/routes.go`. Admin routes live under `/api/admin/...` (or `/api/superadmin/...`). Swagger is at `/swagger/*any` and a health check at `/health` (both outside `/api`).

| Domain Module                  | Public / Self routes (under `/api`)                                                                                     | Admin routes (roles)                                                                                              | Midtrans? |
| :----------------------------- | :---------------------------------------------------------------------------------------------------------------------- | :---------------------------------------------------------------------------------------------------------------- | :-------- |
| `auth`                         | `/auth/register`, `/login`, `/verify-email`, `/resend-verification`, `/forget-password`, `/reset-password`, `/oauth/:provider[/callback]`, `/auth/switch-role` (JWT) | —                                                                                                                 | No        |
| `account`                      | `GET /roles`; `/me` (JWT): `GET`, `PATCH /profile`, `PATCH /password`                                                    | `/admin/accounts` (Social Mgr, Ambulance Mgr): list, `/drivers`, `/foster-parents`, `/:accountId`; `/superadmin/accounts` (Superadmin): list, get, `PATCH /:accountId/ban`, `POST/PATCH /:accountId/roles/:roleId` | No        |
| `foundation_profile`           | `GET /foundation-profile`                                                                                                | `/admin/foundation-profile` (Superadmin): `POST`, `PUT /:id`                                                      | No        |
| `donation_program`             | `GET /donation-programs`, `/:slug`                                                                                       | `/admin/donation-programs` (Finance): list, get, `POST`, `PUT /:id`, `DELETE /:id`, `PATCH /:id/active`, `/:id/archive` | No        |
| `donation_program_transaction` | `POST /donation-programs/:slug/transactions` (optional JWT — guests allowed), `GET /:slug/transactions` (settled only); `/donation-programs/transactions` (Orang Tua Asuh): list/get own | `/admin/donation-programs` (Finance): `/:id/transactions` (+`/monthly-income`, `/export`), `GET /transactions/:id`, `POST /:id/transactions` (offline), `POST /transactions/:id/cancel` (offline only) | **Yes**   |
| `donation_program_expense`     | `GET /donation-programs/:slug/expenses` (+`/export`), `GET /donation-programs/expenses/:id`                              | `/admin/donation-programs` (Finance): `/:id/expenses` (+`/monthly-expense`, `/export`), `POST /:id/expenses`, `DELETE /expenses/:id` | No        |
| `prayer`                       | `GET /donation-programs/:slug/prayers`, `/donation-programs/prayers/:id` (optional JWT); `POST /donation-programs/prayers/:id/amen`, `/report` (Orang Tua Asuh) | `/admin/donation-programs/prayers` (Publication Mgr): reported list, `PATCH /:id/allow`, `DELETE /:id`            | No        |
| `social_program`               | `GET /social-programs`, `/:slug` (optional JWT)                                                                          | `/admin/social-programs`: list/get (Social Mgr, Chairman, Finance); `POST`, `PUT`, `DELETE`, `PATCH /:id/complete` (Social Mgr); `PATCH /:id/approve`, `/reject` (Chairman) | No        |
| `social_program_subscription`  | `POST /social-programs/:id/subscribe`, `PATCH /:id/unsubscribe` (Orang Tua Asuh)                                          | `/admin/social-programs` (Social Mgr, Finance): `/subscribers…`, `/:id/subscriptions`, `/accounts/:account_id/subscriptions`, `/subscriptions/:id`, `POST /:id/subscriptions` (offline), `PATCH /subscriptions/:id/deactivate` | No        |
| `social_program_invoice`       | `/social-programs/subscriptions/invoices/me` (Orang Tua Asuh): list, `/:id`                                               | `/admin/social-programs/subscriptions/invoices` (Social Mgr, Finance): list, `/subscription/:id`                  | No        |
| `social_program_transaction`   | `POST /social-programs/subscriptions/invoices/:id/pay` (Orang Tua Asuh)                                                   | `POST /admin/social-programs/subscriptions/invoices/:id/pay-offline` (Social Mgr)                                 | **Yes**   |
| `social_program_expense`       | `GET /social-programs/:slug/expenses` (+`/export`), `GET /social-programs/expenses/:id`                                   | `/admin/social-programs` (Finance, Social Mgr): `/:id/expenses`, `/expenses/:id`, `POST /:id/expenses`, `DELETE /expenses/:id` | No        |
| `foster_children_candidate`    | `/foster-children/candidates` (Orang Tua Asuh): `POST`, list/get own, `DELETE /:id` (cancel)                              | `/admin/foster-children/candidates` (Social Mgr, Chairman): list, get, `PATCH /:id/accept`, `/reject`              | No        |
| `foster_children`              | `GET /foster-children`, `/:slug`                                                                                         | `/admin/foster-children`: list/get (Social Mgr, Finance); `POST`, `PUT /:id`, `DELETE /:id` (Social Mgr)           | No        |
| `foster_children_transaction`  | `POST /foster-children/:slug/transactions` (optional JWT); `/foster-children/transactions/me` (Orang Tua Asuh): list, `/:id` | `/admin/foster-children` (Social Mgr, Finance): `/:id/transactions`, `/transactions/:id`, `POST /:id/transactions` (offline) | **Yes**   |
| `foster_children_expense`      | `GET /foster-children/:slug/expenses` (+`/export`), `/foster-children/expenses/:id`                                       | `/admin/foster-children` (Finance, Social Mgr): `/:id/expenses` (+`/export`), `/expenses/:id`, `POST /:id/expenses`, `DELETE /expenses/:id` | No        |
| `finance_record`               | `GET /finance-records/summary`                                                                                           | `GET /admin/finance-records/summary`, `/monthly-trend` (Finance, Social Mgr)                                      | No        |
| `payment`                      | `POST /api/webhooks/midtrans/notification` (verified by Midtrans signature)                                              | —                                                                                                                 | **Yes**   |
| `ambulance`                    | `GET /ambulances`, `/:id`                                                                                                | `/admin/ambulances`: list/get (Ambulance Mgr, Driver); `POST`, `PUT /:id`, `DELETE /:id` (Ambulance Mgr)           | No        |
| `ambulance_service_request`    | `/ambulances/requests` (Orang Tua Asuh): `POST`, list/get own, `PATCH /:id/cancel`                                        | Manager: `/admin/ambulances/requests` list/get, `PATCH /:id/accept` (assigns `ambulanceId`), `/reject`. Driver: `/admin/ambulances/requests/assigned` list, `/:id/detail`, `PATCH /:id/start`, `/complete`, `/cancel` | No        |
| `ambulance_history`            | `GET /ambulances/:id/history`, `/:id/history/summary`                                                                    | Manager: `/admin/ambulances/history` (`/:id`, `POST`, `PUT /:id`, `DELETE /:id`, `/summary`, `/monthly-trend`). Driver: `/admin/ambulances/history/driver` (list, `POST`, `PUT`, `DELETE`, `/summary`, `/monthly-trend`) | No        |
| `news`                         | `GET /news`, `/:slug`                                                                                                    | `/admin/news` (Publication Mgr): list, get, `POST`, `PUT`, `DELETE`, `PATCH /:id/publish`, `/archive`              | No        |
| `news_comment`                 | `GET /news/:slug/comments`, `/news/comments/:id`; `POST /news/:slug/comments`, `POST /news/comments/:id/report` (JWT)     | `/admin/news/comments` (Publication Mgr): reported list, `PATCH /:id/allow`, `DELETE /:id`                         | No        |
| `gallery`                      | `GET /galleries`, `/:slug`                                                                                               | `/admin/galleries` (Publication Mgr): list, get, `POST`, `PUT`, `DELETE`, `PATCH /:id/publish`, `/archive`         | No        |
| `log`                          | —                                                                                                                        | `GET /logs` (Superadmin)                                                                                          | No        |
| `backup`                       | —                                                                                                                        | `/admin/backups` (Superadmin): `POST`, `GET`, `GET /:id/download`, `DELETE /:id`, `POST /cleanup`                  | No        |
| `media`                        | _(Internal service only — no handler; used by `gallery`/`news`)_                                                          | —                                                                                                                 | No        |

> `donation_program_expense`, `social_program_expense` and `foster_children_expense` upload proof files directly through the S3 client (`UploadFile`) in their services.

---

## Key Operational Flows

### Flow 1: Authentication, Verification & OAuth Lifecycle (`auth`, `account`)

```text
1. Registration (POST /api/auth/register)
   → `auth/service.go` validates username (3–20 chars, unique), email format/uniqueness, and password strength
   → Hashes password with bcrypt
   → Creates `accounts` row + `user_profiles` row + `account_roles` row (role: Orang Tua Asuh, is_default/is_active)
   → Creates `email_verification_tokens` row (random 32-byte hex, 24h expiry)
   → Sends verification email asynchronously (goroutine) via the email service

2. Email Verification & Login
   → POST /api/auth/verify-email marks `email_verified = true` (resend: POST /resend-verification, rate-limited 3/min)
   → POST /api/auth/login: checks credentials, then `email_verified` (403 if not), then `is_banned` (403 if banned)
   → Collects the account's active roles; the `is_default` role becomes `active_role`
   → Issues a signed JWT (`UserJWTClaims`: account ID, roles, active role, TTL from `config.GetJWTTTL()`)

3. Role Switching (POST /api/auth/switch-role, JWT)
   → Verifies the requested role is active on the account → re-issues JWT with new `active_role`

4. Password Reset
   → POST /forget-password (rate-limited 5/min) creates `password_reset_tokens` and emails the link
   → POST /reset-password consumes the token (`is_used`)

5. OAuth2 Login (`markbates/goth`)
   → GET /api/auth/oauth/:provider (rate-limited 10/min) redirects to the provider
   → GET /api/auth/oauth/:provider/callback → `service.OAuthLogin` → issues JWT
```

---

### Flow 2: One-Time Donation & Foster Sponsorship (`donation_program*`, `foster_children_transaction`, `payment`)

```text
1. Checkout Initiation (online)
   Donor (guest, via optional JWT, or logged-in) POSTs amount to
   `/donation-programs/:slug/transactions` or `/foster-children/:slug/transactions`
   → `*_transaction/service.go` sanitizes inputs (`pkg.SanitizeStrict`), verifies the target exists and (for donations) is `active`
     (`expired`/`completed`/non-active programs are rejected)
   → Defaults donor name to "Hamba Allah" when empty
   → Generates Order ID: `DON-<uuid>` (donation), `FC-<uuid>` (foster child); social invoices use `SPI-<uuid>`
   → Calls Midtrans Snap (`CreateSnapTransaction`) → gets `snap_token` + `redirect_url`
   → Saves transaction (`is_online = true`, `provider = midtrans`, `transaction_status = pending`)
   → Donation only: if the donor included a prayer, creates a `prayers` row with `is_published = false`
   → Returns transaction (incl. snap token/redirect URL)

2. Midtrans Webhook (POST /api/webhooks/midtrans/notification)
   → `app/payment/handler.go` routes by Order ID prefix: `DON-` → donation, `SPI-` → social, `FC-` → foster child
     (unknown prefix → 200 OK, ignored)
   → The service verifies SHA-512 signature: sha512(order_id + status_code + gross_amount + server_key); mismatch → 401

3. Settlement (in each `HandleNotification`)
   → Loads transaction by `order_id`; if `transaction_status` already equals the incoming status → 200 "no change" (idempotency)
   → Updates `transaction_status`, `fraud_status`, `transaction_id`
   → If `settlement`, or `capture` with fraud status != `challenge`:
       sets `paid_at`
       donation: publishes the linked prayer (`is_published = true`)
       inserts an income `finance_records` row (`fund_type`, `fund_id`, `source_type = transaction`, `source_id`, `amount`)
       (collected totals are computed by queries over settled transactions — there is no stored counter to increment)
   → Other statuses (`expire`, `cancel`, `deny`, …) are stored as-is in `transaction_status` with no ledger entry

4. Offline (cash) transactions
   Finance (donation), Social Manager / Finance (foster child) POST `/admin/.../:id/transactions`
   → Stored immediately as `is_online = false`, status `settlement`, with an income `finance_records` row
   → Offline donations can be cancelled (`POST /admin/donation-programs/transactions/:id/cancel`); online ones cannot
```

---

### Flow 3: Social Program Lifecycle, Subscription & Cron Invoicing (`social_program*`, `internal/container`)

```text
1. Program Approval
   Social Manager creates a `social_program` (POST /admin/social-programs)
   → Chairman approves (`PATCH /:id/approve`) or rejects (`PATCH /:id/reject`, with `rejection_reason`)
   → Social Manager can mark it `completed` (`PATCH /:id/complete`)
   (statuses: `pending`, `active`, `completed`, `rejected`)

2. Subscription
   Orang Tua Asuh POSTs `/social-programs/:id/subscribe`
   → Creates `social_program_subscriptions` (`status: active`), or reactivates an existing inactive one
   → Returns 409 if an active subscription already exists
   → No invoice is created at subscribe time — invoices come from the scheduler
   → Unsubscribe: `PATCH /social-programs/:id/unsubscribe` → status `inactive`
   → Admins can create offline subscriptions and deactivate any subscription

3. Cron Jobs (`robfig/cron/v3`, `internal/scheduler` registered in `Container.initScheduler`)
   0 0 * * *  update-expired-donations   → `DonationService.UpdateExpiredDonationProgram`
   0 0 * * *  generate-monthly-invoices  → for subscriptions whose program `billing_day` is today (all later days on the last day of the month),
                                           creates one `social_program_invoices` row per (subscription, billing_period = 1st of month),
                                           `status: pending`, `due_date` = end of month; skips if it already exists (unique index); emails the subscriber
   5 0 * * *  mark-overdue-invoices      → marks past-due `pending` invoices as `overdue`
   0 2 * * *  database-backup            → `BackupService.CreateBackup`
   0 3 * * *  backup-cleanup             → deletes backups older than 7 days

4. Invoice Payment
   Subscriber lists own invoices (`/social-programs/subscriptions/invoices/me`) and POSTs `/…/invoices/:id/pay`
   → `social_program_transaction` is created with a Midtrans Snap token (Order ID prefix `SPI-`)
   → Webhook settlement: updates the transaction, marks the invoice `paid`, increments the subscription's `total_paid_periods`,
     and inserts an income `finance_records` row (`fund_type = social_program`)
   → Social Manager can record an offline payment (`POST /admin/…/invoices/:id/pay-offline`)
```

---

### Flow 4: Expense Disbursement & Financial Transparency (`*_expense`, `finance_record`)

```text
1. Admin Submits Expense
   Finance (donation), Finance / Social Manager (social & foster child) POST multipart/form-data to
   `/admin/donation-programs/:id/expenses`, `/admin/social-programs/:id/expenses`, or `/admin/foster-children/:id/expenses`
   → Includes title, amount, expense date (YYYY-MM-DD), note, optional proof file

2. Validation, Upload & Ledger Sync (`*_expense/service.go`)
   → Validates fields; donation expenses additionally must not exceed available funds (collected − total expenses)
   → Uploads the proof file via the S3 client to a per-domain folder (e.g. `donation-expenses`)
   → Inserts the expense row (`created_by` = acting account)
   → Inserts a matching `finance_records` row (`source_type = expense`, `transaction_date` = expense date)
   → Writes an audit entry via `logService.CreateLog` (action `CREATE`, entity name, old/new values)

3. Public Transparency
   Expense lists and CSV exports are public (`GET /…/:slug/expenses`, `/export`);
   `GET /api/finance-records/summary` exposes the public summary.
```

---

### Flow 5: Ambulance Request & Dispatch (`ambulance*`)

```text
1. Request Submission
   Orang Tua Asuh POSTs `/ambulances/requests` (submitter, patient, pickup date/time, destination, service category…)
   → Saved with status `pending`; the requester may cancel while `pending` (`PATCH /:id/cancel`)

2. Manager Review & Assignment
   Ambulance Manager reviews `/admin/ambulances/requests`
   → `PATCH /:id/accept` with `ambulanceId` (must exist; request must be `pending`) → status `accepted`, `ambulance_id` set, email sent asynchronously
   → or `PATCH /:id/reject` with a required `rejectionReason` → status `rejected`, email sent

3. Driver Execution (`/admin/ambulances/requests/assigned`)
   The driver is the account linked to the assigned ambulance (`ambulances.driver_id`)
   → `PATCH /:id/start`: request must be `accepted`; the ambulance must have no other `in_service` request;
     sets status `in_service`, `picked_up_at`, ambulance status `in use`
   → `PATCH /:id/complete`: request must be `in_service`; sets `done`, `completed_at`, ambulance status back to `available`,
     and auto-creates an `ambulance_histories` row (category copied from the request, note auto-generated)
   → `PATCH /:id/cancel` (driver, with `cancelationReason`): allowed from `accepted`/`in_service`; sets `cancelled` and frees the ambulance

4. Manual History
   Managers and drivers can also create/update/delete `ambulance_histories` entries directly and view summaries / monthly trends.
```

---

### Flow 6: Foster Child Candidate Review (`foster_children_candidate`, `foster_children`)

```text
1. Orang Tua Asuh submits a candidate (`POST /foster-children/candidates`, with document uploads) → status `pending`
   (can be cancelled via `DELETE /:id`)
2. Social Manager `PATCH /:id/accept` (only from `pending`) → `social_manager_accepted`
3. Chairman `PATCH /:id/accept` (only from `social_manager_accepted`) → `accepted`
   → a `foster_childrens` row is created from the candidate and the submitter is emailed; audit log written
4. Either reviewer can `PATCH /:id/reject` with a reason → `rejected`
```

---

### Flow 7: Schema Migration & Seeding Flow (Developer / CI Local Flow)

```text
1. Developer modifies or adds a GORM entity struct in `app/<domain>/<domain>.go`
   → Ensures struct is registered in `models/postgre.go`
   → Updates corresponding seed struct in `cmd/seed/models/` if required fields changed

2. Developer generates versioned SQL migration via Atlas
   → Runs `atlas migrate diff <name> --env gorm` (loads schema via `cmd/migrate/atlas/main.go`)
   → Atlas creates `migrations/<timestamp>.sql` and updates `migrations/atlas.sum`

3. Verification
   → Applies migration to local PostgreSQL (`docker compose up -d`)
   → Runs `go run ./cmd/seed/main.go` and `go vet ./... && go build ./...`

```

---

## Edge Cases & Error Handling

| Scenario                                                 | Handling Mechanism                                                                                                                                                                      |
| -------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Duplicate Midtrans Webhook (`settlement` sent twice)** | `HandleNotification` returns `200 OK` ("no change") when the incoming `transaction_status` equals the stored one, so no second finance record is created. Note: the status update, prayer publish and `finance_records` insert are **not** wrapped in a single `db.Transaction`; a failed finance insert is only logged as a warning. |
| **Spoofed Webhook Payload**                              | Each `*_transaction/service.go` `HandleNotification` verifies the Midtrans SHA-512 signature (`order_id + status_code + gross_amount + server_key`). Mismatch → `401`.                  |
| **Unknown Order ID prefix on webhook**                   | `app/payment/handler.go` returns `200 OK` and ignores the notification (known prefixes: `DON-`, `SPI-`, `FC-`).                                                                         |
| **S3 Upload Fails Mid-Request**                          | Expense services upload the proof first; if upload fails they return `500` and write nothing to the DB. If the later DB insert fails, the uploaded object is **not** cleaned up.        |
| **XSS / Unsafe Input**                                   | Services sanitize user text with `pkg.SanitizeStrict` (bluemonday-based) before persisting (e.g., donor name/email, prayer content).                                                    |
| **Brute-Force on Auth / Public Endpoints**               | `app/middleware/rate_limit.go` (Redis + `ulule/limiter/v3`) applies a global limit; sensitive auth routes add `CustomRateLimitHandler` (e.g., forget-password 5/min, resend-verification 3/min, OAuth 10/min). Excess requests get `429`. |
| **Unhandled Panic in Any Handler/Service**               | Caught by `app/middleware/recovery.go`, logged via `logrus`, returns a safe `500` JSON response.                                                                                         |
| **Unverified or Banned Account Login**                   | `Login` returns `403` ("verify your email" / "account banned"); `SwitchRole` also rejects banned accounts.                                                                              |
| **Insufficient Funds for an Expense**                    | `CreateDonationProgramExpense` rejects an amount greater than collected − total expense with `400`.                                                                                     |
| **Ambulance Not Assigned / Busy**                        | Driver start/complete/cancel return `403` if the request's ambulance isn't the driver's; start returns `400` if the ambulance already has an `in_service` request.                      |
| **Duplicate Subscription**                               | Subscribing with an existing `active` subscription returns `409`; an `inactive` one is reactivated.                                                                                     |
| **Duplicate Invoice for a Period**                       | The cron checks for an existing invoice and a unique index on (`subscription_id`, `billing_period`) prevents duplicates.                                                                |

---

## Middleware & Routing Hierarchy (`internal/app/routes.go`)

```text
Global Middleware Pipeline (in registration order):
├── CORSHandler          (Cross-Origin Resource Sharing)
├── LoggerHandler        (Request logging)
├── RecoveryHandler      (Panic recovery → 500 JSON)
└── RateLimitHandler     (Redis-backed throttling)

Outside /api:
├── GET /health
└── GET /swagger/*any

All domain routes mount on the `/api` group (`Container.RegisterHandlers`).
Per-route middleware: `AuthRequired()`, `AuthOptional()`, `RequireRoles(...)`, `CustomRateLimitHandler(n, period)`.

Route Access Tiers:
├── 1. Public Routes (No JWT)
│   ├── Auth: register, login, verify-email, resend-verification, forget/reset-password, OAuth
│   ├── Midtrans Webhook: POST /api/webhooks/midtrans/notification (signature-verified in the services)
│   ├── Read-only GETs: donation/social programs, foster children, news, galleries, prayers, foundation profile,
│   │   ambulances (+history), expenses (+CSV export), public finance summary
│   └── Guest checkout: POST /donation-programs/:slug/transactions, /foster-children/:slug/transactions (AuthOptional)
│
├── 2. Authenticated Routes (`jwt.go`)
│   ├── Any logged-in user: /me, /auth/switch-role, news comments (post/report)
│   └── Orang Tua Asuh role: own transactions, amen/report prayers, subscribe/unsubscribe & pay invoices,
│       foster child candidates, ambulance requests
│
└── 3. Admin / Staff Routes (`/admin/*`, `/superadmin/*` + `RequireRoles`)
    ├── Chairman:               approve/reject social programs; final accept of foster candidates
    ├── Social Manager:         social programs, foster children, subscriptions, offline payments, candidate review
    ├── Finance (Bendahara):    donation programs, expenses, offline donations, finance summaries
    ├── Ambulance Manager:      ambulances, request accept/reject, histories
    ├── Ambulance Driver:       assigned requests (start/complete/cancel), own histories
    ├── Publication Manager:    news, galleries, reported comments/prayers moderation
    └── Superadmin:             accounts (ban, role assignment), foundation profile, audit logs, backups
```

---

**Last Updated:** 2026-10-18
**Owner:** Vilamuzz  
**Status:** Active Development
