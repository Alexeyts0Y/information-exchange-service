package validation

import (
	"testing"

	apithreat "github.com/AlexeytsOY/information-exchange-service/internal/gen/threat"
)

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }

const validMD5 = "d41d8cd98f00b204e9800998ecf8427e"
const validSHA1 = "da39a3ee5e6b4b0d3255bfef95601890afd80709"
const validSHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

func TestMD5(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid lowercase", validMD5, false},
		{"valid uppercase", "D41D8CD98F00B204E9800998ECF8427E", false},
		{"empty", "", true},
		{"too short", "d41d8cd98f00b204e9800998ecf8427", true},
		{"too long", "d41d8cd98f00b204e9800998ecf8427ee", true},
		{"non-hex", "z41d8cd98f00b204e9800998ecf8427e", true},
		{"with space", " d41d8cd98f00b204e9800998ecf8427e", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MD5(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("MD5(%q) err=%v, wantErr=%v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func validCreateRequest() *apithreat.ThreatCreateRequest {
	return &apithreat.ThreatCreateRequest{Md5: validMD5, Name: "Test"}
}

func TestCreateRequest(t *testing.T) {
	t.Run("valid minimal", func(t *testing.T) {
		if err := CreateRequest(validCreateRequest()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("valid full", func(t *testing.T) {
		req := validCreateRequest()
		req.Sha1 = strPtr(validSHA1)
		req.Sha256 = strPtr(validSHA256)
		req.FileSize = int64Ptr(100)
		req.Confidence = intPtr(50)
		req.Ttl = int64Ptr(3600)
		req.Version = int32Ptr(1)
		refs := []string{"https://example.com/report"}
		req.References = &refs
		sev := apithreat.ThreatCreateRequestSeverityHigh
		req.Severity = &sev
		stat := apithreat.ThreatCreateRequestStatusActive
		req.Status = &stat
		srcType := apithreat.Sandbox
		srcURL := "https://example.com/sandbox"
		req.Source = &apithreat.ThreatSource{Type: &srcType, Url: &srcURL}
		if err := CreateRequest(req); err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("invalid md5", func(t *testing.T) {
		req := validCreateRequest()
		req.Md5 = "short"
		if err := CreateRequest(req); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		req := validCreateRequest()
		req.Name = ""
		if err := CreateRequest(req); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid sha1", func(t *testing.T) {
		req := validCreateRequest()
		req.Sha1 = strPtr("bad")
		if err := CreateRequest(req); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid sha256", func(t *testing.T) {
		req := validCreateRequest()
		req.Sha256 = strPtr("bad")
		if err := CreateRequest(req); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("negative file_size", func(t *testing.T) {
		req := validCreateRequest()
		req.FileSize = int64Ptr(-1)
		if err := CreateRequest(req); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("confidence out of range", func(t *testing.T) {
		for _, v := range []int{-1, 101} {
			req := validCreateRequest()
			req.Confidence = intPtr(v)
			if err := CreateRequest(req); err == nil {
				t.Errorf("confidence=%d: expected error", v)
			}
		}
	})

	t.Run("negative ttl", func(t *testing.T) {
		req := validCreateRequest()
		req.Ttl = int64Ptr(-1)
		if err := CreateRequest(req); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("zero version", func(t *testing.T) {
		req := validCreateRequest()
		req.Version = int32Ptr(0)
		if err := CreateRequest(req); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid reference uri", func(t *testing.T) {
		req := validCreateRequest()
		refs := []string{"no-scheme"}
		req.References = &refs
		if err := CreateRequest(req); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid severity", func(t *testing.T) {
		req := validCreateRequest()
		s := apithreat.ThreatCreateRequestSeverity("bad")
		req.Severity = &s
		if err := CreateRequest(req); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		req := validCreateRequest()
		s := apithreat.ThreatCreateRequestStatus("bad")
		req.Status = &s
		if err := CreateRequest(req); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid source type", func(t *testing.T) {
		req := validCreateRequest()
		bad := apithreat.ThreatSourceType("bad")
		req.Source = &apithreat.ThreatSource{Type: &bad}
		if err := CreateRequest(req); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid source url", func(t *testing.T) {
		req := validCreateRequest()
		bad := "no-scheme"
		req.Source = &apithreat.ThreatSource{Url: &bad}
		if err := CreateRequest(req); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid network ioc type", func(t *testing.T) {
		req := validCreateRequest()
		bad := apithreat.NetworkIOCType("bad")
		iocs := []apithreat.NetworkIOC{{Type: &bad}}
		req.NetworkIocs = &iocs
		if err := CreateRequest(req); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid file ioc type", func(t *testing.T) {
		req := validCreateRequest()
		bad := apithreat.FileIOCType("bad")
		iocs := []apithreat.FileIOC{{Type: &bad}}
		req.FileIocs = &iocs
		if err := CreateRequest(req); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestUpdateRequest(t *testing.T) {
	t.Run("valid minimal", func(t *testing.T) {
		if err := UpdateRequest(&apithreat.ThreatUpdateRequest{Name: "X"}); err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})
	t.Run("empty name", func(t *testing.T) {
		if err := UpdateRequest(&apithreat.ThreatUpdateRequest{}); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("invalid sha1", func(t *testing.T) {
		req := &apithreat.ThreatUpdateRequest{Name: "X", Sha1: strPtr("bad")}
		if err := UpdateRequest(req); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("invalid version", func(t *testing.T) {
		req := &apithreat.ThreatUpdateRequest{Name: "X", Version: int32Ptr(-1)}
		if err := UpdateRequest(req); err == nil {
			t.Fatal("expected error")
		}
	})
}
