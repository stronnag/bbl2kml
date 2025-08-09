package mwpjson

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"geo"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

import (
	"options"
	"types"
)

const (
	PLANE_MODE_MANUAL        = 0
	PLANE_MODE_CIRCLE        = 1
	PLANE_MODE_STABILIZE     = 2
	PLANE_MODE_TRAINING      = 3
	PLANE_MODE_ACRO          = 4
	PLANE_MODE_FLY_BY_WIRE_A = 5
	PLANE_MODE_FLY_BY_WIRE_B = 6
	PLANE_MODE_CRUISE        = 7
	PLANE_MODE_AUTOTUNE      = 8
	PLANE_MODE_AUTO          = 10
	PLANE_MODE_RTL           = 11
	PLANE_MODE_LOITER        = 12
	PLANE_MODE_TAKEOFF       = 13
	PLANE_MODE_AVOID_ADSB    = 14
	PLANE_MODE_GUIDED        = 15
	PLANE_MODE_INITIALIZING  = 16
	PLANE_MODE_QSTABILIZE    = 17
	PLANE_MODE_QHOVER        = 18
	PLANE_MODE_QLOITER       = 19
	PLANE_MODE_QLAND         = 20
	PLANE_MODE_QRTL          = 21
	PLANE_MODE_QAUTOTUNE     = 22
	PLANE_MODE_ENUM_END      = 23
)

const (
	COPTER_MODE_STABILIZE    = 0
	COPTER_MODE_ACRO         = 1
	COPTER_MODE_ALT_HOLD     = 2
	COPTER_MODE_AUTO         = 3
	COPTER_MODE_GUIDED       = 4
	COPTER_MODE_LOITER       = 5
	COPTER_MODE_RTL          = 6
	COPTER_MODE_CIRCLE       = 7
	COPTER_MODE_LAND         = 9
	COPTER_MODE_DRIFT        = 11
	COPTER_MODE_SPORT        = 13
	COPTER_MODE_FLIP         = 14
	COPTER_MODE_AUTOTUNE     = 15
	COPTER_MODE_POSHOLD      = 16
	COPTER_MODE_BRAKE        = 17
	COPTER_MODE_THROW        = 18
	COPTER_MODE_AVOID_ADSB   = 19
	COPTER_MODE_GUIDED_NOGPS = 20
	COPTER_MODE_SMART_RTL    = 21
	COPTER_MODE_ENUM_END     = 22
)

type MWPJSON struct {
	name string
	meta []types.FlightMeta
}

var (
	verbose bool
)

func NewMWPJSONReader(fn string) MWPJSON {
	var l MWPJSON
	l.name = fn
	l.meta = nil
	return l
}

func (o *MWPJSON) LogType() byte {
	return types.LOGMWP
}

func (o *MWPJSON) GetDurations() {
}

func (o *MWPJSON) Dump() {
}

func (o *MWPJSON) GetMetas() ([]types.FlightMeta, error) {
	verbose = (os.Getenv("VERBOSE") != "")
	m, err := types.ReadMetaCache(o.name)
	if err != nil || options.Config.Nocache {
		m, err = metas(o.name)
		types.WriteMetaCache(o.name, m)
	}
	o.meta = m
	return m, err
}

