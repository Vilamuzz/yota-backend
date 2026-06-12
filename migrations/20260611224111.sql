-- Drop index "idx_program_deleted" from table: "donation_program_expenses"
DROP INDEX "idx_program_deleted";
-- Create index "idx_expenses_composite" to table: "donation_program_expenses"
CREATE INDEX "idx_expenses_composite" ON "donation_program_expenses" ("donation_program_id", "expense_date", "created_at", "deleted_at");
