-- Modify "prayers" table
ALTER TABLE "prayers" ADD COLUMN "reported" boolean NULL;
-- Create index "idx_prayers_reported" to table: "prayers"
CREATE INDEX "idx_prayers_reported" ON "prayers" ("reported");
