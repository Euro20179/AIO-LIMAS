ALTER TABLE metadata ADD COLUMN genres STRING;
UPDATE metadata SET genres = '';
UPDATE DBInfo SET version = 4;
