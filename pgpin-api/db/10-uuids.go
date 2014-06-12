BEGIN;

ALTER TABLE pins
DROP CONSTRAINT pins_db_id_references_dbs_id;
	
ALTER TABLE pins
ALTER COLUMN id TYPE uuid USING id::uuid;

ALTER TABLE pins
ALTER COLUMN db_id TYPE uuid USING db_id::uuid;

ALTER TABLE dbs
ALTER COLUMN id TYPE uuid USING id::uuid;

ALTER TABLE pins
ADD CONSTRAINT pins_db_id_references_dbs_id
FOREIGN KEY (db_id)
REFERENCES dbs (id);

COMMIT;
