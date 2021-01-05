package main

import (
	"os"
	"fmt"
	"flag"
	"strings"
	"strconv"
	"path/filepath"
	otx "github.com/stronnag/bbl2kml/pkg/otx"
	options "github.com/stronnag/bbl2kml/pkg/options"
)

var GitCommit = "local"
var GitTag = "0.0.0"

type IFlag struct {
	v   int
	set bool
}

func (v IFlag) String() string {
	return fmt.Sprintf("%d", v.v)
}

func (v *IFlag) Set(s string) error {
	if u, err := strconv.Atoi(s); err != nil {
		return err
	} else {
		v.v = u
		v.set = true
	}
	return nil
}

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
	options.Kml = strings.Contains(defs, "-kml")
	options.Rssi = strings.Contains(defs, "-rssi")

	homealt := IFlag{}

	//	flag.IntVar(&options.HomeAlt, "home-alt", 0, "home altitude")
	flag.Var(&homealt, "home-alt", "Home altitude (m)")
	flag.IntVar(&options.Intvl, "interval", 1000, "Sampling Interval (ms)")
	flag.BoolVar(&options.Kml, "kml", options.Kml, "Generate KML (vice default KMZ)")
	flag.BoolVar(&options.Rssi, "rssi", options.Rssi, "Set RSSI view as default")
	flag.BoolVar(&options.Dms, "dms", options.Dms, "Show positions as DD:MM:SS.s (vice decimal degrees)")
	flag.BoolVar(&options.Extrude, "extrude", options.Extrude, "Extends track points to ground")
	flag.IntVar(&options.SplitTime, "split-time", 120, "Time(s) determining log split, 0 disables")
	flag.StringVar(&options.Mission, "mission", "", "Optional mission file name")
	flag.Parse()

	if homealt.set {
		options.HomeAlt = homealt.v
	}

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
