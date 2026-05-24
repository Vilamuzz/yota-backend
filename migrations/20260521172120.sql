-- Modify "donation_programs" table
ALTER TABLE "donation_programs" ADD COLUMN "collected_expense" numeric NULL;
-- Modify "news_comments" table
ALTER TABLE "news_comments" DROP CONSTRAINT "fk_news_comments_parent_comment", ADD CONSTRAINT "fk_news_comments_replies" FOREIGN KEY ("parent_comment_id") REFERENCES "news_comments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
