-- Create "backups" table
CREATE TABLE "backups" (
  "id" uuid NOT NULL,
  "filename" character varying(255) NOT NULL,
  "size" bigint NOT NULL,
  "duration" bigint NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_backups_filename" UNIQUE ("filename")
);
-- Create index "idx_backups_deleted_at" to table: "backups"
CREATE INDEX "idx_backups_deleted_at" ON "backups" ("deleted_at");
