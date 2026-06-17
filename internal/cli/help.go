package cli

import (
	"fmt"
	"runtime"
)

func PrintHelp(prog string) {
	fmt.Printf(`Usage: %s [OPTION...] [FILE]...
tar-go is a pure-Go reimplementation of GNU tar. It saves many files together
into a single archive, and can restore individual files from the archive.

Examples:
  tar -cf archive.tar foo bar    Create archive.tar from files foo and bar.
  tar -tvf archive.tar           List all files in archive.tar verbosely.
  tar -xf archive.tar            Extract all files from archive.tar.

`, prog)

	mainOpts := [][2]string{
		{"-A, --catenate, --concatenate", "append tar files to an archive"},
		{"-c, --create", "create a new archive"},
		{"-d, --diff, --compare", "find differences between archive and file system"},
		{"--delete", "delete from the archive (not on mag tapes!)"},
		{"-r, --append", "append files to the end of an archive"},
		{"-t, --list", "list the contents of an archive"},
		{"--test-label", "test the archive volume label and exit"},
		{"-u, --update", "only append files newer than copy in archive"},
		{"-x, --extract, --get", "extract files from an archive"},
	}
	fmt.Println(" Main operation mode:")
	for _, o := range mainOpts {
		fmt.Printf("  %-38s %s\n", o[0], o[1])
	}
	fmt.Println()

	otherOpts := [][2]string{
		{"--backup[=CONTROL]", "backup before removal, choose version CONTROL"},
		{"-b, --blocking-factor=BLOCKS", "BLOCKS x 512 bytes per record"},
		{"-B, --read-full-records", "reblock as we read (for 4.2BSD pipes)"},
		{"--checkpoint[=NUMBER]", "display progress messages every NUMBERth record (default 10)"},
		{"--checkpoint-action=ACTION", "execute ACTION on each checkpoint"},
		{"-C, --directory=DIR", "change to directory DIR"},
		{"--delay-directory-restore", "delay setting modification times and permissions of extracted dirs until end of extraction"},
		{"--exclude=PATTERN", "exclude files, given as a PATTERN"},
		{"--exclude-backups", "exclude backup and lock files"},
		{"--exclude-caches", "exclude contents of directories containing CACHEDIR.TAG, except for the tag itself"},
		{"--exclude-from=FILE", "exclude patterns listed in FILE"},
		{"--exclude-vcs", "exclude version control system directories"},
		{"-f, --file=ARCHIVE", "use archive file or device ARCHIVE"},
		{"-F, --info-script=NAME, --new-volume-script=NAME", "run script at end of each tape (implies -M)"},
		{"--force-local", "archive file is local even if it has a colon"},
		{"--full-time", "print file time to its full resolution"},
		{"-g, --listed-incremental=FILE", "handle new GNU-format incremental backup"},
		{"-G, --incremental", "handle old GNU-format incremental backup"},
		{"--group=NAME", "force NAME as group for added files"},
		{"-h, --dereference", "follow symlinks; archive and dump the files they point to"},
		{"--hole-detection=TYPE", "technique to detect holes"},
		{"--help", "show this help message"},
		{"-i, --ignore-zeros", "ignore zeroed blocks in archive (means EOF)"},
		{"-I, --use-compress-program=PROG", "filter through PROG (must accept -d)"},
		{"-k, --keep-old-files", "don't replace existing files when extracting, treat them as errors"},
		{"-K, --starting-file=MEMBER-NAME", "begin at member MEMBER-NAME when reading the archive"},
		{"-l, --check-links", "print a message if not all links are dumped"},
		{"-L, --tape-length=N", "change tape after writing N x 1024 bytes"},
		{"-m, --touch", "don't extract file modified time"},
		{"-M, --multi-volume", "create/list/extract multi-volume archive"},
		{"--mode=CHANGES", "force (symbolic) mode CHANGES for added files"},
		{"--mtime=DATE", "set mtime for added files"},
		{"-n, --seek", "archive is seekable"},
		{"-N, --newer=DATE, --after-date=DATE", "only store files newer than DATE"},
		{"--newer-mtime=DATE", "compare date and time when data changed only"},
		{"--no-delay-directory-restore", "cancel the effect of --delay-directory-restore option"},
		{"--numeric-owner", "always use numbers for user/group names"},
		{"-O, --to-stdout", "extract files to standard output"},
		{"--occurrence[=NUMBER]", "process only the NUMBERth occurrence of each file"},
		{"--old-archive, --portability", "same as --format=v7"},
		{"--one-file-system", "stay in local file system when creating archive"},
		{"--one-top-level[=DIR]", "extract everything into DIR, or create a subdirectory named by the archive basename"},
		{"--overwrite", "overwrite existing files when extracting"},
		{"--owner=NAME", "force NAME as owner for added files"},
		{"-p, --same-permissions, --preserve-permissions", "extract information about file permissions (default for superuser)"},
		{"--pax-option=keyword[[:]=value][,keyword[[:]=value]]...", "control pax keywords"},
		{"-P, --absolute-names", "don't strip leading '/'s from file names"},
		{"-R, --block-number", "show block number within archive with each message"},
		{"--record-size=NUMBER", "NUMBER of bytes per record, multiple of 512"},
		{"--remove-files", "remove files after adding them to the archive"},
		{"--restrict", "disable use of some potentially harmful options"},
		{"--rmt-command=COMMAND", "use given rmt COMMAND instead of rmt"},
		{"--rsh-command=COMMAND", "use remote COMMAND instead of rsh"},
		{"-s, --same-order, --preserve-order", "member arguments are listed in the same order as the files in the archive"},
		{"-S, --sparse", "handle sparse files efficiently"},
		{"--same-owner", "try extracting files with the same ownership as exists in the archive (default for superuser)"},
		{"--show-defaults", "show built-in defaults for various tar options"},
		{"--show-snapshot-ranges", "show ranges of fields in the incremental snapshot file"},
		{"--show-transformed-names", "show file or archive names after transformation"},
		{"--skip-old-files", "don't replace existing files when extracting, silently skip over them"},
		{"--sort=ORDER", "directory sorting order: none (default), name, or inode"},
		{"--strip-components=NUMBER", "strip NUMBER leading components from file names on extraction"},
		{"--suffix=STRING", "backup before removal, override usual suffix (default '~')"},
		{"-T, --files-from=FILE", "get names to extract or create from FILE"},
		{"--to-command=COMMAND", "pipe extracted files to another program"},
		{"--totals[=SIGNAL]", "print total bytes after processing the archive"},
		{"--transform=EXPRESSION", "use sed replace EXPRESSION to transform file names"},
		{"-U, --unlink-first", "remove each file prior to extracting over it"},
		{"--utc", "print file modification times in UTC"},
		{"-v, --verbose", "verbosely list files processed"},
		{"-V, --label=TEXT", "create archive with volume name TEXT; at list/extract, use TEXT as a globbing pattern for volume name"},
		{"--version", "print program version"},
		{"--volno-file=FILE", "use/update the volume number in FILE"},
		{"-w, --interactive, --confirmation", "ask for confirmation for every action"},
		{"-W, --verify", "attempt to verify the archive after writing it"},
		{"--warning=KEYWORD", "warning control"},
		{"--xattrs", "Enable extended attributes support"},
		{"--xattrs-exclude=MASK", "specify the exclude pattern for xattr keys"},
		{"--xattrs-include=MASK", "specify the include pattern for xattr keys"},
		{"-z, --gzip, --gunzip, --ungzip", "filter the archive through gzip"},
		{"-Z, --compress, --uncompress", "filter the archive through compress"},
		{"-j, --bzip2", "filter the archive through bzip2"},
		{"-J, --xz", "filter the archive through xz"},
		{"--lzip", "filter the archive through lzip"},
		{"--lzma", "filter the archive through lzma"},
		{"--lzop", "filter the archive through lzop"},
		{"--zstd", "filter the archive through zstd"},
	}
	otherSection := false
	for _, o := range otherOpts {
		if !otherSection {
			fmt.Println(" Operation modifiers:")
			otherSection = true
		}
		fmt.Printf("  %-38s %s\n", o[0], o[1])
	}
	fmt.Println()

	localOpts := [][2]string{
		{"-a, --auto-compress", "use archive suffix to determine the compression program"},
		{"--add-file=FILE", "add given FILE to the archive (useful if its name starts with a dash)"},
		{"--clamp-mtime", "only set time when the file is more recent than what was given with --mtime"},
		{"--format=FORMAT", "archive format to use for writing (v7, oldgnu, ustar, gnu, posix)"},
		{"--hard-dereference", "follow hard links; archive and dump the files they refer to"},
		{"--keep-directory-symlink", "preserve existing symlinks to directories when extracting"},
		{"--keep-newer-files", "don't replace existing files that are newer than their archive copies"},
		{"-n, --seek", "archive is seekable"},
		{"--no-acls", "Disable the POSIX ACLs support"},
		{"--no-delay-directory-restore", "cancel the effect of --delay-directory-restore option"},
		{"--no-ignore-case", "use case-sensitive matching"},
		{"--no-ignore-command-error", "process non-zero exit codes from child programs on error"},
		{"--no-null", "disable the effect of the previous --null option"},
		{"--no-overwrite-dir", "preserve metadata of existing directories"},
		{"--no-same-owner", "extract files as yourself (default for ordinary users)"},
		{"--no-same-permissions", "apply the user's umask when extracting permissions from the archive (default for ordinary users)"},
		{"--no-seek", "archive is not seekable"},
		{"--no-selinux", "Disable the SELinux context support"},
		{"--no-unquote", "do not unquote input file or member names"},
		{"--no-xattrs", "Disable extended attributes support"},
		{"--null", "instruct subsequent -T options to read null-terminated names, -C to use null-terminated names"},
		{"--owner-map=FILE", "use FILE to map file owner UIDs and names"},
		{"--group-map=FILE", "use FILE to map file owner GIDs and names"},
		{"--pax-option=keyword[[:]=value][,keyword[[:]=value]]...", "control pax keywords"},
		{"--quoting-style=STYLE", "set name quoting style; valid STYLE values are: literal, shell, shell-always, c, c-maybe, escape, locale, clocale"},
		{"--recursive-unlink", "empty hierarchies prior to extracting directory"},
		{"--record-size=SIZE", "record size"},
		{"--transform=EXPRESSION, --xform=EXPRESSION", "use sed replace EXPRESSION to transform file names"},
		{"--xattrs-exclude=MASK", "specify the exclude pattern for xattr keys"},
	}
	for _, o := range localOpts {
		fmt.Printf("  %-38s %s\n", o[0], o[1])
	}
	fmt.Println()

	compOpts := [][2]string{
		{"-j, --bzip2", "filter the archive through bzip2"},
		{"-J, --xz", "filter the archive through xz"},
		{"--lzip", "filter the archive through lzip"},
		{"--lzma", "filter the archive through lzma"},
		{"--lzop", "filter the archive through lzop"},
		{"-z, --gzip, --gunzip, --ungzip", "filter the archive through gzip"},
		{"-Z, --compress, --uncompress", "filter the archive through compress"},
		{"--zstd", "filter the archive through zstd"},
	}
	fmt.Println(" Compression options:")
	for _, o := range compOpts {
		fmt.Printf("  %-38s %s\n", o[0], o[1])
	}
	fmt.Println()

	fmt.Println(` The backup suffix is '~', unless set with --suffix or SIMPLE_BACKUP_SUFFIX.
 The version control may be set with --backup or VERSION_CONTROL, values are:

   none, off         never make backups
   t, numbered       make numbered backups
   nil, existing     numbered if numbered backups exist, simple otherwise
   never, simple     always make simple backups`)
}

func PrintVersion(prog string) {
	fmt.Printf("tar (tar-go) 1.35\n")
	fmt.Printf("Copyright (C) 2026 Harland Wang\n")
	fmt.Printf("License GPLv3+: GNU GPL version 3 or later <https://gnu.org/licenses/gpl.html>.\n")
	fmt.Printf("This is free software: you are free to change and redistribute it.\n")
	fmt.Printf("There is NO WARRANTY, to the extent permitted by law.\n")
	fmt.Printf("\nA pure-Go reimplementation, compatible with GNU tar.\n")
	fmt.Printf("Built with %s\n", runtime.Version())
}

func PrintUsage(prog string) {
	fmt.Printf("Usage: %s [OPTION...] [FILE]...\n", prog)
	fmt.Printf("Try 'tar --help' for more information.\n")
}

func PrintDefaults() {
	fmt.Printf("--format=gnu -f- -b20 --quoting-style=escape --rmt-command=/usr/sbin/rmt\n")
	fmt.Printf("--rsh-command=/usr/bin/rsh\n")
}
