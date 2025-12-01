INSERT INTO relations (uid, left, relation, right)
SELECT uid, itemid, 2, requires FROM entryInfo WHERE requires != 0;

ALTER TABLE entryInfo DROP COLUMN requires;

UPDATE DBInfo SET version = 9;
