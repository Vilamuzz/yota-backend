-- Drop index "idx_donation_program_expenses_deleted_at" from table: "donation_program_expenses"
DROP INDEX "idx_donation_program_expenses_deleted_at";
-- Drop index "idx_donation_program_expenses_donation_program_id" from table: "donation_program_expenses"
DROP INDEX "idx_donation_program_expenses_donation_program_id";
-- Create index "idx_program_deleted" to table: "donation_program_expenses"
CREATE INDEX "idx_program_deleted" ON "donation_program_expenses" ("donation_program_id", "deleted_at");
-- Drop index "idx_donation_program_transactions_donation_program_id" from table: "donation_program_transactions"
DROP INDEX "idx_donation_program_transactions_donation_program_id";
-- Drop index "idx_donation_program_transactions_transaction_status" from table: "donation_program_transactions"
DROP INDEX "idx_donation_program_transactions_transaction_status";
-- Create index "idx_program_status" to table: "donation_program_transactions"
CREATE INDEX "idx_program_status" ON "donation_program_transactions" ("donation_program_id", "transaction_status");
