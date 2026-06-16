package cli

import (
	"os"
	"time"
)

type Subcommand int

const (
	SubNone Subcommand = iota
	SubCreate
	SubExtract
	SubList
	SubAppend
	SubUpdate
	SubConcat
	SubDiff
	SubDelete
	SubTestLabel
)

type ArchiveFormat int

const (
	FormatDefault ArchiveFormat = iota
	FormatV7
	FormatOldGNU
	FormatUstar
	FormatGNU
	FormatPOSIX
)

type OldFiles int

const (
	OldDefault OldFiles = iota
	OldNoOverwriteDir
	OldOverwrite
	OldUnlinkFirst
	OldKeepOldFiles
	OldSkipOldFiles
	OldKeepNewerFiles
)

type AtimePreserve int

const (
	AtimeNone AtimePreserve = iota
	AtimeReplace
	AtimeSystem
)

type SetMtimeMode int

const (
	MtimeFile SetMtimeMode = iota
	MtimeForce
	MtimeClamp
	MtimeCommand
)

type Warning int64

const (
	WarnAloneZeroBlock    Warning = 1 << 0
	WarnBadDumpdir        Warning = 1 << 1
	WarnCachedir          Warning = 1 << 2
	WarnContiguousCast    Warning = 1 << 3
	WarnFileChanged       Warning = 1 << 4
	WarnFileIgnored       Warning = 1 << 5
	WarnFileRemoved       Warning = 1 << 6
	WarnFileShrank        Warning = 1 << 7
	WarnFileUnchanged     Warning = 1 << 8
	WarnFilenameWithNuls  Warning = 1 << 9
	WarnIgnoreArchive     Warning = 1 << 10
	WarnIgnoreNewer       Warning = 1 << 11
	WarnNewDirectory      Warning = 1 << 12
	WarnRenameDirectory   Warning = 1 << 13
	WarnSymlinkCast       Warning = 1 << 14
	WarnTimestamp         Warning = 1 << 15
	WarnUnknownCast       Warning = 1 << 16
	WarnUnknownKeyword    Warning = 1 << 17
	WarnXdev              Warning = 1 << 18
	WarnDecompressProgram Warning = 1 << 19
	WarnExistingFile      Warning = 1 << 20
	WarnXattrWrite        Warning = 1 << 21
	WarnRecordSize        Warning = 1 << 22
	WarnFailedRead        Warning = 1 << 23
	WarnMissingZeroBlocks Warning = 1 << 24
	WarnEmptyTransform    Warning = 1 << 25
	WarnAll               Warning = ^Warning(0)
)

var WarnVerbose = WarnRenameDirectory | WarnNewDirectory | WarnDecompressProgram | WarnExistingFile | WarnRecordSize

type Options struct {
	Subcommand    Subcommand
	ArchiveFormat ArchiveFormat

	ArchiveNames []string
	FileNames    []string

	BlockingFactor  int
	RecordSize      int
	IgnoreZeros     bool
	ReadFullRecords bool

	CompressProgram string
	AutoCompress    bool

	Verbose      int
	ShowTotals   bool
	TotalsSignal string
	CheckLinks   bool
	Utc          bool
	FullTime     bool
	BlockNumber  bool
	Interactive  bool
	Checkpoint   int
	CheckpointAction []string

	IndexFile        string
	ToStdout         bool
	ToCommand        string
	IgnoreCommandErr bool

	KeepOldFiles    OldFiles
	RecursiveUnlink bool
	KeepDirSymlink  bool
	OneTopLevel     string
	OverwriteDir    bool

	SamePermissions bool
	SameOwner       bool
	NoSameOwner     bool
	NumericOwner    bool
	Touch           bool
	Owner           string
	Group           string
	OwnerMap        string
	GroupMap        string
	Mode            string

	PreserveOrder bool

	AbsoluteNames   bool
	Dereference     bool
	HardDereference bool
	OneFileSystem   bool

	RemoveFiles bool
	Seek        bool

	Sparse        bool
	HoleDetection string
	SparseVersion string

	StripComponents int
	Transform       string
	ShowTransformed bool

	Exclude        []string
	ExcludeFrom    string
	ExcludeCaches  bool
	ExcludeBackups bool
	ExcludeVCS     bool

	NewerMtime   time.Time
	AfterDate    bool
	StartingFile string

	MultiVolume bool
	TapeLength  int64
	InfoScript  string
	VolnoFile   string
	VolumeLabel string

	Incremental       bool
	ListedIncremental string
	IncrementalLevel  int
	CheckDevice       bool

	IgnoreFailedRead bool
	Occurrence       int
	Verify           bool
	BackupType       string
	BackupSuffix     string
	Restrict         bool

	Xattrs     bool
	SelinuxCtx bool
	Acls       bool
	PaxOption  []string

	Mtime           time.Time
	SetMtimeMode    SetMtimeMode
	SetMtimeCommand string
	SetMtimeFormat  string

	QuotingStyle string
	QuoteChars   string
	NoQuoteChars string
	OOption      bool

	Warning            Warning
	ShowDefaults       bool
	ShowSnapshotRanges bool

	ForceLocal bool
	RmtCommand string
	RshCommand string

	SortOrder       string
	DelayDirRestore bool

	WarningOption int
}

var (
	Stderr = os.Stderr
	Stdout = os.Stdout
)

var ErrHelpRequested = &exitError{code: 0}

type exitError struct{ code int }

func (e *exitError) Error() string { return "" }
