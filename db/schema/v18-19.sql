ALTER TABLE entryInfo ADD COLUMN metadataId INTEGER;

UPDATE entryInfo SET metadataId = (
    SELECT itemid FROM metadata WHERE metadata.itemid = entryInfo.itemid
);
