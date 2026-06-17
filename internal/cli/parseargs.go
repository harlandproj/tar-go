package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func parseArgs(args []string, opts *Options) error {
	// collect tokens: TAR_OPTIONS env + CLI args
	tokens := collectTokens(args)
	p := &parser{opts: opts}
	p.parse(tokens)

	if p.help {
		PrintHelp(args[0])
		return ErrHelpRequested
	}
	if p.version {
		PrintVersion(args[0])
		return ErrHelpRequested
	}
	if p.usage {
		PrintUsage(args[0])
		return ErrHelpRequested
	}
	if p.showDefaults {
		PrintDefaults()
		return ErrHelpRequested
	}
	if p.errorMsg != "" {
		return fmt.Errorf("%s", p.errorMsg)
	}

	if !p.hasSubcommand && opts.Subcommand == SubNone && len(opts.FileNames) == 0 {
		if !opts.ShowDefaults {
			return fmt.Errorf("You must specify one of the '-Acdtrux', '--delete' or '--test-label' options\nTry 'tar --help' or 'tar --usage' for more information.")
		}
	}

	if opts.AutoCompress && opts.CompressProgram == "" {
		opts.CompressProgram = defaultCompressBySuffix(opts.ArchiveNames)
	}

	return nil
}

func defaultCompressBySuffix(names []string) string {
	if len(names) == 0 {
		return ""
	}
	name := names[0]
	switch {
	case strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz"):
		return "gzip"
	case strings.HasSuffix(name, ".tar.bz2") || strings.HasSuffix(name, ".tbz2") || strings.HasSuffix(name, ".tbz") || strings.HasSuffix(name, ".tb2"):
		return "bzip2"
	case strings.HasSuffix(name, ".tar.xz") || strings.HasSuffix(name, ".txz"):
		return "xz"
	case strings.HasSuffix(name, ".tar.zst") || strings.HasSuffix(name, ".tzst"):
		return "zstd"
	case strings.HasSuffix(name, ".tar.lz"):
		return "lzip"
	case strings.HasSuffix(name, ".tar.lzma") || strings.HasSuffix(name, ".tlz"):
		return "lzma"
	case strings.HasSuffix(name, ".tar.lzo") || strings.HasSuffix(name, ".tzo"):
		return "lzop"
	case strings.HasSuffix(name, ".tar.Z") || strings.HasSuffix(name, ".taz"):
		return "compress"
	default:
		return ""
	}
}

func collectTokens(args []string) []string {
	var tokens []string
	if env := os.Getenv("TAR_OPTIONS"); env != "" {
		for _, tok := range strings.Fields(env) {
			tokens = append(tokens, tok)
		}
	}
	if len(args) > 1 {
		tokens = append(tokens, args[1:]...)
	}
	return tokens
}

type parser struct {
	opts           *Options
	help           bool
	version        bool
	usage          bool
	showDefaults   bool
	hasSubcommand  bool
	stopOptions    bool
	errorMsg       string
}

func (p *parser) parse(tokens []string) {
	for i := 0; i < len(tokens); i++ {
		arg := tokens[i]

		if p.stopOptions {
			p.opts.FileNames = append(p.opts.FileNames, arg)
			continue
		}

		if arg == "--" {
			p.stopOptions = true
			continue
		}

		if len(arg) > 2 && arg[0] == '-' && arg[1] == '-' {
			if !p.parseLongOption(arg, tokens, &i) {
				return
			}
			continue
		}

		if len(arg) > 1 && arg[0] == '-' && isShortOption(arg) {
			if !p.parseShortOptions(arg[1:], tokens, &i) {
				return
			}
			continue
		}

		if arg == "-" || (len(arg) > 1 && arg[0] == '-') {
			p.opts.FileNames = append(p.opts.FileNames, arg)
			continue
		}

		p.opts.FileNames = append(p.opts.FileNames, arg)
	}
}

