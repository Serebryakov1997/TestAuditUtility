package filecheck

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/serebryakov1997/utility/audit"
)

func TestPermissions(t *testing.T) {
	cases := []struct {
		name         string
		mode         os.FileMode
		wantCount    int
		wantSeverity string
	}{
		{
			name:      "owner only",
			mode:      0o600,
			wantCount: 0,
		},
		{
			name:      "others can read",
			mode:      0o644,
			wantCount: 0,
		},
		{
			name:         "group can write",
			mode:         0o660,
			wantCount:    1,
			wantSeverity: audit.Medium,
		},
		{
			name:         "others can write",
			mode:         0o606,
			wantCount:    1,
			wantSeverity: audit.High,
		},
		{
			name:         "group and others can write",
			mode:         0o666,
			wantCount:    1,
			wantSeverity: audit.High,
		},
		{
			name:         "all permissions",
			mode:         0o777,
			wantCount:    1,
			wantSeverity: audit.High,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")

			if err := os.WriteFile(path, []byte(`{}`), 0o600); err != nil {
				t.Fatalf("WriteFile() error: %v", err)
			}

			if err := os.Chmod(path, tt.mode); err != nil {
				t.Fatalf("Chmod() error: %v", err)
			}

			findings, err := Permissions(path)
			if err != nil {
				t.Fatalf("Permissions() error: %v", err)
			}

			if len(findings) != tt.wantCount {
				t.Fatalf(
					"findings count = %d, want %d",
					len(findings),
					tt.wantCount,
				)
			}

			if tt.wantCount == 0 {
				return
			}

			finding := findings[0]

			if finding.RuleID != "file-permissions" {
				t.Errorf(
					"RuleID = %q, want file-permissions",
					finding.RuleID,
				)
			}

			if finding.Path != "@file.permissions" {
				t.Errorf(
					"Path = %q, want @file.permissions",
					finding.Path,
				)
			}
		})
	}
}

func TestPermissions_FileNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")

	_, err := Permissions(path)

	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("error = %v, want os.ErrNotExist", err)
	}
}

func TestPermissions_RejectDirectory(t *testing.T) {
	_, err := Permissions(t.TempDir())
	if err == nil {
		t.Fatal("expected error for directory")
	}
}
