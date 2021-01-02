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
	"path/filepath"
	"time"
	geo "github.com/stronnag/bbl2kml/pkg/geo"
	inav "github.com/stronnag/bbl2kml/pkg/inav"
	options "github.com/stronnag/bbl2kml/pkg/options"
)

type BBLStats struct {
	max_alt          float64
	max_alt_time     uint64
	max_range        float64
	max_range_time   uint64
	max_speed        float64
	max_speed_time   uint64
	max_current      float64
	max_current_time uint64
	distance         float64
	duration         uint64
}

const (
	FM_ACRO = iota
	FM_MANUAL
	FM_HORIZON
	FM_ANGLE
	FM_LAUNCH
	FM_RTH
	FM_WP
	FM_CRUISE3D
	FM_CRUISE2D
	FM_PH
	FM_AH
	FM_EMERG
	FM_FS
)

type BBLRec struct {
	stamp   uint64
	lat     float64
	lon     float64
	alt     float64
	cse     uint32
	spd     float64
	amps    float64
	fix     uint8
	numsat  uint8
	fmode   uint8
	rssi    uint8
	fmtext  string
	utc     time.Time
	fs      bool
	hlat    float64
	hlon    float64
	vrange  float64
	bearing int32 // -ve => not defined
	tdist   float64
}

var hdrs map[string]int

var mNames = []string{"Acro", "Manual", "Horizon", "Angle", "Launch", "RTH", "WP",
	"3CRS", "CRS", "PH", "AH", "EMERG", "F/S"}

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

