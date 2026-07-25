PRAGMA user_version = 15;

CREATE TABLE IF NOT EXISTS entryInfo (
    uid INTEGER,
    itemId INTEGER,
    en_title TEXT,
    native_title TEXT,
    format INTEGER,
    location TEXT,
    collection TEXT,
    type TEXT,
    artStyle INTEGER,
    library INTEGER,
    recommendedBy TEXT,
    priority INTEGER
);

CREATE TABLE IF NOT EXISTS metadata (
    uid INTEGER,
    itemId INTEGER,
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
    genres TEXT,
    country STRING
);

CREATE TABLE IF NOT EXISTS userViewingInfo (
        uid INTEGER,
        itemId INTEGER,
        status TEXT,
        viewCount INTEGER,
        userRating NUMERIC,
        notes TEXT,
        currentPosition TEXT,
        extra TEXT,
        minutes NUMERIC
);

CREATE TABLE IF NOT EXISTS userEventInfo (
    uid INTEGER,
    itemId INTEGER,
    timestamp INTEGER,
    after INTEGER,
    event TEXT,
    timezone TEXT,
    beforeTS INTEGER
);

CREATE TABLE IF NOT EXISTS relations (
    uid INTEGER,
    left INTEGER,
    relation INTEGER,
    right INTEGER
);

CREATE TABLE IF NOT EXISTS transactions (
    uid INTEGER NOT NULL,
    itemId INTEGER NOT NULL,
    eventId INTEGER NOT NULL DEFAULT 0,
    price NUMBER,
    currency STRING
);

