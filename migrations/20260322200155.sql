-- Modify "prayers" table
ALTER TABLE "prayers" ADD CONSTRAINT "fk_prayers_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;
