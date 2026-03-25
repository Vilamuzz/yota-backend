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
