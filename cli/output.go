package cli

import (
	"fmt"
	"io"

	"github.com/serebryakov1997/utility/audit"
)

func writeReport(w io.Writer, findings []audit.Finding) error {
	if len(findings) == 0 {
		_, err := fmt.Fprintln(w, "Not found problems.")
		return err
	}

	for _, f := range findings {
		if _, err := fmt.Fprintf(w, "%s %s [%s]\n%s\nRecommendation: %s\n\n",
			f.Severity, f.Path, f.RuleID, f.Message, f.Recommendation); err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(w, "Found problems: %d\n", len(findings))
	return err
}

const usage = `audit-file - проверка конфигураций JSON/YAML

Usage:
  audit-file [flags] <file.json|file.yaml> [flags]
  audit-file --stdin [flags]

Flags:
  -s, --silent		return 0 if found; report is unvisible
  --stdin			read stdin instead of a file
  --format FORMAT	auto, json or yaml (default auto)
  -h, --help		show help

Completion codes: 0 - success, 1 - found problems, 2 - exec error
`
