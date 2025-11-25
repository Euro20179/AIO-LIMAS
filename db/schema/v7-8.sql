INSERT INTO relations (uid, left, relation, right)
SELECT uid, itemid, 3, copyOf FROM entryInfo WHERE copyOf != 0;

DELETE FROM relations as top WHERE relation = 3 and top.left = (
    SELECT right FROM relations WHERE relation = 3 and right = top.left
);

ALTER TABLE entryInfo DROP COLUMN copyOf;

UPDATE DBInfo SET version = 8;
