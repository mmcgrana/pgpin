CREATE TABLE dbs (
    id         char(12) PRIMARY KEY,
    name       text NOT NULL,
    url        text NOT NULL,
    added_at   timestamptz NOT NULL,
    removed_at timestamptz NOT NULL
);
