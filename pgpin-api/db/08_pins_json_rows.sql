BEGIN;

ALTER TABLE pins
ALTER COLUMN results_rows_json TYPE json USING to_json(results_rows_json);

ALTER TABLE pins
RENAME COLUMN results_rows_json TO results_rows;

COMMIT;
