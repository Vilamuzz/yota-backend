-- Rename a column from "account_id" to "submitted_by"
ALTER TABLE "ambulance_service_requests" RENAME COLUMN "account_id" TO "submitted_by";
-- Rename a column from "applicant_name" to "submitter_name"
ALTER TABLE "ambulance_service_requests" RENAME COLUMN "applicant_name" TO "submitter_name";
-- Rename a column from "applicant_phone" to "submitter_phone"
ALTER TABLE "ambulance_service_requests" RENAME COLUMN "applicant_phone" TO "submitter_phone";
-- Rename a column from "applicant_address" to "submitter_id_card"
ALTER TABLE "ambulance_service_requests" RENAME COLUMN "applicant_address" TO "submitter_id_card";
-- Rename a column from "description" to "patient_name"
ALTER TABLE "ambulance_service_requests" RENAME COLUMN "description" TO "patient_name";
-- Rename a column from "request_date" to "pickup_date"
ALTER TABLE "ambulance_service_requests" RENAME COLUMN "request_date" TO "pickup_date";
-- Rename a column from "request_reason" to "patient_address"
ALTER TABLE "ambulance_service_requests" RENAME COLUMN "request_reason" TO "patient_address";
-- Modify "ambulance_service_requests" table
ALTER TABLE "ambulance_service_requests" ADD COLUMN "patient_age" bigint NULL, ADD COLUMN "is_infectious" boolean NULL, ADD COLUMN "disease" text NULL, ADD COLUMN "is_able_to_sit" boolean NULL, ADD COLUMN "pickup_time" timestamptz NULL, ADD COLUMN "destination" text NULL, ADD COLUMN "note" text NULL, ADD CONSTRAINT "fk_ambulance_service_requests_account" FOREIGN KEY ("submitted_by") REFERENCES "accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