func isShortOption(arg string) bool {
	if len(arg) < 2 || arg[0] != '-' {
		return false
	}
	if arg[1] == '-' {
		return false
	}
	if arg[1] >= '0' && arg[1] <= '9' {
		return false
	}
	return true
}

func (p *parser) parseShortOptions(opts string, rest []string, idx *int) bool {
	for j := 0; j < len(opts); j++ {
		ch := opts[j]

		var needsArg bool
		var val string

		needsArg, val = p.shortOption(ch, rest, idx, j == len(opts)-1)
		if val == "__ERROR__" {
			return false
		}
		if needsArg && val == "" && j == len(opts)-1 {
			(*idx)++
			if *idx < len(rest) {
				val = rest[*idx]
			}
		}
		if needsArg || val != "" {
			p.applyShortArg(ch, val)
		}
	}
	return true
}

func (p *parser) shortOption(ch byte, rest []string, idx *int, isLast bool) (needsArg bool, val string) {
	switch ch {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return false, ""

	case 'A':
		p.setSubcommand(SubConcat)
	case 'c':
		p.setSubcommand(SubCreate)
	case 'd':
		p.setSubcommand(SubDiff)
	case 't':
		p.setSubcommand(SubList)
	case 'r':
		p.setSubcommand(SubAppend)
	case 'u':
		p.setSubcommand(SubUpdate)
	case 'x':
		p.setSubcommand(SubExtract)

	case 'z':
		p.opts.CompressProgram = "gzip"
	case 'j':
		p.opts.CompressProgram = "bzip2"
	case 'Z':
		p.opts.CompressProgram = "compress"
	case 'J':
		p.opts.CompressProgram = "xz"
	case 'a':
		p.opts.AutoCompress = true
	case 'v':
		p.opts.Verbose++
	case 'k':
		p.opts.KeepOldFiles = OldKeepOldFiles
	case 'S':
		p.opts.Sparse = true
	case 'p':
		p.opts.SamePermissions = true
	case 'm':
		p.opts.Touch = true
	case 'P':
		p.opts.AbsoluteNames = true
	case 'h':
		p.opts.Dereference = true
	case 'O':
		p.opts.ToStdout = true
	case 'R':
		p.opts.BlockNumber = true
	case 'l':
		p.opts.CheckLinks = true
	case 'w':
		p.opts.Interactive = true
	case 'W':
		p.opts.Verify = true
	case 'M':
		p.opts.MultiVolume = true
	case 'G':
		p.opts.Incremental = true
	case 's':
		p.opts.PreserveOrder = true
	case 'n':
		p.opts.Seek = true
	case 'i':
		p.opts.IgnoreZeros = true
	case 'B':
		p.opts.ReadFullRecords = true
	case 'U':
		p.opts.KeepOldFiles = OldUnlinkFirst

	case 'f':
		return true, ""
	case 'b':
		return true, ""
	case 'C':
		return true, ""
	case 'L':
		return true, ""
	case 'H':
		return true, ""
	case 'g':
		return true, ""
	case 'N':
		return true, ""
	case 'K':
		return true, ""
	case 'F':
		return true, ""
	case 'I':
		return true, ""
	case 'V':
		return true, ""
	case 'T':
		return true, ""

	default:
		p.errorMsg = fmt.Sprintf("Invalid option -- '%c'\nTry 'tar --help' or 'tar --usage' for more information.", ch)
		return false, "__ERROR__"
	}
	return false, ""
}

func (p *parser) applyShortArg(ch byte, val string) {
	switch ch {
	case 'f':
		p.opts.ArchiveNames = append(p.opts.ArchiveNames, val)
	case 'b':
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			p.opts.BlockingFactor = n
			p.opts.RecordSize = n * 512
		}
	case 'C':
		p.opts.FileNames = append(p.opts.FileNames, "-C", val)
	case 'L':
		p.opts.TapeLength = parseSize(val)
	case 'H':
		p.parseFormat(val)
	case 'g':
		p.opts.ListedIncremental = val
	case 'N':
		p.parseNewer(val, false)
	case 'K':
		p.opts.StartingFile = val
	case 'F':
		p.opts.InfoScript = val
		p.opts.MultiVolume = true
	case 'I':
		p.opts.CompressProgram = val
	case 'V':
		p.opts.VolumeLabel = val
	case 'T':
		p.opts.FilesFrom = val
	}
}

