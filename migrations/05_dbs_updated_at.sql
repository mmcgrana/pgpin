ALTER TABLE dbs
ADD COLUMN updated_at timestamptz NOT NULL DEFAULT now();
