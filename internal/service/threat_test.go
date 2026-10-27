package service

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexeytsOY/information-exchange-service/internal/entity"
	apithreat "github.com/AlexeytsOY/information-exchange-service/internal/gen/threat"
	"github.com/AlexeytsOY/information-exchange-service/internal/repository"
)

type mockRepo struct {
	createFn func(ctx context.Context, t *entity.Threat) (*entity.Threat, error)
	getFn    func(ctx context.Context, md5 string) (*entity.Threat, error)
	updateFn func(ctx context.Context, t *entity.Threat) (*entity.Threat, error)
}

func (m *mockRepo) Create(ctx context.Context, t *entity.Threat) (*entity.Threat, error) {
	return m.createFn(ctx, t)
}
func (m *mockRepo) GetByMD5(ctx context.Context, md5 string) (*entity.Threat, error) {
	return m.getFn(ctx, md5)
}
func (m *mockRepo) Update(ctx context.Context, t *entity.Threat) (*entity.Threat, error) {
	return m.updateFn(ctx, t)
}

func TestNewThreatService_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	NewThreatService(nil)
}

func TestService_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockRepo{
			createFn: func(ctx context.Context, th *entity.Threat) (*entity.Threat, error) {
				if th.MD5 != "d41d8cd98f00b204e9800998ecf8427e" {
					t.Errorf("md5 not normalized: %q", th.MD5)
				}
				th.Version = 1
				return th, nil
			},
		}
		svc := NewThreatService(repo)
		got, err := svc.Create(context.Background(), &apithreat.ThreatCreateRequest{
			Md5:  "D41D8CD98F00B204E9800998ECF8427E",
			Name: "n",
		})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if got.Md5 != "d41d8cd98f00b204e9800998ecf8427e" {
			t.Errorf("got: %+v", got)
		}
	})

	t.Run("nil context", func(t *testing.T) {
		svc := NewThreatService(&mockRepo{})
		_, err := svc.Create(nil, &apithreat.ThreatCreateRequest{})
		if !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("err: %v", err)
		}
	})

	t.Run("nil request", func(t *testing.T) {
		svc := NewThreatService(&mockRepo{})
		_, err := svc.Create(context.Background(), nil)
		if !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("err: %v", err)
		}
	})

	t.Run("repo error propagated", func(t *testing.T) {
		repo := &mockRepo{createFn: func(context.Context, *entity.Threat) (*entity.Threat, error) {
			return nil, repository.ErrAlreadyExists
		}}
		svc := NewThreatService(repo)
		_, err := svc.Create(context.Background(), &apithreat.ThreatCreateRequest{Md5: "x", Name: "n"})
		if !errors.Is(err, repository.ErrAlreadyExists) {
			t.Fatalf("err: %v", err)
		}
	})
}

func TestService_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockRepo{
			getFn: func(ctx context.Context, md5 string) (*entity.Threat, error) {
				if md5 != "d41d8cd98f00b204e9800998ecf8427e" {
					t.Errorf("md5 not lowercased: %q", md5)
				}
				return &entity.Threat{MD5: md5, Name: "n"}, nil
			},
		}
		svc := NewThreatService(repo)
		got, err := svc.Get(context.Background(), "D41D8CD98F00B204E9800998ECF8427E")
		if err != nil || got == nil {
			t.Fatalf("err=%v got=%v", err, got)
		}
	})

	t.Run("nil context", func(t *testing.T) {
		svc := NewThreatService(&mockRepo{})
		_, err := svc.Get(nil, "x")
		if !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("err: %v", err)
		}
	})

	t.Run("empty md5", func(t *testing.T) {
		svc := NewThreatService(&mockRepo{})
		_, err := svc.Get(context.Background(), "")
		if !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("err: %v", err)
		}
	})

	t.Run("not found propagated", func(t *testing.T) {
		repo := &mockRepo{getFn: func(context.Context, string) (*entity.Threat, error) {
			return nil, repository.ErrNotFound
		}}
		svc := NewThreatService(repo)
		_, err := svc.Get(context.Background(), "x")
		if !errors.Is(err, repository.ErrNotFound) {
			t.Fatalf("err: %v", err)
		}
	})
}

