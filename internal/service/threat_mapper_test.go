package service

import (
	"testing"
	"time"

	"github.com/AlexeytsOY/information-exchange-service/internal/entity"
	apithreat "github.com/AlexeytsOY/information-exchange-service/internal/gen/threat"
)

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }

func TestFromCreateRequest(t *testing.T) {
	sha1 := "DA39A3EE5E6B4B0D3255BFEF95601890AFD80709"
	sha256 := "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855"
	aliases := []string{"a1", "a2"}
	platform := []string{"Windows"}
	sev := apithreat.ThreatCreateRequestSeverityHigh
	stat := apithreat.ThreatCreateRequestStatusActive
	srcType := apithreat.Sandbox
	srcURL := "https://example.com"
	src := &apithreat.ThreatSource{Name: strPtr("sb"), Type: &srcType, Url: &srcURL}
	netType := apithreat.Domain
	nets := []apithreat.NetworkIOC{{Type: &netType, Value: strPtr("evil.com")}}
	fileType := apithreat.Path
	files := []apithreat.FileIOC{{Type: &fileType, Value: strPtr("C:\\x")}}
	det := []apithreat.Detection{{Engine: strPtr("K")}}

	req := &apithreat.ThreatCreateRequest{
		Md5:         "D41D8CD98F00B204E9800998ECF8427E",
		Name:        "T",
		Sha1:        &sha1,
		Sha256:      &sha256,
		Aliases:     &aliases,
		Platform:    &platform,
		Severity:    &sev,
		Status:      &stat,
		Source:      src,
		NetworkIocs: &nets,
		FileIocs:    &files,
		Detections:  &det,
	}
	got := fromCreateRequest(req)

	if got.MD5 != "d41d8cd98f00b204e9800998ecf8427e" {
		t.Errorf("MD5 not lowercased: %q", got.MD5)
	}
	if got.SHA1 == nil || *got.SHA1 != "da39a3ee5e6b4b0d3255bfef95601890afd80709" {
		t.Errorf("SHA1 not lowercased: %v", got.SHA1)
	}
	if got.SHA256 == nil || *got.SHA256 != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Errorf("SHA256 not lowercased: %v", got.SHA256)
	}
	if got.Name != "T" || len(got.Aliases) != 2 || len(got.Platform) != 1 {
		t.Errorf("fields not mapped: %+v", got)
	}
	if got.Severity == nil || *got.Severity != "high" {
		t.Errorf("severity: %v", got.Severity)
	}
	if got.Source == nil || got.Source.Type == nil || *got.Source.Type != "sandbox" {
		t.Errorf("source: %+v", got.Source)
	}
	if len(got.NetworkIOCs) != 1 || got.NetworkIOCs[0].Type == nil || *got.NetworkIOCs[0].Type != "domain" {
		t.Errorf("network iocs: %+v", got.NetworkIOCs)
	}
	if len(got.Detections) != 1 || got.Detections[0].Engine == nil || *got.Detections[0].Engine != "K" {
		t.Errorf("detections: %+v", got.Detections)
	}
}

func TestFromCreateRequest_Nils(t *testing.T) {
	got := fromCreateRequest(&apithreat.ThreatCreateRequest{Md5: "D41D8CD98F00B204E9800998ECF8427E", Name: "n"})
	if got.SHA1 != nil || got.Aliases != nil || got.Source != nil || got.Detections != nil {
		t.Errorf("nil inputs should map to nil/empty: %+v", got)
	}
	if got.MD5 != "d41d8cd98f00b204e9800998ecf8427e" {
		t.Errorf("md5 lowercase: %q", got.MD5)
	}
}

