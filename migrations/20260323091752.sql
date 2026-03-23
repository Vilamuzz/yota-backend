-- Rename a column from "like_count" to "amen_count"
ALTER TABLE "prayers" RENAME COLUMN "like_count" TO "amen_count";
-- Create "prayer_amens" table
CREATE TABLE "prayer_amens" (
  "id" text NOT NULL,
  "prayer_id" text NOT NULL,
  "user_id" text NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "prayer_reports" table
CREATE TABLE "prayer_reports" (
  "id" text NOT NULL,
  "prayer_id" text NOT NULL,
  "user_id" text NOT NULL,
  "reason" text NOT NULL,
  PRIMARY KEY ("id")
);
