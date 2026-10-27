# 05_IMPLEMENTATION_PLAN — Database Schema Remediation

> This document turns the schema audit of `docs/04_DATABASE_SCHEMA.md` into an executable, phased implementation plan for `vilamuzz-yota-backend`.
> Complements `AGENTS.md`, `docs/02_TRD.md`, and `docs/04_DATABASE_SCHEMA.md`.
> **Rule:** all schema changes are generated through Atlas from GORM models. Never edit committed `migrations/*.sql` or `migrations/atlas.sum` by hand.

---

## 1. Objectives

1. Close financial consistency gaps (webhook idempotency, atomic ledger writes, duplicate subscriptions).
2. Enforce data integrity at the database layer (unique keys, foreign keys, CHECK constraints, correct soft delete).
3. Remove performance traps (text UUIDs, missing FK/slug indexes, redundant or unusable indexes, `EXTRACT(YEAR)` scans).
4. Remove dead schema (computed columns never written) and ambiguous fields.
5. Keep the public API contract unchanged unless a decision gate below is explicitly approved.

## 2. Baseline and Status Legend

The audit identified the following findings. Status reflects the current working tree.

| ID    | Finding                                                                    | Severity | Status                 |
| :---- | :------------------------------------------------------------------------- | :------- | :--------------------- |
| DB-01 | `finance_records` lacked DB-level idempotency and webhooks were non-atomic | Critical | Implemented, migration applied (uncommitted) |
| DB-02 | `prayers.deleted_at` is `NOT NULL time.Time`, so prayer deletion is a hard delete | Critical | Implemented, migration applied (uncommitted) |
| DB-03 | Missing unique keys: subscription `(account_id, social_program_id)`, all slugs, two `transaction_id` columns | Critical | Planned                |
| DB-04 | Missing foreign keys on expenses, ambulance histories, auth tokens, amen/report accounts | High     | Planned                |
| DB-05 | Computed columns are real but never written (always `NULL`)                 | Medium   | Planned                |
| DB-06 | Finance summary counts soft-deleted programs/children                       | High     | Planned                |
| DB-07 | UUID PK/FK columns stored as `text`                                         | High     | Decision gate          |
| DB-08 | Missing indexes on FK/filter/order columns (subscriptions, ambulance, candidates, tokens, finance) | High | Planned |
| DB-09 | Slugs are unindexed or non-unique; public detail lookups scan                | High     | Planned                |
| DB-10 | Money stored as Go `float64` → unconstrained `numeric`                      | High     | Decision gate          |
| DB-11 | No CHECK/enum constraints for statuses, categories, amounts, date ranges     | Medium   | Planned                |
| DB-12 | `billing_period` is `timestamptz` but queried with `DATE()`                  | Medium   | Planned                |
| DB-13 | Redundant indexes and wrong composite order                                 | Medium   | Planned                |
| DB-14 | `EXTRACT(YEAR FROM ...)` prevents index usage in reports                    | Medium   | Planned                |
| DB-15 | Three near-identical transaction tables have drifted                         | Medium   | Decision gate          |
| DB-16 | `foundation_profiles` singleton is not enforced                             | Low      | Planned                |
| DB-17 | `pickup_date` + `pickup_time` and `requested_at` duplicate concepts         | Low      | Planned                |
| DB-18 | Auth token/log tables grow without retention/index support                  | Low      | Planned                |
| DB-19 | `04_DATABASE_SCHEMA.md` misstates GORM soft-delete and computed-column behavior | Low  | Planned                |

### Already implemented (Phase 0 — migration applied 2026-10-25)

Changes exist in the working tree and are documented in `docs/CHANGELOG.md`:

- `finance_records` unique index `uni_finance_records_source (source_type, source_id)` via `migrations/20261025191259.sql`.
- `finance_record.CreateIdempotent` with `ON CONFLICT DO NOTHING`.
- Donation, foster children, and social program notifications use row-locked `ApplyNotification` transactions.
- Offline transaction creation writes the transaction and ledger atomically.
- Social program settlement increments `total_paid_periods` only once.
- Ledger failures now abort the webhook instead of being logged and ignored.

**Phase 0 progress (2026-10-25)**