func TestService_Update(t *testing.T) {
	t.Run("uses existing version when request has none", func(t *testing.T) {
		var gotVer int32
		repo := &mockRepo{
			getFn: func(_ context.Context, md5 string) (*entity.Threat, error) {
				return &entity.Threat{MD5: md5, Name: "old", Version: 7}, nil
			},
			updateFn: func(_ context.Context, th *entity.Threat) (*entity.Threat, error) {
				gotVer = th.Version
				th.Version = th.Version + 1
				return th, nil
			},
		}
		svc := NewThreatService(repo)
		_, err := svc.Update(context.Background(), "D41D8CD98F00B204E9800998ECF8427E",
			&apithreat.ThreatUpdateRequest{Name: "new"})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if gotVer != 7 {
			t.Errorf("expected version 7, got %d", gotVer)
		}
	})

	t.Run("uses request version when provided", func(t *testing.T) {
		var gotVer int32
		repo := &mockRepo{
			getFn: func(_ context.Context, md5 string) (*entity.Threat, error) {
				return &entity.Threat{MD5: md5, Name: "old", Version: 1}, nil
			},
			updateFn: func(_ context.Context, th *entity.Threat) (*entity.Threat, error) {
				gotVer = th.Version
				return th, nil
			},
		}
		svc := NewThreatService(repo)
		req := &apithreat.ThreatUpdateRequest{Name: "new", Version: int32Ptr(42)}
		_, err := svc.Update(context.Background(), "d41d8cd98f00b204e9800998ecf8427e", req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if gotVer != 42 {
			t.Errorf("expected 42, got %d", gotVer)
		}
	})

	t.Run("zero/negative request version is ignored", func(t *testing.T) {
		var gotVer int32
		repo := &mockRepo{
			getFn: func(_ context.Context, md5 string) (*entity.Threat, error) {
				return &entity.Threat{MD5: md5, Name: "old", Version: 5}, nil
			},
			updateFn: func(_ context.Context, th *entity.Threat) (*entity.Threat, error) {
				gotVer = th.Version
				return th, nil
			},
		}
		svc := NewThreatService(repo)
		req := &apithreat.ThreatUpdateRequest{Name: "new", Version: int32Ptr(0)}
		_, _ = svc.Update(context.Background(), "d41d8cd98f00b204e9800998ecf8427e", req)
		if gotVer != 5 {
			t.Errorf("expected 5, got %d", gotVer)
		}
	})

	t.Run("nil ctx", func(t *testing.T) {
		svc := NewThreatService(&mockRepo{})
		_, err := svc.Update(nil, "x", &apithreat.ThreatUpdateRequest{Name: "n"})
		if !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("err: %v", err)
		}
	})
	t.Run("empty md5", func(t *testing.T) {
		svc := NewThreatService(&mockRepo{})
		_, err := svc.Update(context.Background(), "", &apithreat.ThreatUpdateRequest{Name: "n"})
		if !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("err: %v", err)
		}
	})
	t.Run("nil req", func(t *testing.T) {
		svc := NewThreatService(&mockRepo{})
		_, err := svc.Update(context.Background(), "x", nil)
		if !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("err: %v", err)
		}
	})
	t.Run("not found", func(t *testing.T) {
		repo := &mockRepo{getFn: func(context.Context, string) (*entity.Threat, error) {
			return nil, repository.ErrNotFound
		}}
		svc := NewThreatService(repo)
		_, err := svc.Update(context.Background(), "x", &apithreat.ThreatUpdateRequest{Name: "n"})
		if !errors.Is(err, repository.ErrNotFound) {
			t.Fatalf("err: %v", err)
		}
	})
	t.Run("version conflict", func(t *testing.T) {
		repo := &mockRepo{
			getFn: func(_ context.Context, md5 string) (*entity.Threat, error) {
				return &entity.Threat{MD5: md5, Name: "old", Version: 1}, nil
			},
			updateFn: func(context.Context, *entity.Threat) (*entity.Threat, error) {
				return nil, repository.ErrVersionConflict
			},
		}
		svc := NewThreatService(repo)
		_, err := svc.Update(context.Background(), "d41d8cd98f00b204e9800998ecf8427e",
			&apithreat.ThreatUpdateRequest{Name: "n"})
		if !errors.Is(err, repository.ErrVersionConflict) {
			t.Fatalf("err: %v", err)
		}
	})
}
