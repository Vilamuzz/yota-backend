-- Modify "news_comment_reports" table
ALTER TABLE "news_comment_reports" DROP COLUMN "reason", DROP COLUMN "created_at";
-- Modify "prayer_reports" table
ALTER TABLE "prayer_reports" DROP COLUMN "reason";
