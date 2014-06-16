BEGIN;

ALTER TABLE dbs
RENAME COLUMN added_at TO created_at;

ALTER TABLE dbs
RENAME COLUMN removed_at TO deleted_at;

COMMIT;
