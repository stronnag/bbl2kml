package main

import (
	"fmt"
	"os"
	"path/filepath"
	otx "github.com/stronnag/bbl2kml/pkg/otx"
	bbl "github.com/stronnag/bbl2kml/pkg/bbl"
	blt "github.com/stronnag/bbl2kml/pkg/bltreader"
	ltom "github.com/stronnag/bbl2kml/pkg/log2mission"
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
	files, _ := options.ParseCLI(getVersion)
	if len(files) == 0 {
		options.Usage()
		os.Exit(1)
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
		} else if ftype == types.IS_BLT {
			blfr := blt.NewBLTReader(fn)
			lfr = &blfr
		} else {
			continue
		}
		metas, err := lfr.GetMetas()
		if err == nil {
			if options.Config.Idx <= len(metas) {
				if options.Config.Idx < 1 {
					options.Config.Idx = 1
				}
				if metas[options.Config.Idx-1].Flags&types.Is_Valid != 0 {
					for k, v := range metas[options.Config.Idx-1].Summary() {
						fmt.Printf("%-8.8s : %s\n", k, v)
					}
					if metas[options.Config.Idx-1].Flags&types.Is_Suspect != 0 {
						fmt.Println("Warning  : Log entry may be corrupt")
					}
					ls, res := lfr.Reader(metas[options.Config.Idx-1])
					if res {
						for k, v := range ls.M {
							fmt.Printf("%-8.8s : %s\n", k, v)
						}
						ltom.Generate_mission(ls, metas[options.Config.Idx-1])
					} else {
						fmt.Fprintf(os.Stderr, "*** skipping generation for log  with no valid geospatial data\n")
					}
					fmt.Println()
				} else {
					fmt.Println("Not valid")
				}
			}
		}
	}
}
