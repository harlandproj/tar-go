package app

import (
	"fmt"
	"os"

	"github.com/harlandproj/tar-go/internal/cli"
)

func dispatch(opts *cli.Options, version string) int {
	switch opts.Subcommand {
	case cli.SubNone:
		fmt.Fprintf(os.Stderr, "tar: You must specify one of the '-Acdtrux', '--delete' or '--test-label' options\n")
		fmt.Fprintf(os.Stderr, "Try 'tar --help' or 'tar --usage' for more information.\n")
		return 2
	default:
		return 0
	}
}
