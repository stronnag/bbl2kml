package main

import (
	"os"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	otx "github.com/stronnag/bbl2kml/pkg/otx"
	bbl "github.com/stronnag/bbl2kml/pkg/bbl"
	options "github.com/stronnag/bbl2kml/pkg/options"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	geo "github.com/stronnag/bbl2kml/pkg/geo"
)

var GitCommit = "local"
var GitTag = "0.0.0"

func getVersion() string {
	return fmt.Sprintf("%s %s, commit: %s", filepath.Base(os.Args[0]), GitTag, GitCommit)
}

func main() {
	files := options.ParseCLI(getVersion)
	if len(files) == 0 {
		options.Usage()
		os.Exit(1)
	}

	geo.Frobnicate_init()

	var lfr types.FlightLog
	for _, fn := range files {
		ext := filepath.Ext(fn)
		if strings.EqualFold(ext, ".csv") {
			olfr := otx.NewOTXReader(fn)
			lfr = &olfr
		} else {
			blfr := bbl.NewBBLReader(fn)
			lfr = &blfr
		}
		metas, err := lfr.GetMetas()
		if err == nil {
			if options.Dump {
				lfr.Dump()
				os.Exit(0)
			}
			for _, b := range metas {
				if (options.Idx == 0 || options.Idx == b.Index) && b.Flags&types.Is_Valid != 0 {
					for k, v := range b.Summary() {
						fmt.Printf("%-8.8s : %s\n", k, v)
					}
					smap, res := lfr.Reader(b)
					for k, v := range smap {
						fmt.Printf("%-8.8s : %s\n", k, v)
					}
					if s, ok := b.ShowDisarm(); ok {
						fmt.Printf("%-8.8s : %s\n", "Disarm", s)
					}
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
