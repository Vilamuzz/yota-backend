-- Modify "foster_children_transactions" table
ALTER TABLE "foster_children_transactions" ADD COLUMN "fee" numeric NULL DEFAULT 0, ADD COLUMN "net_amount" numeric NULL DEFAULT 0;
-- Modify "social_program_transactions" table
ALTER TABLE "social_program_transactions" ADD COLUMN "fee" numeric NULL DEFAULT 0, ADD COLUMN "net_amount" numeric NULL DEFAULT 0;
