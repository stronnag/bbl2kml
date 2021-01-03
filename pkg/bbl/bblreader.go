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

var INAV_vers int

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

func get_bbl_line(r []string, have_origin bool) types.BBLRec {
	b := types.BBLRec{}
	s, ok := get_rec_value(r, "amperage (A)")
	if ok {
		b.Amps, _ = strconv.ParseFloat(s, 64)
	}
	s, ok = get_rec_value(r, "navPos[2]")
	if ok {
		b.Alt, _ = strconv.ParseFloat(s, 64)
		b.Alt = b.Alt / 100.0
	}

	s, ok = get_rec_value(r, "GPS_numSat")
	if ok {
		i64, _ := strconv.Atoi(s)
		b.Numsat = uint8(i64)
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
		if inav.IsCruise3d(INAV_vers, int(i64)) {
			md = types.FM_CRUISE3D
		} else if inav.IsCruise2d(INAV_vers, int(i64)) {
			md = types.FM_CRUISE2D
		} else if inav.IsRTH(INAV_vers, int(i64)) {
			md = types.FM_RTH
		} else if inav.IsWP(INAV_vers, int(i64)) {
			md = types.FM_WP
		} else if inav.IsLaunch(INAV_vers, int(i64)) {
			md = types.FM_LAUNCH
		} else if inav.IsPH(INAV_vers, int(i64)) {
			md = types.FM_PH
		} else if inav.IsAH(INAV_vers, int(i64)) {
			md = types.FM_AH
		} else if inav.IsEmerg(INAV_vers, int(i64)) {
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

func Reader(bbfile string, meta types.BBLSummary) bool {
	idx := meta.Index
	cmd := exec.Command(options.Blackbox_decode,
		"--datetime", "--merge-gps", "--stdout", "--index",
		strconv.Itoa(idx), bbfile)
	out, err := cmd.StdoutPipe()
	defer cmd.Wait()
	defer out.Close()
	var homes types.HomeRec
	var recs []types.BBLRec

	r := csv.NewReader(out)
	r.TrimLeadingSpace = true

	err = cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start err=%v", err)
		os.Exit(1)
	}

	bblsmry := types.BBLStats{false, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	var llat, llon float64
	var dt, st, lt uint64
	var basetime time.Time
	have_origin := false

	INAV_vers = 0
	fwvers := strings.Split(meta.Firmware, " ")
	if len(fwvers) == 4 {
		parts := strings.Split(fwvers[1], ".")
		if len(parts) == 3 {
			mask := (1 << 16)
			for _, p := range parts {
				v, _ := strconv.Atoi(p)
				INAV_vers = INAV_vers + (v * mask)
				mask = mask >> 8
			}
		}
	}

	for i := 0; ; i++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if i == 0 {
			hdrs = get_headers(record)
			if options.Dump {
				dump_headers(hdrs)
				return true
			}
		}

		br := get_bbl_line(record, have_origin)

		if !have_origin {
			if br.Fix > 1 && br.Numsat > 5 {
				have_origin = true
				llat = br.Lat
				llon = br.Lon
				st = br.Stamp
				homes.HomeLat = br.Lat
				homes.HomeLon = br.Lon
				homes.HomeAlt = br.GAlt
				homes.Flags = types.HOME_ARM | types.HOME_ALT
				if br.Bearing == -2 {
					homes.SafeLat = br.Hlat
					homes.SafeLon = br.Hlon
					homes.Flags |= types.HOME_SAFE
				} else if br.Bearing > -1 {
					hlat, hlon := geo.Posit(br.Lat, br.Lon, float64(br.Bearing), br.Vrange/1852.0, true)
					homes.SafeLat = hlat
					homes.SafeLon = hlon
					homes.Flags |= types.HOME_SAFE
				}
			}
			if br.Utc.IsZero() {
				basetime, _ = time.Parse("Jan 2 2006 15:04:05", meta.Fwdate)
			}
		} else {
			us := br.Stamp
			if us > st {
				var d float64
				var c float64
				// Do the plot every 100ms
				if (us - dt) > 1000*uint64(options.Intvl) {
					if br.Utc.IsZero() {
						br.Utc = basetime.Add(time.Duration(us) * time.Microsecond)
					}
					c, d = geo.Csedist(homes.HomeLat, homes.HomeLon, br.Lat, br.Lon)
					br.Bearing = int32(c)
					br.Vrange = d * 1852.0

					if d > bblsmry.Max_range {
						bblsmry.Max_range = d
						bblsmry.Max_range_time = us - st
					}

					if llat != br.Lat && llon != br.Lon {
						_, d = geo.Csedist(llat, llon, br.Lat, br.Lon)
						bblsmry.Distance += d
					}

					br.Tdist = (bblsmry.Distance * 1852.0)

					llat = br.Lat
					llon = br.Lon
					dt = us
					recs = append(recs, br)
				}

				if br.Alt > bblsmry.Max_alt {
					bblsmry.Max_alt = br.Alt
					bblsmry.Max_alt_time = us - st
				}

				if br.Spd < 400 && br.Spd > bblsmry.Max_speed {
					bblsmry.Max_speed = br.Spd
					bblsmry.Max_speed_time = us - st
				}

				if br.Amps > bblsmry.Max_current {
					bblsmry.Max_current = br.Amps
					bblsmry.Max_current_time = us - st
				}
				lt = us
			}
		}
		if err != nil {
			log.Fatal(err)
		}
	}
	bblsmry.Duration = lt - st
	bblsmry.Max_range *= 1852.0
	bblsmry.Distance *= 1852.0
	fmt.Printf("Altitude : %.1f m at %s\n", bblsmry.Max_alt, Show_time(bblsmry.Max_alt_time))
	fmt.Printf("Speed    : %.1f m/s at %s\n", bblsmry.Max_speed, Show_time(bblsmry.Max_speed_time))
	fmt.Printf("Range    : %.0f m at %s\n", bblsmry.Max_range, Show_time(bblsmry.Max_range_time))
	if bblsmry.Max_current > 0 {
		fmt.Printf("Current  : %.1f A at %s\n", bblsmry.Max_current, Show_time(bblsmry.Max_current_time))
	}
	fmt.Printf("Distance : %.0f m\n", bblsmry.Distance)
	fmt.Printf("Duration : %s\n", Show_time(bblsmry.Duration))

	if homes.Flags != 0 && len(recs) > 0 {
		meta.Valid = true
		bblsmry.Valid = true
		outfn := kmlgen.GenKmlName(bbfile, idx)
		kmlgen.GenerateKML(homes, recs, outfn, meta, bblsmry)
		return true
	}
	return false
}

func Show_time(t uint64) string {
	secs := t / 1000000
	m := secs / 60
	s := secs % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}
