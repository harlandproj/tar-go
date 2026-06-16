package app

import (
	"fmt"
	"os"

	"github.com/harlandproj/tar-go/internal/cli"
)

func Run(args []string, version string) int {
	opts := &cli.Options{
		BlockingFactor: 20,
		RecordSize:     10240,
		ArchiveFormat:  cli.FormatGNU,
	}
	if err := cli.Parse(args, opts); err != nil {
		if err == cli.ErrHelpRequested {
			return 0
		}
		fmt.Fprintf(os.Stderr, "tar: %v\n", err)
		return 2
	}
	return dispatch(opts, version)
}
