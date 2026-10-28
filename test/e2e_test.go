//go:build integration

package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/AlexeytsOY/information-exchange-service/internal/config"
	"github.com/AlexeytsOY/information-exchange-service/internal/database"
	threatapi "github.com/AlexeytsOY/information-exchange-service/internal/gen/threat"
	"github.com/AlexeytsOY/information-exchange-service/internal/handler"
	"github.com/AlexeytsOY/information-exchange-service/internal/repository"
	"github.com/AlexeytsOY/information-exchange-service/internal/service"
	"github.com/AlexeytsOY/information-exchange-service/internal/testutil"
	"github.com/gin-gonic/gin"
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

func setupServer(t *testing.T) (*httptest.Server, func()) {
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

	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := repository.NewThreatRepository(db)
	svc := service.NewThreatService(repo)
	h := handler.NewThreatHandler(svc)
	r.POST("/info", h.CreateThreat)
	r.GET("/info/:md5", func(c *gin.Context) { h.GetInfoByMd5(c, c.Param("md5")) })
	r.PUT("/info/:md5", func(c *gin.Context) { h.UpdateThreat(c, c.Param("md5")) })

	srv := httptest.NewServer(r)
	cleanup := func() {
		srv.Close()
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	return srv, cleanup
}

func TestE2E_CreateGetUpdateFlow(t *testing.T) {
	srv, cleanup := setupServer(t)
	defer cleanup()

	md5 := "00000000000000000000000000000002"
	createBody := map[string]any{
		"md5":     md5,
		"name":    "Trojan.Test",
		"aliases": []string{"TT"},
	}

	// CREATE
	b, _ := json.Marshal(createBody)
	resp, err := http.Post(srv.URL+"/info", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status=%d", resp.StatusCode)
	}
	var created threatapi.Threat
	_ = json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()
	if created.Md5 != md5 || created.Version == nil || *created.Version != 1 {
		t.Fatalf("created=%+v", created)
	}

	// GET
	getResp, err := http.Get(fmt.Sprintf("%s/info/%s", srv.URL, md5))
	if err != nil {
		t.Fatal(err)
	}
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get status=%d", getResp.StatusCode)
	}
	getResp.Body.Close()

	// CREATE duplicate -> 409
	resp, _ = http.Post(srv.URL+"/info", "application/json", bytes.NewReader(b))
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("dup create status=%d", resp.StatusCode)
	}
	resp.Body.Close()

	// UPDATE
	updBody, _ := json.Marshal(map[string]any{"name": "NewName", "version": 1})
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/info/%s", srv.URL, md5), bytes.NewReader(updBody))
	req.Header.Set("Content-Type", "application/json")
	updResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if updResp.StatusCode != http.StatusOK {
		t.Fatalf("update status=%d", updResp.StatusCode)
	}
	var updated threatapi.Threat
	_ = json.NewDecoder(updResp.Body).Decode(&updated)
	updResp.Body.Close()
	if updated.Name != "NewName" || updated.Version == nil || *updated.Version != 2 {
		t.Errorf("updated=%+v", updated)
	}

	// UPDATE stale version -> 409
	req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("%s/info/%s", srv.URL, md5), bytes.NewReader(updBody))
	req.Header.Set("Content-Type", "application/json")
	conflictResp, _ := http.DefaultClient.Do(req)
	if conflictResp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", conflictResp.StatusCode)
	}
	conflictResp.Body.Close()

	// GET not found -> 404
	nfResp, _ := http.Get(fmt.Sprintf("%s/info/%s", srv.URL, "ffffffffffffffffffffffffffffffff"))
	if nfResp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", nfResp.StatusCode)
	}
	nfResp.Body.Close()

	// GET invalid md5 -> 400
	badResp, _ := http.Get(srv.URL + "/info/zzz")
	if badResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", badResp.StatusCode)
	}
	badResp.Body.Close()
}
