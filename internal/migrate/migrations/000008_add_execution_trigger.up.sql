-- Add trigger column to track execution source (http or cron)
ALTER TABLE executions ADD COLUMN trigger TEXT DEFAULT 'http';
