-- Modify "donation_transactions" table
ALTER TABLE "donation_transactions" DROP COLUMN "prayer_content";
-- Modify "finance_records" table
ALTER TABLE "finance_records" DROP COLUMN "updated_at";
-- Modify "social_programs" table
ALTER TABLE "social_programs" DROP COLUMN "image", ADD COLUMN "image_url" text NOT NULL;
-- Modify "prayers" table
ALTER TABLE "prayers" ADD COLUMN "donation_transaction_id" text NOT NULL, ADD COLUMN "status" boolean NULL DEFAULT false, ADD CONSTRAINT "fk_donation_transactions_prayer" FOREIGN KEY ("donation_transaction_id") REFERENCES "donation_transactions" ("id") ON UPDATE CASCADE ON DELETE CASCADE;
-- Modify "prayer_amens" table
ALTER TABLE "prayer_amens" ADD CONSTRAINT "fk_prayers_prayer_ames" FOREIGN KEY ("prayer_id") REFERENCES "prayers" ("id") ON UPDATE CASCADE ON DELETE CASCADE;
-- Modify "prayer_reports" table
ALTER TABLE "prayer_reports" ADD CONSTRAINT "fk_prayers_prayer_reports" FOREIGN KEY ("prayer_id") REFERENCES "prayers" ("id") ON UPDATE CASCADE ON DELETE CASCADE;
