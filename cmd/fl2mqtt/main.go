package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

import (
	"aplog"
	"bbl"
	"bltlog"
	"bltmqtt"
	"geo"
	"ltmgen"
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
	defer types.RemoveTmpDir()

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

						switch {
						case strings.HasPrefix(app, "fl2mqtt"):
							ls, res := lfr.Reader(metas[options.Config.Idx-1], nil)
							if res {
								bltmqtt.MQTTGen(ls, metas[options.Config.Idx-1])
							}
						case strings.HasPrefix(app, "fl2ltm"):
							ch := make(chan interface{})
							go lfr.Reader(metas[options.Config.Idx-1], ch)
							ltmgen.LTMGen(ch, metas[options.Config.Idx-1])
						}
						//						fmt.Println()
					} else {
						fmt.Println("Log: Not valid")
					}
				}
			}
		} else {
			log.Fatalf("fl2mqtt: %+v\n", err)
		}
	}
}
