# tar-go

[![Go Report Card](https://goreportcard.com/badge/github.com/harlandproj/tar-go)](https://goreportcard.com/report/github.com/harlandproj/tar-go)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.22-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-GPLv3-blue.svg)](LICENSE)
[![CI](https://github.com/harlandproj/tar-go/actions/workflows/ci.yml/badge.svg)](https://github.com/harlandproj/tar-go/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/harlandproj/tar-go/branch/main/graph/badge.svg)](https://codecov.io/gh/harlandproj/tar-go)

A complete, pure-Go reimplementation of GNU tar. Zero C dependencies, cross-platform, fully self-contained.

## Features

- **Full GNU tar compatibility** — same CLI, same output, same behavior
- **All compression formats** — gzip, bzip2, xz, zstd, lzma, lzip, lzop (all pure Go)
- **All archive formats** — V7, oldgnu, ustar (POSIX.1-1988), gnu, pax (POSIX.1-2001)
- **All core operations** — create, extract, list, append, update, concatenate, diff, delete
- **Advanced features** — multi-volume, incremental backups, sparse files, name transformation
- **Cross-platform** — Windows (amd64/arm64), Linux (amd64/arm64)
- **CGO_ENABLED=0** — single static binary, no dynamic library dependencies

## Quick Start

```bash
# Build
make build

# Create archive
./bin/tar -cf archive.tar file1 file2 directory/

# Create compressed archive
./bin/tar -czvf archive.tar.gz directory/

# List archive contents
./bin/tar -tvf archive.tar

# Extract archive
./bin/tar -xvf archive.tar
```

## Cross-Platform Build

```bash
make build-all
```

Produces:
- `bin/tar-windows-amd64.exe`
- `bin/tar-windows-arm64.exe`
- `bin/tar-linux-amd64`
- `bin/tar-linux-arm64`

## Supported Options

| Category | Options |
|----------|---------|
| **Operations** | `-c`, `-x`, `-t`, `-r`, `-u`, `-A`, `-d`, `--delete`, `--test-label` |
| **Compression** | `-z` (gzip), `-j` (bzip2), `-J` (xz), `--zstd`, `--lzip`, `--lzma`, `--lzop`, `-a` (auto) |
| **Archive** | `-f FILE`, `-b BLOCKS`, `--format=FORMAT`, `-i`, `-B`, `--record-size` |
| **File Control** | `--exclude=`, `--exclude-from=`, `--exclude-caches`, `--exclude-vcs`, `-C DIR` |
| **Overwrite** | `-k`, `--keep-newer-files`, `--overwrite`, `-U`, `--skip-old-files` |
| **Permissions** | `-p`, `--same-owner`, `--numeric-owner`, `-m`, `--owner=`, `--group=` |
| **Transform** | `--strip-components=N`, `--transform=EXPR`, `--show-transformed-names` |
| **Multi-Volume** | `-M`, `-L SIZE`, `-F SCRIPT`, `--volno-file=` |
| **Incremental** | `-G`, `-g FILE`, `--level=N`, `--listed-incremental=FILE` |
| **Sparse** | `-S`, `--sparse`, `--hole-detection=` |
| **Output** | `-v`, `--totals`, `--checkpoint`, `-R`, `-O`, `--to-command=`, `--index-file=` |
| **Misc** | `--help`, `--version`, `--usage`, `--show-defaults`, `--backup`, `--warning=` |

## Project Structure

```
├── cmd/tar/                     # Entry point & integration tests
├── internal/
│   ├── app/                     # Application orchestration
│   ├── cli/                     # GNU getopt-style CLI parser
│   ├── compress/                # Pure-Go compression codecs
│   ├── filters/                 # File exclusion & name transformation
│   ├── increm/                  # Incremental backup snapshots
│   ├── misc/                    # Shared utilities
│   ├── ops/                     # Core operations (create/extract/list/...)
│   └── vol/                     # Multi-volume support
└── scripts/                     # Build scripts
```

## License

GPLv3 — see [LICENSE](LICENSE)
