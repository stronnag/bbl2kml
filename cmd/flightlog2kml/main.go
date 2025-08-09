package main

import (
	"fmt"
	"github.com/yookoala/realpath"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

import (
	"aplog"
	"bbl"
	"bltlog"
	"flsql"
	"geo"
	"kmlgen"
	"mwpjson"
	"options"
	"otx"
	"sqlreader"
	"types"
)

var GitCommit = "local"
var GitTag = "0.0.0"

func GetVersion() string {
	return fmt.Sprintf("%s %s commit:%s", filepath.Base(os.Args[0]), GitTag, GitCommit)
}

func main() {
	dump_log := os.Getenv("DUMP_LOG") != ""
	files, _ := options.ParseCLI(GetVersion)
	geo.Frobnicate_init()
	if len(files) == 0 {
		if len(options.Config.Mission) > 0 {
			outms := kmlgen.GenKmlName(options.Config.Mission, options.Config.MissionIndex)
			kmlgen.GenerateMissionOnly(outms, GetVersion)
			show_output(outms)
		} else if len(options.Config.Cli) > 0 {
			outms := kmlgen.GenKmlName(options.Config.Cli, 0)
			kmlgen.GenerateCliOnly(outms, GetVersion)
			show_output(outms)
		} else {
			options.Usage()
		}
		os.Exit(1)
	}

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
		case types.IS_SQL:
			l := sqlreader.NewSQLReader(fn)
			lfr = &l
		default:
			log.Fatalf("%s: unknown log format\n", fn)
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

			var db flsql.DBL
			use_db := false
			if options.Config.Sql != "" {
				db = flsql.NewSQLliteDB(options.Config.Sql)
				use_db = true
				defer db.Close()
			}

			for _, b := range metas {
				outfn := ""
				if (options.Config.Idx == 0 || options.Config.Idx == b.Index) && b.Flags&types.Is_Valid != 0 {
					if use_db {
						db.Reset()
					} else {
						for k, v := range b.Summary() {
							fmt.Printf("%-8.8s : %s\n", k, v)
						}
					}
					ls, res := lfr.Reader(b, nil)
					if res {
						if dump_log {
							for _, bi := range ls.L.Items {
								fmt.Fprintf(os.Stderr, "%+v\n", bi)
							}
						} else if use_db {
							n := len(ls.L.Items)
							ns := uint64(0)
							if n > 0 {
								ns = uint64(ls.L.Items[n-1].Stamp - ls.L.Items[0].Stamp)
								b.Duration = time.Duration(ns) * time.Microsecond
							}
							ndelay := uint64(100 * 1000) // 100 millsecs
							if ns > 10*60*1000*1000 {    // > 10 mins
								ndelay = ns / uint64(6000)
							}

							db.Begin()
							dt := uint64(0)
							nx := 0
							for _, bi := range ls.L.Items {
								ut := bi.Stamp
								if (ut - dt) >= ndelay {
									db.Writelog(b.Index, nx, bi)
									nx += 1
									dt = ut
								}
							}
							if dt != ls.L.Items[n-1].Stamp {
								db.Writelog(b.Index, nx, ls.L.Items[n-1])
								nx += 1
							}
							db.Commit()
							fmt.Printf("%d\t%s\t%.1f\t%d", b.Index, b.Date, b.Duration.Seconds(), nx)
							if ls.S != "" {
								fmt.Printf("\t*")
							}
							fmt.Println()

							if ls.S != "" {
								db.Begin()
								db.Writeerr(b.Index, ls.S)
								db.Commit()
							}
						} else if options.Config.Summary == false {
							outfn = kmlgen.GenKmlName(b.Logname, b.Index)
							kmlgen.GenerateKML(ls.H, ls.L, outfn, b, ls.M, GetVersion)
						}
						if use_db {
							db.Begin()
							db.Writemeta(b)
							db.Commit()
						}
					}
					if !use_db {
						for k, v := range ls.M {
							fmt.Printf("%-8.8s : %s\n", k, v)
						}
						if s, ok := b.ShowDisarm(); ok {
							fmt.Printf("%-8.8s : %s\n", "Disarm", s)
						}
						if !res {
							fmt.Fprintf(os.Stderr, "*** skipping KML/Z for log  with no valid geospatial data\n")
						} else {
							show_output(outfn)
						}
						fmt.Println()
					}
				}
			}
		} else {
			log.Fatalf("fl2x: %+v\n", err)
		}
	}
}

func show_output(outfn string) {
	if outfn != "" {
		rp, err := realpath.Realpath(outfn)
		if err != nil || rp == "" {
			fmt.Printf("%-8.8s : <%s> <%s>\n", "RealPath", rp, err)
			rp = outfn
		}
		fmt.Printf("%-8.8s : %s\n", "Output", rp)
	}
}