func (p *parser) parseLongOption(arg string, rest []string, idx *int) bool {
	opt := arg[2:]

	var name, val string
	if eq := strings.IndexByte(opt, '='); eq >= 0 {
		name = opt[:eq]
		val = opt[eq+1:]
	} else {
		name = opt
	}

	if !p.longOption(name, val, rest, idx) {
		return false
	}
	return true
}

func (p *parser) longOption(name string, val string, rest []string, idx *int) bool {
	optArg := func(def string) string {
		if val != "" {
			return val
		}
		if *idx+1 < len(rest) {
			(*idx)++
			return rest[*idx]
		}
		return def
	}

	optInt := func() int {
		s := optArg("")
		if s != "" {
			if n, err := strconv.Atoi(s); err == nil {
				return n
			}
		}
		return 0
	}

	optStr := func(def string) string {
		return optArg(def)
	}

	switch name {
	case "create":
		p.setSubcommand(SubCreate)
	case "extract", "get":
		p.setSubcommand(SubExtract)
	case "list":
		p.setSubcommand(SubList)
	case "append":
		p.setSubcommand(SubAppend)
	case "update":
		p.setSubcommand(SubUpdate)
	case "catenate", "concatenate":
		p.setSubcommand(SubConcat)
	case "diff", "compare":
		p.setSubcommand(SubDiff)
	case "delete":
		p.setSubcommand(SubDelete)
	case "test-label":
		p.setSubcommand(SubTestLabel)

	case "file":
		p.opts.ArchiveNames = append(p.opts.ArchiveNames, optStr(""))
	case "blocking-factor":
		n := optInt()
		if n > 0 {
			p.opts.BlockingFactor = n
			p.opts.RecordSize = n * 512
		}
	case "record-size":
		n := optInt()
		if n > 0 {
			p.opts.RecordSize = n
		}
	case "directory":
		p.opts.FileNames = append(p.opts.FileNames, "-C", optStr(""))
	case "tape-length":
		p.opts.TapeLength = parseSize(optStr("0"))
	case "format":
		p.parseFormat(optStr("gnu"))
	case "old-archive", "portability":
		p.opts.ArchiveFormat = FormatV7
	case "posix":
		fallthrough
	case "pax":
		p.opts.ArchiveFormat = FormatPOSIX

	case "gzip", "gunzip", "ungzip":
		p.opts.CompressProgram = "gzip"
	case "bzip2":
		p.opts.CompressProgram = "bzip2"
	case "xz":
		p.opts.CompressProgram = "xz"
	case "zstd":
		p.opts.CompressProgram = "zstd"
	case "compress", "uncompress":
		p.opts.CompressProgram = "compress"
	case "lzip":
		p.opts.CompressProgram = "lzip"
	case "lzma":
		p.opts.CompressProgram = "lzma"
	case "lzop":
		p.opts.CompressProgram = "lzop"
	case "auto-compress":
		p.opts.AutoCompress = true
	case "use-compress-program":
		p.opts.CompressProgram = optStr("")

	case "verbose":
		p.opts.Verbose++
	case "ignore-zeros":
		p.opts.IgnoreZeros = true
	case "read-full-records":
		p.opts.ReadFullRecords = true

	case "keep-old-files":
		p.opts.KeepOldFiles = OldKeepOldFiles
	case "skip-old-files":
		p.opts.KeepOldFiles = OldSkipOldFiles
	case "keep-newer-files":
		p.opts.KeepOldFiles = OldKeepNewerFiles
	case "overwrite":
		p.opts.KeepOldFiles = OldOverwrite
	case "unlink-first":
		p.opts.KeepOldFiles = OldUnlinkFirst
	case "remove-files":
		p.opts.RemoveFiles = true
	case "sparse":
		p.opts.Sparse = true

	case "same-permissions", "preserve-permissions":
		p.opts.SamePermissions = true
	case "same-owner":
		p.opts.SameOwner = true
	case "no-same-owner":
		p.opts.NoSameOwner = true
	case "numeric-owner":
		p.opts.NumericOwner = true
	case "touch":
		p.opts.Touch = true
	case "owner":
		p.opts.Owner = optStr("")
	case "group":
		p.opts.Group = optStr("")
	case "mode":
		p.opts.Mode = optStr("")

	case "absolute-names":
		p.opts.AbsoluteNames = true
	case "dereference":
		p.opts.Dereference = true
	case "to-stdout":
		p.opts.ToStdout = true
	case "block-number":
		p.opts.BlockNumber = true
	case "check-links":
		p.opts.CheckLinks = true
	case "interactive", "confirmation":
		p.opts.Interactive = true
	case "verify":
		p.opts.Verify = true

	case "multi-volume":
		p.opts.MultiVolume = true
	case "info-script":
		p.opts.InfoScript = optStr("")
		p.opts.MultiVolume = true
	case "label":
		p.opts.VolumeLabel = optStr("")
	case "incremental":
		p.opts.Incremental = true
	case "listed-incremental":
		p.opts.ListedIncremental = optStr("")

	case "same-order", "preserve-order":
		p.opts.PreserveOrder = true
	case "seek":
		p.opts.Seek = true
	case "no-seek":
		p.opts.Seek = false
	case "strip-components":
		p.opts.StripComponents = optInt()
	case "transform":
		p.opts.Transform = optStr("")
	case "show-transformed-names":
		p.opts.ShowTransformed = true

	case "exclude":
		p.opts.Exclude = append(p.opts.Exclude, optStr(""))
	case "files-from":
		p.opts.FilesFrom = optStr("")
	case "exclude-from":
		p.opts.ExcludeFrom = optStr("")
	case "exclude-caches":
		p.opts.ExcludeCaches = true
	case "exclude-backups":
		p.opts.ExcludeBackups = true
	case "exclude-vcs":
		p.opts.ExcludeVCS = true

	case "newer", "newer-mtime", "after-date":
		p.parseNewer(optStr(""), false)
	case "starting-file":
		p.opts.StartingFile = optStr("")

	case "totals":
		if val != "" {
			p.opts.ShowTotals = true
			p.opts.TotalsSignal = val
		} else {
			p.opts.ShowTotals = true
		}
	case "checkpoint":
		if val != "" {
			if n, err := strconv.Atoi(val); err == nil {
				p.opts.Checkpoint = n
			}
		} else {
			p.opts.Checkpoint = 10
		}
	case "checkpoint-action":
		p.opts.CheckpointAction = append(p.opts.CheckpointAction, optStr(""))

	case "utc":
		p.opts.Utc = true
	case "full-time":
		p.opts.FullTime = true
	case "index-file":
		p.opts.IndexFile = optStr("")
	case "to-command":
		p.opts.ToCommand = optStr("")
	case "warning":
		p.parseWarning(optStr(""))
	case "quoting-style":
		p.opts.QuotingStyle = optStr("escape")

	case "backup":
		if val != "" {
			p.opts.BackupType = val
		} else {
			p.opts.BackupType = "existing"
		}
	case "suffix":
		p.opts.BackupSuffix = optStr("")

	case "one-file-system":
		p.opts.OneFileSystem = true
	case "one-top-level":
		if val != "" {
			p.opts.OneTopLevel = val
		} else {
			p.opts.OneTopLevel = "."
		}
	case "ignore-failed-read":
		p.opts.IgnoreFailedRead = true
	case "occurrence":
		if val != "" {
			if n, err := strconv.Atoi(val); err == nil {
				p.opts.Occurrence = n
			}
		} else {
			p.opts.Occurrence = 1
		}
	case "restrict":
		p.opts.Restrict = true
	case "force-local":
		p.opts.ForceLocal = true

	case "xattrs":
		p.opts.Xattrs = true
	case "no-xattrs":
		p.opts.Xattrs = false
	case "selinux":
		p.opts.SelinuxCtx = true
	case "no-selinux":
		p.opts.SelinuxCtx = false
	case "acls":
		p.opts.Acls = true
	case "no-acls":
		p.opts.Acls = false
	case "pax-option":
		p.opts.PaxOption = append(p.opts.PaxOption, optStr(""))

	case "mtime":
		p.parseMtime(optStr(""))
	case "clamp-mtime":
		p.opts.SetMtimeMode = MtimeClamp

	case "sort":
		p.opts.SortOrder = optStr("")

	case "delay-directory-restore":
		p.opts.DelayDirRestore = true
	case "no-delay-directory-restore":
		p.opts.DelayDirRestore = false

	case "owner-map":
		p.opts.OwnerMap = optStr("")
	case "group-map":
		p.opts.GroupMap = optStr("")

	case "volno-file":
		p.opts.VolnoFile = optStr("")
	case "level":
		p.opts.IncrementalLevel = optInt()
	case "hole-detection":
		p.opts.HoleDetection = optStr("")

	case "help":
		p.help = true
	case "version":
		p.version = true
	case "usage":
		p.usage = true
	case "show-defaults":
		p.showDefaults = true

	case "xform":
		p.opts.Transform = optStr("")
	case "hard-dereference":
		p.opts.HardDereference = true
	case "recursive-unlink":
		p.opts.RecursiveUnlink = true
	case "keep-directory-symlink":
		p.opts.KeepDirSymlink = true
	case "overwrite-dir":
		p.opts.OverwriteDir = true
	case "check-device":
		p.opts.CheckDevice = true
	case "show-snapshot-ranges":
		p.opts.ShowSnapshotRanges = true
	case "newer-mtime-option":
		p.parseNewer(optStr(""), true)
	case "ignore-command-error":
		p.opts.IgnoreCommandErr = true
	case "quoting-characters":
		p.opts.QuoteChars = optStr("")
	case "no-quoting-characters":
		p.opts.NoQuoteChars = optStr("")
	case "option":
		p.opts.OOption = true
	case "rmt-command":
		p.opts.RmtCommand = optStr("")
	case "rsh-command":
		p.opts.RshCommand = optStr("")
	case "add-file":
		p.opts.FileNames = append(p.opts.FileNames, optStr(""))
	case "no-ignore-case":
	case "no-ignore-command-error":
		p.opts.IgnoreCommandErr = false
	case "no-null":
	case "no-overwrite-dir":
		p.opts.OverwriteDir = false
	case "no-same-permissions":
		p.opts.SamePermissions = false
	case "no-unquote":
	case "warning-option":
		p.parseWarningOption(optStr(""))
	default:
		p.errorMsg = fmt.Sprintf("unrecognized option '--%s'\nTry 'tar --help' or 'tar --usage' for more information.", name)
		return false
	}
	return true
}

