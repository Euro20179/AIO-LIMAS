INSERT INTO relations (uid, left, relation, right)
SELECT uid, itemid, 1, parentid FROM entryInfo WHERE parentId != 0;

ALTER TABLE entryInfo DROP COLUMN parentId;

UPDATE DBInfo SET version = 7;
