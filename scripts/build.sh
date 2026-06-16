#!/bin/bash
set -e

CGO_ENABLED=0
LDFLAGS="-s -w"

PLATFORMS=(
    "windows/amd64"
    "windows/arm64"
    "linux/amd64"
    "linux/arm64"
)

mkdir -p bin

for platform in "${PLATFORMS[@]}"; do
    GOOS="${platform%/*}"
    GOARCH="${platform#*/}"
    output="bin/tar-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output+='.exe'
    fi
    echo "Building $output..."
    GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=$CGO_ENABLED go build -ldflags="$LDFLAGS" -o "$output" ./cmd/tar
done

echo "All builds complete."
