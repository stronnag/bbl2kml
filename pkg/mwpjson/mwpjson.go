package mwpjson

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"geo"
	"os"
	"path/filepath"
	"time"
	"types"
)

type MWPJSON struct {
	name string
	meta []types.FlightMeta
}

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
	m, err := types.ReadMetaCache(o.name)
	if err != nil {
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

	var metas []types.FlightMeta
	r, err := os.Open(logfile)
	if err == nil {
		bp := filepath.Base(logfile)
		mt := types.FlightMeta{Logname: bp, Index: 1, Start: 1}
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			l := scanner.Text()
			var o map[string]interface{}
			json.Unmarshal([]byte(l), &o)
			lt = o["utime"].(float64)
			switch o["type"] {
			case "environment":
				mt.Craft = "No-name"
				mt.Firmware = "INAV"
				mt.Fwdate = "unknown"
				sec := int64(lt)
				nsec := int64((lt - float64(sec)) * 1e9)
				baseutc = time.Unix(sec, nsec)
				mt.Date = baseutc
			case "armed":
				if (o["armed"]).(bool) {
					if st == 0 {
						st = (o["utime"].(float64))
					}
					id += 1
				}
			default:
			}
		}
		mt.End = id
		mt.Duration = time.Duration(lt-st) * time.Second
		metas = append(metas, mt)
	}

	for j, mx := range metas {
		if mx.End-mx.Start > 64 {
			metas[j].Flags |= types.Has_Start | types.Is_Valid | types.Has_Craft
		}
	}
	if len(metas) == 0 {
		err = errors.New("No records in MWP JSON log")
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

func parse_json(l string, b *types.LogItem) (bool, uint16) {
	cap := uint16(0)
	done := false

	var o map[string]interface{}
	json.Unmarshal([]byte(l), &o)
	lt = o["utime"].(float64)
	switch o["type"] {
	case "environment":
		st = 0
		hlat = -999
		hlon = -999
		b.Status = 0
		b.Tdist = 0
	case "armed":
		if (o["armed"]).(bool) {
			if st == 0 {
				st = (o["utime"].(float64))
				b.Stamp = uint64((lt - st) * 1000 * 1000)
			}

			b.Status |= 1
			if b.Cse == 0xffff {
				b.Cse = b.Cog
			}
			done = true
		}
	case "analog2":
		b.Volts = o["voltage"].(float64)
		b.Amps = o["amps"].(float64) / 100.0
		b.Rssi = uint8(o["rssi"].(float64) * 100 / 1023)
		cap |= (types.CAP_RSSI_VALID | types.CAP_VOLTS | types.CAP_AMPS)

	case "status":
		b.NavMode = byte(o["nav_mode"].(float64))
		b.ActiveWP = uint8(o["wp_number"].(float64))
		switch b.NavMode {
		case 1, 2: // RTH
			b.Status |= (13 << 2)
			b.Fmode = types.FM_RTH
		case 3, 4: // PH
			b.Status |= (9 << 2)
			b.Fmode = types.FM_PH
		case 5, 6, 7: // WP
			b.Status |= (10 << 2)
			b.Fmode = types.FM_WP
			cap |= types.CAP_WPNO
		case 8, 10, 11, 12, 13, 14: // Land
			b.Status |= (15 << 2)
			b.Fmode = types.FM_LAND
		default:
			b.Fmode = types.FM_ACRO
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
		b.Hdop = uint16(o["hdop"].(float64))
		b.Cog = uint32(o["cse"].(float64))
		b.Spd = o["spd"].(float64)
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
		cap |= types.CAP_SPEED

	case "comp_gps":
		b.Vrange = o["range"].(float64)
		b.Bearing = int32(o["bearing"].(float64))

	case "ltm_xframe":
		if o["sensorok"].(float64) != 0 {
			b.HWfail = true
		}

	case "attitude":
		b.Cse = uint32(o["heading"].(float64))
		b.Roll = int16(o["angx"].(float64))
		b.Pitch = int16(o["angy"].(float64))

	case "altitude":
		cap |= types.CAP_ALTITUDE
		b.Alt = o["estalt"].(float64)
		// FIXME vario (mwp update)

	case "ltm_raw_sframe":
		// FIXME more fields (mwp update)
		b.Status = uint8(o["flags"].(float64))
		b.Volts = o["vbat"].(float64) / 1000.0
		b.Amps = o["vcurr"].(float64) / 1000.0
		b.Rssi = uint8(o["rssi"].(float64) * 100 / 255)
		cap |= (types.CAP_RSSI_VALID | types.CAP_VOLTS | types.CAP_AMPS)

	default:
	}
	return done, cap
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

	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		l := scanner.Text()
		done, cap := parse_json(l, &b)
		rec.Cap |= cap
		if done {
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
						b.Effic = b.Amps * 1000 / (3.6 * b.Spd) // efficiency
						leffic = b.Effic
						b.Whkm = b.Amps * b.Volts / (3.6 * b.Spd)
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
				_, dx := geo.Csedist(b.Lat, b.Lon, llat, llon)
				b.Tdist += (dx * 1852)
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