func metas(logfile string) ([]types.FlightMeta, error) {
	var lt float64
	var st float64
	var id int
	var baseutc time.Time

	st = 0
	id = 0

	have_start := bool(false)

	var metas []types.FlightMeta
	var mt types.FlightMeta

	r, err := os.Open(logfile)
	if err == nil {
		bp := filepath.Base(logfile)
		nl := 0
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			l := scanner.Text()
			var o map[string]interface{}
			json.Unmarshal([]byte(l), &o)
			if utx, ok := o["utime"]; ok {
				nl += 1
				lt = utx.(float64)
				switch o["type"] {
				case "init", "environment":
					if have_start == false {
						sec := int64(lt)
						nsec := int64((lt - float64(sec)) * 1e9)
						baseutc = time.Unix(sec, nsec)
						have_start = true
						if id != 0 {
							mt.End = nl - 1
							mt.Duration = time.Duration(lt-st) * time.Second
							if verbose {
								fmt.Fprintf(os.Stderr, ":DBG: %+v %+v %+v\n", id, lt, st)
							}
							metas = append(metas, mt)
						}
						mt = types.FlightMeta{Logname: bp, Index: id + 1, Start: nl, Date: baseutc}
						id += 1
						st = 0
					}

					if o["type"] == "init" {
						s, ok := o["vname"].(string)
						if ok {
							mt.Craft = s
						} else {
							mt.Craft = "NO-NAME"
						}
						s, ok = o["fc_var"].(string)
						if ok {
							var sb strings.Builder
							sb.WriteString(s)
							sb.WriteString(" ")
							s, ok = o["fc_vers_str"].(string)
							if !ok {
								x, ok := o["fc_vers"].(float64)
								if ok {
									xi := int(x)
									a0 := xi & 0xff
									a1 := (xi >> 8) & 0xff
									a2 := (xi >> 16) & 0xff
									s = fmt.Sprintf("%d.%d.%d", a2, a1, a0)
								}
							}
							sb.WriteString(s)
							sb.WriteString(" (")
							s, ok = o["git_info"].(string)
							sb.WriteString(s)
							sb.WriteString(") ")
							s, ok = o["fcname"].(string)
							sb.WriteString(s)
							mt.Firmware = sb.String()
						} else {
							mt.Firmware = "INAV"
						}
						s, ok = o["fcdate"].(string)
						if ok {
							mt.Fwdate = s
						} else {
							mt.Fwdate = "unknown"
						}
						val, ok := o["sensors"].(float64)
						if ok {
							mt.Sensors = uint16(val)
						} else {
							mt.Sensors = types.Has_GPS | types.Has_Baro | types.Has_Acc
						}
						val, ok = o["features"].(float64)
						if ok {
							mt.Features = uint32(val)
						}
						val, ok = o["capability"].(float64)
						if ok {
							mt.Features = uint32(val)
						}
					}

				case "v0:armed", "armed":
					if (o["armed"]).(bool) {
						if st == 0 {
							st = (o["utime"].(float64))
						}
					}

				case "v0:gps", "raw_gps", "mavlink_gps_raw_int":
					et := (o["utime"].(float64))
					if st > 0 && et-st > 30 {
						have_start = false
					}
				default:
				}
			}
		}

		if st > 0 {
			mt.Sensors |= types.Has_GPS | types.Has_Baro | types.Has_Acc
			mt.End = nl
			mt.Duration = time.Duration(lt-st) * time.Second
			if verbose {
				fmt.Fprintf(os.Stderr, ":DBG: %+v %+v %+v\n", id, lt, st)
			}
			metas = append(metas, mt)
		}
	}

	for j, mx := range metas {
		if mx.End-mx.Start > 64 {
			metas[j].Flags |= types.Has_Start | types.Is_Valid | types.Has_Craft
		}
		if verbose {
			fmt.Fprintf(os.Stderr, ":DBG: %+v\n", metas[j])
		}
	}
	if len(metas) == 0 {
		err = errors.New("No usable records in MWP JSON log")
	}
	return metas, err
}

var (
	lt         float64
	st         float64
	id         int
	hlat, hlon float64
	homes      types.HomeRec
)

func fm_ltm(ltm uint8) uint8 {
	var fm uint8
	switch ltm {
	case 0:
		fm = types.FM_MANUAL
	case 1:
		fm = types.FM_ACRO
	case 2:
		fm = types.FM_ANGLE
	case 3:
		fm = types.FM_HORIZON
	case 8:
		fm = types.FM_AH
	case 9:
		fm = types.FM_PH
	case 10:
		fm = types.FM_WP
	case 13:
		fm = types.FM_RTH
	case 15:
		fm = types.FM_LAND
	case 18:
		fm = types.FM_CRUISE3D
	case 19:
		fm = types.FM_EMERG
	case 20:
		fm = types.FM_LAUNCH
	default:
		fm = types.FM_ACRO
	}
	return fm
}

