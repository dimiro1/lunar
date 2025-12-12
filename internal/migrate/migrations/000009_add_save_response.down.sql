-- Remove save_response setting from functions and response_json from executions
-- Note: SQLite does not support DROP COLUMN in older versions
-- This migration may require recreating the tables in older SQLite versions
ALTER TABLE functions DROP COLUMN save_response;
ALTER TABLE executions DROP COLUMN response_json;
