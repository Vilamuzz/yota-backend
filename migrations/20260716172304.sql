-- Modify "ambulance_service_requests" table
ALTER TABLE "ambulance_service_requests" ADD COLUMN "requested_at" timestamptz NULL, ADD COLUMN "picked_up_at" timestamptz NULL, ADD COLUMN "completed_at" timestamptz NULL;
