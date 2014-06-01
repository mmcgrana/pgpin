BEGIN;

ALTER TABLE pins
DROP COLUMN results_at;

ALTER TABLE pins
ADD COLUMN query_started_at timestamptz;

ALTER TABLE pins
ADD COLUMN query_finished_at timestamptz;

COMMIT;
