-- Modify "donation_transactions" table
ALTER TABLE "donation_transactions" DROP COLUMN IF EXISTS "prayer_content";
-- Modify "finance_records" table
ALTER TABLE "finance_records" DROP COLUMN IF EXISTS "updated_at";
-- Modify "social_programs" table
ALTER TABLE "social_programs" DROP COLUMN IF EXISTS "image", ADD COLUMN IF NOT EXISTS "image_url" text NOT NULL;
-- Modify "prayers" table
ALTER TABLE "prayers" ADD COLUMN IF NOT EXISTS "donation_transaction_id" text NOT NULL, ADD COLUMN IF NOT EXISTS "status" boolean NULL DEFAULT false;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_donation_transactions_prayer') THEN
        ALTER TABLE "prayers" ADD CONSTRAINT "fk_donation_transactions_prayer" FOREIGN KEY ("donation_transaction_id") REFERENCES "donation_transactions" ("id") ON UPDATE CASCADE ON DELETE CASCADE;
    END IF;
END $$;

-- Modify "prayer_amens" table
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_prayers_prayer_ames') THEN
        ALTER TABLE "prayer_amens" ADD CONSTRAINT "fk_prayers_prayer_ames" FOREIGN KEY ("prayer_id") REFERENCES "prayers" ("id") ON UPDATE CASCADE ON DELETE CASCADE;
    END IF;
END $$;

-- Modify "prayer_reports" table
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_prayers_prayer_reports') THEN
        ALTER TABLE "prayer_reports" ADD CONSTRAINT "fk_prayers_prayer_reports" FOREIGN KEY ("prayer_id") REFERENCES "prayers" ("id") ON UPDATE CASCADE ON DELETE CASCADE;
    END IF;
END $$;
