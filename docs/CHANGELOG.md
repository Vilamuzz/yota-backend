# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Fixed

- Enforce DB-level idempotency for finance ledger entries with a unique index on
  `finance_records (source_type, source_id)` and `ON CONFLICT DO NOTHING` inserts.
- Make Midtrans webhook processing atomic for donation, foster children, and
  social program transactions: transaction status, related side effects
  (invoice, subscription, prayer) and ledger insert now run in a single DB
  transaction guarded by a row lock.
- Propagate ledger insert failures instead of logging and continuing, so the
  webhook returns an error and Midtrans retries.
- Make offline transaction creation and expense creation atomic with their
  corresponding finance records. Social offline payments previously wrote the
  ledger outside the transaction because they used a separate DB connection.
- Prevent duplicate `total_paid_periods` increments when the same social program
  settlement notification is delivered more than once.
- Make prayer deletion a soft delete: `prayers.deleted_at` is now nullable
  (`gorm.DeletedAt`) and GORM filters deleted prayers from all repository reads
  automatically. Amen and report rows are preserved when a prayer is deleted.

### Added

- Migration `20261025191259.sql` adding the `uni_finance_records_source` unique
  index. The duplicate pre-flight passed before it was applied.
- Migration `20261027005657_prayer_soft_delete.sql` making
  `prayers.deleted_at` nullable and backfilling legacy zero timestamps to
  `NULL`.

### Migration Notes

- The duplicate pre-flight for `(source_type, source_id)` passed and
  `20261025191259.sql` was applied without errors on 2026-10-25.
- Keep the pre-flight query in the deploy runbook for other environments; if
  duplicates are present, clean them up before applying the unique index.
- `20261027005657_prayer_soft_delete.sql` must run before deploying the new
  `gorm.DeletedAt` code. Old binaries write the zero timestamp for new prayers
  during a rolling deploy; re-run the backfill
  (`deleted_at < '1970-01-01'`) after the rollout to catch any rows created in
  the gap.
