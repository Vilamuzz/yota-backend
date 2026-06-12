-- Modify "news_comments" table
ALTER TABLE "news_comments" ADD COLUMN "report_count" bigint NULL DEFAULT 0, ADD COLUMN "reported" boolean NULL, ADD CONSTRAINT "fk_news_comments_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Create index "idx_news_comments_reported" to table: "news_comments"
CREATE INDEX "idx_news_comments_reported" ON "news_comments" ("reported");
