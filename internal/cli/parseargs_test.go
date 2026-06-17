package cli

import (
	"os"
	"testing"
)

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

func TestParseWarningNone(t *testing.T) {
	opts := &Options{Warning: WarnAll}
	err := parseArgs([]string{"tar", "--warning=none", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.Warning != 0 {
		t.Error("expected zero warning flags after none")
	}
}

func TestParseWarningSpecific(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--warning=file-changed", "-cf", "a.tar"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.Warning&WarnFileChanged == 0 {
		t.Error("expected WarnFileChanged flag")
	}
}

func TestDefaultCompressBySuffix(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"archive.tar.gz", "gzip"},
		{"archive.tar.bz2", "bzip2"},
		{"archive.tar.xz", "xz"},
		{"archive.tar.zst", "zstd"},
		{"archive.tar.lz", "lzip"},
		{"archive.tar.lzma", "lzma"},
		{"archive.tar.lzo", "lzop"},
		{"archive.tar.Z", "compress"},
		{"archive.tgz", "gzip"},
		{"archive.tbz2", "bzip2"},
		{"archive.txz", "xz"},
		{"archive.txt", ""},
	}
	for _, tt := range tests {
		got := defaultCompressBySuffix([]string{tt.name})
		if got != tt.want {
			t.Errorf("defaultCompressBySuffix(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestParseNewer(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--newer=2024-01-01"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.NewerMtime.IsZero() {
		t.Error("expected newer mtime to be set")
	}
}

func TestParseMtime(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--mtime=2024-06-15"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.Mtime.IsZero() {
		t.Error("expected mtime to be set")
	}
	if opts.SetMtimeMode != MtimeForce {
		t.Errorf("expected MtimeForce, got %d", opts.SetMtimeMode)
	}
}

func TestParseMtimeClamp(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--clamp-mtime"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.SetMtimeMode != MtimeClamp {
		t.Errorf("expected MtimeClamp, got %d", opts.SetMtimeMode)
	}
}

