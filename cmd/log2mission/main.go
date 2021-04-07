package main

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/deet/simpleline"
	otx "github.com/stronnag/bbl2kml/pkg/otx"
	bbl "github.com/stronnag/bbl2kml/pkg/bbl"
	blt "github.com/stronnag/bbl2kml/pkg/bltreader"
	options "github.com/stronnag/bbl2kml/pkg/options"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	//	geo "github.com/stronnag/bbl2kml/pkg/geo"
	mission "github.com/stronnag/bbl2kml/pkg/mission"
)

var GitCommit = "local"
var GitTag = "0.0.0"

func getVersion() string {
	return fmt.Sprintf("%s %s, commit: %s", filepath.Base(os.Args[0]), GitTag, GitCommit)
}

func generate_filename(m types.FlightMeta) string {
	outfn := filepath.Base(m.Logname)
	ext := filepath.Ext(outfn)
	if len(ext) < len(outfn) {
		outfn = outfn[0 : len(outfn)-len(ext)]
	}
	ext = fmt.Sprintf(".%d.mission", m.Index)
	outfn = outfn + ext
	return outfn
}


func generate_mission(seg types.LogSegment, meta types.FlightMeta) {
	points := []simpleline.Point{}
	var b types.LogItem
	for _, b = range seg.L.Items {
		pt := simpleline.Point3d{X: b.Lon, Y: b.Lat, Z: b.Alt}
		points = append(points, &pt)
	}
	res, err := simpleline.RDP(points, options.Config.Epsilon, simpleline.Euclidean, true)
	if err != nil {
		fmt.Printf("Simplify error:  %v\n", err)
		os.Exit(1)
	}
	var ms mission.Mission
	for i, p := range res {
		v := p.Vector()
		mi := mission.MissionItem{No: i + 1, Lat: v[1], Lon: v[0],
			Alt: int32(v[2]), Action: "WAYPOINT"}
		ms.MissionItems = append(ms.MissionItems, mi)
	}
	fmt.Printf("Mission  : %v points\n", len(res))
	ms.To_MWXML(generate_filename(meta))
}

func main() {
	files, _ := options.ParseCLI(getVersion)
	if len(files) == 0 {
		options.Usage()
		os.Exit(1)
	}

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
						generate_mission(ls, metas[options.Config.Idx-1])
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
