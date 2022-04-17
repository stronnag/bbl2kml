package main

import (
	"flag"
	"fmt"
	mission "github.com/stronnag/bbl2kml/pkg/mission"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var GitCommit = "local"
var GitTag = "0.0.0"

var (
	dms     bool
	homepos string
	idx     int
)

func getVersion() string {
	return fmt.Sprintf("%s %s commit:%s", filepath.Base(os.Args[0]), GitTag, GitCommit)
}

func split(s string, separators []rune) []string {
	f := func(r rune) bool {
		for _, s := range separators {
			if r == s {
				return true
			}
		}
		return false
	}
	return strings.FieldsFunc(s, f)
}

func main() {
	flag.Usage = func() {
		extra := `The home location is given as decimal degrees latitude and
longitude and optional altitude. The values should be separated by a single
separator, one of "/:; ,". If space is used, then the values must be enclosed
in quotes.

In locales where comma is used as decimal "point", then comma should not be
used as a separator.

If a syntactically valid home position is given, without altitude, an online
elevation service is used to adjust mission elevations in the KML.

Examples:
    -home 54.353974/-4.5236
    --home 48,9975:2,5789/104
    -home 54.353974;-4.5236
    --home "48,9975 2,5789"
    -home 54.353974,-4.5236,24
`
		fmt.Fprintf(os.Stderr, "Usage of %s [options] mission_file\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, extra)
		fmt.Fprintln(os.Stderr, getVersion())
	}

	defs := os.Getenv("BBL2KML_OPTS")
	dms = strings.Contains(defs, "-dms")

	flag.BoolVar(&dms, "dms", dms, "Show positions as DMS (vice decimal degrees)")
	flag.StringVar(&homepos, "home", homepos, "Use home location")
	flag.IntVar(&idx, "mission-index", 1, "Mission Index")
	flag.Parse()
	files := flag.Args()
	if len(files) == 0 {
		flag.Usage()
		os.Exit(-1)
	}

	var home []float64
	var v float64

	if len(homepos) > 0 {
		parts := split(homepos, []rune{'/', ':', ';', ' ', ','})
		if len(parts) >= 2 {
			var err error
			v, err = strconv.ParseFloat(parts[0], 64)
			if err == nil {
				home = append(home, v)
				v, err = strconv.ParseFloat(parts[1], 64)
				if err == nil {
					home = append(home, v)
					if len(parts) == 3 {
						v, err = strconv.ParseFloat(parts[2], 64)
						if err == nil {
							home = append(home, v)
						}
					}
				}
			}
		}
	}

	_, m, err := mission.Read_Mission_File_Index(files[0], idx)
	if m != nil && err == nil {
		if len(home) == 0 {
			if m.Metadata.Homey != 0 && m.Metadata.Homex != 0 {
				home = append(home, m.Metadata.Homey, m.Metadata.Homex)
			}
		}

		m.Dump(dms, home...)
	}
	if err != nil {
		log.Fatalf("mission2kmk: %+v\n", err)
	}
}
