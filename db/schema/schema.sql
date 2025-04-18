CREATE TABLE IF NOT EXISTS entryInfo (
    itemId INTEGER,
    en_title TEXT,
    native_title TEXT,
    format INTEGER,
    location TEXT,
    purchasePrice NUMERIC,
    collection TEXT,
    type TEXT,
    parentId INTEGER,
    copyOf INTEGER,
    artStyle INTEGER,
    library INTEGER
);

CREATE TABLE IF NOT EXISTS metadata (
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
    providerID TEXT
);

CREATE TABLE IF NOT EXISTS userViewingInfo (
        itemId INTEGER,
        status TEXT,
        viewCount INTEGER,
        userRating NUMERIC,
        notes TEXT,
        currentPosition TEXT,
        extra TEXT
);

CREATE TABLE IF NOT EXISTS userEventInfo (
    itemId INTEGER,
    timestamp INTEGER,
    after INTEGER,
    event TEXT,
    timezone TEXT
);
