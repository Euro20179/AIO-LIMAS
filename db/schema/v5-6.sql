/*Transitional schema for the relations table*/

CREATE TABLE relations (uid NUMBER, left NUMBER, relation NUMBER, right NUMBER);

UPDATE DBInfo SET version = 6;
