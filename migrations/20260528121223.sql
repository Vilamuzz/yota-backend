-- Rename a column from "image" to "foundation_name"
ALTER TABLE "foundation_profiles" RENAME COLUMN "image" TO "foundation_name";
-- Rename a column from "type" to "logo"
ALTER TABLE "foundation_profiles" RENAME COLUMN "type" TO "logo";
-- Modify "foundation_profiles" table
ALTER TABLE "foundation_profiles" DROP COLUMN "order", ADD COLUMN "icon" text NULL, ADD COLUMN "organization_structure" text NULL, ADD COLUMN "hero_image_one" text NULL, ADD COLUMN "hero_image_two" text NULL, ADD COLUMN "hero_image_three" text NULL, ADD COLUMN "hero_image_four" text NULL;
