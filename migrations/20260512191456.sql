-- Modify "prayers" table
ALTER TABLE "prayers" ADD COLUMN "amen_count" bigint NULL DEFAULT 0, ADD COLUMN "report_count" bigint NULL DEFAULT 0;
-- Modify "social_programs" table
ALTER TABLE "social_programs" ADD COLUMN "subscription_id" text NULL;
