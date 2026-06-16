package cli

import (
	"fmt"
)

func parseArgs(args []string, opts *Options) error {
	if len(args) < 2 {
		return fmt.Errorf("not enough arguments")
	}
	return nil
}
