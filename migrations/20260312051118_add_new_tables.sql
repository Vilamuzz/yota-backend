-- Create "donation_expenses" table
CREATE TABLE "donation_expenses" (
  "id" text NOT NULL,
  "donation_id" text NULL,
  "title" text NULL,
  "amount" numeric NULL,
  "date" timestamptz NULL,
  "note" text NULL,
  "proof_file" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "finance_records" table
CREATE TABLE "finance_records" (
  "id" text NOT NULL,
  "fund_type" text NULL,
  "fund_id" text NULL,
  "source_type" text NULL,
  "source_id" text NULL,
  "amount" numeric NULL,
  "transaction_date" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "prayers" table
CREATE TABLE "prayers" (
  "id" text NOT NULL,
  "donation_id" text NOT NULL,
  "user_id" text NOT NULL,
  "content" text NOT NULL,
  "like_count" bigint NOT NULL,
  "is_reported" boolean NOT NULL,
  "deleted_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
