-- Rename a column from "amount" to "minimum_amount"
ALTER TABLE "social_program_invoices" RENAME COLUMN "amount" TO "minimum_amount";
-- Modify "donation_program_transactions" table
ALTER TABLE "donation_program_transactions" ADD CONSTRAINT "fk_donation_program_transactions_donation_program" FOREIGN KEY ("donation_program_id") REFERENCES "donation_programs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "foster_children_transactions" table
ALTER TABLE "foster_children_transactions" ADD CONSTRAINT "fk_foster_children_transactions_foster_children" FOREIGN KEY ("foster_children_id") REFERENCES "foster_childrens" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