func parse_json(o map[string]interface{}, b *types.LogItem) (bool, uint16) {
	cap := uint16(0)
	done := false

	if utm, ok := o["utime"]; ok {
		lt = utm.(float64)
		switch o["type"] {
		case "environment", "init":
			st = 0
			hlat = -999
			hlon = -999
			b.Status = 0
			b.Tdist = 0
			if s, ok := o["mission"]; ok {
				options.MissionFile = s.(string)
			}

		case "text", "v0:text":
			if s, ok := o["id"]; ok {
				sid := s.(string)
				switch sid {
				case "geozone":
					options.GeoZone = o["content"].(string)
				default:
				}
			}

		case "fwa":
			loiter_radius := int(o["loiter_radius"].(float64))
			approach_length := int(o["approach_length"].(float64))
			options.Fwastr = fmt.Sprintf("loiter_radius=%d approach_length=%d", loiter_radius, approach_length)

		case "armed", "v0:armed":
			if (o["armed"]).(bool) {
				if st == 0 {
					st = (o["utime"].(float64))
				} else {
					b.Stamp = uint64((lt - st) * 1000 * 1000)
					b.Status |= 1
					if b.Cse == 0xffff {
						b.Cse = b.Cog
					}
					done = true
				}
			}

		case "v0:nav-status":
			b.Navmode = byte(o["nav_mode"].(float64))
			b.ActiveWP = uint8(o["wp_number"].(float64))
			if s, ok := o["gps_mode"]; ok {
				gmo := byte(0)
				act := byte(0)
				gmo = byte(s.(float64))
				b.Navextra = gmo
				if s, ok := o["action"]; ok {
					act = byte(s.(float64))
					b.Navextra |= (act << 4)
				}
			} else {
				b.Navextra = 0
			}

		case "v0:origin":
			b.Hlat = o["lat"].(float64)
			b.Hlon = o["lon"].(float64)
			homes.Flags |= types.HOME_ARM | types.HOME_ALT
			homes.HomeLat = b.Hlat
			homes.HomeLon = b.Hlon
			homes.HomeAlt = o["alt"].(float64)

		case "v0:mode-flags":
			b.Status = uint8(o["flags"].(float64))
			ltmmode := uint8(o["ltmmode"].(float64))
			b.Fmode = fm_ltm(ltmmode)

		case "v0:attitude":
			b.Cse = uint32(o["yaw"].(float64))
			b.Roll = int16(o["roll"].(float64))
			b.Pitch = int16(o["pitch"].(float64))

		case "v0:altitude", "altitude":
			cap |= types.CAP_ALTITUDE
			b.Alt = o["estalt"].(float64)
			// b.Vario = o["vario"].(float64)

		case "v0:power":
			b.Volts = o["voltage"].(float64)
			b.Amps = o["amps"].(float64)
			b.Energy = o["power"].(float64)
			b.Rssi = uint8(o["rssi"].(float64) * 100 / 255)
			cap |= (types.CAP_RSSI_VALID | types.CAP_VOLTS | types.CAP_AMPS)

		case "v0:gps":
			b.Stamp = uint64((lt - st) * 1000 * 1000)
			sec := int64(lt)
			nsec := int64((lt - float64(sec)) * 1e9)
			b.Utc = time.Unix(sec, nsec)
			b.Lat = o["lat"].(float64)
			b.Lon = o["lon"].(float64)
			b.Cog = uint32(o["cog"].(float64))
			b.Spd = o["speed"].(float64)
			b.GAlt = o["alt"].(float64)
			b.Fix = uint8(o["fix"].(float64))
			b.Numsat = uint8(o["numsat"].(float64))
			b.Hdop = uint16(o["hdop"].(float64))
			cap |= types.CAP_SPEED
			if (b.Status & 1) != 0 {
				if hlat == -999 && hlon == -999 {
					hlat = b.Lat
					hlon = b.Lon
					homes.Flags |= types.HOME_ARM | types.HOME_ALT
					homes.HomeLat = hlat
					homes.HomeLon = hlon
					homes.HomeAlt = b.GAlt
				}
				b.Hlat = hlat
				b.Hlon = hlon
			}

		case "v0:range-bearing", "comp_gps":
			b.Vrange = o["range"].(float64)
			b.Bearing = int32(o["bearing"].(float64))

		case "v0:sensor-reason", "ltm_xframe":
			if s, ok := o["sensorok"].(float64); ok {
				b.HWfail = (s != 0)
			} else {
				b.HWfail = false
			}

			/*******************/

		case "analog2":
			b.Volts = o["voltage"].(float64)
			b.Amps = o["amps"].(float64) / 100.0
			b.Rssi = uint8(o["rssi"].(float64) * 100 / 1023)
			cap |= (types.CAP_RSSI_VALID | types.CAP_VOLTS | types.CAP_AMPS)

		case "status":
			b.Navmode = byte(o["nav_mode"].(float64))
			b.ActiveWP = uint8(o["wp_number"].(float64))
			switch b.Navmode {
			case 1, 2: // RTH
				b.Status = ((13 << 2) | (b.Status & 3))
				b.Fmode = types.FM_RTH
			case 3, 4: // PH
				b.Status = ((9 << 2) | (b.Status & 3))
				b.Fmode = types.FM_PH
			case 5, 6, 7: // WP
				b.Status = ((10 << 2) | (b.Status & 3))
				b.Fmode = types.FM_WP
				cap |= types.CAP_WPNO
			case 8, 9, 10, 11, 12, 13, 14: // Land
				b.Status = ((15 << 2) | (b.Status & 3))
				b.Fmode = types.FM_LAND
			default:
			}

		case "raw_gps":
			b.Stamp = uint64((lt - st) * 1000 * 1000)
			sec := int64(lt)
			nsec := int64((lt - float64(sec)) * 1e9)
			b.Utc = time.Unix(sec, nsec)
			b.Lat = o["lat"].(float64)
			b.Lon = o["lon"].(float64)
			b.GAlt = o["alt"].(float64)
			b.Fix = uint8(o["fix"].(float64))
			b.Numsat = uint8(o["numsat"].(float64))
			if _, ok := o["hdop"]; ok {
				b.Hdop = uint16(o["hdop"].(float64))
			}
			b.Cog = uint32(o["cse"].(float64))
			b.Spd = o["spd"].(float64)
			cap |= types.CAP_SPEED
			if (b.Status & 1) != 0 {
				if hlat == -999 && hlon == -999 {
					hlat = b.Lat
					hlon = b.Lon
					homes.Flags |= types.HOME_ARM | types.HOME_ALT
					homes.HomeLat = hlat
					homes.HomeLon = hlon
					homes.HomeAlt = b.GAlt
				}
				b.Hlat = hlat
				b.Hlon = hlon
			}

		case "attitude":
			mhead := o["heading"].(float64)
			if mhead < 0 {
				mhead += 360
			}
			b.Cse = uint32(mhead)
			b.Roll = int16(o["angx"].(float64))
			b.Pitch = int16(o["angy"].(float64))

		case "ltm_raw_sframe":
			b.Volts = o["vbat"].(float64) / 1000.0
			b.Energy = o["vcurr"].(float64)
			b.Rssi = uint8(o["rssi"].(float64) * 100 / 255)
			cap |= (types.CAP_RSSI_VALID | types.CAP_VOLTS | types.CAP_AMPS)
			if b.Navmode == 0 {
				ltmmode := b.Status >> 2
				b.Fmode = fm_ltm(ltmmode)
				b.Status = uint8(o["flags"].(float64))
			}

		case "mavlink_attitude":
			cse := int(o["yaw"].(float64) * 57.29578)
			if cse < 0 {
				cse += 360
			}
			b.Cse = uint32(cse)
			b.Roll = int16(o["roll"].(float64) * 57.29578)
			b.Pitch = int16(-o["pitch"].(float64) * 57.29578)

		case "mavlink_vfr_hud":
			b.Alt = o["alt"].(float64)

		case "mavlink_gps_raw_int":
			b.Stamp = uint64((lt - st) * 1000 * 1000)
			b.Lat = o["lat"].(float64) / 1e7
			b.Lon = o["lon"].(float64) / 1e7
			b.GAlt = o["alt"].(float64) / 1000

			fix := o["fix_type"].(float64)
			b.Fix = uint8(math.Min(fix, 3))
			eph := o["eph"].(float64)
			if eph != 65535 {
				b.Hdop = uint16(eph / 100)
			}
			b.Numsat = uint8(o["satellites_visible"].(float64))
			ival := uint(o["vel"].(float64))
			if ival != 0xffff {
				b.Spd = float64(ival) / 100.0
			}
			ival = uint(o["cog"].(float64))
			if ival != 0xffff {
				b.Cog = uint32(ival) / 100
			}

			if (b.Status & 1) != 0 {
				if hlat == -999 && hlon == -999 {
					hlat = b.Lat
					hlon = b.Lon
					homes.Flags |= types.HOME_ARM | types.HOME_ALT
					homes.HomeLat = hlat
					homes.HomeLon = hlon
					homes.HomeAlt = b.GAlt
				}
				b.Hlat = hlat
				b.Hlon = hlon
			}

			if (homes.Flags & (types.HOME_ARM | types.HOME_ALT)) != 0 {
				cs, dx := geo.Csedist(b.Hlat, b.Hlon, b.Lat, b.Lon)
				b.Bearing = int32(cs)
				b.Vrange = dx * 1852
			}

		case "mavlink_gps_global_origin":
			b.Hlat = o["latitude"].(float64) / 1e7
			b.Hlon = o["longitude"].(float64) / 1e7
			homes.Flags |= types.HOME_ARM | types.HOME_ALT
			homes.HomeLat = b.Hlat
			homes.HomeLon = b.Hlon
			homes.HomeAlt = o["altitude"].(float64) / 1000.0

		case "mavlink_heartbeat":
			mavtype := int(o["mavtype"].(float64))
			mavmode := int(o["custom_mode"].(float64))
			var ltmflags uint8
			if o["utime"].(float64) > 1607040000 {
				ltmflags = mav2ltm(mavmode, (mavtype == 1))
			} else {
				ltmflags = xmav2ltm(mavmode, (mavtype == 1))
			}
			b.Fmode = fm_ltm(ltmflags)
			b.Status |= (ltmflags << 2)

		default:
		}
	}
	return done, cap
}

