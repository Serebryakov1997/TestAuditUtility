package main

import (
	"os"

	"github.com/serebryakov1997/utility/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
