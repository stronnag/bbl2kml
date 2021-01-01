package main

import (
	"os"
	"fmt"
	"log"
	"flag"
)

var GitCommit = "local"
var GitTag = "0.0.0"

var Options = struct {
	dms             bool
	dump            bool
	kml             bool
	rssi            bool
	intvl           int
	idx             int
	blackbox_decode string
}{false, false, false, false, 1000, 0, "blackbox_decode"}


func GetVersion() string {
	return fmt.Sprintf("bbl2kml %s, commit: %s", GitTag, GitCommit)
}

func Show_size(sz int64) string {
	var s string
	switch {
	case sz > 1024*1024:
		s = fmt.Sprintf("%.2f MB", float64(sz)/(1024*1024))
	case sz > 10*1024:
		s = fmt.Sprintf("%.1f KB", float64(sz)/1024)
	default:
		s = fmt.Sprintf("%d B", sz)
	}
	return s
}

func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of bb2kml [options] file...\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintln(os.Stderr, GetVersion())
	}

	flag.IntVar(&Options.idx, "index", 0, "Log index")
	flag.IntVar(&Options.intvl, "interval", 1000, "Sampling Interval (ms)")
	flag.BoolVar(&Options.kml, "kml", false, "Generate KML (vice KMZ)")
	flag.BoolVar(&Options.rssi, "rssi", false, "Set RSSI view as default")
	flag.BoolVar(&Options.dump, "dump", false, "Dump headers and exit")
	flag.BoolVar(&Options.dms, "dms", false, "Show positions as DMS (vice decimal degrees)")
	flag.Parse()

	decoder := os.Getenv("BLACKBOX_DECODE")
	if len(decoder) > 0 {
		Options.blackbox_decode = decoder
	}

	files := flag.Args()
	if len(files) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if Options.dump {
		bblreader(files[0], BBLSummary{index: 1})
		os.Exit(1)
	}

	for _, fn := range files {
		bmeta, err := GetBBLMeta(fn)
		if err == nil {
			for _, b := range bmeta {
				if (Options.idx == 0 || Options.idx == b.index) && b.size > 4096 {
					fmt.Printf("Log      : %s / %d\n", b.logname, b.index)
					fmt.Printf("Craft    : %s on %s\n", b.craft, b.cdate)
					fmt.Printf("Firmware : %s of %s\n", b.firmware, b.fwdate)
					fmt.Printf("Size     : %s\n", Show_size(b.size))
					res := bblreader(fn, b)
					fmt.Printf("Disarm   : %s\n", b.disarm)
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
