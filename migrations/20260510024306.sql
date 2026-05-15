-- Modify "galleries" table
ALTER TABLE "galleries" DROP CONSTRAINT "uni_galleries_slug";
-- Modify "news" table
ALTER TABLE "news" DROP CONSTRAINT "uni_news_slug";
