ALTER TABLE userViewingInfo ADD COLUMN minutes NUMERIC;
UPDATE userViewingInfo SET minutes = 0;
INSERT INTO DBInfo (version) VALUES (1);
