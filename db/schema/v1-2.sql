ALTER TABLE userEventInfo ADD COLUMN beforeTS INTEGER;
UPDATE userEventInfo SET beforeTS = 0;
UPDATE DBInfo SET version = 2;
