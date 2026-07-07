-- Modify "donation_program_transactions" table
ALTER TABLE "donation_program_transactions" ADD COLUMN "fee" numeric NULL DEFAULT 0, ADD COLUMN "net_amount" numeric NULL DEFAULT 0;
-- Create "payment_methods" table
CREATE TABLE "payment_methods" (
  "id" bigserial NOT NULL,
  "code" text NOT NULL,
  "name" text NOT NULL,
  "fee_type" text NOT NULL,
  "fee_value" numeric NOT NULL,
  "is_active" boolean NOT NULL,
  PRIMARY KEY ("id")
);
