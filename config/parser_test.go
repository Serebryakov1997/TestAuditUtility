package config

import (
	"strings"
	"testing"
)

func TestParseNeedFormats(t *testing.T) {
	json := `{"services":[{"log":{"level":"debug"}}]}`
	yaml := "version: 2.2\nstorage:\n    digest-algorithm: MD5\n"

	_, err := Parse(strings.NewReader(json), JSON)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Parse(strings.NewReader(yaml), YAML)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseValidData(t *testing.T) {
	cases := []struct{
		name string
		input string
		format string
	}{
		{"empty JSON object", `{}`, JSON},
		{"empty YAML object", `{}`, YAML},
		{"JSON array config", `[{"debug":true}]`, JSON},
		{"YAML array config", "- debug: true", YAML},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Parse(strings.NewReader(tc.input), tc.format); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestParseInvalidData(t *testing.T) {
	cases := []struct{
		name string
		input string
		format string
	}{
		{"empty", " \n", JSON},
		{"JSON broken", `{"password":"secret"`, JSON},
		{"JSON duplicate", `{"level":"debug", "level":"info"}`, JSON},
		{"JSON scalar", `50`, JSON},
		{"JSON extra document", `{} {}`, JSON},
		{"YAML broken", "level: [", YAML},
		{"YAML duplicate", "debug: true\ndebug: false", YAML},
		{"YAML not string key", "1: debug", YAML},
		{"YAML alias", "level: &x {debug: true}\nb: *x", YAML},
		{"YAML extra document", "version: 1.2\n---\ndebug: true", YAML},
		{"YAML scalar", "secret", YAML},
		{"YAML empty document", "---", YAML},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Parse(strings.NewReader(tc.input), tc.format); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}