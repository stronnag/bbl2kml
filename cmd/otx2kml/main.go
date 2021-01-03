package main

import (
	"os"
	"fmt"
	"flag"
	"strings"
	"path/filepath"
	otx "github.com/stronnag/bbl2kml/pkg/otx"
	options "github.com/stronnag/bbl2kml/pkg/options"
	//types "github.com/stronnag/bbl2kml/pkg/api/types"
)

var GitCommit = "local"
var GitTag = "0.0.0"

func getVersion() string {
	return fmt.Sprintf("%s %s, commit: %s", filepath.Base(os.Args[0]), GitTag, GitCommit)
}

func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s [options] file...\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintln(os.Stderr, getVersion())
	}

	defs := os.Getenv("BBL2KML_OPTS")
	options.Dms = strings.Contains(defs, "-dms")
	options.Elev = strings.Contains(defs, "-elev")
	options.Kml = strings.Contains(defs, "-kml")
	options.Rssi = strings.Contains(defs, "-rssi")

	flag.IntVar(&options.Intvl, "interval", 1000, "Sampling Interval (ms)")
	flag.BoolVar(&options.Kml, "kml", false, "Generate KML (vice default KMZ)")
	flag.BoolVar(&options.Rssi, "rssi", false, "Set RSSI view as default")
	flag.BoolVar(&options.Dms, "dms", false, "Show positions as DD:MM:SS.s (vice decimal degrees)")
	flag.StringVar(&options.Mission, "mission", "", "Optional mission file name")
	flag.BoolVar(&options.Elev, "elev", false, "Use an online elevation service to adjust mission evelations to above terrain")
	flag.Parse()

	files := flag.Args()
	if len(files) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	for _, fn := range files {
		res := otx.Reader(fn, true)
		if !res {
			fmt.Fprintf(os.Stderr, "*** skipping OTX with no valid geospatial data\n")
		}
	}
}
