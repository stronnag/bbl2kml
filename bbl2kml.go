package main

import (
	"os"
	"fmt"
	"log"
	"flag"
)

var GitCommit = "local"
var GitTag = "0.0.0"
var BlackboxDecode = "blackbox_decode"

func GetVersion() string {
	return fmt.Sprintf("bbl2kml %s, commit: %s", GitTag, GitCommit)
}

func show_size(sz int64) string {
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

	dump := false
	compress := false
	colrssi := false
	intvl := 1000
	idx := 0

	flag.IntVar(&idx, "index", 0, "Log index")
	flag.IntVar(&intvl, "interval", 1000, "Sampling Interval (ms)")
	flag.BoolVar(&compress, "kmz", false, "Generate KMZ (vice KML)")
	flag.BoolVar(&colrssi, "rssi", false, "Shade according to RSSI%")
	flag.BoolVar(&dump, "dump", false, "Dump headers and exit")
	flag.Parse()

	decoder := os.Getenv("BLACKBOX_DECODE")
	if len(decoder) > 0 {
		BlackboxDecode = decoder
	}

	files := flag.Args()
	if len(files) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if dump {
		bblreader(files[0], 1, 0, true, false, false)
		os.Exit(1)
	}

	for _, fn := range files {
		bmeta, err := GetBBLMeta(fn)
		if err == nil {
			for _, b := range bmeta {
				if (idx == 0 || idx == b.index) && b.size > 4096 {
					fmt.Printf("Log      : %s / %d\n", b.logname, b.index)
					fmt.Printf("Craft    : %s on %s\n", b.craft, b.cdate)
					fmt.Printf("Firmware : %s of %s\n", b.firmware, b.fwdate)
					fmt.Printf("Size     : %s\n", show_size(b.size))
					bblreader(fn, b.index, intvl, false, compress, colrssi)
					fmt.Printf("Disarm   : %s\n\n", b.disarm)
				}
			}
		} else {
			log.Fatal(err)
		}
	}
}
