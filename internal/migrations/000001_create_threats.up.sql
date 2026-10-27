CREATE TABLE threats (
    md5                varchar(32)  PRIMARY KEY
        CHECK (md5 ~ '^[a-fA-F0-9]{32}$'),

    sha1               varchar(40)
        CHECK (sha1 IS NULL OR sha1 ~ '^[a-fA-F0-9]{40}$'),
    sha256             varchar(64)
        CHECK (sha256 IS NULL OR sha256 ~ '^[a-fA-F0-9]{64}$'),

    name               text         NOT NULL,

    aliases            jsonb        NOT NULL DEFAULT '[]'::jsonb,
    file_name          jsonb        NOT NULL DEFAULT '[]'::jsonb,
    platform           jsonb        NOT NULL DEFAULT '[]'::jsonb,
    tags               jsonb        NOT NULL DEFAULT '[]'::jsonb,
    "references"       jsonb        NOT NULL DEFAULT '[]'::jsonb,
    behaviors          jsonb        NOT NULL DEFAULT '[]'::jsonb,
    related_hashes     jsonb        NOT NULL DEFAULT '[]'::jsonb,

    file_size          bigint
        CHECK (file_size IS NULL OR file_size >= 0),
    file_type          text,
    mime_type          text,

    threat_type        text,
    category           text,
    family             text,

    severity           varchar(16)
        CHECK (severity IS NULL OR severity IN ('low','medium','high','critical')),
    status             varchar(16)
        CHECK (status IS NULL OR status IN ('active','resolved','false_positive','expired')),

    confidence         integer
        CHECK (confidence IS NULL OR confidence BETWEEN 0 AND 100),

    first_seen         timestamptz,
    last_seen          timestamptz,

    source             jsonb,

    description        text,
    detection_ratio    varchar(16),
    recommended_action text,

    detections         jsonb        NOT NULL DEFAULT '[]'::jsonb,
    network_iocs       jsonb        NOT NULL DEFAULT '[]'::jsonb,
    file_iocs          jsonb        NOT NULL DEFAULT '[]'::jsonb,

    is_expired         boolean,
    ttl                bigint
        CHECK (ttl IS NULL OR ttl >= 0),

    reporter           text,

    version            integer      NOT NULL DEFAULT 1
        CHECK (version >= 1),

    created_at         timestamptz  NOT NULL DEFAULT now(),
    updated_at         timestamptz  NOT NULL DEFAULT now()
);

-- Обычные индексы под фильтры/поиск
CREATE INDEX idx_threats_name          ON threats (name);
CREATE INDEX idx_threats_threat_type   ON threats (threat_type);
CREATE INDEX idx_threats_severity      ON threats (severity);
CREATE INDEX idx_threats_status        ON threats (status);
CREATE INDEX idx_threats_first_seen    ON threats (first_seen);
CREATE INDEX idx_threats_last_seen     ON threats (last_seen);

-- Уникальность вторичных хешей (partial: NULL разрешён)
CREATE UNIQUE INDEX uq_threats_sha1   ON threats (sha1)   WHERE sha1   IS NOT NULL;
CREATE UNIQUE INDEX uq_threats_sha256 ON threats (sha256) WHERE sha256 IS NOT NULL;

-- GIN-индексы для фильтрации по JSONB-массивам (@>, ?)
CREATE INDEX gin_threats_tags     ON threats USING gin (tags);
CREATE INDEX gin_threats_aliases  ON threats USING gin (aliases);
CREATE INDEX gin_threats_platform ON threats USING gin (platform);