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

	writeFile(t, filepath.Join(root, "app.json"), `{"debug":true}`)
	writeFile(t, filepath.Join(nestedDir, "server.yaml"), "debug: true\n")
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

	if analyzer.calls != 2 {
		t.Fatalf("Analyze() calls = %d, want 2", analyzer.calls)
	}

	if reports[0].Path != "app.json" {
		t.Fatalf("reports[0].Path = %q, want app.json", reports[0].Path)
	}

	if reports[1].Path != "nested/server.yaml" {
		t.Fatalf(
			"reports[1].Path = %q, want nested/server.yaml",
			reports[1].Path,
		)
	}
}

func TestScannerReturnsFileErrorAndContinues(t *testing.T) {
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "valid.json"), `{"debug":true}`)
	writeFile(t, filepath.Join(root, "broken.yaml"), "debug: [")

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