func (p *parser) setSubcommand(sub Subcommand) {
	if p.hasSubcommand && p.opts.Subcommand != sub {
		p.errorMsg = "You may not specify more than one '-Acdtrux', '--delete' or '--test-label' option"
	} else {
		p.opts.Subcommand = sub
		p.hasSubcommand = true
	}
}

func (p *parser) parseFormat(fmt string) {
	switch strings.ToLower(fmt) {
	case "v7":
		p.opts.ArchiveFormat = FormatV7
	case "oldgnu":
		p.opts.ArchiveFormat = FormatOldGNU
	case "ustar":
		p.opts.ArchiveFormat = FormatUstar
	case "gnu":
		p.opts.ArchiveFormat = FormatGNU
	case "posix", "pax":
		p.opts.ArchiveFormat = FormatPOSIX
	}
}

func (p *parser) parseNewer(date string, mtimeOnly bool) {
	if date == "" {
		return
	}
	t, err := parseDateTime(date)
	if err != nil {
		return
	}
	p.opts.NewerMtime = t
	p.opts.AfterDate = !mtimeOnly
}

func (p *parser) parseWarning(w string) {
	if w == "" {
		return
	}
	for _, kw := range strings.Split(w, ",") {
		switch strings.TrimSpace(kw) {
		case "all":
			p.opts.Warning |= WarnAll
		case "none":
			p.opts.Warning = 0
		case "alone-zero-block":
			p.opts.Warning |= WarnAloneZeroBlock
		case "bad-dumpdir":
			p.opts.Warning |= WarnBadDumpdir
		case "cachedir":
			p.opts.Warning |= WarnCachedir
		case "contiguous-cast":
			p.opts.Warning |= WarnContiguousCast
		case "file-changed":
			p.opts.Warning |= WarnFileChanged
		case "file-ignored":
			p.opts.Warning |= WarnFileIgnored
		case "file-removed":
			p.opts.Warning |= WarnFileRemoved
		case "file-shrank":
			p.opts.Warning |= WarnFileShrank
		case "file-unchanged":
			p.opts.Warning |= WarnFileUnchanged
		case "filename-with-nuls":
			p.opts.Warning |= WarnFilenameWithNuls
		case "ignore-archive":
			p.opts.Warning |= WarnIgnoreArchive
		case "ignore-newer":
			p.opts.Warning |= WarnIgnoreNewer
		case "new-directory":
			p.opts.Warning |= WarnNewDirectory
		case "rename-directory":
			p.opts.Warning |= WarnRenameDirectory
		case "symlink-cast":
			p.opts.Warning |= WarnSymlinkCast
		case "timestamp":
			p.opts.Warning |= WarnTimestamp
		case "unknown-cast":
			p.opts.Warning |= WarnUnknownCast
		case "unknown-keyword":
			p.opts.Warning |= WarnUnknownKeyword
		case "xdev":
			p.opts.Warning |= WarnXdev
		case "decompress-program":
			p.opts.Warning |= WarnDecompressProgram
		case "existing-file":
			p.opts.Warning |= WarnExistingFile
		case "xattr-write":
			p.opts.Warning |= WarnXattrWrite
		case "record-size":
			p.opts.Warning |= WarnRecordSize
		case "failed-read":
			p.opts.Warning |= WarnFailedRead
		case "missing-zero-blocks":
			p.opts.Warning |= WarnMissingZeroBlocks
		case "empty-transform":
			p.opts.Warning |= WarnEmptyTransform
		}
	}
}

