package main

import (
	"os"

	"github.com/harlandproj/tar-go/internal/app"
)

var version = "1.35"

func main() {
	os.Exit(app.Run(os.Args, version))
}
