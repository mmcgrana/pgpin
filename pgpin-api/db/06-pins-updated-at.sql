ALTER TABLE pins
ADD COLUMN updated_at timestamptz NOT NULL DEFAULT now();
