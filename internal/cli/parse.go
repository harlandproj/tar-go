package cli

func Parse(args []string, opts *Options) error {
	return parseArgs(args, opts)
}