func TestApplyReplace(t *testing.T) {
	existing := &entity.Threat{
		MD5:     "old",
		Name:    "old",
		SHA1:    strPtr("oldsha"),
		Aliases: []string{"x"},
		Version: 5,
	}
	applyReplace(existing, &apithreat.ThreatUpdateRequest{Name: "new"})

	if existing.Name != "new" {
		t.Errorf("Name = %q", existing.Name)
	}
	if existing.SHA1 != nil {
		t.Errorf("SHA1 should be cleared, got %v", *existing.SHA1)
	}
	if existing.Aliases != nil {
		t.Errorf("Aliases should be cleared, got %v", existing.Aliases)
	}
	if existing.MD5 != "old" || existing.Version != 5 {
		t.Errorf("server fields must not change: md5=%q version=%d", existing.MD5, existing.Version)
	}
}

func TestToAPI(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	sha := "abc"
	sev := "high"
	srcType := "sandbox"
	netType := "domain"
	in := &entity.Threat{
		MD5:         "m",
		Name:        "n",
		SHA1:        &sha,
		Severity:    &sev,
		Source:      &entity.ThreatSource{Type: &srcType},
		NetworkIOCs: []entity.NetworkIOC{{Type: &netType}},
		Detections:  []entity.Detection{{Engine: strPtr("K")}},
		Aliases:     []string{"a"},
		Version:     3,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	out := toAPI(in)
	if out == nil {
		t.Fatal("nil output")
	}
	if out.Md5 != "m" || out.Name != "n" {
		t.Errorf("basic fields: %+v", out)
	}
	if out.Sha1 == nil || *out.Sha1 != "abc" {
		t.Errorf("sha1: %v", out.Sha1)
	}
	if out.Severity == nil || *out.Severity != apithreat.ThreatSeverityHigh {
		t.Errorf("severity: %v", out.Severity)
	}
	if out.Source == nil || out.Source.Type == nil || *out.Source.Type != apithreat.Sandbox {
		t.Errorf("source: %+v", out.Source)
	}
	if out.Version == nil || *out.Version != 3 {
		t.Errorf("version: %v", out.Version)
	}
	if out.CreatedAt == nil || !out.CreatedAt.Equal(now) {
		t.Errorf("created_at: %v", out.CreatedAt)
	}
	if out.Aliases == nil || len(*out.Aliases) != 1 {
		t.Errorf("aliases: %v", out.Aliases)
	}
}

func TestToAPI_Nil(t *testing.T) {
	if toAPI(nil) != nil {
		t.Fatal("toAPI(nil) should return nil")
	}
}

func TestHelpers(t *testing.T) {
	// enumPtrString / stringPtrEnum
	sev := apithreat.ThreatSeverityHigh
	got := enumPtrString(&sev)
	if got == nil || *got != "high" {
		t.Errorf("enumPtrString: %v", got)
	}
	back := stringPtrEnum[apithreat.ThreatSeverity](strPtr("low"))
	if back == nil || *back != apithreat.ThreatSeverityLow {
		t.Errorf("stringPtrEnum: %v", back)
	}
	if enumPtrString[apithreat.ThreatSeverity](nil) != nil {
		t.Error("enumPtrString(nil) should be nil")
	}
	if stringPtrEnum[apithreat.ThreatSeverity](nil) != nil {
		t.Error("stringPtrEnum(nil) should be nil")
	}

	// lowerString*
	if s := lowerString("ABC"); s == nil || *s != "abc" {
		t.Errorf("lowerString: %v", s)
	}
	if lowerStringPtr(nil) != nil {
		t.Error("lowerStringPtr(nil) should be nil")
	}
	if s := lowerStringPtr(strPtr("XYZ")); s == nil || *s != "xyz" {
		t.Errorf("lowerStringPtr: %v", s)
	}

	// slice ptr
	in := []string{"a", "b"}
	p := ptrSlice(in)
	if p == nil || len(*p) != 2 || (*p)[0] != "a" {
		t.Errorf("ptrSlice: %v", p)
	}
	got2 := sliceFromPtr(p)
	if len(got2) != 2 || got2[1] != "b" {
		t.Errorf("sliceFromPtr: %v", got2)
	}
	if ptrSlice(nil) != nil || sliceFromPtr(nil) != nil {
		t.Error("nil slices")
	}
}
