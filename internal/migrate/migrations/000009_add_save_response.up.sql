-- Add save_response setting to functions and response_json to executions
ALTER TABLE functions ADD COLUMN save_response BOOLEAN DEFAULT 0;
ALTER TABLE executions ADD COLUMN response_json TEXT;
