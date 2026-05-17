-- Modify "foster_children_candidates" table
ALTER TABLE "foster_children_candidates" ADD COLUMN "school_name" text NULL, ADD COLUMN "education_level" bigint NULL;
-- Modify "foster_childrens" table
ALTER TABLE "foster_childrens" ADD COLUMN "school_name" text NULL, ADD COLUMN "education_level" bigint NULL;
-- Modify "social_program_invoices" table
ALTER TABLE "social_program_invoices" ADD CONSTRAINT "fk_social_program_invoices_subscription" FOREIGN KEY ("subscription_id") REFERENCES "social_program_subscriptions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
