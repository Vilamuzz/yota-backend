-- Modify "users" table
ALTER TABLE "users" DROP COLUMN "role_id";
-- Create "user_roles" table
CREATE TABLE "user_roles" (
  "user_id" text NOT NULL,
  "role_id" smallint NOT NULL,
  PRIMARY KEY ("user_id", "role_id"),
  CONSTRAINT "fk_user_roles_role" FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_user_roles_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
