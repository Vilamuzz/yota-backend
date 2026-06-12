-- Modify "news" table
ALTER TABLE "news" ALTER COLUMN "category" DROP NOT NULL, ALTER COLUMN "cover_image" DROP NOT NULL, ALTER COLUMN "deleted_at" DROP NOT NULL;
