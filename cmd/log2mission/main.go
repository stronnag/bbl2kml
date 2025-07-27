package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

import (
	"aplog"
	"bbl"
	"bltlog"
	"geo"
	ltom "log2mission"
	"mwpjson"
	"options"
	"otx"
	"types"
)

var GitCommit = "local"
var GitTag = "0.0.0"

func getVersion() string {
	return fmt.Sprintf("%s %s commit:%s", filepath.Base(os.Args[0]), GitTag, GitCommit)
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
		switch ftype {
		case types.IS_OTX:
			l := otx.NewOTXReader(fn)
			lfr = &l
		case types.IS_BBL:
			l := bbl.NewBBLReader(fn)
			lfr = &l
		case types.IS_BLT:
			l := bltlog.NewBLTReader(fn)
			lfr = &l
		case types.IS_AP:
			l := aplog.NewAPReader(fn)
			lfr = &l
		case types.IS_MWP:
			l := mwpjson.NewMWPJSONReader(fn)
			lfr = &l
		default:
			log.Fatal("Unknown log format")
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
					ls, res := lfr.Reader(metas[options.Config.Idx-1], nil)
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
