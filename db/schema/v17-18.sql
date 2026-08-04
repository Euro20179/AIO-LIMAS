CREATE TABLE entrySettings (
    itemid INTEGER PRIMARY KEY NOT NULL,
    permissions INTEGER NOT NULL
);

/* set permissions to PERM_READ (publid read) by default */
INSERT INTO entrySettings
SELECT itemid, 1 FROM entryInfo;
