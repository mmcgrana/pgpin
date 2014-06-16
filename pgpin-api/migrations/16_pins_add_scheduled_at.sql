ALTER TABLE pins
ADD COLUMN scheduled_at timestamptz NOT NULL DEFAULT now();
