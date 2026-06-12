-- Modify "donation_program_expenses" table
ALTER TABLE "donation_program_expenses" ADD COLUMN "deleted_at" timestamptz NULL;
-- Create index "idx_donation_program_expenses_deleted_at" to table: "donation_program_expenses"
CREATE INDEX "idx_donation_program_expenses_deleted_at" ON "donation_program_expenses" ("deleted_at");
-- Modify "finance_records" table
ALTER TABLE "finance_records" ADD COLUMN "deleted_at" timestamptz NULL;
-- Create index "idx_finance_records_deleted_at" to table: "finance_records"
CREATE INDEX "idx_finance_records_deleted_at" ON "finance_records" ("deleted_at");
-- Modify "galleries" table
ALTER TABLE "galleries" ALTER COLUMN "category" DROP NOT NULL, ALTER COLUMN "cover_image" DROP NOT NULL, ALTER COLUMN "description" DROP NOT NULL, DROP COLUMN "published_at";
-- Create "ambulance_service_requests" table
CREATE TABLE "ambulance_service_requests" (
  "id" text NOT NULL,
  "account_id" text NOT NULL,
  "applicant_name" text NOT NULL,
  "applicant_phone" text NOT NULL,
  "applicant_address" text NOT NULL,
  "description" text NOT NULL,
  "request_date" timestamptz NOT NULL,
  "request_reason" text NOT NULL,
  "status" text NOT NULL,
  "rejection_reason" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Drop "ambulance_requests" table
DROP TABLE "ambulance_requests";
