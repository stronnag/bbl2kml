package log2mission

import (
	"os"
	"fmt"
	"time"
	"strings"
	"log"
	"path/filepath"
	"github.com/deet/simpleline"
	mission "github.com/stronnag/bbl2kml/pkg/mission"
	options "github.com/stronnag/bbl2kml/pkg/options"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
)

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


func Generate_mission(seg types.LogSegment, meta types.FlightMeta) {
	points := []simpleline.Point{}
	var b types.LogItem
	var st, et time.Time
	if options.Config.StartOff > 0 {
		diff := (time.Duration(options.Config.StartOff) * time.Second)
		st = seg.L.Items[0].Utc.Add(diff)
	}
	if options.Config.EndOff < 0 {
		diff := (time.Duration(options.Config.EndOff) * time.Second)
		lidx := len(seg.L.Items) - 1
		et = seg.L.Items[lidx].Utc.Add(diff)
	} else if options.Config.EndOff > 0 {
		diff := (time.Duration(options.Config.EndOff) * time.Second)
		et = seg.L.Items[0].Utc.Add(diff)
	}

	mfilter := byte(0)
	if strings.Contains(options.Config.Modefilter, "cruise") {
		mfilter |= 1
	}
	if strings.Contains(options.Config.Modefilter, "wp") {
		mfilter |= 2
	}

	for _, b = range seg.L.Items {
		if !st.IsZero() && b.Utc.Before(st) {
			continue
		}
		if !et.IsZero() && b.Utc.After(et) {
			continue
		}
		if (mfilter&1 == 1 && b.Fmode != types.FM_CRUISE2D && b.Fmode != types.FM_CRUISE3D) ||
			(mfilter&2 == 2 && b.Fmode != types.FM_WP) {
			continue
		}
		pt := simpleline.Point3d{X: b.Lon, Y: b.Lat, Z: b.Alt}
		points = append(points, &pt)
	}

	nmi := 0
	ntry := 0
	needrth := !et.IsZero() && options.Config.Modefilter == ""
	var res []simpleline.Point
	var err error
	ep := options.Config.Epsilon
	for {
		res, err = simpleline.RDP(points, ep, simpleline.Euclidean, true)
		if err != nil {
			fmt.Printf("Simplify error:  %v\n", err)
			os.Exit(1)
		}
		nmi = len(res)
		if needrth {
			nmi += 1
		}
		if nmi > 60 {
			ep += float64(float64(nmi-60) * ep * 0.02) // 0.00025
			ntry += 1
			if ntry > 42 {
				log.Fatal("Failed to generate an aceeptable mission after 42 iterations")
			}
		} else if len(res) == 2 {
			ep = ep / 15.0
			ntry += 1
			if ntry > 5 {
				fmt.Fprintln(os.Stderr, "Giving up with minimal mission")
				break
			}
		} else {
			break
		}
	}

	var ms mission.Mission
	for i, p := range res {
		v := p.Vector()
		mi := mission.MissionItem{No: i + 1, Lat: v[1], Lon: v[0], Alt: int32(v[2]), Action: "WAYPOINT"}
		ms.MissionItems = append(ms.MissionItems, mi)
	}
	if needrth {
		ms.MissionItems = append(ms.MissionItems,
			mission.MissionItem{No: len(res), Lat: 0.0, Lon: 0.0, Alt: int32(0.0), Action: "RTH"})
	}
	fmt.Printf("Mission  : %d points", nmi)
	if ntry > 0 {
		fmt.Printf(" (reprocess: %d, epsilon: %.6f)", ntry, ep)
	}
	fmt.Println()
	ms.To_MWXML(generate_filename(meta))
}
