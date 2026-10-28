ALTER TABLE threats
    ALTER COLUMN aliases        DROP DEFAULT,
    ALTER COLUMN file_name      DROP DEFAULT,
    ALTER COLUMN platform       DROP DEFAULT,
    ALTER COLUMN tags           DROP DEFAULT,
    ALTER COLUMN "references"   DROP DEFAULT,
    ALTER COLUMN behaviors      DROP DEFAULT,
    ALTER COLUMN related_hashes DROP DEFAULT,
    ALTER COLUMN detections     DROP DEFAULT,
    ALTER COLUMN network_iocs   DROP DEFAULT,
    ALTER COLUMN file_iocs      DROP DEFAULT;