- Duplicate pre-flight passed: no conflicting `(source_type, source_id)` rows.
- `20261025191259.sql` applied without errors; `atlas.sum` verified.
- `go build ./...`, `go vet ./...`, and `go test ./...` pass.

**Phase 0 exit checklist**

- [x] Run duplicate pre-flight from the migration header before applying (passed in the environment where the migration was applied).
- [x] Apply `20261025191259.sql` and confirm `atlas.sum` checks succeed.
- [x] Run `go vet ./... && go build ./...` (also `go test ./...`).
- [x] Replay duplicate Midtrans notifications and confirm no duplicate ledger rows or period increments.
- [x] Commit Phase 0 separately before starting Phase 1.

---

## 3. Phased Plan

Each phase produces one or more Atlas migrations and must remain independently reviewable and deployable. Forward-only migrations are assumed; rollbacks are performed with corrective forward migrations.

### Phase 1 — Integrity and Correctness (P0)

#### DB-02: Fix prayer soft delete

Files: `app/prayer/prayer.go`, `app/prayer/repository.go`, `cmd/seed/models/prayers.go`, `migrations/*` (new).

Implemented decisions:

- `Prayer.DeletedAt` uses `gorm.DeletedAt` with `index`. GORM automatically adds
  the `deleted_at IS NULL` filter to `FindOnePrayer`, `FindAllPrayers`, and
  `CountPrayers`, so no manual read filters are required.
- `DeletePrayer` soft-deletes only the prayer row. `prayer_amens` and
  `prayer_reports` are preserved rather than hard-deleted.
- Existing zero timestamps are backfilled to `NULL` in the migration, after
  dropping `NOT NULL`. Real deletion timestamps, if any, are preserved:
  `UPDATE prayers SET deleted_at = NULL WHERE deleted_at < '1970-01-01';`
- `cmd/seed/models/prayers.go` uses `Unscoped()` for its existence check so a
  soft-deleted seed prayer is not re-inserted.

Acceptance: deleted prayers disappear from all reads, rows remain in the table, and `deleted_at` is `NULL` for active rows.

**DB-02 progress (2026-10-27)**

- Migration `20261027005657_prayer_soft_delete.sql` applied to the local dev
  database.
- `atlas migrate validate`, `go vet ./...`, `go build ./...`, and
  `go test ./...` pass.
- A rolled-back runtime check confirmed the row survives deletion, `deleted_at`
  becomes non-null, `FindOnePrayer` returns `record not found`, and
  `CountPrayers` excludes the deleted row.

#### DB-03: Add missing unique constraints

Pre-flight queries must return zero rows before applying:

```sql
SELECT account_id, social_program_id, count(*) FROM social_program_subscriptions
GROUP BY 1, 2 HAVING count(*) > 1;

SELECT slug, count(*) FROM donation_programs GROUP BY 1 HAVING count(*) > 1;
SELECT slug, count(*) FROM news GROUP BY 1 HAVING count(*) > 1;
SELECT slug, count(*) FROM galleries GROUP BY 1 HAVING count(*) > 1;
SELECT slug, count(*) FROM social_programs GROUP BY 1 HAVING count(*) > 1;
SELECT slug, count(*) FROM foster_childrens GROUP BY 1 HAVING count(*) > 1;

SELECT transaction_id, count(*) FROM donation_program_transactions
WHERE transaction_id IS NOT NULL GROUP BY 1 HAVING count(*) > 1;
SELECT transaction_id, count(*) FROM foster_children_transactions
WHERE transaction_id IS NOT NULL GROUP BY 1 HAVING count(*) > 1;
```

Changes:

- Replace `idx_slug` on `donation_programs` with a unique index.
- Add unique indexes for `news.slug`, `galleries.slug`, `social_programs.slug`, `foster_childrens.slug`.
- Add `uniqueIndex` tags for subscription `(account_id, social_program_id)`.
- Mark `donation_program_transactions.transaction_id` and `foster_children_transactions.transaction_id` unique (nullable unique is valid in PostgreSQL; duplicate `NULL`s remain allowed).

Acceptance: duplicate inserts fail at the DB layer, and concurrent subscription creation returns one active subscription.

