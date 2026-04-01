-- Modify "users" table
ALTER TABLE "users" ADD COLUMN "default_role_id" smallint NOT NULL, ADD CONSTRAINT "fk_users_default_role" FOREIGN KEY ("default_role_id") REFERENCES "roles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