func (p *parser) parseWarningOption(w string) {
	if w == "" {
		return
	}
	if n, err := strconv.Atoi(w); err == nil {
		p.opts.WarningOption = n
	}
}

func (p *parser) parseMtime(m string) {
	if m == "" {
		return
	}
	if m == "atime" {
		p.opts.SetMtimeMode = MtimeCommand
		p.opts.SetMtimeCommand = "atime"
		return
	}
	if m == "ctime" {
		p.opts.SetMtimeMode = MtimeCommand
		p.opts.SetMtimeCommand = "ctime"
		return
	}
	t, err := parseDateTime(m)
	if err == nil {
		p.opts.Mtime = t
		p.opts.SetMtimeMode = MtimeForce
		return
	}
}

func parseDateTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05.999999999 -0700 MST",
		time.RFC3339,
		time.RFC3339Nano,
		time.UnixDate,
		time.RubyDate,
		time.RFC1123,
		time.RFC1123Z,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(n, 0), nil
	}
	return time.Time{}, fmt.Errorf("unparseable date: %s", s)
}

func parseSize(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	mult := int64(1)
	last := s[len(s)-1]
	switch last {
	case 'k', 'K':
		mult = 1024
		s = s[:len(s)-1]
	case 'm', 'M':
		mult = 1024 * 1024
		s = s[:len(s)-1]
	case 'g', 'G':
		mult = 1024 * 1024 * 1024
		s = s[:len(s)-1]
	case 't', 'T':
		mult = 1024 * 1024 * 1024 * 1024
		s = s[:len(s)-1]
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n * mult
	}
	return 0
}
