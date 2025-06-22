ALTER TABLE entryInfo ADD COLUMN requires INTEGER;
UPDATE entryInfo SET requires = 0;
UPDATE DBInfo SET version = 3;
