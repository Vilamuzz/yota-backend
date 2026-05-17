-- Modify "social_program_subscriptions" table
ALTER TABLE "social_program_subscriptions" ADD COLUMN "total_paid_periods" bigint NOT NULL DEFAULT 0;
