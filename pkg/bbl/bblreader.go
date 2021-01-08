package bbl

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"encoding/csv"
	"sort"
	"strconv"
	"strings"
	"time"
	geo "github.com/stronnag/bbl2kml/pkg/geo"
	inav "github.com/stronnag/bbl2kml/pkg/inav"
	options "github.com/stronnag/bbl2kml/pkg/options"
	kmlgen "github.com/stronnag/bbl2kml/pkg/kmlgen"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
)

var hdrs map[string]int

var inav_vers int

func get_rec_value(r []string, key string) (string, bool) {
	var s string
	i, ok := hdrs[key]
	if ok {
		if i < len(r) {
			s = r[i]
		} else {
			return "", false
		}
	}
	return s, ok
}

func dataCapability() uint8 {
	var ret uint8 = 0
	if _, ok := hdrs["amperage (A)"]; ok {
		ret |= types.CAP_AMPS
	}
	if _, ok := hdrs["vbat (V)"]; ok {
		ret |= types.CAP_VOLTS
	}
	if _, ok := hdrs["energyCumulative (mAh)"]; ok {
		ret |= types.CAP_ENERGY
	}
	return ret
}

func get_bbl_line(r []string, have_origin bool) types.LogItem {
	b := types.LogItem{}

	s, ok := get_rec_value(r, "GPS_numSat")
	if ok {
		i64, _ := strconv.Atoi(s)
		b.Numsat = uint8(i64)
	}

	if s, ok = get_rec_value(r, "vbat (V)"); ok {
		b.Volts, _ = strconv.ParseFloat(s, 64)
	} else if s, ok = get_rec_value(r, "vbatLatest (V)"); ok {
		b.Volts, _ = strconv.ParseFloat(s, 64)
	}

	s, ok = get_rec_value(r, "navPos[2]")
	if ok {
		b.Alt, _ = strconv.ParseFloat(s, 64)
		b.Alt = b.Alt / 100.0
	}

	s, ok = get_rec_value(r, "GPS_fixType")
	if ok {
		i64, _ := strconv.Atoi(s)
		b.Fix = uint8(i64)
	} else {
		if b.Numsat > 5 {
			b.Fix = 2
		} else if b.Numsat > 0 {
			b.Fix = 1
		} else {
			b.Fix = 0
		}
	}
	s, ok = get_rec_value(r, "GPS_coord[0]")
	if ok {
		b.Lat, _ = strconv.ParseFloat(s, 64)
	}
	s, ok = get_rec_value(r, "GPS_coord[1]")
	if ok {
		b.Lon, _ = strconv.ParseFloat(s, 64)
	}

	s, ok = get_rec_value(r, "GPS_altitude")
	if ok {
		b.GAlt, _ = strconv.ParseFloat(s, 64)
	}

	s, ok = get_rec_value(r, "GPS_speed (m/s)")
	if ok {
		b.Spd, _ = strconv.ParseFloat(s, 64)
	}
	s, ok = get_rec_value(r, "time (us)")
	if ok {
		i64, _ := strconv.ParseInt(s, 10, 64)
		b.Stamp = uint64(i64)
	}

	md := uint8(0)
	s0, ok := get_rec_value(r, "flightModeFlags (flags)")
	s, ok = get_rec_value(r, "navState")
	if ok {
		i64, _ := strconv.ParseInt(s, 10, 64)
		if inav.IsCruise3d(inav_vers, int(i64)) {
			md = types.FM_CRUISE3D
		} else if inav.IsCruise2d(inav_vers, int(i64)) {
			md = types.FM_CRUISE2D
		} else if inav.IsRTH(inav_vers, int(i64)) {
			md = types.FM_RTH
		} else if inav.IsWP(inav_vers, int(i64)) {
			md = types.FM_WP
		} else if inav.IsLaunch(inav_vers, int(i64)) {
			md = types.FM_LAUNCH
		} else if inav.IsPH(inav_vers, int(i64)) {
			md = types.FM_PH
		} else if inav.IsAH(inav_vers, int(i64)) {
			md = types.FM_AH
		} else if inav.IsEmerg(inav_vers, int(i64)) {
			md = types.FM_EMERG
		} else {
			if strings.Contains(s0, "MANUAL") {
				md = types.FM_MANUAL
			} else if strings.Contains(s0, "ANGLE") {
				md = types.FM_ANGLE
			} else if strings.Contains(s0, "HORIZON") {
				md = types.FM_HORIZON
			}
		}
	}
	// fallback for old inav bug
	if strings.Contains(s0, "NAVRTH") {
		md = types.FM_RTH
	}

	b.Fmode = md
	b.Fmtext = types.Mnames[md]

	s, ok = get_rec_value(r, "failsafePhase (flags)")
	if ok {
		b.Fs = !strings.Contains(s, "IDLE")
	}

	if !have_origin {
		b.Hlat = 0
		b.Hlon = 0
		b.Vrange = -1
		b.Bearing = -1
		s, ok = get_rec_value(r, "GPS_home_lat")
		if ok {
			b.Hlat, _ = strconv.ParseFloat(s, 64)
		}
		s, ok = get_rec_value(r, "GPS_home_lon")
		if ok {
			b.Hlon, _ = strconv.ParseFloat(s, 64)
			b.Bearing = -2
		} else {
			s, ok = get_rec_value(r, "homeDirection")
			if ok {
				i64, _ := strconv.Atoi(s)
				b.Bearing = int32(i64)
			} else {
				s, ok = get_rec_value(r, "Azimuth")
				if ok {
					i64, _ := strconv.Atoi(s)
					b.Bearing = int32((i64 + 180) % 360)
				}
			}

			if b.Bearing != -1 {
				s, ok = get_rec_value(r, "Distance (m)")
				if ok {
					b.Vrange, _ = strconv.ParseFloat(s, 64)
				}
			}
		}
	}

	s, ok = get_rec_value(r, "attitude[2]")
	if ok {
		i64, _ := strconv.Atoi(s)
		b.Cse = uint32(i64 / 10)
	}

	s, ok = get_rec_value(r, "rssi")
	if ok {
		i64, _ := strconv.Atoi(s)
		b.Rssi = uint8(i64 * 100 / 1023)
	}

	s, ok = get_rec_value(r, "dateTime")
	if ok {
		b.Utc, _ = time.Parse(time.RFC3339Nano, s)
	}

	s, ok = get_rec_value(r, "amperage (A)")
	if ok {
		b.Amps, _ = strconv.ParseFloat(s, 64)
	}

	if s, ok = get_rec_value(r, "energyCumulative (mAh)"); ok {
		b.Energy, _ = strconv.ParseFloat(s, 64)
	}

	return b
}

