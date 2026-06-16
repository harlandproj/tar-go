package app

import (
	"fmt"
	"os"

	"github.com/harlandproj/tar-go/internal/cli"
	"github.com/harlandproj/tar-go/internal/ops"
)

func dispatch(opts *cli.Options, version string) int {
	switch opts.Subcommand {
	case cli.SubCreate:
		if err := ops.Create(opts); err != nil {
			fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			return 2
		}
		return 0
	case cli.SubList:
		if err := ops.List(opts); err != nil {
			fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			return 2
		}
		return 0
	case cli.SubExtract:
		if err := ops.Extract(opts); err != nil {
			fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			return 2
		}
		return 0
	case cli.SubAppend:
		if err := ops.Append(opts); err != nil {
			if e, ok := err.(*cli.ExitError); ok {
				fmt.Fprintf(os.Stderr, "tar: %s\n", e.Message)
				return e.Code
			}
			fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			return 2
		}
		return 0
	case cli.SubUpdate:
		if err := ops.Update(opts); err != nil {
			if e, ok := err.(*cli.ExitError); ok {
				fmt.Fprintf(os.Stderr, "tar: %s\n", e.Message)
				return e.Code
			}
			fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			return 2
		}
		return 0
	case cli.SubConcat:
		if err := ops.Concat(opts); err != nil {
			if e, ok := err.(*cli.ExitError); ok {
				fmt.Fprintf(os.Stderr, "tar: %s\n", e.Message)
				return e.Code
			}
			fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			return 2
		}
		return 0
	case cli.SubDiff:
		if err := ops.Diff(opts); err != nil {
			if e, ok := err.(*cli.ExitError); ok {
				fmt.Fprintf(os.Stderr, "tar: %s\n", e.Message)
				return e.Code
			}
			fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			return 2
		}
		return 0
	case cli.SubDelete:
		if err := ops.Delete(opts); err != nil {
			if e, ok := err.(*cli.ExitError); ok {
				fmt.Fprintf(os.Stderr, "tar: %s\n", e.Message)
				return e.Code
			}
			fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			return 2
		}
		return 0
	case cli.SubTestLabel:
		if err := ops.TestLabel(opts); err != nil {
			if e, ok := err.(*cli.ExitError); ok {
				fmt.Fprintf(os.Stderr, "tar: %s\n", e.Message)
				return e.Code
			}
			fmt.Fprintf(os.Stderr, "tar: %v\n", err)
			return 2
		}
		return 0
	case cli.SubNone:
		fmt.Fprintf(os.Stderr, "tar: You must specify one of the '-Acdtrux', '--delete' or '--test-label' options\n")
		fmt.Fprintf(os.Stderr, "Try 'tar --help' or 'tar --usage' for more information.\n")
		return 2
	default:
		return 0
	}
}
