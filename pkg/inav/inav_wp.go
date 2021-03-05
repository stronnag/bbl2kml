package inav

import (
	//	"fmt"
	//	"os"
	"time"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	options "github.com/stronnag/bbl2kml/pkg/options"
	geo "github.com/stronnag/bbl2kml/pkg/geo"
	mission "github.com/stronnag/bbl2kml/pkg/mission"
)

var phtime time.Time

func get_next_wp(ms *mission.Mission, k int) int {
	tgt := 0
	if ms != nil && k < len(ms.MissionItems)-1 {
		switch ms.MissionItems[k+1].Action {
		case "JUMP":
			if ms.MissionItems[k+1].P3 == -1 {
				tgt = int(ms.MissionItems[k+1].P1)
			} else {
				if ms.MissionItems[k+1].P3 == 0 {
					if k < len(ms.MissionItems)-2 {
						tgt = int(ms.MissionItems[k+2].No)
						ms.MissionItems[k+1].P3 = ms.MissionItems[k+1].P2
					}
				} else {
					tgt = int(ms.MissionItems[k+1].P1)
					ms.MissionItems[k+1].P3 -= 1
				}
			}
		case "RTH":
			tgt = int(ms.MissionItems[k+1].No)
		case "SET_HEAD", "SET_POI":
			if k < len(ms.MissionItems)-1 {
				tgt = int(ms.MissionItems[k+2].No)
			}
		default:
			tgt = int(ms.MissionItems[k+1].No)
		}
	}
	return tgt
}

var isTimed bool

func WP_state(ms *mission.Mission, b types.LogItem, tgt int) (int, int) {
	k := tgt - 1
	if isTimed {
		if b.Utc.After(phtime) {
			tgt = get_next_wp(ms, k)
			isTimed = false
		} else {
			b.NavMode = 4
		}
	} else {
		cdist := 1.25 * b.Spd * float64(options.Intvl/1000.0)
		if cdist < 30 {
			cdist = 30
		}
		cdist /= 1852.0
		mi := ms.MissionItems[k]
		if mi.Is_GeoPoint() {
			brg, d := geo.Csedist(b.Lat, b.Lon, mi.Lat, mi.Lon)
			if d < cdist {
				// relative heading, independent of which is greaer & 359<->0
				// sign depends on whether target is to port or starboard
				bdiff := (int(brg)-int(b.Cse)+540)%360 - 180
				//				fmt.Fprintf(os.Stderr, "Around WP %d brg=%.0f cse=%d d=%.1f (%d) [%.1f]\n", mi.No, brg, b.Cse, d*1852, bdiff, cdist*1852)
				if bdiff > 90 || bdiff < -90 {
					//					fmt.Fprintf(os.Stderr, "Reached %d %s\n", k, ms.MissionItems[k].Action)
					if ms.MissionItems[k].Action == "POSHOLD_TIME" {
						var phwait time.Duration
						mwaitms := int(ms.MissionItems[k].P1) * 1000
						if mwaitms > options.Intvl/2000 {
							phwait = time.Duration(mwaitms-options.Intvl/2) * time.Millisecond
						} else {
							phwait = time.Duration(ms.MissionItems[k].P1) * time.Second
						}
						phtime = b.Utc.Add(phwait)
						isTimed = true
						b.NavMode = 4
					} else {
						tgt = get_next_wp(ms, k)
						//						fmt.Fprintf(os.Stderr, "New target WP %d %d (%s)\n", tgt, nvs, ms.MissionItems[k+1].Action)
					}
				}
			}
		}
	}
	act, _ := mission.ActionMap[ms.MissionItems[k].Action]
	return tgt, act
}
