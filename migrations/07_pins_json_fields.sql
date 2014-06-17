BEGIN;

ALTER TABLE pins
ALTER COLUMN results_fields_json TYPE json USING to_json(results_fields_json);

ALTER TABLE pins
RENAME COLUMN results_fields_json TO results_fields;

COMMIT;
