CREATE TABLE dbs (
    id       char(12) PRIMARY KEY,
    name     text NOT NULL,
    added_at timestamptz NOT NULL,
    url      text NOT NULL
);
