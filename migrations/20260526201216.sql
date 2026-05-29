-- Modify "ambulance_histories" table
ALTER TABLE "ambulance_histories" ADD CONSTRAINT "fk_ambulance_histories_driver" FOREIGN KEY ("driver_id") REFERENCES "accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
