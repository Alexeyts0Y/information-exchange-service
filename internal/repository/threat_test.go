//go:build integration

package repository_test

import (
	"context"
	"os"
	"testing"

	"github.com/AlexeytsOY/information-exchange-service/internal/config"
	"github.com/AlexeytsOY/information-exchange-service/internal/database"
	"github.com/AlexeytsOY/information-exchange-service/internal/entity"
	"github.com/AlexeytsOY/information-exchange-service/internal/repository"
	"github.com/AlexeytsOY/information-exchange-service/internal/testutil"
)

var testDBConfig config.DB

func TestMain(m *testing.M) {
	cfg, cleanup, err := testutil.StartPostgres()
	if err != nil {
		panic("testutil: " + err.Error())
	}
	testDBConfig = cfg
	code := m.Run()
	cleanup()
	os.Exit(code)
}

const testMD5 = "00000000000000000000000000000001"

func setupRepo(t *testing.T) (repository.ThreatRepository, func()) {
	t.Helper()
	db, err := database.New(context.Background(), testDBConfig)
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec("TRUNCATE TABLE threats").Error; err != nil {
		t.Fatalf("truncate: %v", err)
	}
	cleanup := func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	return repository.NewThreatRepository(db), cleanup
}

func strPtr(s string) *string { return &s }

func TestRepo_CreateGet(t *testing.T) {
	repo, cleanup := setupRepo(t)
	defer cleanup()

	ctx := context.Background()
	sha := "da39a3ee5e6b4b0d3255bfef95601890afd80709"
	created, err := repo.Create(ctx, &entity.Threat{
		MD5:         testMD5,
		Name:        "Trojan",
		SHA1:        &sha,
		Aliases:     []string{"a"},
		NetworkIOCs: []entity.NetworkIOC{{Type: strPtr("domain"), Value: strPtr("evil.com")}},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.Version != 1 {
		t.Errorf("version=%d", created.Version)
	}

	got, err := repo.GetByMD5(ctx, testMD5)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Trojan" || len(got.NetworkIOCs) != 1 || got.NetworkIOCs[0].Value == nil {
		t.Errorf("got: %+v", got)
	}
}

func TestRepo_Create_Duplicate(t *testing.T) {
	repo, cleanup := setupRepo(t)
	defer cleanup()
	ctx := context.Background()
	if _, err := repo.Create(ctx, &entity.Threat{MD5: testMD5, Name: "a"}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := repo.Create(ctx, &entity.Threat{MD5: testMD5, Name: "b"})
	if err != repository.ErrAlreadyExists {
		t.Fatalf("err=%v", err)
	}
}

func TestRepo_Get_NotFound(t *testing.T) {
	repo, cleanup := setupRepo(t)
	defer cleanup()
	_, err := repo.GetByMD5(context.Background(), testMD5)
	if err != repository.ErrNotFound {
		t.Fatalf("err=%v", err)
	}
}

func TestRepo_Update_OptimisticLock(t *testing.T) {
	repo, cleanup := setupRepo(t)
	defer cleanup()
	ctx := context.Background()

	if _, err := repo.Create(ctx, &entity.Threat{MD5: testMD5, Name: "old"}); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByMD5(ctx, testMD5)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	got.Name = "new"
	updated, err := repo.Update(ctx, got)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Version != got.Version+1 {
		t.Errorf("version=%d want=%d", updated.Version, got.Version+1)
	}

	got.Name = "third"
	got.Version = 1 // stale
	if _, err := repo.Update(ctx, got); err != repository.ErrVersionConflict {
		t.Fatalf("expected ErrVersionConflict, got %v", err)
	}
}
