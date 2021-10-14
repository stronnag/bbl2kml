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
	ltmgen "github.com/stronnag/bbl2kml/pkg/ltmgen"
	blt "github.com/stronnag/bbl2kml/pkg/bltreader"
	ap "github.com/stronnag/bbl2kml/pkg/aplog"
)

var GitCommit = "local"
var GitTag = "0.0.0"

func getVersion() string {
	return fmt.Sprintf("%s %s, commit: %s", filepath.Base(os.Args[0]), GitTag, GitCommit)
}

func main() {
	files, app := options.ParseCLI(getVersion)
	if len(files) == 0 || (len(options.Config.Mqttopts) == 0 && len(options.Config.Outdir) == 0 &&
		options.Config.Dump == false && len(options.Config.LTMdev) == 0 && options.Config.Metas == false) {
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
			l := blt.NewBLTReader(fn)
			lfr = &l
		case types.IS_AP:
			l := ap.NewAPReader(fn)
			lfr = &l
		default:
			log.Fatal("Unknown log format")
		}

		metas, err := lfr.GetMetas()
		if err == nil {
			if options.Config.Dump {
				lfr.Dump()
			} else if options.Config.Metas {
				lfr.GetDurations()
				for _, mx := range metas {
					fmt.Printf("%d,%s,%s,%d,%d,%.0f,%x\n", mx.Index, mx.Logname, mx.Date, mx.Start, mx.End, mx.Duration.Seconds(), mx.Flags)
				}
			} else {
				if options.Config.Idx < 1 {
					options.Config.Idx = 1
				}
				if options.Config.Idx <= len(metas) {
					if metas[options.Config.Idx-1].Flags&types.Is_Valid != 0 {
						for k, v := range metas[options.Config.Idx-1].Summary() {
							fmt.Printf("%-8.8s : %s\n", k, v)
						}
						if metas[options.Config.Idx-1].Flags&types.Is_Suspect != 0 {
							fmt.Println("Warning  : Log entry may be corrupt")
						}
						ls, res := lfr.Reader(metas[options.Config.Idx-1])
						if res {
							switch app {
							case "fl2mqtt":
								mqttgen.MQTTGen(ls, metas[options.Config.Idx-1])
							case "fl2ltm":
								ltmgen.LTMGen(ls, metas[options.Config.Idx-1])
							}
						} else {
							fmt.Fprintf(os.Stderr, "*** skipping generation for log  with no valid geospatial data\n")
						}
						fmt.Println()
					} else {
						fmt.Println("Not valid")
					}
				}
			}
		} else {
			log.Fatalf("fl2mqtt: %+v\n", err)
		}
	}
}
