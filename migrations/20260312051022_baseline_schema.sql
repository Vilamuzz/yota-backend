-- Create "category_media" table
CREATE TABLE "category_media" (
  "id" smallserial NOT NULL,
  "category" text NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "donation_transactions" table
CREATE TABLE "donation_transactions" (
  "id" text NOT NULL,
  "donation_id" text NOT NULL,
  "order_id" text NULL,
  "user_id" text NULL,
  "donor_name" text NULL,
  "donor_email" text NULL,
  "source" boolean NULL,
  "gross_amount" numeric NULL,
  "fraud_status" text NULL,
  "transaction_status" text NULL,
  "provider" text NULL,
  "transaction_id" text NULL,
  "snap_token" text NULL,
  "snap_redirect_url" text NULL,
  "prayer_content" text NULL,
  "paid_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_donation_transactions_order_id" to table: "donation_transactions"
CREATE UNIQUE INDEX "idx_donation_transactions_order_id" ON "donation_transactions" ("order_id");
-- Create "donations" table
CREATE TABLE "donations" (
  "id" text NOT NULL,
  "title" text NOT NULL,
  "slug" text NOT NULL,
  "description" text NOT NULL,
  "image_url" text NULL,
  "category" text NOT NULL,
  "fund_target" numeric NOT NULL,
  "collected_fund" numeric NULL,
  "status" character varying(20) NOT NULL DEFAULT 'active',
  "date_end" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_donations_slug" UNIQUE ("slug")
);
-- Create "email_verification_tokens" table
CREATE TABLE "email_verification_tokens" (
  "id" text NOT NULL,
  "user_id" text NOT NULL,
  "token" text NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "used" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_email_verification_tokens_token" UNIQUE ("token")
);
-- Create "media" table
CREATE TABLE "media" (
  "id" text NOT NULL,
  "entity_id" text NOT NULL,
  "entity_type" text NOT NULL,
  "type" text NOT NULL,
  "url" text NOT NULL,
  "alt_text" text NULL,
  "order" bigint NOT NULL DEFAULT 0,
  PRIMARY KEY ("id")
);
-- Create "news" table
CREATE TABLE "news" (
  "id" text NOT NULL,
  "title" text NOT NULL,
  "category" text NOT NULL,
  "content" text NOT NULL,
  "image" text NULL,
  "status" character varying(20) NOT NULL DEFAULT 'draft',
  "views" bigint NOT NULL DEFAULT 0,
  "published_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "password_reset_tokens" table
CREATE TABLE "password_reset_tokens" (
  "id" text NOT NULL,
  "user_id" text NOT NULL,
  "token" text NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "used" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_password_reset_tokens_token" UNIQUE ("token")
);
-- Create "social_programs" table
CREATE TABLE "social_programs" (
  "id" text NOT NULL,
  "title" text NOT NULL,
  "description" text NOT NULL,
  "image" text NULL,
  "status" character varying(20) NOT NULL DEFAULT 'active',
  "minimum_amount" numeric NOT NULL,
  "billing_day" bigint NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "galleries" table
CREATE TABLE "galleries" (
  "id" text NOT NULL,
  "title" text NOT NULL,
  "slug" text NOT NULL,
  "category_id" smallint NOT NULL,
  "description" text NOT NULL,
  "views" bigint NOT NULL DEFAULT 0,
  "published_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_galleries_slug" UNIQUE ("slug"),
  CONSTRAINT "fk_galleries_category_media" FOREIGN KEY ("category_id") REFERENCES "category_media" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "roles" table
CREATE TABLE "roles" (
  "id" smallserial NOT NULL,
  "role" character varying(20) NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_roles_role" UNIQUE ("role")
);
-- Create "users" table
CREATE TABLE "users" (
  "id" text NOT NULL,
  "username" text NOT NULL,
  "email" text NOT NULL,
  "password" text NOT NULL,
  "role_id" smallint NOT NULL DEFAULT 1,
  "status" boolean NOT NULL DEFAULT true,
  "email_verified" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_users_email" UNIQUE ("email"),
  CONSTRAINT "uni_users_username" UNIQUE ("username"),
  CONSTRAINT "fk_users_role" FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
