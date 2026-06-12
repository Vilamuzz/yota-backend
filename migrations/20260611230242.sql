-- Drop index "idx_donation_program_transactions_account_id" from table: "donation_program_transactions"
DROP INDEX "idx_donation_program_transactions_account_id";
-- Drop index "idx_program_status" from table: "donation_program_transactions"
DROP INDEX "idx_program_status";
-- Create index "idx_transaction_composite" to table: "donation_program_transactions"
CREATE INDEX "idx_transaction_composite" ON "donation_program_transactions" ("donation_program_id", "account_id", "transaction_status");
