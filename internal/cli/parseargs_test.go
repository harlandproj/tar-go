package cli

import "testing"

func TestParseAddFile(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--add-file=test.txt", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Errorf("--add-file should be accepted: %v", err)
	}
	found := false
	for _, f := range opts.FileNames {
		if f == "test.txt" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected test.txt in FileNames, got %v", opts.FileNames)
	}
}

func TestParseNoIgnoreCase(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--no-ignore-case", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Errorf("--no-ignore-case should be accepted: %v", err)
	}
}

func TestParseNoIgnoreCommandError(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--no-ignore-command-error", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Errorf("--no-ignore-command-error should be accepted: %v", err)
	}
}

func TestParseNoNull(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--no-null", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Errorf("--no-null should be accepted: %v", err)
	}
}

func TestParseNoOverwriteDir(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--no-overwrite-dir", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Errorf("--no-overwrite-dir should be accepted: %v", err)
	}
}

func TestParseNoSamePermissions(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--no-same-permissions", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Errorf("--no-same-permissions should be accepted: %v", err)
	}
}

func TestParseNoUnquote(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--no-unquote", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Errorf("--no-unquote should be accepted: %v", err)
	}
}

func TestParseShortSubcommands(t *testing.T) {
	tests := []struct {
		arg  string
		want Subcommand
	}{
		{"-c", SubCreate},
		{"-x", SubExtract},
		{"-t", SubList},
		{"-r", SubAppend},
		{"-u", SubUpdate},
		{"-A", SubConcat},
		{"-d", SubDiff},
	}
	for _, tt := range tests {
		opts := &Options{}
		err := parseArgs([]string{"tar", tt.arg, "-f", "a.tar"}, opts)
		if err != nil {
			t.Errorf("%s: %v", tt.arg, err)
		}
		if opts.Subcommand != tt.want {
			t.Errorf("%s: got %d, want %d", tt.arg, opts.Subcommand, tt.want)
		}
	}
}

func TestParseCompressionShort(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{"-z", "gzip"},
		{"-j", "bzip2"},
		{"-J", "xz"},
		{"-Z", "compress"},
	}
	for _, tt := range tests {
		opts := &Options{}
		err := parseArgs([]string{"tar", "-cf", "a.tar", tt.arg}, opts)
		if err != nil {
			t.Errorf("%s: %v", tt.arg, err)
		}
		if opts.CompressProgram != tt.want {
			t.Errorf("%s: got %q, want %q", tt.arg, opts.CompressProgram, tt.want)
		}
	}
}

func TestParseVerbose(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "-vvv"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.Verbose != 3 {
		t.Errorf("expected verbose=3, got %d", opts.Verbose)
	}
}

func TestParseExclude(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--exclude=*.log", "--exclude=*.tmp"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(opts.Exclude) != 2 {
		t.Errorf("expected 2 excludes, got %d", len(opts.Exclude))
	}
}

func TestParseStripComponents(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-xf", "a.tar", "--strip-components=2"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.StripComponents != 2 {
		t.Errorf("expected 2, got %d", opts.StripComponents)
	}
}

func TestParseFilesFrom(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "-T", "list.txt"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.FilesFrom != "list.txt" {
		t.Errorf("expected list.txt, got %q", opts.FilesFrom)
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		arg  string
		want ArchiveFormat
	}{
		{"--format=gnu", FormatGNU},
		{"--format=ustar", FormatUstar},
		{"--format=posix", FormatPOSIX},
		{"--format=v7", FormatV7},
		{"--format=oldgnu", FormatOldGNU},
	}
	for _, tt := range tests {
		opts := &Options{}
		err := parseArgs([]string{"tar", "-cf", "a.tar", tt.arg}, opts)
		if err != nil {
			t.Errorf("%s: %v", tt.arg, err)
		}
		if opts.ArchiveFormat != tt.want {
			t.Errorf("%s: got %d, want %d", tt.arg, opts.ArchiveFormat, tt.want)
		}
	}
}

func TestParseMutualExclusion(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "-xf", "a.tar"}, opts)
	if err == nil {
		t.Error("expected error for conflicting subcommands")
	}
}

func TestParseInvalidOption(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--nonexistent-option"}, opts)
	if err == nil {
		t.Error("expected error for invalid option")
	}
}

func TestParseDoubleDash(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--", "-file.txt"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range opts.FileNames {
		if f == "-file.txt" {
			found = true
		}
	}
	if !found {
		t.Error("expected -file.txt after --")
	}
}

func TestParseWarning(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--warning=all", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Errorf("--warning should be accepted: %v", err)
	}
	if opts.Warning == 0 {
		t.Error("expected non-zero warning flags")
	}
}
