package config

import (
	"strings"
	"testing"
)

func largeJSON(size int) string {
	const prefix = `{"payload":"`
	const suffix = `"}`

	return prefix +
		strings.Repeat("a", size-len(prefix)-len(suffix)) +
		suffix
}

func largeYAML(size int) string {
	const prefix = "payload: \""
	const suffix = "\"\n"

	return prefix +
		strings.Repeat("a", size-len(prefix)-len(suffix)) +
		suffix
}

func TestParseLargeConfig(t *testing.T) {
	payload := strings.Repeat("a", 9<<20)

	cases := []struct {
		name  string
		input string
		yaml  bool
	}{
		{
			name:  "json",
			input: `{"payload":"` + payload + `"}`,
		},
		{
			name:  "yaml",
			input: "payload: \"" + payload + "\"\n",
			yaml:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			format := JSON
			if tc.yaml {
				format = YAML
			}

			parseConfig, err := Parse(strings.NewReader(tc.input), format)
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}

			object, ok := parseConfig.(map[string]any)
			if !ok {
				t.Fatalf("root type = %T, want map[string]any", parseConfig)
			}

			got, ok := object["payload"].(string)
			if !ok {
				t.Fatalf(
					"payload type = %T, want string",
					object["payload"],
				)
			}

			if got != payload {
				t.Fatalf(
					"payload differs: got length %d, want %d",
					len(got),
					len(payload),
				)
			}
		})
	}
}

func TestParseSizeBoundary(t *testing.T) {
	cases := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{
			name: "below limit",
			size: int(MaxBytes) - 1,
		},
		{
			name: "exact limit",
			size: int(MaxBytes),
		},
		{
			name:    "above limit",
			size:    int(MaxBytes) + 1,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := largeJSON(tc.size)

			if len(input) != tc.size {
				t.Fatalf(
					"input size = %d, want %d",
					len(input),
					tc.size,
				)
			}

			parseConfig, err := Parse(strings.NewReader(input), JSON)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected size limit error")
				}

				if parseConfig != nil {
					t.Fatal("expected nil config")
				}

				return
			}

			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}

			object, ok := parseConfig.(map[string]any)
			if !ok {
				t.Fatalf("config type = %T, want map[string]any", parseConfig)
			}

			payload, ok := object["payload"].(string)
			if !ok {
				t.Fatal("payload must be a string")
			}

			const overhead = len(`{"payload":""}`)
			if len(payload) != tc.size-overhead {
				t.Fatal("payload was truncated")
			}
		})
	}
}

func TestParseLargeInvalidJSON(t *testing.T) {
	valid := largeJSON(int(MaxBytes))
	invalid := valid[:len(valid)-1]

	parseConfig, err := Parse(strings.NewReader(invalid), JSON)
	if err == nil {
		t.Fatal("expected JSON parsing error")
	}

	if parseConfig != nil {
		t.Fatal("expected nil config")
	}
}