#### DB-04: Add missing foreign keys and delete policies

Add FKs for at least:

- `donation_program_expenses.donation_program_id → donation_programs`
- `social_program_expenses.social_program_id → social_programs`
- `foster_children_expenses.foster_children_id → foster_childrens`
- `ambulance_histories.ambulance_id → ambulances`
- `email_verification_tokens.account_id`, `password_reset_tokens.account_id → accounts`
- `prayer_amens.account_id`, `prayer_reports.account_id → accounts`
- `news_comment_reports.account_id → accounts`

Policy decisions:

- Parent-owned children (`media`, `prayer_amens`, `prayer_reports`, `news_comment_reports`, `achivements`) should use `ON DELETE CASCADE` where hard deletion is supported.
- Financial history (`*_expenses`, `*_transactions`, `finance_records`) should use `ON DELETE RESTRICT`; these tables must not lose audit rows.
- Soft-deleted parents need no cascade; their children are filtered by parent visibility in queries.

Pre-flight orphan checks are required for every new FK before applying. Add indexes for each FK column in the same phase (see DB-08).

Acceptance: inserting a child with a non-existent parent fails; deleting a parent with audit children is blocked.

#### DB-06: Fix finance summary filters

Files: `app/finance_record/repository.go`.

- Add `deleted_at IS NULL` to the donation program, social program, and foster children counts in `Summary` (`app/finance_record/repository.go:97-109`).
- Keep the existing behavior of excluding draft/archived and pending/rejected records.

Acceptance: soft-deleted programs and children no longer appear in summary totals.

#### DB-11: Add CHECK constraints and NOT NULL defaults

Add CHECK constraints for:

- `social_programs.billing_day BETWEEN 1 AND 31`
- All amount columns `>= 0` (`fund_target`, `minimum_amount`, `gross_amount`, `amount`)
- `donation_programs.start_date <= end_date` when both are set
- `social_program_invoices.due_date >= billing_period`
- `ambulance_service_requests.patient_age >= 0` when set
- `media`: exactly one of `news_id` / `gallery_id` is non-null
- Enum-like text columns: use CHECK constraints or PostgreSQL enums for statuses, categories, genders, fund types, source types, and service categories

Also make boolean/status fields `NOT NULL` with explicit defaults where the domain has no third state. Review every nullable boolean (`is_online`, `reported`, `is_infectious`, etc.) and either document the tri-state meaning or make it non-null.

Change `social_programs.status` default from `'active'` to `'pending'` to match the approval workflow.

Acceptance: invalid enum values, negative amounts, and malformed media rows are rejected by PostgreSQL.

#### DB-16/DB-17: Singleton and redundant fields

- Enforce the `foundation_profiles` singleton with a fixed primary key or a `CHECK`/unique constraint on a constant column.
- Replace `ambulance_service_requests.pickup_date` + `pickup_time` with one `pickup_at timestamptz` (or `pickup_date date` + `pickup_time time` if the product needs them separately).
- Remove `requested_at` if it is identical to `created_at`, or keep it and document that `created_at` is the row-insert time while `requested_at` is the business request time.

These changes affect API responses; confirm with the frontend before renaming fields. Keep JSON compatibility aliases if required.

---

### Phase 2 — Performance (P1)

#### DB-09: Slug lookup indexes

Already covered by unique indexes in DB-03. Verify with `EXPLAIN` that public detail endpoints use index scans.

#### DB-08: Add missing indexes

Proposed indexes (final names follow Atlas conventions):

