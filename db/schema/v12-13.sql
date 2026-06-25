ALTER TABLE entryInfo ADD COLUMN priority INTEGER;
UPDATE entryInfo SET priority = 0;
UPDATE DBInfo SET version = 13
