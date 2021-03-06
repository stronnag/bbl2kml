package main

import (
	"os"
	"fmt"
	"log"
	"path/filepath"
	"io/ioutil"
	otx "github.com/stronnag/bbl2kml/pkg/otx"
	bbl "github.com/stronnag/bbl2kml/pkg/bbl"
	blt "github.com/stronnag/bbl2kml/pkg/bltreader"
	options "github.com/stronnag/bbl2kml/pkg/options"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	geo "github.com/stronnag/bbl2kml/pkg/geo"
	kmlgen "github.com/stronnag/bbl2kml/pkg/kmlgen"
)

var GitCommit = "local"
var GitTag = "0.0.0"

func getVersion() string {
	return fmt.Sprintf("%s %s, commit: %s", filepath.Base(os.Args[0]), GitTag, GitCommit)
}

func main() {
	dump_log := os.Getenv("DUMP_LOG") != ""
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
			if options.Config.Dump {
				lfr.Dump()
				os.Exit(0)
			}
			options.Config.Tmpdir, err = ioutil.TempDir("", ".fl2x")
			if err != nil {
				log.Fatalf("fl2x: %+v\n", err)
			}
			defer os.RemoveAll(options.Config.Tmpdir)

			for _, b := range metas {
				if (options.Config.Idx == 0 || options.Config.Idx == b.Index) && b.Flags&types.Is_Valid != 0 {
					for k, v := range b.Summary() {
						fmt.Printf("%-8.8s : %s\n", k, v)
					}
					ls, res := lfr.Reader(b)
					if res {
						if dump_log {
							for _, b := range ls.L.Items {
								fmt.Fprintf(os.Stderr, "%+v\n", b)
							}
						} else if options.Config.Summary == false {
							outfn := kmlgen.GenKmlName(b.Logname, b.Index)
							kmlgen.GenerateKML(ls.H, ls.L, outfn, b, ls.M)
						}
					}
					for k, v := range ls.M {
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
			log.Fatalf("fl2x: %+v\n", err)
		}
	}
}
