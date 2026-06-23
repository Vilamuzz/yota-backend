-- Modify "foster_children_expenses" table
ALTER TABLE "foster_children_expenses" ADD COLUMN "deleted_at" timestamptz NULL;
-- Modify "social_program_expenses" table
ALTER TABLE "social_program_expenses" ADD COLUMN "updated_at" timestamptz NULL, ADD COLUMN "deleted_at" timestamptz NULL;
