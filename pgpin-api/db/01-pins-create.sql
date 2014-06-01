CREATE TABLE pins (
    id                   char(32) PRIMARY KEY,
    resource_id          text NOT NULL,
    name                 text NOT NULL,
    sql                  text NOT NULL,
    user_id              text NOT NULL,
    created_at           timestamptz NOT NULL,
    resource_url         text NOT NULL,
    results_fields_json  text,
    results_rows_json    text,
    results_at           timestamptz,
    deleted_at           timestamptz
);
