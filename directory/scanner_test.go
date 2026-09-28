package directory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/serebryakov1997/utility/audit"
)

type mockAnalyzer struct {
	calls int
}

func findFinding(
	findings []audit.Finding,
	ruleID string,
) (audit.Finding, bool) {
	for _, finding := range findings {
		if finding.RuleID == ruleID {
			return finding, true
		}
	}

	return audit.Finding{}, false
}

func (f *mockAnalyzer) Analyze(config any) ([]audit.Finding, error) {
	f.calls++

	return []audit.Finding{
		{
			RuleID:   "test-rule",
			Severity: audit.Low,
			Path:     "$.test",
		},
	}, nil
}

func TestScannerAnalyzeRecursiveDir(t *testing.T) {
	root := t.TempDir()

	nestedDir := filepath.Join(root, "nested")
	if err := os.Mkdir(nestedDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	appPath := filepath.Join(root, "app.json")
	writeFile(t, appPath, `{"debug":true}`)

	serverPath := filepath.Join(nestedDir, "server.yaml")
	writeFile(t, filepath.Join(nestedDir, "server.yaml"), "debug: true\n")

	for _, path := range []string{appPath, serverPath} {
		if err := os.Chmod(path, 0o666); err != nil {
			t.Fatalf("chmod %q: %v", path, err)
		}
	}

	writeFile(t, filepath.Join(root, "README.txt"), "ignored")

	analyzer := &mockAnalyzer{}
	scanner, err := New(analyzer)
	if err != nil {
		t.Fatalf("scanner init: %v", err)
	}

	reports, err := scanner.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error: %v", err)
	}

	if len(reports) != 2 {
		t.Fatalf("reports count = %d, want 2", len(reports))
	}

	byPath := make(map[string]FileReport, len(reports))

	for _, report := range reports {
		byPath[report.Path] = report
	}

	cases := []struct {
		name string
		path string
	}{
		{
			name: "root json",
			path: "app.json",
		},
		{
			name: "nested yaml",
			path: "nested/server.yaml",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			report, ok := byPath[tt.path]
			if !ok {
				t.Fatalf("report for %q is missing", tt.path)
			}

			if report.Error != "" {
				t.Fatalf(
					"report %q has unexpected error: %q",
					tt.path,
					report.Error,
				)
			}

			finding, ok := findFinding(
				report.Findings,
				"file-permissions",
			)

			if !ok {
				t.Fatalf(
					"report %q does not contain file-permissions finding",
					tt.path,
				)
			}

			if finding.Path != "@file.permissions" {
				t.Errorf(
					"finding path = %q, want %q",
					finding.Path,
					"@file.permissions",
				)
			}

			if finding.Severity != audit.High {
				t.Errorf(
					"finding severity = %q, want %q",
					finding.Severity,
					audit.High,
				)
			}
		})
	}

	if analyzer.calls != 2 {
		t.Fatalf("Analyze() calls = %d, want 2", analyzer.calls)
	}
}

func TestScannerReturnsFileErrorAndContinues(t *testing.T) {
	root := t.TempDir()

	validPath := filepath.Join(root, "valid.json")
	writeFile(t, validPath, `{"debug":true}`)

	invalidPath := filepath.Join(root, "broken.yaml")
	writeFile(t, invalidPath, "debug: [")

	if err := os.Chmod(invalidPath, 0o666); err != nil {
		t.Fatalf("Chmod() %q: %v", invalidPath, err)
	}

	analyzer := &mockAnalyzer{}
	scanner, err := New(analyzer)
	if err != nil {
		t.Fatalf("scanner init error: %v", err)
	}

	reports, err := scanner.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error: %v", err)
	}

	if len(reports) != 2 {
		t.Fatalf("reports count = %d, want 2", len(reports))
	}

	byPath := make(map[string]FileReport, len(reports))
	for _, report := range reports {
		byPath[report.Path] = report
	}

	valid, ok := byPath["valid.json"]
	if !ok {
		t.Fatal("report for valid.json is missing")
	}

	if valid.Error != "" {
		t.Fatalf("valid file has error: %q", valid.Error)
	}

	broken, ok := byPath["broken.yaml"]
	if !ok {
		t.Fatal("report for broken.yaml is missing")
	}

	if broken.Error == "" {
		t.Fatal("broken file must contain an error")
	}

	finding, ok := findFinding(
		broken.Findings,
		"file-permissions",
	)

	if !ok {
		t.Fatal("file-permissions finding not found")
	}

	if finding.Path != "@file.permissions" {
		t.Errorf(
			"finding path = %q, want %q",
			finding.Path,
			"@file.permissions",
		)
	}

	if finding.Severity != audit.High {
		t.Errorf(
			"finding severity = %q, want %q",
			finding.Severity,
			audit.High,
		)
	}

	if analyzer.calls != 1 {
		t.Fatalf("Analyze() calls = %d, want 1", analyzer.calls)
	}
}

func TestScannerRejectsFileAsRoot(t *testing.T) {
	file := filepath.Join(t.TempDir(), "config.json")
	writeFile(t, file, `{}`)

	scanner, err := New(&mockAnalyzer{})
	if err != nil {
		t.Fatalf("scanner init error: %v", err)
	}

	_, err = scanner.Analyze(file)
	if err == nil {
		t.Fatal("expected error for file path")
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
