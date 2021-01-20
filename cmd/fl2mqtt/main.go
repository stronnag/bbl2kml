package main

import (
	"os"
	"fmt"
	"log"
	"path/filepath"
	otx "github.com/stronnag/bbl2kml/pkg/otx"
	bbl "github.com/stronnag/bbl2kml/pkg/bbl"
	options "github.com/stronnag/bbl2kml/pkg/options"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	geo "github.com/stronnag/bbl2kml/pkg/geo"
	mqttgen "github.com/stronnag/bbl2kml/pkg/bltmqtt"
)

var GitCommit = "local"
var GitTag = "0.0.0"

func getVersion() string {
	return fmt.Sprintf("%s %s, commit: %s", filepath.Base(os.Args[0]), GitTag, GitCommit)
}

func main() {
	files := options.ParseCLI(getVersion)
	if len(files) == 0 || len(options.Mqttopts) == 0 {
		options.Usage()
		os.Exit(1)
	}

	if options.Idx == 0 {
		options.Idx = 1
	}

	geo.Frobnicate_init()
	var lfr types.FlightLog
	for _, fn := range files {
		ftype := types.EvinceFileType(fn)
		if ftype == types.IS_OTX {
			olfr := otx.NewOTXReader(fn)
			lfr = &olfr
		} else if ftype == types.IS_BBL {
			blfr := bbl.NewBBLReader(fn)
			lfr = &blfr
		} else {
			continue
		}
		metas, err := lfr.GetMetas()
		if err == nil {
			if options.Dump {
				lfr.Dump()
				os.Exit(0)
			}
			if options.Idx <= len(metas) {
				if metas[options.Idx-1].Flags&types.Is_Valid != 0 {
					for k, v := range metas[options.Idx-1].Summary() {
						fmt.Printf("%-8.8s : %s\n", k, v)
					}
					ls, res := lfr.Reader(metas[options.Idx-1])
					if res {
						mqttgen.MQTTGen(ls)
					} else {
						fmt.Fprintf(os.Stderr, "*** skipping generation for log  with no valid geospatial data\n")
					}
					fmt.Println()
				}
			}
		} else {
			log.Fatal(err)
		}
	}
}