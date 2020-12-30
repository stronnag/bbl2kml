package main

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
	utc     string
	fs      bool
	hlat    float64
	hlon    float64
	vrange  float64
	bearing int32 // -ve => not defined
	tdist   float64
}

var hdrs map[string]int

var mNames = []string{"Acro", "Manual", "Horizon", "Angle", "Launch", "RTH", "WP",
	"3CRS", "CRS", "PH", "AH", "F/S"}

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
	s, ok = get_rec_value(r, "GPS_fixType")
	if ok {
		i64, _ := strconv.Atoi(s)
		b.fix = uint8(i64)
	}
	s, ok = get_rec_value(r, "GPS_numSat")
	if ok {
		i64, _ := strconv.Atoi(s)
		b.numsat = uint8(i64)
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
		switch i64 {
		case 29, 30, 31:
			md = FM_CRUISE2D
		case 32, 33, 34:
			md = FM_CRUISE3D
		case 8, 9, 10, 11, 12, 13, 14, 36:
			md = FM_RTH
		case 15, 16, 17, 18, 19, 20, 21, 35, 37:
			md = FM_WP
		case 25, 26, 28:
			md = FM_LAUNCH
		case 6, 7:
			md = FM_PH
		case 2, 3:
			md = FM_AH
		default:
			if strings.Contains(s0, "MANUAL") {
				md = FM_MANUAL
			}
			if strings.Contains(s0, "ANGLE") {
				md = FM_ANGLE
			}
			if strings.Contains(s0, "HORIZON") {
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
		b.utc = s
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

func bblreader(bbfile string, meta BBLSummary) {
	idx := meta.index
	cmd := exec.Command(Options.blackbox_decode,
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

	have_origin := false

	for i := 0; ; i++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if i == 0 {
			hdrs = get_headers(record)
			if Options.dump {
				dump_headers(hdrs)
				return
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
					home_lat, home_lon = Posit(br.lat, br.lon, float64(br.bearing), br.vrange/1852.0, true)
					homes = append(homes, home_lat, home_lon)
				}
			}
		} else {
			us := br.stamp
			var d float64
			var c float64
			// Do the plot every 100ms
			if (us - dt) > 1000*uint64(Options.intvl) {
				c, d = Csedist(home_lat, home_lon, br.lat, br.lon)
				br.bearing = int32(c)
				br.vrange = d * 1852.0

				if d > bblsmry.max_range {
					bblsmry.max_range = d
					bblsmry.max_range_time = us - st
				}

				if llat != br.lat && llon != br.lon {
					_, d = Csedist(llat, llon, br.lat, br.lon)
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
	if Options.kml {
		ext = fmt.Sprintf(".%d.kml", idx)
	} else {
		ext = fmt.Sprintf(".%d.kmz", idx)
	}
	outfn = outfn + ext
	GenerateKML(homes, recs, outfn, meta, bblsmry)
}

func Show_time(t uint64) string {
	secs := t / 1000000
	m := secs / 60
	s := secs % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}
