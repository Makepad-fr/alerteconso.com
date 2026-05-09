DO $$
BEGIN
  IF to_regclass('recalls') IS NOT NULL AND EXISTS (
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

DO $$
BEGIN
  IF to_regclass('recalls') IS NOT NULL AND EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'recalls'
      AND column_name = 'date_publication'
  ) THEN
    IF EXISTS (SELECT 1 FROM recalls WHERE date_publication IS NULL) THEN
      RAISE EXCEPTION 'recalls.date_publication contains NULL values; populate or remove those rows before applying 001_recall_publication_timestamptz.sql';
    END IF;

    ALTER TABLE recalls
      ALTER COLUMN date_publication SET NOT NULL;
  END IF;
END $$;