func TestParseMtimeAtime(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--mtime=atime"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.SetMtimeMode != MtimeCommand {
		t.Errorf("expected MtimeCommand, got %d", opts.SetMtimeMode)
	}
	if opts.SetMtimeCommand != "atime" {
		t.Errorf("expected atime, got %q", opts.SetMtimeCommand)
	}
}

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"2024-01-01", true},
		{"2024-01-01 12:00:00", true},
		{"2024-06-15T10:30:00", true},
		{"1704067200", true},
		{"not-a-date", false},
	}
	for _, tt := range tests {
		_, err := parseDateTime(tt.input)
		if (err == nil) != tt.valid {
			t.Errorf("parseDateTime(%q) valid=%v, want %v", tt.input, err == nil, tt.valid)
		}
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"100", 100},
		{"10k", 10240},
		{"5M", 5 * 1024 * 1024},
		{"2G", 2 * 1024 * 1024 * 1024},
		{"", 0},
		{"abc", 0},
	}
	for _, tt := range tests {
		got := parseSize(tt.input)
		if got != tt.want {
			t.Errorf("parseSize(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestCollectTokens(t *testing.T) {
	os.Setenv("TAR_OPTIONS", "--verbose --exclude=*.log")
	defer os.Unsetenv("TAR_OPTIONS")
	tokens := collectTokens([]string{"tar", "-cf", "a.tar"})
	if len(tokens) < 3 {
		t.Errorf("expected at least 3 tokens, got %d", len(tokens))
	}
}

func TestCollectTokensNoEnv(t *testing.T) {
	os.Unsetenv("TAR_OPTIONS")
	tokens := collectTokens([]string{"tar", "-cf", "a.tar"})
	if len(tokens) != 2 {
		t.Errorf("expected 2 tokens, got %d", len(tokens))
	}
}

func TestIsShortOption(t *testing.T) {
	tests := []struct {
		arg  string
		want bool
	}{
		{"-c", true},
		{"-cf", true},
		{"--create", false},
		{"-", false},
		{"-1", false},
	}
	for _, tt := range tests {
		got := isShortOption(tt.arg)
		if got != tt.want {
			t.Errorf("isShortOption(%q) = %v, want %v", tt.arg, got, tt.want)
		}
	}
}

func TestParseBlockingFactor(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "-b", "50"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.BlockingFactor != 50 {
		t.Errorf("expected 50, got %d", opts.BlockingFactor)
	}
	if opts.RecordSize != 50*512 {
		t.Errorf("expected %d, got %d", 50*512, opts.RecordSize)
	}
}

func TestParseRecordSize(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--record-size=10240"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.RecordSize != 10240 {
		t.Errorf("expected 10240, got %d", opts.RecordSize)
	}
}

func TestParseLongSubcommands(t *testing.T) {
	tests := []struct {
		arg  string
		want Subcommand
	}{
		{"--create", SubCreate},
		{"--extract", SubExtract},
		{"--list", SubList},
		{"--append", SubAppend},
		{"--update", SubUpdate},
		{"--catenate", SubConcat},
		{"--concatenate", SubConcat},
		{"--diff", SubDiff},
		{"--compare", SubDiff},
		{"--delete", SubDelete},
		{"--test-label", SubTestLabel},
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

func TestParseCompressionLong(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{"--gzip", "gzip"},
		{"--bzip2", "bzip2"},
		{"--xz", "xz"},
		{"--zstd", "zstd"},
		{"--lzip", "lzip"},
		{"--lzma", "lzma"},
		{"--lzop", "lzop"},
		{"--compress", "compress"},
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

func TestParseAutoCompress(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar.gz", "-a"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.AutoCompress {
		t.Error("expected auto-compress")
	}
}

func TestParseShortFlags(t *testing.T) {
	tests := []struct {
		args []string
		check func(*Options) bool
	}{
		{[]string{"tar", "-cf", "a.tar", "-S"}, func(o *Options) bool { return o.Sparse }},
		{[]string{"tar", "-cf", "a.tar", "-p"}, func(o *Options) bool { return o.SamePermissions }},
		{[]string{"tar", "-cf", "a.tar", "-m"}, func(o *Options) bool { return o.Touch }},
		{[]string{"tar", "-cf", "a.tar", "-P"}, func(o *Options) bool { return o.AbsoluteNames }},
		{[]string{"tar", "-cf", "a.tar", "-O"}, func(o *Options) bool { return o.ToStdout }},
		{[]string{"tar", "-cf", "a.tar", "-R"}, func(o *Options) bool { return o.BlockNumber }},
		{[]string{"tar", "-cf", "a.tar", "-l"}, func(o *Options) bool { return o.CheckLinks }},
		{[]string{"tar", "-cf", "a.tar", "-W"}, func(o *Options) bool { return o.Verify }},
		{[]string{"tar", "-cf", "a.tar", "-M"}, func(o *Options) bool { return o.MultiVolume }},
		{[]string{"tar", "-cf", "a.tar", "-i"}, func(o *Options) bool { return o.IgnoreZeros }},
		{[]string{"tar", "-cf", "a.tar", "-B"}, func(o *Options) bool { return o.ReadFullRecords }},
		{[]string{"tar", "-cf", "a.tar", "-n"}, func(o *Options) bool { return o.Seek }},
		{[]string{"tar", "-cf", "a.tar", "-k"}, func(o *Options) bool { return o.KeepOldFiles == OldKeepOldFiles }},
	}
	for _, tt := range tests {
		opts := &Options{}
		err := parseArgs(tt.args, opts)
		if err != nil {
			t.Errorf("%v: %v", tt.args, err)
		}
		if !tt.check(opts) {
			t.Errorf("%v: check failed", tt.args)
		}
	}
}

func TestParseDelete(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--delete", "-f", "a.tar"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.Subcommand != SubDelete {
		t.Errorf("expected SubDelete, got %d", opts.Subcommand)
	}
}

func TestParseTestLabel(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--test-label", "-f", "a.tar"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.Subcommand != SubTestLabel {
		t.Errorf("expected SubTestLabel, got %d", opts.Subcommand)
	}
}

func TestParseUseCompressProgram(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "-I", "pigz"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.CompressProgram != "pigz" {
		t.Errorf("expected pigz, got %q", opts.CompressProgram)
	}
}

func TestParseVolumeLabel(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "-V", "MYVOL"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.VolumeLabel != "MYVOL" {
		t.Errorf("expected MYVOL, got %q", opts.VolumeLabel)
	}
}

func TestParseTapeLength(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "-L", "1024"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.TapeLength != 1024 {
		t.Errorf("expected 1024, got %d", opts.TapeLength)
	}
}

func TestParseStartingFile(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "-K", "file.txt"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.StartingFile != "file.txt" {
		t.Errorf("expected file.txt, got %q", opts.StartingFile)
	}
}

func TestParseOwnerGroupLong(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--owner=root", "--group=wheel"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.Owner != "root" {
		t.Errorf("expected root, got %q", opts.Owner)
	}
	if opts.Group != "wheel" {
		t.Errorf("expected wheel, got %q", opts.Group)
	}
}

func TestParseOneFileSystem(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--one-file-system"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.OneFileSystem {
		t.Error("expected one-file-system")
	}
}

func TestParseDereference(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--dereference"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.Dereference {
		t.Error("expected dereference")
	}
}