| Table                          | Index                                                              | Reason |
| :----------------------------- | :----------------------------------------------------------------- | :----- |
| `social_program_subscriptions` | `(account_id, status)`, `(social_program_id, status)`, `(created_at)` | Listing, duplicate checks, billing sweep |
| `social_program_transactions`  | `(account_id, transaction_status)`                                 | Account history and reporting |
| `ambulance_service_requests`   | `(submitted_by, created_at)`, `(ambulance_id, created_at)`, `(status, created_at)`, `(service_category, created_at)` | Queue/list filters |
| `ambulance_histories`          | `(ambulance_id, created_at)`, `(driver_id, created_at)`, `(created_at)` | History and monthly reports |
| `foster_children_candidates`   | `(status, created_at)`, `(submitted_by, created_at)`               | Admin queues and submitter history |
| `news_comments`                | `(parent_comment_id)`                                              | Threaded replies |
| `account_roles`                | `(role_id)`                                                        | Reverse role lookups |
| `finance_records`              | `(fund_type, source_type, transaction_date)`, `(fund_id, transaction_date)` | Summary, trend, cursor lists |
| `email_verification_tokens`    | `(account_id)`, `(expired_at)`, `(is_used, expired_at)`            | Resend/revoke and cleanup |
| `password_reset_tokens`        | `(account_id)`, `(expired_at)`, `(is_used, expired_at)`            | Resend/revoke and cleanup |
| `logs`                         | `(entity_type, entity_id)`, `(created_at)`                         | Audit lookups and retention |
| `backups`                      | `(created_at)`                                                     | Retention sweep |
| `donation_program_transactions`| `(donation_program_id, transaction_status, created_at)`            | Replace the current composite whose middle `account_id` blocks status filters |
| `donation_programs`            | `(status, deleted_at, created_at)` instead of standalone `idx_status` | Public list and scheduling |

Use partial indexes where queries always filter active rows, for example:

```sql
CREATE INDEX idx_donation_program_expenses_active
ON donation_program_expenses (donation_program_id, expense_date)
WHERE deleted_at IS NULL;
```

Apply the same pattern to `social_program_expenses` and `foster_children_expenses`.

#### DB-13: Remove redundant and unusable indexes

- Drop `idx_social_program_transactions_social_program_invoice_id`; the unique constraint already provides an index.
- Review `idx_prayers_donation_program_transaction_id` × unique constraint duplication.
- Replace `idx_expenses_composite` with the partial index above, or reorder it to `(donation_program_id, deleted_at, expense_date)`.
- Keep `idx_news_published_at` only if queries do not always filter `status`; otherwise replace with `(status, published_at)`.

#### DB-12/DB-14: Query-friendly columns and predicates

- Change `social_program_invoices.billing_period` to `date` and update the model/API mapping. Replace `DATE(billing_period) = ?` with `billing_period = ?`.
- Replace `EXTRACT(YEAR FROM x) = ?` with range predicates (`x >= start AND x < nextYearStart`) in:
  - `app/finance_record/repository.go:160`
  - `app/donation_program_transaction/repository.go:165`
  - `app/ambulance_history/repository.go:192`, `:292`
- Recheck the monthly trend `GROUP BY` query after the index changes.

#### DB-05: Remove or materialize computed columns

Preferred: delete the physical columns and keep read-only API fields populated by query aliases.

- Tag computed fields with `gorm:"->;-:migration"` (verify Atlas emits `DROP COLUMN`).
- Keep subquery aliases in repositories so JSON output is unchanged.
- Affected fields: `donation_programs.collected_fund/total_expense`, `social_programs.total_subscribers/is_subscribed/subscription_id/collected_fund/total_expense`, `social_program_subscriptions.total_donation`, `social_program_invoices.snap_token`, `foster_childrens.collected_fund/total_expense`, `prayers.is_amen`.
- If materialization is preferred instead, use generated columns or triggers and document refresh semantics. Do not leave them as always-`NULL` columns.

Acceptance: `\d+` shows no dead computed columns, and API responses retain the same fields and values.

---

### Phase 3 — Decision-Gated and Longer-Term Work

These require explicit owner approval because they change storage types, public API shapes, or module boundaries.

#### DB-07: Native UUID migration

- Add `gorm:"type:uuid"` to all `uuid.UUID` PK/FK fields.
- Generate a migration that alters `text → uuid` with `USING id::uuid` in dependency order (parents before children).
- Expect table rewrites and locks on large tables; schedule a maintenance window or use a table-rebuild strategy.
- Switch new PK generation to `uuid.NewV7()` for time-ordered inserts where request semantics allow it.
- Update `04_DATABASE_SCHEMA.md` after the change.

Decision required because it touches every table and all FK columns.

#### DB-10: Money type

