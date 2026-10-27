package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/AlexeytsOY/information-exchange-service/internal/entity"
	"gorm.io/gorm"
)

var (
	ErrNotFound        = errors.New("repository: threat not found")
	ErrAlreadyExists   = errors.New("repository: threat already exists")
	ErrVersionConflict = errors.New("repository: version conflict")
)

type ThreatRepository interface {
	Create(ctx context.Context, t *entity.Threat) (*entity.Threat, error)
	GetByMD5(ctx context.Context, md5 string) (*entity.Threat, error)
	Update(ctx context.Context, t *entity.Threat) (*entity.Threat, error)
}

type threatRepository struct {
	db *gorm.DB
}

func NewThreatRepository(db *gorm.DB) ThreatRepository {
	return &threatRepository{db: db}
}

func (r *threatRepository) Create(ctx context.Context, t *entity.Threat) (*entity.Threat, error) {
	if t.Version == 0 {
		t.Version = 1
	}
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrAlreadyExists
		}
		return nil, fmt.Errorf("repository: create threat: %w", err)
	}
	return t, nil
}

func (r *threatRepository) GetByMD5(ctx context.Context, md5 string) (*entity.Threat, error) {
	var t entity.Threat
	if err := r.db.WithContext(ctx).Where("md5 = ?", md5).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get threat: %w", err)
	}
	return &t, nil
}

func (r *threatRepository) Update(ctx context.Context, t *entity.Threat) (*entity.Threat, error) {
	if t.Version <= 0 {
		return nil, fmt.Errorf("repository: update threat: invalid expected version %d", t.Version)
	}

	expectedVersion := t.Version

	updates := map[string]any{
		"sha1":               t.SHA1,
		"sha256":             t.SHA256,
		"name":               t.Name,
		"aliases":            t.Aliases,
		"file_name":          t.FileName,
		"platform":           t.Platform,
		"tags":               t.Tags,
		"references":         t.References,
		"behaviors":          t.Behaviors,
		"related_hashes":     t.RelatedHashes,
		"file_size":          t.FileSize,
		"file_type":          t.FileType,
		"mime_type":          t.MimeType,
		"threat_type":        t.ThreatType,
		"category":           t.Category,
		"family":             t.Family,
		"severity":           t.Severity,
		"status":             t.Status,
		"confidence":         t.Confidence,
		"first_seen":         t.FirstSeen,
		"last_seen":          t.LastSeen,
		"source":             t.Source,
		"description":        t.Description,
		"detection_ratio":    t.DetectionRatio,
		"recommended_action": t.RecommendedAction,
		"detections":         t.Detections,
		"network_iocs":       t.NetworkIOCs,
		"file_iocs":          t.FileIOCs,
		"is_expired":         t.IsExpired,
		"ttl":                t.TTL,
		"reporter":           t.Reporter,
		"version":            expectedVersion + 1,
	}

	res := r.db.WithContext(ctx).
		Model(&entity.Threat{}).
		Where("md5 = ? AND version = ?", t.MD5, expectedVersion).
		Updates(updates)

	if res.Error != nil {
		return nil, fmt.Errorf("repository: update threat: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return nil, ErrVersionConflict
	}

	// Перечитываем, чтобы отдать актуальные version/updated_at.
	return r.GetByMD5(ctx, t.MD5)
}
