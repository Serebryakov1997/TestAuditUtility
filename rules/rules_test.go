package rules

import (
	"strings"
	"testing"

	"github.com/serebryakov1997/utility/audit"
	"github.com/serebryakov1997/utility/config"
)

func analyze(t *testing.T, input, format string) []audit.Finding {
	t.Helper()
	parseConfig, err := config.Parse(strings.NewReader(input), format)
	if err != nil {
		t.Fatal(err)
	}

	result, err := audit.New(RegisteredRules()...).Analyze(parseConfig)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestRules(t *testing.T) {
	cases := []struct {
		name string
		input string
		id string
		path string
		severity string
	}{
		{"debug bool", `{"debug":true}`, "debug-enabled", "$.debug", audit.Low},
		{"logging", `{"log":{"level":"debug"}}`, "debug-enabled", "$.log.level", audit.Low},
		{"password", `{"password": "test_password"}`, "plaintext-password", "$.password", audit.High},
		{"numeric password", `{"password":123456}`, "plaintext-password", "$.password", audit.High},
		{"bind", `{"host":"0.0.0.0"}`, "all-interfaces", "$.host", audit.Medium},
		{"bind port", `{"listen":"0.0.0.0:8080"}`, "all-interfaces", "$.listen", audit.Medium},
		{"bind ipv6", `{"listen":"[::]:8080"}`, "all-interfaces", "$.listen", audit.Medium},
		{"tls disabled", `{"tls":{"enabled":false}}`, "tls-unsafe", "$.tls.enabled", audit.High},
		{"ssl disabled", `{"ssl":false}`, "tls-unsafe", "$.ssl", audit.High},
		{"MD5", `{"storage":{"digest-algorithm":"MD5"}}`, "weak-algorithm", `$.storage["digest-algorithm"]`, audit.High},
		{"file permissions", `{"permissions":"0777"}`, "broad-permissions", "$.permissions", audit.Medium},
		{"numeric permissions", `{"file_mode":511}`, "broad-permissions", "$.file_mode", audit.Medium},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := analyze(t, tc.input, config.JSON)
			if len(got) != 1 {
				t.Fatalf("expected 1, got %d: %+v", len(got), got)
			}

			f := got[0]
			if f.RuleID != tc.id || f.Path != tc.path || f.Severity != tc.severity {
				t.Fatalf("unexpected find: %+v", f)
			}

			if f.Message == "" || f.Recommendation == "" {
				t.Fatal("no explanations or recommendations")
			}
		})
	}
}

func TestNoFindings(t *testing.T) {
	inputs := []string{
		`{}`,
		`{"log":{"level":"info"},"debug":false}`,
		`{"game":{"level":"debug"}}`,
		`{"password":"${DB_PASSWORD}"}`,
		`{"password":""}`,
		`{"storage":{"digest-algorithm":"SHA256"}}`,
		`{"host":"127.0.0.1","listen":"127.0.0.1:8080"}`,
		`{"tls":{"enabled":true, "verify":true}}`,
		`{"file_mode":"0644"}`,
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			if got := analyze(t, input, config.JSON); len(got) != 0 {
				t.Fatalf("false alarm: %+v", got)
			}
		})
	}
}