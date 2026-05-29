-- Modify "ambulance_service_requests" table
ALTER TABLE "ambulance_service_requests" ADD CONSTRAINT "fk_ambulance_service_requests_ambulance" FOREIGN KEY ("ambulance_id") REFERENCES "ambulances" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
