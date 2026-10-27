-- Modify "prayers" table
ALTER TABLE "prayers" ALTER COLUMN "deleted_at" DROP NOT NULL;

-- Backfill legacy zero timestamps written by the previous non-nullable time.Time field.
-- Rows with a real deletion timestamp (if any) are preserved.
UPDATE "prayers"
SET "deleted_at" = NULL
WHERE "deleted_at" IS NOT NULL
  AND "deleted_at" < '1970-01-01 00:00:00+00'::timestamptz;
