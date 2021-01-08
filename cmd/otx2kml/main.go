package main

import (
	"os"
	"fmt"
	"path/filepath"
	otx "github.com/stronnag/bbl2kml/pkg/otx"
	options "github.com/stronnag/bbl2kml/pkg/options"
)

var GitCommit = "local"
var GitTag = "0.0.0"

func getVersion() string {
	return fmt.Sprintf("%s %s, commit: %s", filepath.Base(os.Args[0]), GitTag, GitCommit)
}

func main() {
	files := options.ParseCLI(getVersion, true)

	for _, fn := range files {
		res := otx.Reader(fn, true)
		if !res {
			fmt.Fprintf(os.Stderr, "*** skipping OTX with no valid geospatial data\n")
		}
	}
}
