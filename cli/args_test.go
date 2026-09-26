package cli

import (
	"testing"

	"github.com/serebryakov1997/utility/config"
)


func TestParseArgs(t *testing.T) {
	cases := []struct{
		name string
		args []string
		want options
	}{
		{"file", []string{"file.yaml"}, options{path: "file.yaml", format: config.Auto}},
		{"silent after", []string{"file.yaml", "--silent"}, options{path: "file.yaml", silent: true, format: config.Auto}},
		{"silent before", []string{"-s", "file.json"}, options{path: "file.json", silent: true, format: config.Auto}},
		{"stdin", []string{"--stdin"}, options{stdin: true, format: config.Auto}},
		{"explicit format", []string{"--stdin", "--format=yaml"}, options{stdin: true, format: config.YAML}},
		{"split format", []string{"file.txt", "--format", "json"}, options{path: "file.txt", format: config.JSON}},
		{"help", []string{"--help"}, options{help: true, format: config.Auto}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseArgs(tc.args)
			if err != nil || got != tc.want {
				t.Fatalf("got %+v, %v; expected %+v", got, err, tc.want)
			}
		})
	}
}

func TestParseArgsErrors(t *testing.T) {
	cases := [][]string{
		nil, {""}, {"file1.json", "file2.json"}, {"--stdin", "file1.json"},
		{"--unknown"}, {"--stdin", "--format"}, {"--stdin", "--format=xml"},
		{"--stdin", "--format="},
	}

	for _, args := range cases {
		if _, err := parseArgs(args); err == nil {
			t.Fatalf("args %q gor without errors", args)
		}
	}
}