func xmav2ltm(mavmode int, is_fw bool) uint8 {
	ltmmode := uint8(0)
	if is_fw {
		// I don't believe the old iNav mapping for FW ...
		switch mavmode {
		case 0:
			ltmmode = types.Ltm_MANUAL // manual
			break
		case 4:
			ltmmode = types.Ltm_ACRO // acro
			break
		case 2:
			ltmmode = types.Ltm_HORIZON // angle / horiz
			break
		case 5:
			ltmmode = types.Ltm_ALTHOLD // alth
			break
		case 1:
			ltmmode = types.Ltm_POSHOLD // posh
			break
		case 11:
			ltmmode = types.Ltm_RTH // rth
			break
		case 10:
			ltmmode = types.Ltm_WAYPOINTS // wp
			break
		case 15:
			ltmmode = types.Ltm_LAUNCH // launch
			break
		default:
			ltmmode = types.Ltm_ACRO
			break
		}
	} else {
		switch mavmode {
		case 1:
			ltmmode = types.Ltm_ACRO // acro / manual
			break
		case 0:
			ltmmode = types.Ltm_HORIZON // angle / horz
			break
		case 2:
			ltmmode = types.Ltm_ALTHOLD // alth
			break
		case 16:
			ltmmode = types.Ltm_POSHOLD // posh
			break
		case 6:
			ltmmode = types.Ltm_RTH // rth
			break
		case 3:
			ltmmode = types.Ltm_WAYPOINTS // wp
			break
		case 18:
			ltmmode = types.Ltm_LAUNCH // launch
			break
		default:
			ltmmode = types.Ltm_ACRO
			break
		}
	}
	return ltmmode
}

