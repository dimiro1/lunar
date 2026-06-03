-- Add the language a version's code is written in. Existing rows predate
-- multi-language support and are all Lua, so default to 'lua'.
ALTER TABLE function_versions ADD COLUMN language TEXT NOT NULL DEFAULT 'lua';
