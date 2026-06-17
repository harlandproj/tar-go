# CI Pipeline Design

## Goal

Add GitHub Actions CI pipeline with test, lint, and coverage reporting via Codecov.

## Decisions

- **Trigger**: Push to all branches + all PRs
- **Platform matrix**: `ubuntu-latest` + `windows-latest`
- **Go version**: 1.22 only (matches go.mod)
- **Coverage**: Codecov, generated only on Linux
- **Approach**: Single workflow with matrix (Option A)

## Workflow Structure

**File**: `.github/workflows/ci.yml`

### Job: test (matrix: ubuntu-latest, windows-latest)

Steps:
1. `actions/checkout@v4`
2. `actions/setup-go@v5` with go-version `1.22` and `cache: true`
3. `go vet ./...`
4. `go test -race ./...`
5. (Linux only) `go test -race -coverprofile=coverage.out -covermode=atomic ./...`
6. (Linux only) `codecov/codecov-action@v4`

### README Badges

Add after existing badges:
- CI: `[![CI](https://github.com/harlandproj/tar-go/actions/workflows/ci.yml/badge.svg)]`
- Coverage: `[![Coverage](https://codecov.io/gh/harlandproj/tar-go/branch/main/graph/badge.svg)]`

## Notes

- Windows only runs lint + test (no coverage due to path issues with coverprofile)
- Codecov token: set `CODECOV_TOKEN` as repository secret (recommended for v4 action)
- `go vet` failure stops the job before test runs