func mav2ltm(mavmode int, is_fw bool) uint8 {
	ltmmode := uint8(0)
	if is_fw {
		switch mavmode {
		case PLANE_MODE_MANUAL:
			ltmmode = types.Ltm_MANUAL
			break
		case PLANE_MODE_ACRO:
			ltmmode = types.Ltm_ACRO
			break
		case PLANE_MODE_FLY_BY_WIRE_A:
			ltmmode = types.Ltm_ANGLE
			break
		case PLANE_MODE_STABILIZE:
			ltmmode = types.Ltm_HORIZON
			break
		case PLANE_MODE_FLY_BY_WIRE_B:
			ltmmode = types.Ltm_ALTHOLD
			break
		case PLANE_MODE_LOITER:
			ltmmode = types.Ltm_POSHOLD
			break
		case PLANE_MODE_RTL:
			ltmmode = types.Ltm_RTH
			break
		case PLANE_MODE_AUTO:
			ltmmode = types.Ltm_WAYPOINTS
			break
		case PLANE_MODE_CRUISE:
			ltmmode = types.Ltm_CRUISE
			break
		case PLANE_MODE_TAKEOFF:
			ltmmode = types.Ltm_LAUNCH
			break
		default:
			ltmmode = types.Ltm_ACRO
			break
		}
	} else {
		switch mavmode {
		case COPTER_MODE_ACRO:
			ltmmode = types.Ltm_ACRO
			break
		case COPTER_MODE_STABILIZE:
			ltmmode = types.Ltm_HORIZON
			break
		case COPTER_MODE_ALT_HOLD:
			ltmmode = types.Ltm_ALTHOLD
			break
		case COPTER_MODE_POSHOLD:
			ltmmode = types.Ltm_POSHOLD
			break
		case COPTER_MODE_RTL:
			ltmmode = types.Ltm_RTH
			break
		case COPTER_MODE_AUTO:
			ltmmode = types.Ltm_WAYPOINTS
			break
		default:
			ltmmode = types.Ltm_ACRO
			break
		}
	}
	return ltmmode
}

