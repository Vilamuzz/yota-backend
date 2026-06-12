-- Create index "idx_slug" to table: "donation_programs"
CREATE INDEX "idx_slug" ON "donation_programs" ("slug");
-- Rename an index from "idx_donation_programs_status" to "idx_status"
ALTER INDEX "idx_donation_programs_status" RENAME TO "idx_status";
