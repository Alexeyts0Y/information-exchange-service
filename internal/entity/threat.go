package entity

import "time"

type Threat struct {
	MD5 string `gorm:"primaryKey;column:md5;type:varchar(32)"`

	SHA1   *string `gorm:"column:sha1;type:varchar(40)"`
	SHA256 *string `gorm:"column:sha256;type:varchar(64)"`

	Name string `gorm:"column:name;type:text;not null"`

	Aliases       []string `gorm:"column:aliases;type:jsonb;serializer:json"`
	FileName      []string `gorm:"column:file_name;type:jsonb;serializer:json"`
	Platform      []string `gorm:"column:platform;type:jsonb;serializer:json"`
	Tags          []string `gorm:"column:tags;type:jsonb;serializer:json"`
	References    []string `gorm:"column:references;type:jsonb;serializer:json"`
	Behaviors     []string `gorm:"column:behaviors;type:jsonb;serializer:json"`
	RelatedHashes []string `gorm:"column:related_hashes;type:jsonb;serializer:json"`

	FileSize *int64  `gorm:"column:file_size"`
	FileType *string `gorm:"column:file_type;type:text"`
	MimeType *string `gorm:"column:mime_type;type:text"`

	ThreatType *string `gorm:"column:threat_type;type:text"`
	Category   *string `gorm:"column:category;type:text"`
	Family     *string `gorm:"column:family;type:text"`

	Severity *string `gorm:"column:severity;type:varchar(16)"`
	Status   *string `gorm:"column:status;type:varchar(16)"`

	Confidence *int `gorm:"column:confidence"`

	FirstSeen *time.Time `gorm:"column:first_seen"`
	LastSeen  *time.Time `gorm:"column:last_seen"`

	// Вложенные value objects.
	Source *ThreatSource `gorm:"column:source;type:jsonb;serializer:json"`

	Description       *string `gorm:"column:description;type:text"`
	DetectionRatio    *string `gorm:"column:detection_ratio;type:varchar(16)"`
	RecommendedAction *string `gorm:"column:recommended_action;type:text"`

	Detections  []Detection  `gorm:"column:detections;type:jsonb;serializer:json"`
	NetworkIOCs []NetworkIOC `gorm:"column:network_iocs;type:jsonb;serializer:json"`
	FileIOCs    []FileIOC    `gorm:"column:file_iocs;type:jsonb;serializer:json"`

	IsExpired *bool  `gorm:"column:is_expired"`
	TTL       *int64 `gorm:"column:ttl"`

	Reporter *string `gorm:"column:reporter;type:text"`

	Version int32 `gorm:"column:version;not null"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName задаёт явное имя таблицы.
func (Threat) TableName() string { return "threats" }

// ThreatSource — источник данных об угрозе (value object).
type ThreatSource struct {
	Name *string `json:"name,omitempty"`
	Type *string `json:"type,omitempty"`
	URL  *string `json:"url,omitempty"`
}

// Detection — детект антивирусного движка (value object).
type Detection struct {
	Engine    *string `json:"engine,omitempty"`
	Version   *string `json:"version,omitempty"`
	Signature *string `json:"signature,omitempty"`
}

// NetworkIOC — сетевой индикатор компрометации (value object).
type NetworkIOC struct {
	Type        *string `json:"type,omitempty"`
	Value       *string `json:"value,omitempty"`
	Description *string `json:"description,omitempty"`
}

// FileIOC — файловый индикатор компрометации (value object).
type FileIOC struct {
	Type        *string `json:"type,omitempty"`
	Value       *string `json:"value,omitempty"`
	Description *string `json:"description,omitempty"`
}
