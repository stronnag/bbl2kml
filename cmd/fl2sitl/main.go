package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

import (
	"bbl"
	"geo"
	"options"
	"sitlgen"
	"types"
)

var GitCommit = "local"
var GitTag = "0.0.0"

func getVersion() string {
	return fmt.Sprintf("%s %s commit:%s", filepath.Base(os.Args[0]), GitTag, GitCommit)
}

func main() {
	files, app := options.ParseCLI(getVersion)
	if len(files) == 0 {
		if options.Config.SitlMinimal == false {
			options.Usage()
			os.Exit(1)
		} else {
			stl := sitlgen.NewSITL()
			stl.Faker()
		}
	}
	geo.Frobnicate_init()
	var lfr types.FlightLog
	for _, fn := range files {
		ftype := types.EvinceFileType(fn)
		switch ftype {
		case types.IS_BBL:
			l := bbl.NewBBLReader(fn)
			lfr = &l
		default:
			log.Fatal("Unknown log format")
		}

		metas, err := lfr.GetMetas()
		if err == nil {
			if metas[0].Acc1G == 0 {
				// Old file, refresh the cache
				currentTime := time.Now().Local()
				err = os.Chtimes(fn, currentTime, currentTime)
				if err == nil {
					metas, err = lfr.GetMetas()
				}
			}

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
						stl := sitlgen.NewSITL()
						ch := make(chan interface{})
						go lfr.Reader(metas[options.Config.Idx-1], ch)
						stl.Run(ch, metas[options.Config.Idx-1])
					} else {
						fmt.Println("Log: Not valid")
					}
				}
			}
		} else {
			log.Fatalf("%s: %+v\n", app, err)
		}
	}
}
