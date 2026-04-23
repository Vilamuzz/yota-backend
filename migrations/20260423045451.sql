-- Create "donation_programs" table
CREATE TABLE "donation_programs" (
  "id" text NOT NULL,
  "title" text NOT NULL,
  "slug" text NOT NULL,
  "cover_image" text NULL,
  "category" text NOT NULL,
  "description" text NOT NULL,
  "fund_target" numeric NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'active',
  "start_date" timestamptz NOT NULL,
  "end_date" timestamptz NOT NULL,
  "published_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_donation_programs_slug" UNIQUE ("slug")
);
-- Create index "idx_donation_programs_deleted_at" to table: "donation_programs"
CREATE INDEX "idx_donation_programs_deleted_at" ON "donation_programs" ("deleted_at");
-- Create index "idx_donation_programs_status" to table: "donation_programs"
CREATE INDEX "idx_donation_programs_status" ON "donation_programs" ("status");
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
  PRIMARY KEY ("id")
);
-- Create "roles" table
CREATE TABLE "roles" (
  "id" bigserial NOT NULL,
  "name" character varying(20) NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_roles_name" UNIQUE ("name")
);
-- Create "email_verification_tokens" table
CREATE TABLE "email_verification_tokens" (
  "id" text NOT NULL,
  "account_id" text NOT NULL,
  "token" text NOT NULL,
  "expired_at" timestamptz NOT NULL,
  "is_used" boolean NULL DEFAULT false,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_email_verification_tokens_token" UNIQUE ("token")
);
-- Create "ambulance_requests" table
CREATE TABLE "ambulance_requests" (
  "id" text NOT NULL,
  "account_id" text NOT NULL,
  "applicant_name" text NOT NULL,
  "applicant_phone" text NOT NULL,
  "applicant_address" text NOT NULL,
  "description" text NOT NULL,
  "request_date" timestamptz NOT NULL,
  "request_reason" text NOT NULL,
  "status" text NOT NULL,
  "rejection_reason" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "ambulances" table
CREATE TABLE "ambulances" (
  "id" text NOT NULL,
  "driver_id" text NULL,
  "image" text NOT NULL,
  "plate_number" text NOT NULL,
  "phone" text NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "logs" table
CREATE TABLE "logs" (
  "id" text NOT NULL,
  "user_id" text NULL,
  "action" text NOT NULL,
  "entity_type" text NOT NULL,
  "entity_id" text NOT NULL,
  "old_value" text NULL,
  "new_value" text NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_logs_user_id" to table: "logs"
CREATE INDEX "idx_logs_user_id" ON "logs" ("user_id");
-- Create "foundation_profiles" table
CREATE TABLE "foundation_profiles" (
  "id" text NOT NULL,
  "image" text NULL,
  "type" text NULL,
  "order" bigint NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "ambulance_histories" table
CREATE TABLE "ambulance_histories" (
  "id" text NOT NULL,
  "ambulance_id" text NOT NULL,
  "driver_id" text NOT NULL,
  "service_category" text NOT NULL,
  "note" text NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "password_reset_tokens" table
CREATE TABLE "password_reset_tokens" (
  "id" text NOT NULL,
  "account_id" text NOT NULL,
  "token" text NOT NULL,
  "expired_at" timestamptz NOT NULL,
  "is_used" boolean NULL DEFAULT false,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_password_reset_tokens_token" UNIQUE ("token")
);
-- Create "accounts" table
CREATE TABLE "accounts" (
  "id" text NOT NULL,
  "email" text NOT NULL,
  "password" text NOT NULL,
  "is_banned" boolean NOT NULL DEFAULT false,
  "email_verified" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_accounts_email" UNIQUE ("email")
);
-- Create "account_roles" table
CREATE TABLE "account_roles" (
  "account_id" text NOT NULL,
  "role_id" bigint NOT NULL,
  "is_default" boolean NULL DEFAULT false,
  "is_active" boolean NULL DEFAULT true,
  PRIMARY KEY ("account_id", "role_id"),
  CONSTRAINT "fk_account_roles_role" FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_accounts_account_roles" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "foster_childrens" table
CREATE TABLE "foster_childrens" (
  "id" text NOT NULL,
  "slug" text NOT NULL,
  "name" text NOT NULL,
  "profile_picture" text NOT NULL,
  "gender" text NOT NULL,
  "is_graduated" boolean NOT NULL,
  "category" text NULL,
  "birth_date" timestamptz NULL,
  "birth_place" text NULL,
  "address" text NULL,
  "family_card" text NOT NULL,
  "sktm" text NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_foster_childrens_deleted_at" to table: "foster_childrens"
CREATE INDEX "idx_foster_childrens_deleted_at" ON "foster_childrens" ("deleted_at");
-- Create "achivements" table
CREATE TABLE "achivements" (
  "id" text NOT NULL,
  "foster_children_id" text NOT NULL,
  "url" text NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_foster_childrens_achivements" FOREIGN KEY ("foster_children_id") REFERENCES "foster_childrens" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "donation_program_expenses" table
CREATE TABLE "donation_program_expenses" (
  "id" text NOT NULL,
  "donation_program_id" text NOT NULL,
  "title" text NOT NULL,
  "amount" numeric NOT NULL,
  "expense_date" timestamptz NOT NULL,
  "note" text NOT NULL,
  "proof_file" text NULL,
  "created_by" text NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_donation_program_expenses_account" FOREIGN KEY ("created_by") REFERENCES "accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_donation_program_expenses_donation_program_id" to table: "donation_program_expenses"
CREATE INDEX "idx_donation_program_expenses_donation_program_id" ON "donation_program_expenses" ("donation_program_id");
-- Create "donation_program_transactions" table
CREATE TABLE "donation_program_transactions" (
  "id" text NOT NULL,
  "donation_program_id" text NOT NULL,
  "order_id" text NULL,
  "account_id" text NULL,
  "donor_name" text NULL,
  "donor_email" text NULL,
  "is_online" boolean NULL,
  "gross_amount" numeric NULL,
  "fraud_status" text NULL,
  "transaction_status" text NULL,
  "provider" text NULL,
  "transaction_id" text NULL,
  "snap_token" text NULL,
  "snap_redirect_url" text NULL,
  "paid_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_donation_program_transactions_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_donation_program_transactions_account_id" to table: "donation_program_transactions"
CREATE INDEX "idx_donation_program_transactions_account_id" ON "donation_program_transactions" ("account_id");
-- Create index "idx_donation_program_transactions_donation_program_id" to table: "donation_program_transactions"
CREATE INDEX "idx_donation_program_transactions_donation_program_id" ON "donation_program_transactions" ("donation_program_id");
-- Create index "idx_donation_program_transactions_order_id" to table: "donation_program_transactions"
CREATE UNIQUE INDEX "idx_donation_program_transactions_order_id" ON "donation_program_transactions" ("order_id");
-- Create index "idx_donation_program_transactions_transaction_status" to table: "donation_program_transactions"
CREATE INDEX "idx_donation_program_transactions_transaction_status" ON "donation_program_transactions" ("transaction_status");
-- Create "foster_children_candidates" table
CREATE TABLE "foster_children_candidates" (
  "id" text NOT NULL,
  "name" text NOT NULL,
  "profile_picture" text NOT NULL,
  "gender" text NOT NULL,
  "category" text NULL,
  "birth_date" timestamptz NULL,
  "birth_place" text NULL,
  "address" text NULL,
  "family_card" text NOT NULL,
  "sktm" text NOT NULL,
  "submitter_name" text NULL,
  "submitter_phone" text NULL,
  "submitter_address" text NULL,
  "submitter_id_card" text NULL,
  "submitted_by" text NOT NULL,
  "status" text NULL,
  "rejection_reason" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_foster_children_candidates_account" FOREIGN KEY ("submitted_by") REFERENCES "accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "foster_children_expenses" table
CREATE TABLE "foster_children_expenses" (
  "id" text NOT NULL,
  "foster_children_id" text NOT NULL,
  "title" text NOT NULL,
  "amount" numeric NOT NULL,
  "expense_date" timestamptz NOT NULL,
  "note" text NOT NULL,
  "proof_file" text NULL,
  "created_by" text NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_foster_children_expenses_account" FOREIGN KEY ("created_by") REFERENCES "accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "foster_children_transactions" table
CREATE TABLE "foster_children_transactions" (
  "id" text NOT NULL,
  "foster_children_id" text NOT NULL,
  "order_id" text NULL,
  "account_id" text NULL,
  "donor_name" text NULL,
  "donor_email" text NULL,
  "is_online" boolean NULL,
  "gross_amount" numeric NULL,
  "fraud_status" text NULL,
  "transaction_status" text NULL,
  "provider" text NULL,
  "transaction_id" text NULL,
  "snap_token" text NULL,
  "snap_redirect_url" text NULL,
  "paid_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_foster_children_transactions_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_foster_children_transactions_order_id" to table: "foster_children_transactions"
CREATE UNIQUE INDEX "idx_foster_children_transactions_order_id" ON "foster_children_transactions" ("order_id");
-- Create "galleries" table
CREATE TABLE "galleries" (
  "id" text NOT NULL,
  "title" text NOT NULL,
  "slug" text NOT NULL,
  "category" text NOT NULL,
  "cover_image" text NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'draft',
  "description" text NOT NULL,
  "views" bigint NOT NULL DEFAULT 0,
  "published_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_galleries_slug" UNIQUE ("slug")
);
-- Create index "idx_galleries_deleted_at" to table: "galleries"
CREATE INDEX "idx_galleries_deleted_at" ON "galleries" ("deleted_at");
-- Create index "idx_galleries_published_at" to table: "galleries"
CREATE INDEX "idx_galleries_published_at" ON "galleries" ("published_at");
-- Create "news" table
CREATE TABLE "news" (
  "id" text NOT NULL,
  "title" text NOT NULL,
  "slug" text NOT NULL,
  "category" text NOT NULL,
  "cover_image" text NOT NULL,
  "content" text NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'draft',
  "views" bigint NOT NULL DEFAULT 0,
  "published_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_news_slug" UNIQUE ("slug")
);
-- Create index "idx_news_deleted_at" to table: "news"
CREATE INDEX "idx_news_deleted_at" ON "news" ("deleted_at");
-- Create index "idx_news_published_at" to table: "news"
CREATE INDEX "idx_news_published_at" ON "news" ("published_at");
-- Create "media" table
CREATE TABLE "media" (
  "id" text NOT NULL,
  "news_id" text NULL,
  "gallery_id" text NULL,
  "type" text NOT NULL,
  "url" text NOT NULL,
  "alt_text" text NULL,
  "order" bigint NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_galleries_media" FOREIGN KEY ("gallery_id") REFERENCES "galleries" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_news_media" FOREIGN KEY ("news_id") REFERENCES "news" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_media_gallery_id" to table: "media"
CREATE INDEX "idx_media_gallery_id" ON "media" ("gallery_id");
-- Create index "idx_media_news_id" to table: "media"
CREATE INDEX "idx_media_news_id" ON "media" ("news_id");
-- Create "news_comments" table
CREATE TABLE "news_comments" (
  "id" text NOT NULL,
  "news_id" text NOT NULL,
  "parent_comment_id" text NULL,
  "account_id" text NOT NULL,
  "content" text NOT NULL,
  "created_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_news_comments_news" FOREIGN KEY ("news_id") REFERENCES "news" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_news_comments_parent_comment" FOREIGN KEY ("parent_comment_id") REFERENCES "news_comments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_news_comments_account_id" to table: "news_comments"
CREATE INDEX "idx_news_comments_account_id" ON "news_comments" ("account_id");
-- Create index "idx_news_comments_deleted_at" to table: "news_comments"
CREATE INDEX "idx_news_comments_deleted_at" ON "news_comments" ("deleted_at");
-- Create index "idx_news_comments_news_id" to table: "news_comments"
CREATE INDEX "idx_news_comments_news_id" ON "news_comments" ("news_id");
-- Create "news_comment_reports" table
CREATE TABLE "news_comment_reports" (
  "account_id" text NOT NULL,
  "news_comment_id" text NOT NULL,
  "reason" text NOT NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("account_id", "news_comment_id"),
  CONSTRAINT "fk_news_comments_news_comment_reports" FOREIGN KEY ("news_comment_id") REFERENCES "news_comments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "prayers" table
CREATE TABLE "prayers" (
  "id" text NOT NULL,
  "donation_transaction_id" text NOT NULL,
  "content" text NOT NULL,
  "is_published" boolean NULL DEFAULT false,
  "created_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_prayers_donation_program_transaction" FOREIGN KEY ("donation_transaction_id") REFERENCES "donation_program_transactions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_prayers_deleted_at" to table: "prayers"
CREATE INDEX "idx_prayers_deleted_at" ON "prayers" ("deleted_at");
-- Create index "idx_prayers_donation_transaction_id" to table: "prayers"
CREATE UNIQUE INDEX "idx_prayers_donation_transaction_id" ON "prayers" ("donation_transaction_id");
-- Create index "idx_prayers_is_published" to table: "prayers"
CREATE INDEX "idx_prayers_is_published" ON "prayers" ("is_published");
-- Create "prayer_amens" table
CREATE TABLE "prayer_amens" (
  "prayer_id" text NOT NULL,
  "account_id" text NOT NULL,
  PRIMARY KEY ("prayer_id", "account_id"),
  CONSTRAINT "fk_prayers_prayer_amens" FOREIGN KEY ("prayer_id") REFERENCES "prayers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "prayer_reports" table
CREATE TABLE "prayer_reports" (
  "prayer_id" text NOT NULL,
  "account_id" text NOT NULL,
  "reason" text NOT NULL,
  PRIMARY KEY ("prayer_id", "account_id"),
  CONSTRAINT "fk_prayers_prayer_reports" FOREIGN KEY ("prayer_id") REFERENCES "prayers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "social_program_expenses" table
CREATE TABLE "social_program_expenses" (
  "id" text NOT NULL,
  "social_program_id" text NOT NULL,
  "title" text NOT NULL,
  "amount" numeric NOT NULL,
  "expense_date" timestamptz NOT NULL,
  "note" text NOT NULL,
  "proof_file" text NULL,
  "created_by" text NOT NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_social_program_expenses_account" FOREIGN KEY ("created_by") REFERENCES "accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_social_program_expenses_social_program_id" to table: "social_program_expenses"
CREATE INDEX "idx_social_program_expenses_social_program_id" ON "social_program_expenses" ("social_program_id");
-- Create "social_programs" table
CREATE TABLE "social_programs" (
  "id" text NOT NULL,
  "slug" text NOT NULL,
  "title" text NOT NULL,
  "description" text NOT NULL,
  "cover_image" text NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'active',
  "minimum_amount" numeric NOT NULL,
  "billing_day" bigint NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_social_programs_deleted_at" to table: "social_programs"
CREATE INDEX "idx_social_programs_deleted_at" ON "social_programs" ("deleted_at");
-- Create "social_program_subscriptions" table
CREATE TABLE "social_program_subscriptions" (
  "id" text NOT NULL,
  "social_program_id" text NOT NULL,
  "account_id" text NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'active',
  "amount" numeric NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_social_program_subscriptions_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_social_program_subscriptions_social_program" FOREIGN KEY ("social_program_id") REFERENCES "social_programs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "social_program_invoices" table
CREATE TABLE "social_program_invoices" (
  "id" text NOT NULL,
  "subscription_id" text NOT NULL,
  "billing_period" timestamptz NOT NULL,
  "amount" numeric NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'active',
  "due_date" timestamptz NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_status_due_date" to table: "social_program_invoices"
CREATE INDEX "idx_status_due_date" ON "social_program_invoices" ("status", "due_date");
-- Create index "idx_subscription_billing" to table: "social_program_invoices"
CREATE UNIQUE INDEX "idx_subscription_billing" ON "social_program_invoices" ("subscription_id", "billing_period");
-- Create "social_program_transactions" table
CREATE TABLE "social_program_transactions" (
  "id" text NOT NULL,
  "social_program_invoice_id" text NOT NULL,
  "order_id" text NULL,
  "account_id" text NOT NULL,
  "is_online" boolean NULL,
  "gross_amount" numeric NULL,
  "fraud_status" text NULL,
  "transaction_status" text NULL,
  "provider" text NULL,
  "transaction_id" text NULL,
  "snap_token" text NULL,
  "snap_redirect_url" text NULL,
  "paid_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_social_program_transactions_order_id" UNIQUE ("order_id"),
  CONSTRAINT "uni_social_program_transactions_social_program_invoice_id" UNIQUE ("social_program_invoice_id"),
  CONSTRAINT "uni_social_program_transactions_transaction_id" UNIQUE ("transaction_id"),
  CONSTRAINT "fk_social_program_transactions_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_social_program_transactions_social_program_invoice" FOREIGN KEY ("social_program_invoice_id") REFERENCES "social_program_invoices" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_social_program_transactions_social_program_invoice_id" to table: "social_program_transactions"
CREATE INDEX "idx_social_program_transactions_social_program_invoice_id" ON "social_program_transactions" ("social_program_invoice_id");
-- Create "user_profiles" table
CREATE TABLE "user_profiles" (
  "id" text NOT NULL,
  "account_id" text NOT NULL,
  "username" text NULL,
  "phone" text NULL,
  "address" text NULL,
  "profile_picture" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_user_profiles_account_id" UNIQUE ("account_id"),
  CONSTRAINT "uni_user_profiles_phone" UNIQUE ("phone"),
  CONSTRAINT "uni_user_profiles_username" UNIQUE ("username"),
  CONSTRAINT "fk_accounts_user_profile" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