- Target `bigint` (IDR minor units are not used, so whole rupiah) or `numeric(18,2)` with `CHECK (amount >= 0)`.
- `float64` cannot guarantee exact accounting arithmetic. If the API must stay numeric, expose numbers but store exact values.
- Add an explicit conversion migration and update all Go arithmetic/serialization.

Decision required because it changes API field types and aggregation code.

#### DB-15: Consolidate transaction tables

The three transaction tables duplicate Midtrans fields and have drifted:

- `donation_program_transactions`
- `foster_children_transactions`
- `social_program_transactions`

Proposed: one `payment_transactions` table with `payable_type`/`payable_id`, common Midtrans fields, and typed owner tables. This reduces webhook branching and keeps idempotency consistent.

Decision required because it is a domain-model refactor, not a migration.

#### DB-18: Retention and operational hygiene

- Add a scheduler job to purge expired/used auth tokens and old logs/backups.
- Consider `jsonb` for `logs.old_value`/`new_value` if structured diffs are needed.
- Add an index supporting the retention `DELETE` predicates.
- Document retention periods.

#### DB-19: Documentation corrections

Update `docs/04_DATABASE_SCHEMA.md` after the schema work:

- Correct the soft-delete note: GORM auto-filters only with `gorm.DeletedAt`; `*time.Time` is filtered manually.
- Correct the computed-column note: stored columns are unused; values come from query aliases.
- Document all new constraints, indexes, and FK delete policies.
- Update the ER diagram if delete policies or transaction consolidation change.

---

## 4. Migration Workflow

Execute one logical group per Atlas diff so each migration is small and reviewable.

```bash
# 1. Update GORM structs/tags first
# 2. Generate the migration
atlas migrate diff prayer_soft_delete --env gorm

# 3. Inspect the generated SQL; confirm no destructive or locking surprise
# 4. Re-hash only when Atlas requires it after generation
atlas migrate hash

# 5. Verify application compilation
go vet ./... && go build ./...

# 6. Apply to a dev/staging database
atlas migrate apply --env gorm
```

Required order:

1. Phase 0 applied (`20261025191259`) — commit before starting Phase 1.
2. `prayer_soft_delete` — nullable `deleted_at`, normalize zero timestamps.
3. `unique_constraints` — after duplicate pre-flight cleanup.
4. `foreign_keys` — after orphan pre-flight cleanup and index creation.
5. `indexes` — additive, low risk.
6. `dead_computed_columns` — destructive, after confirming API parity.
7. `check_constraints_and_defaults` — may fail on legacy invalid rows; clean data first.
8. `billing_period_date` and query rewrites.
9. Decision-gated migrations (`uuid`, `money`, transaction consolidation) as separate projects.

Do not combine steps 2–7 into one migration; a failure in one constraint would roll back unrelated work and make staging diagnosis harder.

---

## 5. Pre-Flight Data Cleanup

Run these before applying integrity migrations. Save results as evidence in the PR.

```sql
-- Duplicate subscriptions
SELECT account_id, social_program_id, count(*) FROM social_program_subscriptions
GROUP BY 1, 2 HAVING count(*) > 1;

-- Duplicate slugs (run per table)
SELECT slug, count(*) FROM news GROUP BY 1 HAVING count(*) > 1;
SELECT slug, count(*) FROM galleries GROUP BY 1 HAVING count(*) > 1;
SELECT slug, count(*) FROM social_programs GROUP BY 1 HAVING count(*) > 1;
SELECT slug, count(*) FROM foster_childrens GROUP BY 1 HAVING count(*) > 1;
SELECT slug, count(*) FROM donation_programs GROUP BY 1 HAVING count(*) > 1;

-- Duplicate Midtrans transaction IDs
SELECT transaction_id, count(*) FROM donation_program_transactions
WHERE transaction_id IS NOT NULL GROUP BY 1 HAVING count(*) > 1;
SELECT transaction_id, count(*) FROM foster_children_transactions
WHERE transaction_id IS NOT NULL GROUP BY 1 HAVING count(*) > 1;

-- Orphans for new FKs
SELECT e.id FROM donation_program_expenses e
LEFT JOIN donation_programs p ON p.id = e.donation_program_id WHERE p.id IS NULL;
SELECT e.id FROM social_program_expenses e
LEFT JOIN social_programs p ON p.id = e.social_program_id WHERE p.id IS NULL;
SELECT e.id FROM foster_children_expenses e
LEFT JOIN foster_childrens c ON c.id = e.foster_children_id WHERE c.id IS NULL;
SELECT h.id FROM ambulance_histories h
LEFT JOIN ambulances a ON a.id = h.ambulance_id WHERE a.id IS NULL;

-- Invalid enum/amount values
SELECT id, status FROM social_programs WHERE billing_day < 1 OR billing_day > 31;
SELECT id FROM donation_programs WHERE fund_target < 0;
```

