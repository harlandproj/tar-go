package main

import (
	"os"

	"github.com/harlandproj/tar-go/internal/app"
)

var version = "0.1.0"

func main() {
	os.Exit(app.Run(os.Args, version))
}
