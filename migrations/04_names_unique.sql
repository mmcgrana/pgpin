BEGIN;

CREATE UNIQUE INDEX pins_name_unique
ON pins (name)
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX dbs_name_unique
ON dbs (name)
WHERE removed_at IS NULL;

COMMIT;