Cleanup requires its own reviewed script or migration. Do not delete financial rows; soft-delete or reconcile them with the owner.

---

## 6. Verification Matrix

| Check | Command / Method | Required |
| :---- | :--------------- | :------- |
| Compilation and vet | `go vet ./... && go build ./...` | Every phase |
| Migration validity | `atlas migrate validate --env gorm` | Every phase |
| Migration checksum | `atlas migrate hash` output matches `atlas.sum` | Every phase |
| Seeder still runs | `go run ./cmd/seed/main.go` against dev DB | Schema-touching phases |
| Duplicate webhook replay | Send the same Midtrans payload twice; assert one ledger row and one period increment | Phase 0/1 |
| Constraint behavior | Insert duplicate slug/subscription and expect a DB error | Phase 1 |
| FK behavior | Insert child with missing parent and expect a DB error | Phase 1 |
| Soft delete | Delete a prayer; assert it is hidden but still present with `deleted_at` | Phase 1 |
| Query plans | `EXPLAIN (ANALYZE, BUFFERS)` on public detail and admin list queries | Phase 2 |
| API contract | Compare Swagger responses before/after for computed fields | Phase 2 |
| Retention | Dry-run cleanup counts for tokens/logs | Phase 3 |

---

## 7. Risks and Mitigations

| Risk | Mitigation |
| :--- | :--------- |
| Duplicate existing data blocks a unique index | Run pre-flight queries, reconcile duplicates, then apply. |
| New FK blocks on orphan rows | Clean orphans with owner-approved scripts before applying. |
| Table rewrite/lock during `text → uuid` conversion | Separate decision-gated migration; schedule maintenance; test timing on a production-sized copy. |
| Dropping computed columns breaks direct SQL consumers | Confirm only repositories reference them; keep API aliases and update dashboards first. |
| CHECK constraints reject historical invalid rows | Audit invalid rows first; add constraints as `NOT VALID`, clean data, then `VALIDATE CONSTRAINT`. |
| Breaking API field types during money/UUID changes | Treat as versioned API changes; require owner approval and frontend coordination. |
| Atlas checksum mismatch | Never edit committed SQL; regenerate and inspect a new migration. |
| Soft-delete behavior changes hide rows unexpectedly | Standardize on `gorm.DeletedAt`, add explicit filters, and add regression checks for each repository. |

---

## 8. Definition of Done

A phase is done only when all of the following hold:

1. GORM tags/models and repository queries are updated.
2. A reviewed Atlas migration exists; no historical migration was edited.
3. Pre-flight cleanup evidence is attached for integrity changes.
4. `go vet ./...`, `go build ./...`, and `atlas migrate validate --env gorm` pass.
5. The seeder runs, or the PR explains why no seed change is required.
6. The verification matrix items for that phase pass on staging.
7. `docs/04_DATABASE_SCHEMA.md` and `docs/CHANGELOG.md` are updated.
8. No secrets, generated files, or unrelated diffs are included.

---

## 9. Out of Scope

- Rewriting historical migrations.
- Changing `app/middleware/jwt.go`, `rate_limit.go`, `Dockerfile`, `docker-compose.yml`, `atlas.hcl`, or CI workflows without separate approval.
- Introducing new third-party dependencies; the plan uses GORM, Atlas, and existing packages only.
- Business-rule changes to donation, sponsorship, subscription, or ambulance state machines beyond the schema constraints listed here.

---

**Last Updated:** 2026-10-27
**Owner:** Vilamuzz
**Status:** Active
