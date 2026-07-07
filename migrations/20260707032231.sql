-- Modify "donation_program_transactions" table
ALTER TABLE "donation_program_transactions" ADD COLUMN "ppn_percentage" numeric NULL DEFAULT 0, ADD COLUMN "ppn_amount" numeric NULL DEFAULT 0;
-- Modify "foster_children_transactions" table
ALTER TABLE "foster_children_transactions" ADD COLUMN "ppn_percentage" numeric NULL DEFAULT 0, ADD COLUMN "ppn_amount" numeric NULL DEFAULT 0;
-- Modify "foundation_profiles" table
ALTER TABLE "foundation_profiles" ADD COLUMN "ppn_percentage" numeric NULL DEFAULT 11;
-- Modify "social_program_transactions" table
ALTER TABLE "social_program_transactions" ADD COLUMN "ppn_percentage" numeric NULL DEFAULT 0, ADD COLUMN "ppn_amount" numeric NULL DEFAULT 0;
