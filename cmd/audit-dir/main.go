package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/serebryakov1997/utility/audit"
	"github.com/serebryakov1997/utility/directory"
	"github.com/serebryakov1997/utility/rules"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: audit-dir <directory>")
		os.Exit(2)
	}

	analyzer := audit.New(rules.RegisteredRules()...)

	scanner, err := directory.New(analyzer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scanner init failed: %v\n", err)
		os.Exit(2)
	}

	reports, err := scanner.Analyze(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "directory analysis failed: %v\n", err)
		os.Exit(2)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", " ")

	if err := encoder.Encode(reports); err != nil {
		fmt.Fprintf(os.Stderr, "write report: %v\n", err)
		os.Exit(2)
	}
}
