CREATE TABLE temp_entryInfo AS SELECT * FROM entryInfo;
CREATE TABLE temp_metadata AS SELECT * FROM metadata;
CREATE TABLE temp_userInfo AS SELECT * FROM userViewingInfo;
CREATE TABLE temp_eventInfo AS SELECT * FROM userEventInfo;
CREATE TABLE temp_relations AS SELECT * FROM relations;

DROP TABLE entryInfo;
DROP TABLE metadata;
DROP TABLE userViewingInfo;
DROP TABLE userEventInfo;
DROP TABLE relations;

CREATE TABLE entryInfo (
    uid INTEGER,
    oldid INTEGER,
    itemid INTEGER PRIMARY KEY ASC,
    en_title TEXT,
    native_title TEXT,
    format INTEGER,
    location TEXT,
    purchasePrice NUMERIC,
    collection TEXT,
    type TEXT,
    artStyle INTEGER,
    library INTEGER,
    recommendedBy TEXT
);

CREATE TABLE metadata (
    uid INTEGER,
    itemid INTEGER PRIMARY KEY ASC,
    rating NUMERIC,
    description TEXT,
    releaseYear INTEGER,
    thumbnail TEXT,
    mediaDependant TEXT,
    dataPoints TEXT,
    title TEXT,
    native_title TEXT,
    ratingMax NUMERIC,
    provider TEXT,
    providerID TEXT,
    genres TEXT
);

INSERT INTO entryInfo
SELECT
    uid,
    itemid,
    null,
    en_title ,
    native_title ,
    format,
    location,
    purchasePrice ,
    collection ,
    type,
    artStyle ,
    library ,
    recommendedBy
FROM temp_entryInfo ;

INSERT INTO metadata
SELECT
    entryInfo.uid,
    entryInfo.itemid,
    rating ,
    description ,
    releaseYear ,
    thumbnail ,
    mediaDependant ,
    dataPoints ,
    title ,
    temp_metadata.native_title ,
    ratingMax ,
    provider ,
    providerID ,
    genres 
FROM temp_metadata JOIN entryInfo on temp_metadata.itemid = entryInfo.oldid;

CREATE TABLE userViewingInfo AS SELECT
    entryInfo.uid,
    entryInfo.itemid,
    status,
    viewCount,
    userRating,
    notes,
    currentPosition,
    extra,
    minutes
FROM temp_userInfo JOIN entryInfo ON temp_userInfo.itemid = entryInfo.oldid;

CREATE TABLE userEventInfo AS SELECT
    entryInfo.uid,
    entryInfo.itemid,
    timestamp,
    after,
    event,
    timezone,
    beforeTS
FROM temp_eventInfo JOIN entryInfo ON temp_eventInfo.itemid = entryInfo.oldid;

CREATE TABLE relations (
    uid INTEGER,
    left INTEGER,
    relation INTEGER,
    right INTEGER
);

INSERT INTO relations
SELECT
    temp_relations.uid,
    (SELECT itemid FROM entryInfo WHERE oldid = left),
    relation,
    (SELECT itemid from entryinfo where oldid = right)
FROM temp_relations;

DROP TABLE temp_entryInfo;
DROP TABLE temp_metadata;
DROP TABLE temp_userInfo;
DROP TABLE temp_eventInfo;
DROP TABLE temp_relations;

ALTER TABLE entryInfo DROP COLUMN oldid;

UPDATE DBInfo SET version = 10;
