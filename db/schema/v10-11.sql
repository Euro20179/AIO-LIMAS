ALTER TABLE metadata ADD COLUMN country STRING;
UPDATE metadata set country = '';
UPDATE DBInfo SET version = 11;
