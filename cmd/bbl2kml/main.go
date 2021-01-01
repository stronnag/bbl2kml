package main

import (
	"os"
	"fmt"
	"log"
	"flag"
	bbl "github.com/stronnag/bbl2kml/pkg/bbl"
	options "github.com/stronnag/bbl2kml/pkg/options"
)

var GitCommit = "local"
var GitTag = "0.0.0"

func getVersion() string {
	return fmt.Sprintf("bbl2kml %s, commit: %s", GitTag, GitCommit)
}

func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of bb2kml [options] file...\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintln(os.Stderr, getVersion())
	}

	flag.IntVar(&options.Idx, "index", 0, "Log index")
	flag.IntVar(&options.Intvl, "interval", 1000, "Sampling Interval (ms)")
	flag.BoolVar(&options.Kml, "kml", false, "Generate KML (vice KMZ)")
	flag.BoolVar(&options.Rssi, "rssi", false, "Set RSSI view as default")
	flag.BoolVar(&options.Dump, "dump", false, "Dump headers and exit")
	flag.BoolVar(&options.Dms, "dms", false, "Show positions as DMS (vice decimal degrees)")
	flag.StringVar(&options.Mission, "mission", "", "Mission file name")
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
		bbl.Reader(files[0], bbl.BBLSummary{Index: 1})
		os.Exit(1)
	}

	for _, fn := range files {
		bmeta, err := bbl.Meta(fn)
		if err == nil {
			for _, b := range bmeta {
				if (options.Idx == 0 || options.Idx == b.Index) && b.Size > 4096 {
					fmt.Printf("Log      : %s / %d\n", b.Logname, b.Index)
					fmt.Printf("Craft    : %s on %s\n", b.Craft, b.Cdate)
					fmt.Printf("Firmware : %s of %s\n", b.Firmware, b.Fwdate)
					fmt.Printf("Size     : %s\n", bbl.Show_size(b.Size))
					res := bbl.Reader(fn, b)
					fmt.Printf("Disarm   : %s\n", b.Disarm)
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
