ALTER TABLE pins
ADD CONSTRAINT pins_db_id_references_dbs_id
FOREIGN KEY (db_id)
REFERENCES dbs (id);
