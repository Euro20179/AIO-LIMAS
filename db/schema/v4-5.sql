ALTER TABLE entryInfo ADD COLUMN recommendedBy STRING;
UPDATE entryInfo SET recommendedBy = '';
UPDATE DBInfo SET version = 5;
