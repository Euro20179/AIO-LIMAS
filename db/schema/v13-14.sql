CREATE TABLE IF NOT EXISTS transactions (
    uid INTEGER NOT NULL,
    itemId INTEGER NOT NULL,
    eventId INTEGER NOT NULL DEFAULT 0,
    price NUMBER,
    currency STRING
);

INSERT INTO transactions
SELECT e.uid, e.itemid, ue.rowid, purchasePrice, 'USD'
    FROM entryInfo as e
    LEFT JOIN userEventInfo ue
        ON ue.itemId = e.itemId AND ue.event = 'Purchased'
    WHERE ue.rowid IS NOT NULL AND purchasePrice > 0;

INSERT INTO transactions
SELECT e.uid, e.itemid, 0, purchasePrice, 'USD'
    FROM entryInfo as e
    LEFT JOIN userEventInfo ue
        ON ue.itemId = e.itemId AND ue.event = 'Purchased'
    WHERE ue.rowid IS NULL AND purchasePrice > 0;

ALTER TABLE entryInfo DROP COLUMN purchasePrice;

UPDATE DBInfo SET version = 14;
