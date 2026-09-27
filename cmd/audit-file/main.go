package main

import (
	"fmt"
	"os"

	"github.com/serebryakov1997/utility/cli"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		if err := runServer(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	os.Exit(cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
