CREATE TABLE pins (
    id                 char(12) PRIMARY KEY,
    name               text NOT NULL,
    db_id              char(12),
    query              text NOT NULL,
    created_at         timestamptz NOT NULL,
    query_started_at   timestamptz,
    query_finished_at  timestamptz,
    result_fields_json text,
    result_rows_json   text,
    result_error       text
);