func TestParseHardDereference(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--hard-dereference"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.HardDereference {
		t.Error("expected hard-dereference")
	}
}

func TestParseNumericOwner(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--numeric-owner"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.NumericOwner {
		t.Error("expected numeric-owner")
	}
}

func TestParseRemoveFiles(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--remove-files"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.RemoveFiles {
		t.Error("expected remove-files")
	}
}

func TestParseRecursiveUnlink(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-xf", "a.tar", "--recursive-unlink"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.RecursiveUnlink {
		t.Error("expected recursive-unlink")
	}
}

func TestParseSameOwner(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-xf", "a.tar", "--same-owner"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.SameOwner {
		t.Error("expected same-owner")
	}
}

func TestParseTransformLong(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--transform=s/old/new/"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.Transform != "s/old/new/" {
		t.Errorf("expected s/old/new/, got %q", opts.Transform)
	}
}

func TestParseExcludeFromShort(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "-X", "excludes.txt"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.ExcludeFrom != "excludes.txt" {
		t.Errorf("expected excludes.txt, got %q", opts.ExcludeFrom)
	}
}

func TestParseExcludeVCS(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--exclude-vcs"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.ExcludeVCS {
		t.Error("expected exclude-vcs")
	}
}

func TestParseExcludeCaches(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--exclude-caches"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.ExcludeCaches {
		t.Error("expected exclude-caches")
	}
}

func TestParseSort(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--sort=name"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.SortOrder != "name" {
		t.Errorf("expected name, got %q", opts.SortOrder)
	}
}

func TestParseOneTopLevel(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-xf", "a.tar", "--one-top-level=dir"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.OneTopLevel != "dir" {
		t.Errorf("expected dir, got %q", opts.OneTopLevel)
	}
}

func TestParseDelayDirRestore(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-xf", "a.tar", "--delay-directory-restore"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.DelayDirRestore {
		t.Error("expected delay-directory-restore")
	}
}

func TestParseCheckpoint(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--checkpoint=5"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.Checkpoint != 5 {
		t.Errorf("expected 5, got %d", opts.Checkpoint)
	}
}

func TestParseBackupType(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-xf", "a.tar", "--backup=numbered"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.BackupType != "numbered" {
		t.Errorf("expected numbered, got %q", opts.BackupType)
	}
}

func TestParseIncremental(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "-G"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.Incremental {
		t.Error("expected incremental")
	}
}

func TestParseListedIncremental(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "-g", "snap.file"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.ListedIncremental != "snap.file" {
		t.Errorf("expected snap.file, got %q", opts.ListedIncremental)
	}
}

func TestParseSkipOldFiles(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-xf", "a.tar", "--skip-old-files"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.KeepOldFiles != OldSkipOldFiles {
		t.Errorf("expected OldSkipOldFiles, got %d", opts.KeepOldFiles)
	}
}

func TestParseOverwrite(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-xf", "a.tar", "--overwrite"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.KeepOldFiles != OldOverwrite {
		t.Errorf("expected OldOverwrite, got %d", opts.KeepOldFiles)
	}
}

func TestParseUnlinkFirst(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-xf", "a.tar", "-U"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.KeepOldFiles != OldUnlinkFirst {
		t.Errorf("expected OldUnlinkFirst, got %d", opts.KeepOldFiles)
	}
}

func TestParseXattrs(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--xattrs"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.Xattrs {
		t.Error("expected xattrs")
	}
}

func TestParseForceLocal(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "--force-local"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.ForceLocal {
		t.Error("expected force-local")
	}
}

func TestParseHelp(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--help"}, opts)
	if err != ErrHelpRequested {
		t.Errorf("expected ErrHelpRequested, got %v", err)
	}
}

func TestParseVersion(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "--version"}, opts)
	if err != ErrHelpRequested {
		t.Errorf("expected ErrHelpRequested, got %v", err)
	}
}

func TestParseNoSubcommand(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar"}, opts)
	if err == nil {
		t.Error("expected error for no subcommand")
	}
}

func TestParseInfoScript(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-cf", "a.tar", "-F", "script.sh"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.InfoScript != "script.sh" {
		t.Errorf("expected script.sh, got %q", opts.InfoScript)
	}
	if !opts.MultiVolume {
		t.Error("expected multi-volume implied by -F")
	}
}

func TestParseKeepNewerFiles(t *testing.T) {
	opts := &Options{}
	err := parseArgs([]string{"tar", "-xf", "a.tar", "--keep-newer-files"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if opts.KeepOldFiles != OldKeepNewerFiles {
		t.Errorf("expected OldKeepNewerFiles, got %d", opts.KeepOldFiles)
	}
}