func get_headers(r []string) map[string]int {
	m := make(map[string]int)
	for i, s := range r {
		m[s] = i
	}
	return m
}

func dump_headers(m map[string]int) {
	n := map[int][]string{}
	var a []int
	for k, v := range m {
		n[v] = append(n[v], k)
	}
	for k := range n {
		a = append(a, k)
	}
	sort.Sort(sort.IntSlice(a))
	for _, k := range a {
		for _, s := range n[k] {
			fmt.Printf("%s, %d\n", s, k)
		}
	}
}

func Reader(bbfile string, meta BBLMeta) bool {
	idx := meta.Index
	cmd := exec.Command(options.Blackbox_decode,
		"--datetime", "--merge-gps", "--stdout", "--index",
		strconv.Itoa(idx), bbfile)
	out, err := cmd.StdoutPipe()
	defer cmd.Wait()
	defer out.Close()
	var homes types.HomeRec
	var rec types.LogRec

	r := csv.NewReader(out)
	r.TrimLeadingSpace = true

	err = cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start err=%v", err)
		os.Exit(1)
	}

	stats := types.LogStats{}

	var llat, llon float64
	var dt, st, lt uint64
	var basetime time.Time
	have_origin := false

	inav_vers = 0
	fwvers := strings.Split(meta.Firmware, " ")
	if len(fwvers) == 4 {
		parts := strings.Split(fwvers[1], ".")
		if len(parts) == 3 {
			mask := (1 << 16)
			for _, p := range parts {
				v, _ := strconv.Atoi(p)
				inav_vers = inav_vers + (v * mask)
				mask = mask >> 8
			}
		}
	}

	leffic := 0.0
	for i := 0; ; i++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if i == 0 {
			hdrs = get_headers(record)
			rec.Cap = dataCapability()
			if options.Dump {
				dump_headers(hdrs)
				return true
			}
		}

		b := get_bbl_line(record, have_origin)

		if !have_origin {
			if b.Fix > 1 && b.Numsat > 5 {
				have_origin = true
				llat = b.Lat
				llon = b.Lon
				st = b.Stamp
				homes.HomeLat = b.Lat
				homes.HomeLon = b.Lon
				homes.HomeAlt = b.GAlt
				homes.Flags = types.HOME_ARM | types.HOME_ALT
				if b.Bearing == -2 {
					_, dh := geo.Csedist(b.Hlat, b.Hlon, b.Lat, b.Lon)
					if dh > 2.0/1852.0 {
						homes.SafeLat = b.Hlat
						homes.SafeLon = b.Hlon
						homes.Flags |= types.HOME_SAFE
					}
				} else if b.Bearing > -1 {
					hlat, hlon := geo.Posit(b.Lat, b.Lon, float64(b.Bearing), b.Vrange/1852.0, true)
					homes.SafeLat = hlat
					homes.SafeLon = hlon
					homes.Flags |= types.HOME_SAFE
				}
			}
			if b.Utc.IsZero() {
				basetime, _ = time.Parse("Jan 2 2006 15:04:05", meta.Fwdate)
			}
		} else {
			us := b.Stamp
			if us > st {
				var d float64
				var c float64
				// Do the plot every 100ms
				if (us - dt) > 1000*uint64(options.Intvl) {
					if b.Utc.IsZero() {
						b.Utc = basetime.Add(time.Duration(us) * time.Microsecond)
					}
					c, d = geo.Csedist(homes.HomeLat, homes.HomeLon, b.Lat, b.Lon)
					b.Bearing = int32(c)
					b.Vrange = d * 1852.0

					if d > stats.Max_range {
						stats.Max_range = d
						stats.Max_range_time = us - st
					}

					if llat != b.Lat || llon != b.Lon {
						_, d = geo.Csedist(llat, llon, b.Lat, b.Lon)
						stats.Distance += d
					}
					b.Tdist = (stats.Distance * 1852.0)
					llat = b.Lat
					llon = b.Lon

					if (rec.Cap & types.CAP_AMPS) == types.CAP_AMPS {
						if d > 0 {
							deltat := float64((us - dt)) / 1000000.0 // seconds
							aspd := d * 1852 / deltat                // m/s
							b.Effic = b.Amps * 1000 / (3.6 * aspd)   // efficiency
							leffic = b.Effic
						} else {
							b.Effic = leffic
						}
					}
					if b.Rssi > 0 {
						rec.Cap |= types.CAP_RSSI_VALID
					}

					rec.Items = append(rec.Items, b)
					dt = us
				}

				if b.Alt > stats.Max_alt {
					stats.Max_alt = b.Alt
					stats.Max_alt_time = us - st
				}

				if b.Spd < 400 && b.Spd > stats.Max_speed {
					stats.Max_speed = b.Spd
					stats.Max_speed_time = us - st
				}

				if b.Amps > stats.Max_current {
					stats.Max_current = b.Amps
					stats.Max_current_time = us - st
				}
				lt = us
			}
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	stats.ShowSummary(lt - st)
	if homes.Flags != 0 && len(rec.Items) > 0 {
		outfn := kmlgen.GenKmlName(bbfile, idx)
		kmlgen.GenerateKML(homes, rec, outfn, &meta, stats)
		return true
	}
	return false
}
