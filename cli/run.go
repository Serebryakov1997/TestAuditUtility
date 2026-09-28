package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/serebryakov1997/utility/audit"
	"github.com/serebryakov1997/utility/config"
	"github.com/serebryakov1997/utility/filecheck"
	"github.com/serebryakov1997/utility/rules"
)

const (
	ExitOK       = 0
	ExitFindings = 1
	ExitError    = 2
)

func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	opts, err := parseArgs(args)
	if err != nil {
		return fail(stderr, err)
	}
	if opts.help {
		if _, err := io.WriteString(stdout, usage); err != nil {
			return fail(stderr, err)
		}
		return ExitOK
	}

	input := stdin
	format := opts.format
	var permissionFindings []audit.Finding
	if !opts.stdin {
		permissionFindings, err = filecheck.Permissions(opts.path)
		if err != nil {
			return fail(stderr, err)
		}

		file, err := os.Open(opts.path)
		if err != nil {
			return fail(stderr, fmt.Errorf("open file: %w", err))
		}
		defer file.Close()
		input = file
	}

	parseConfig, err := config.Parse(input, format)
	if err != nil {
		return fail(stderr, err)
	}

	analyzer := audit.New(rules.RegisteredRules()...)
	findings, err := analyzer.Analyze(parseConfig)
	if err != nil {
		return fail(stderr, err)
	}

	findings = append(permissionFindings, findings...)

	if err := writeReport(stdout, findings); err != nil {
		return fail(stderr, fmt.Errorf("report entry: %w", err))
	}

	if len(findings) > 0 && !opts.silent {
		return ExitFindings
	}

	return ExitOK
}

func fail(stderr io.Writer, err error) int {
	_, _ = fmt.Fprintf(stderr, "Ошибка: %v\n", err)
	return ExitError
}
