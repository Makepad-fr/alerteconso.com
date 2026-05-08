DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'recalls'
      AND column_name = 'date_publication'
      AND data_type = 'timestamp without time zone'
  ) THEN
    ALTER TABLE recalls
      ALTER COLUMN date_publication TYPE TIMESTAMPTZ
      USING date_publication AT TIME ZONE 'UTC';
  END IF;
END $$;

ALTER TABLE recalls
  ALTER COLUMN date_publication SET NOT NULL;
