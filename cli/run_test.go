package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunStdin(t *testing.T) {
	cases := []struct{
		name string
		args []string
		input string
		code int
		fragment string
	}{
		{"safe", []string{"--stdin"}, `{}`, ExitOK, "Not found problems"},
		{"unsafe", []string{"--stdin"}, `{"debug":true}`, ExitFindings, "LOW $.debug"},
		{"silent", []string{"--stdin", "--silent"}, `{"debug":true}`, ExitOK, "LOW $.debug"},
		{"short silent", []string{"-s", "--stdin"}, `{"debug":true}`, ExitOK, "LOW $.debug"},
		{"YAML", []string{"--stdin"}, "log:\n  level: debug", ExitFindings, "$.log.level"},
		{"flow YAML", []string{"--stdin", "--format=yaml"}, "{debug: true}", ExitFindings, "$.debug"},
		{"help", []string{"--help"}, "", ExitOK, "Usage"},
		{"invalid", []string{"--stdin"}, `{"password":"test_password"`, ExitError, ""},
		{"invalid silent", []string{"--stdin", "--silent"}, "log: [", ExitError, ""},
		{"missing input", nil, "", ExitError, ""},
		{"conflicting input", []string{"--stdin", "file.yaml"}, "{}", ExitError, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			code := Run(tc.args, strings.NewReader(tc.input), &out, &errOut)
			if code != tc.code {
				t.Fatalf("code=%d, expected %d, stderr=%s", code, tc.code, errOut.String())
			}

			if !strings.Contains(out.String(), tc.fragment) {
				t.Fatalf("no %q in %q", tc.fragment, out.String())
			}

			if tc.code == ExitError && (errOut.Len() == 0 || out.Len() != 0) {
				t.Fatal("error should go only to stderr")
			}

			if tc.code != ExitError && errOut.Len() != 0 {
				t.Fatalf("unexpected stderr: %s", errOut.String())
			}
		})
	}
}

func TestRunFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.json")
	
	err := os.WriteFile(path, []byte(`{"password":"test_password"}`), 0600);
	if err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	code := Run([]string{path, "--silent"}, strings.NewReader("invalid stdin must not be read"), &out, &errOut)
	if code != ExitOK || !strings.Contains(out.String(), "HIGH $.password") {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func TestRunMissingFile(t *testing.T) {
	var out, errOut bytes.Buffer
	path := filepath.Join(t.TempDir(), "missing.json")
	
	code := Run([]string{path, "-s"}, strings.NewReader(""), &out, &errOut)
	if code != ExitError {
		t.Fatalf("code=%d", code)
	}
}