func (lg *MWPJSON) Reader(m types.FlightMeta, ch chan interface{}) (types.LogSegment, bool) {
	stats := types.LogStats{}
	ls := types.LogSegment{}

	fh, err := os.Open(lg.name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "log file %s\n", err)
		return ls, false
	}
	defer fh.Close()

	rec := types.LogRec{}
	b := types.LogItem{}

	leffic := 0.0
	lwhkm := 0.0
	whacc := 0.0
	blt := 0.0
	llat := -999.0
	llon := 0.0

	var o map[string]interface{}
	scanner := bufio.NewScanner(fh)
	nl := int(0)

	for scanner.Scan() {
		l := scanner.Text()
		nl += 1
		if nl > m.End {
			break
		}
		if nl >= m.Start {
			json.Unmarshal([]byte(l), &o)
			done, cap := parse_json(o, &b)
			rec.Cap |= cap

			if done && b.Fix != 0 {
				if b.Vrange > stats.Max_range {
					stats.Max_range = b.Vrange
					stats.Max_range_time = uint64(lt-st) * 1000000
				}

				if b.Alt > stats.Max_alt {
					stats.Max_alt = b.Alt
					stats.Max_alt_time = uint64(lt-st) * 1000000
				}

				if b.Spd > 0 && b.Spd < 400 {
					if b.Spd > stats.Max_speed {
						stats.Max_speed = b.Spd
						stats.Max_speed_time = uint64(lt-st) * 1000000
					}
				}

				if blt > 0 {
					deltat := lt - blt
					if deltat > 0 {
						if (rec.Cap & types.CAP_AMPS) == types.CAP_AMPS {
							if b.Spd > 1 {
								b.Whkm = b.Amps * b.Volts / (3.6 * b.Spd)
								b.Effic = b.Amps * 1000 / (3.6 * b.Spd) // efficiency
							}
							leffic = b.Effic
							whacc += b.Amps * b.Volts * deltat / 3600
							b.WhAcc = whacc
							lwhkm = b.Whkm
						} else {
							b.Effic = leffic
							b.Whkm = lwhkm
						}
					}
				}
				blt = lt

				if homes.Flags != 0 {
					if llat == 999 {
						llat = b.Hlat
						llon = b.Hlon
					}
					if llat != 0 && llon != 0 {
						_, dx := geo.Csedist(b.Lat, b.Lon, llat, llon)
						dx *= 1852
						b.Tdist += dx
					}
				}
				llat = b.Lat
				llon = b.Lon

				if b.Amps > stats.Max_current {
					stats.Max_current = b.Amps
					stats.Max_current_time = uint64(lt-st) * 1000000
				}

				if ch != nil {
					ch <- b
				} else {
					rec.Items = append(rec.Items, b)
				}
				stats.Distance = b.Tdist / 1852.0
				b.Cse = 0xffff
			}
			if verbose {
				fmt.Fprintf(os.Stderr, ":DBG: nl=%d. %+v\n", nl, b)
			}
		}
	}
	stats.Max_range /= 1852.0
	srec := stats.Summary(uint64(lt-st) * 1000000)
	if ch != nil {
		ch <- srec
		return ls, true
	} else {
		ok := homes.Flags != 0 && len(rec.Items) > 0
		if ok {
			ls.L = rec
			ls.H = homes
			ls.M = srec
		}
		return ls, ok
	}
}
