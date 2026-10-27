ALTER TABLE threats
    ALTER COLUMN aliases        SET DEFAULT '[]'::jsonb,
    ALTER COLUMN file_name      SET DEFAULT '[]'::jsonb,
    ALTER COLUMN platform       SET DEFAULT '[]'::jsonb,
    ALTER COLUMN tags           SET DEFAULT '[]'::jsonb,
    ALTER COLUMN "references"   SET DEFAULT '[]'::jsonb,
    ALTER COLUMN behaviors      SET DEFAULT '[]'::jsonb,
    ALTER COLUMN related_hashes SET DEFAULT '[]'::jsonb,
    ALTER COLUMN detections     SET DEFAULT '[]'::jsonb,
    ALTER COLUMN network_iocs   SET DEFAULT '[]'::jsonb,
    ALTER COLUMN file_iocs      SET DEFAULT '[]'::jsonb;