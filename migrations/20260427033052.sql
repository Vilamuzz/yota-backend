-- Drop index "idx_prayers_donation_transaction_id" from table: "prayers"
DROP INDEX "idx_prayers_donation_transaction_id";
-- Rename a column from "donation_transaction_id" to "donation_program_transaction_id"
ALTER TABLE "prayers" RENAME COLUMN "donation_transaction_id" TO "donation_program_transaction_id";
-- Create index "idx_prayers_donation_program_transaction_id" to table: "prayers"
CREATE UNIQUE INDEX "idx_prayers_donation_program_transaction_id" ON "prayers" ("donation_program_transaction_id");
-- Modify "roles" table
ALTER TABLE "roles" ALTER COLUMN "name" TYPE character varying(30);
