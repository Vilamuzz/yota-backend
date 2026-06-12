-- Modify "donation_programs" table
ALTER TABLE "donation_programs" ALTER COLUMN "category" DROP NOT NULL, ALTER COLUMN "description" DROP NOT NULL, ALTER COLUMN "fund_target" DROP NOT NULL, ALTER COLUMN "status" SET DEFAULT 'draft', ALTER COLUMN "start_date" DROP NOT NULL, ALTER COLUMN "end_date" DROP NOT NULL, DROP COLUMN "published_at";
