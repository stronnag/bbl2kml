package main

import (
	"os"
	"fmt"
	"log"
	"flag"
	"strings"
	"path/filepath"
	bbl "github.com/stronnag/bbl2kml/pkg/bbl"
	options "github.com/stronnag/bbl2kml/pkg/options"
)

var GitCommit = "local"
var GitTag = "0.0.0"

func getVersion() string {
	return fmt.Sprintf("%s %s, commit: %s", filepath.Base(os.Args[0]), GitTag, GitCommit)
}

func main() {

	flag.Usage = func() {
		app := filepath.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage of %s [options] file...\n", app)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintln(os.Stderr, getVersion())
	}

	defs := os.Getenv("BBL2KML_OPTS")
	options.Dms = strings.Contains(defs, "-dms")
	options.Kml = strings.Contains(defs, "-kml")
	options.Rssi = strings.Contains(defs, "-rssi")
	options.Extrude = strings.Contains(defs, "-extrude")

	flag.IntVar(&options.Idx, "index", 0, "Log index")
	flag.IntVar(&options.Intvl, "interval", 1000, "Sampling Interval (ms)")
	flag.BoolVar(&options.Kml, "kml", options.Kml, "Generate KML (vice default KMZ)")
	flag.BoolVar(&options.Rssi, "rssi", options.Rssi, "Set RSSI view as default")
	flag.BoolVar(&options.Extrude, "extrude", options.Extrude, "Extends track points to ground")
	flag.BoolVar(&options.Dump, "dump", false, "Dump log headers and exit")
	flag.BoolVar(&options.Dms, "dms", options.Dms, "Show positions as DD:MM:SS.s (vice decimal degrees)")
	flag.StringVar(&options.Mission, "mission", "", "Optional mission file name")
	flag.Parse()

	decoder := os.Getenv("BLACKBOX_DECODE")
	if len(decoder) > 0 {
		options.Blackbox_decode = decoder
	}

	files := flag.Args()
	if len(files) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if options.Dump {
		bbl.Reader(files[0], bbl.BBLMeta{Index: 1})
		os.Exit(1)
	}

	for _, fn := range files {
		bmeta, err := bbl.Meta(fn)
		if err == nil {
			for _, b := range bmeta {
				if (options.Idx == 0 || options.Idx == b.Index) && b.Size > 4096 {
					m := b.MetaData()
					for _, k := range []string{"Log", "Flight", "Firmware", "Size"} {
						if v, ok := m[k]; ok {
							fmt.Printf("%-8.8s : %s\n", k, v)
						}
					}
					res := bbl.Reader(fn, b)
					fmt.Printf("%-8.8s : %s\n", "Disarm", m["Disarm"])
					if !res {
						fmt.Fprintf(os.Stderr, "*** skipping KML/Z for log  with no valid geospatial data\n")
					}
					fmt.Println()
				}
			}
		} else {
			log.Fatal(err)
		}
	}
}
