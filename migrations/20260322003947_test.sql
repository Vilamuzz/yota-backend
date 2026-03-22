-- Modify "prayers" table
ALTER TABLE "prayers" ALTER COLUMN "like_count" DROP NOT NULL, ALTER COLUMN "like_count" SET DEFAULT 0, DROP COLUMN "is_reported", DROP COLUMN "deleted_at", ADD COLUMN "report_count" bigint NULL DEFAULT 0;
