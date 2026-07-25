PRAGMA user_version = 16;
ALTER TABLE entryInfo
ADD COLUMN format_modifiers INTEGER NOT NULL DEFAULT 0;

UPDATE entryInfo SET
    format_modifiers = 1,
    format = format - 0x1000
WHERE format > 0xFFF;
