-- Create index "uni_finance_records_source" to table: "finance_records"
-- NOTE: Run the duplicate pre-flight check before applying:
--   SELECT source_type, source_id, count(*) FROM finance_records
--   GROUP BY 1,2 HAVING count(*) > 1;
CREATE UNIQUE INDEX "uni_finance_records_source" ON "finance_records" ("source_type", "source_id");
