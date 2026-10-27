package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	apithreat "github.com/AlexeytsOY/information-exchange-service/internal/gen/threat"
	"github.com/AlexeytsOY/information-exchange-service/internal/repository"
)

var ErrInvalidArgument = errors.New("service: invalid argument")

type ThreatService interface {
	Create(ctx context.Context, req *apithreat.ThreatCreateRequest) (*apithreat.Threat, error)
	Get(ctx context.Context, md5 string) (*apithreat.Threat, error)
	Update(ctx context.Context, md5 string, req *apithreat.ThreatUpdateRequest) (*apithreat.Threat, error)
}

type threatService struct {
	repo repository.ThreatRepository
}

func NewThreatService(repo repository.ThreatRepository) ThreatService {
	if repo == nil {
		panic("service: NewThreatService: nil repository")
	}
	return &threatService{repo: repo}
}

func (s *threatService) Create(ctx context.Context, req *apithreat.ThreatCreateRequest) (*apithreat.Threat, error) {
	if ctx == nil {
		return nil, fmt.Errorf("%w: nil context", ErrInvalidArgument)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: nil create request", ErrInvalidArgument)
	}

	created, err := s.repo.Create(ctx, fromCreateRequest(req))
	if err != nil {
		return nil, err // ErrAlreadyExists и прочее пробрасываем как есть
	}
	return toAPI(created), nil
}

func (s *threatService) Get(ctx context.Context, md5 string) (*apithreat.Threat, error) {
	if ctx == nil {
		return nil, fmt.Errorf("%w: nil context", ErrInvalidArgument)
	}
	if md5 == "" {
		return nil, fmt.Errorf("%w: empty md5", ErrInvalidArgument)
	}

	t, err := s.repo.GetByMD5(ctx, strings.ToLower(md5))
	if err != nil {
		return nil, err // ErrNotFound пробросится в handler → 404
	}
	return toAPI(t), nil
}

func (s *threatService) Update(ctx context.Context, md5 string, req *apithreat.ThreatUpdateRequest) (*apithreat.Threat, error) {
	if ctx == nil {
		return nil, fmt.Errorf("%w: nil context", ErrInvalidArgument)
	}
	if md5 == "" {
		return nil, fmt.Errorf("%w: empty md5", ErrInvalidArgument)
	}
	if req == nil {
		return nil, fmt.Errorf("%w: nil update request", ErrInvalidArgument)
	}

	md5 = strings.ToLower(md5)

	existing, err := s.repo.GetByMD5(ctx, md5)
	if err != nil {
		return nil, err
	}

	expectedVersion := existing.Version
	if req.Version != nil && *req.Version > 0 {
		expectedVersion = *req.Version
	}

	applyReplace(existing, req)
	existing.MD5 = md5
	existing.Version = expectedVersion

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return nil, err
	}
	return toAPI(updated), nil
}
