package cli

import (
	"errors"
	"strings"

	"github.com/serebryakov1997/utility/config"
)

type options struct {
	path string
	stdin bool
	silent bool
	help bool
	format string
}

func parseArgs(args []string) (options, error) {
	opts := options{format: config.Auto}
	positionalOnly := false
	pathSeen := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !positionalOnly {
			switch arg {
			case "-h", "--help":
				opts.help = true
				continue
			case "-s", "--silent":
				opts.silent = true
				continue
			case "--stdin":
				opts.stdin = true
				continue
			case "--format":
				if i+1 == len(args) {
					return opts, errors.New("after --format requires auto, json or yaml")
				}
				i++
				opts.format = args[i]
				continue
			}

			if strings.HasPrefix(arg, "--format=") {
				opts.format = strings.TrimPrefix(arg, "--format=")
				continue
			}

			if strings.HasPrefix(arg, "-") {
				return opts, errors.New("unknown flag; use --help")
			}
		}

		if pathSeen || arg == "" {
			return opts, errors.New("required only one not empty path to file")
		}
		pathSeen = true
		opts.path = arg
	}

	if opts.format != config.Auto && opts.format != config.JSON && opts.format != config.YAML {
		return opts, errors.New("--format requires auto, json or yaml")
	}

	if opts.help {
		return opts, nil
	}

	if opts.stdin == pathSeen {
		return opts, errors.New("set path to file or --stdin")
	}

	return opts, nil
}