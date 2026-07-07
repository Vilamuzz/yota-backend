-- Modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" ADD COLUMN "new_email" text NULL DEFAULT '';