func get_bbl_line(r []string, have_origin bool) BBLRec {
	b := BBLRec{}
	s, ok := get_rec_value(r, "amperage (A)")
	if ok {
		b.amps, _ = strconv.ParseFloat(s, 64)
	}
	s, ok = get_rec_value(r, "navPos[2]")
	if ok {
		b.alt, _ = strconv.ParseFloat(s, 64)
		b.alt = b.alt / 100.0
	}

	s, ok = get_rec_value(r, "GPS_numSat")
	if ok {
		i64, _ := strconv.Atoi(s)
		b.numsat = uint8(i64)
	}

	s, ok = get_rec_value(r, "GPS_fixType")
	if ok {
		i64, _ := strconv.Atoi(s)
		b.fix = uint8(i64)
	} else {
		if b.numsat > 5 {
			b.fix = 2
		} else if b.numsat > 0 {
			b.fix = 1
		} else {
			b.fix = 0
		}
	}
	s, ok = get_rec_value(r, "GPS_coord[0]")
	if ok {
		b.lat, _ = strconv.ParseFloat(s, 64)
	}
	s, ok = get_rec_value(r, "GPS_coord[1]")
	if ok {
		b.lon, _ = strconv.ParseFloat(s, 64)
	}
	s, ok = get_rec_value(r, "GPS_speed (m/s)")
	if ok {
		b.spd, _ = strconv.ParseFloat(s, 64)
	}
	s, ok = get_rec_value(r, "time (us)")
	if ok {
		i64, _ := strconv.ParseInt(s, 10, 64)
		b.stamp = uint64(i64)
	}

	md := uint8(0)
	s0, ok := get_rec_value(r, "flightModeFlags (flags)")
	s, ok = get_rec_value(r, "navState")
	if ok {
		i64, _ := strconv.ParseInt(s, 10, 64)
		switch {
		case inav.IsCruise2d(INAV_vers, int(i64)):
			md = FM_CRUISE2D
		case inav.IsCruise3d(INAV_vers, int(i64)):
			md = FM_CRUISE3D
		case inav.IsRTH(INAV_vers, int(i64)):
			md = FM_RTH
		case inav.IsWP(INAV_vers, int(i64)):
			md = FM_WP
		case inav.IsLaunch(INAV_vers, int(i64)):
			md = FM_LAUNCH
		case inav.IsPH(INAV_vers, int(i64)):
			md = FM_PH
		case inav.IsAH(INAV_vers, int(i64)):
			md = FM_AH
		case inav.IsEmerg(INAV_vers, int(i64)):
			md = FM_EMERG
		default:
			if strings.Contains(s0, "MANUAL") {
				md = FM_MANUAL
			} else if strings.Contains(s0, "ANGLE") {
				md = FM_ANGLE
			} else if strings.Contains(s0, "HORIZON") {
				md = FM_HORIZON
			}
		}
		if strings.Contains(s0, "NAVRTH") {
			md = FM_RTH
		}
	}
	b.fmode = md
	b.fmtext = mNames[md]

	s, ok = get_rec_value(r, "failsafePhase (flags)")
	if ok {
		b.fs = !strings.Contains(s, "IDLE")
	}

	if !have_origin {
		b.hlat = 0
		b.hlon = 0
		b.vrange = -1
		b.bearing = -1
		s, ok = get_rec_value(r, "GPS_home_lat")
		if ok {
			b.hlat, _ = strconv.ParseFloat(s, 64)
		}
		s, ok = get_rec_value(r, "GPS_home_lon")
		if ok {
			b.hlon, _ = strconv.ParseFloat(s, 64)
			b.bearing = -2
		} else {
			s, ok = get_rec_value(r, "homeDirection")
			if ok {
				i64, _ := strconv.Atoi(s)
				b.bearing = int32(i64)
			} else {
				s, ok = get_rec_value(r, "Azimuth")
				if ok {
					i64, _ := strconv.Atoi(s)
					b.bearing = int32((i64 + 180) % 360)
				}
			}

			if b.bearing != -1 {
				s, ok = get_rec_value(r, "Distance (m)")
				if ok {
					b.vrange, _ = strconv.ParseFloat(s, 64)
				}
			}
		}
	}

	s, ok = get_rec_value(r, "attitude[2]")
	if ok {
		i64, _ := strconv.Atoi(s)
		b.cse = uint32(i64 / 10)
	}

	s, ok = get_rec_value(r, "rssi")
	if ok {
		i64, _ := strconv.Atoi(s)
		b.rssi = uint8(i64 * 100 / 1023)
	}

	s, ok = get_rec_value(r, "dateTime")
	if ok {
		b.utc, _ = time.Parse(time.RFC3339Nano, s)
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

func Reader(bbfile string, meta BBLSummary) bool {
	idx := meta.Index
	cmd := exec.Command(options.Blackbox_decode,
		"--datetime", "--merge-gps", "--stdout", "--index",
		strconv.Itoa(idx), bbfile)
	out, err := cmd.StdoutPipe()
	defer cmd.Wait()
	defer out.Close()
	var homes []float64
	var recs []BBLRec

	r := csv.NewReader(out)
	r.TrimLeadingSpace = true

	err = cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start err=%v", err)
		os.Exit(1)
	}

	bblsmry := BBLStats{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	var home_lat, home_lon, llat, llon float64
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
			if br.fix > 1 && br.numsat > 5 {
				have_origin = true
				llat = br.lat
				llon = br.lon
				st = br.stamp
				homes = append(homes, br.lat, br.lon)
				if br.bearing == -1 {
					home_lat = br.lat
					home_lon = br.lon
				} else if br.bearing == -2 {
					home_lat = br.hlat
					home_lon = br.hlon
					homes = append(homes, home_lat, home_lon)
				} else {
					home_lat, home_lon = geo.Posit(br.lat, br.lon, float64(br.bearing), br.vrange/1852.0, true)
					homes = append(homes, home_lat, home_lon)
				}
			}
			if br.utc.IsZero() {
				basetime, _ = time.Parse("Jan 2 2006 15:04:05", meta.Fwdate)
			}
		} else {
			us := br.stamp
			if us > st {
				var d float64
				var c float64
				// Do the plot every 100ms
				if (us - dt) > 1000*uint64(options.Intvl) {
					if br.utc.IsZero() {
						br.utc = basetime.Add(time.Duration(us) * time.Microsecond)
					}
					c, d = geo.Csedist(home_lat, home_lon, br.lat, br.lon)
					br.bearing = int32(c)
					br.vrange = d * 1852.0

					if d > bblsmry.max_range {
						bblsmry.max_range = d
						bblsmry.max_range_time = us - st
					}

					if llat != br.lat && llon != br.lon {
						_, d = geo.Csedist(llat, llon, br.lat, br.lon)
						bblsmry.distance += d
						br.tdist = (bblsmry.distance * 1852.0)
					}

					llat = br.lat
					llon = br.lon
					dt = us
					recs = append(recs, br)
				}

				if br.alt > bblsmry.max_alt {
					bblsmry.max_alt = br.alt
					bblsmry.max_alt_time = us - st
				}

				if br.spd < 400 && br.spd > bblsmry.max_speed {
					bblsmry.max_speed = br.spd
					bblsmry.max_speed_time = us - st
				}

				if br.amps > bblsmry.max_current {
					bblsmry.max_current = br.amps
					bblsmry.max_current_time = us - st
				}
				lt = us
			}
		}
		if err != nil {
			log.Fatal(err)
		}
	}
	bblsmry.duration = lt - st
	bblsmry.max_range *= 1852.0
	bblsmry.distance *= 1852.0
	fmt.Printf("Altitude : %.1f m at %s\n", bblsmry.max_alt, Show_time(bblsmry.max_alt_time))
	fmt.Printf("Speed    : %.1f m/s at %s\n", bblsmry.max_speed, Show_time(bblsmry.max_speed_time))
	fmt.Printf("Range    : %.0f m at %s\n", bblsmry.max_range, Show_time(bblsmry.max_range_time))
	if bblsmry.max_current > 0 {
		fmt.Printf("Current  : %.1f A at %s\n", bblsmry.max_current, Show_time(bblsmry.max_current_time))
	}
	fmt.Printf("Distance : %.0f m\n", bblsmry.distance)
	fmt.Printf("Duration : %s\n", Show_time(bblsmry.duration))

	outfn := filepath.Base(bbfile)
	ext := filepath.Ext(outfn)
	if len(ext) < len(outfn) {
		outfn = outfn[0 : len(outfn)-len(ext)]
	}
	if options.Kml {
		ext = fmt.Sprintf(".%d.kml", idx)
	} else {
		ext = fmt.Sprintf(".%d.kmz", idx)
	}
	outfn = outfn + ext
	if len(homes) > 0 && len(recs) > 0 {
		GenerateKML(homes, recs, outfn, meta, bblsmry)
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
