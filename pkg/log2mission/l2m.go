package log2mission

import (
	"fmt"
	"geo"
	"github.com/deet/simpleline"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

import (
	"mission"
	"options"
	"types"
)

func generate_filename(m types.FlightMeta) string {
	var outfn string
	if options.Config.Mission == "" {
		outfn = filepath.Base(m.Logname)
		ext := filepath.Ext(outfn)
		if len(ext) < len(outfn) {
			outfn = outfn[0 : len(outfn)-len(ext)]
		}
		ext = fmt.Sprintf(".%d.mission", m.Index)
		outfn = outfn + ext
	} else {
		outfn = options.Config.Mission
	}
	return outfn
}

func Generate_mission(seg types.LogSegment, meta types.FlightMeta) {
	mfilter := byte(0)

	if options.Config.Modefilter == "any" {
		mfilter = 0xff
	} else {
		if strings.Contains(options.Config.Modefilter, "cruise") {
			mfilter |= 1
		}
		if strings.Contains(options.Config.Modefilter, "wp") {
			mfilter |= 2
		}
	}
	if (mfilter == 0) && (seg.L.Cap&types.CAP_WPNO) == types.CAP_WPNO {
		generate_from_active(seg, meta)
	} else {
		generate_from_path(seg, meta, mfilter)
	}
}

func generate_from_path(seg types.LogSegment, meta types.FlightMeta, mfilter byte) {
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

	for _, b = range seg.L.Items {
		if !st.IsZero() && b.Utc.Before(st) {
			continue
		}
		if !et.IsZero() && b.Utc.After(et) {
			continue
		}
		if mfilter != 0xff {
			if ((mfilter&1 == 1) && b.Fmode != types.FM_CRUISE2D && b.Fmode != types.FM_CRUISE3D) ||
				(mfilter&2 == 2 && b.Fmode != types.FM_WP) {
				continue
			}
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
		if nmi > options.Config.MaxWP {
			ep += float64(float64(nmi-options.Config.MaxWP) * ep * 0.02) // 0.00025
			ntry += 1
			if ntry > 42 {
				log.Fatalln("l2m: Failed to generate an aceeptable mission after 42 iterations")
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

	generate_log_mission(res, generate_filename(meta), needrth, seg.H)
	fmt.Printf("Mission  : %d points, epsilon: %.6f", nmi, ep)
	if ntry > 0 {
		fmt.Printf(" (reprocess: %d, epsilon: %.6f)", ntry, ep)
	}
	fmt.Println()
	fmt.Println("Note: Increase epsilon to decrease the number of mission points,")
	fmt.Println("      decrease epsilon to increase the number of mission points.")

}

func generate_log_mission(res []simpleline.Point, mfn string, needrth bool, homes types.HomeRec) {
	var ms mission.Mission

	ms.Version.Value = "latest"

	ms.Metadata.Homey = homes.HomeLat
	ms.Metadata.Homex = homes.HomeLon
	fb := geo.Getfrobnication()

	if fb != nil {
		fb.Set_origin(homes.HomeLat, homes.HomeLon, homes.HomeAlt)
		ms.Metadata.Homey, ms.Metadata.Homex, _ = fb.Relocate(ms.Metadata.Homey, ms.Metadata.Homex, 0)
	}

	for i, p := range res {
		v := p.Vector()
		la := v[1]
		lo := v[0]
		alt := v[2]
		if fb != nil {
			la, lo, alt = fb.Relocate(la, lo, alt)
		}
		mi := mission.MissionItem{No: i + 1, Lat: la, Lon: lo, Alt: int32(alt), Action: "WAYPOINT"}
		ms.MissionItems = append(ms.MissionItems, mi)
	}
	if needrth {
		ms.MissionItems = append(ms.MissionItems,
			mission.MissionItem{No: len(res), Lat: 0.0, Lon: 0.0, Alt: int32(0.0), Action: "RTH"})
	}

	n := len(ms.MissionItems)
	if n > 0 {
		ms.MissionItems[n-1].Flag = 165
		ms.To_MWXML(mfn)
	}
}

func generate_from_active(seg types.LogSegment, meta types.FlightMeta) {
	points := []simpleline.Point{}
	navm := false
	lnavm := false
	lwpno := uint8(0)
	llat := 0.0
	llon := 0.0
	lalt := 0.0
	for _, b := range seg.L.Items {
		navm = b.Fmode == types.FM_WP
		if navm != lnavm {
			if lnavm {
				pt := simpleline.Point3d{X: b.Lon, Y: b.Lat, Z: b.Alt}
				points = append(points, &pt)
			}
		}
		if lwpno != b.ActiveWP {
			if lwpno != 0 {
				pt := simpleline.Point3d{X: b.Lon, Y: b.Lat, Z: b.Alt}
				points = append(points, &pt)
			}
		}
		lnavm = navm
		lwpno = b.ActiveWP
		llat = b.Lat
		llon = b.Lon
		lalt = b.Alt
	}
	if navm {
		pt := simpleline.Point3d{X: llon, Y: llat, Z: lalt}
		points = append(points, &pt)
	}

	mfn := generate_filename(meta)
	generate_log_mission(points, mfn, false, seg.H)
	fmt.Printf("Mission  : %d active points\n", len(points))
	fmt.Printf("Output   : %s\n", mfn)
}
