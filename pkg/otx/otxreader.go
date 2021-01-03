package otx

import (
	"fmt"
	"io"
	"os"
	"encoding/csv"
	"sort"
	"strconv"
	"strings"
	"time"
	"math"
	"regexp"
	geo "github.com/stronnag/bbl2kml/pkg/geo"
	options "github.com/stronnag/bbl2kml/pkg/options"
	kmlgen "github.com/stronnag/bbl2kml/pkg/kmlgen"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
)

type hdrrec struct {
	i int
	u string
}

var hdrs map[string]hdrrec

const is_ARMED uint8 = 1
const is_CRSF uint8 = 2

func read_headers(r []string) {
	hdrs = make(map[string]hdrrec)
	rx := regexp.MustCompile(`(\w+)\(([A-Za-z/@]*)\)`)
	var k string
	var u string
	for i, s := range r {
		m := rx.FindAllStringSubmatch(s, -1)
		if len(m) > 0 {
			k = m[0][1]
			u = m[0][2]
		} else {
			k = s
			u = ""
		}
		hdrs[k] = hdrrec{i, u}
	}
}

func dump_headers() {
	var s string
	n := map[int][]string{}
	var a []int
	for k, v := range hdrs {
		if v.u == "" {
			s = k
		} else {
			s = fmt.Sprintf("%s(%s)", k, v.u)
		}
		n[v.i] = append(n[v.i], s)
	}

	for k := range n {
		a = append(a, k)
	}
	sort.Sort(sort.IntSlice(a))
	for _, k := range a {
		for _, s := range n[k] {
			fmt.Printf("%3d: %s\n", k, s)
		}
	}
}

func get_rec_value(r []string, key string) (string, string, bool) {
	var s string
	v, ok := hdrs[key]
	if ok {
		if v.i < len(r) {
			s = r[v.i]
		} else {
			ok = false
		}
	}
	return s, v.u, ok
}

func normalise_speed(v float64, u string) float64 {
	switch u {
	case "kmh":
		v = v / 3.6
	case "mph":
		v = v * 0.44704
	case "kts":
		v = v * 0.51444444
	}
	return v
}

func get_otx_line(r []string) (types.BBLRec, uint8) {
	b := types.BBLRec{}
	status := uint8(0)
	if s, _, ok := get_rec_value(r, "Tmp2"); ok {
		tmp2, _ := strconv.ParseInt(s, 10, 16)
		b.Numsat = uint8(tmp2 % 100)
		gfix := tmp2 / 1000
		if (gfix & 1) == 1 {
			b.Fix = 3
		} else if b.Numsat > 0 {
			b.Fix = 1
		} else {
			b.Fix = 0
		}
	}

	if s, _, ok := get_rec_value(r, "GPS"); ok {
		lstr := strings.Split(s, " ")
		if len(lstr) == 2 {
			b.Lat, _ = strconv.ParseFloat(lstr[0], 64)
			b.Lon, _ = strconv.ParseFloat(lstr[1], 64)
		}
	}

	if s, _, ok := get_rec_value(r, "Date"); ok {
		if s1, _, ok := get_rec_value(r, "Time"); ok {
			var sb strings.Builder
			sb.WriteString(s)
			sb.WriteByte(' ')
			sb.WriteString(s1)
			b.Utc, _ = time.Parse("2006-01-02 15:04:05.000", sb.String())
		}
	}

	if s, u, ok := get_rec_value(r, "Alt"); ok {
		b.Alt, _ = strconv.ParseFloat(s, 64)
		if u == "ft" {
			b.Alt = b.Alt * 0.3048
		}
	}

	if s, units, ok := get_rec_value(r, "GSpd"); ok {
		spd, _ := strconv.ParseFloat(s, 64)
		spd = normalise_speed(spd, units)
		if spd > 255 || spd < 0 {
			spd = 0
		}
		b.Spd = spd
	}

	if s, _, ok := get_rec_value(r, "Hdg"); ok {
		v, _ := strconv.ParseFloat(s, 64)
		b.Cse = uint32(v)
	}

	md := uint8(0)

	if s, _, ok := get_rec_value(r, "Tmp1"); ok {
		tmp1, _ := strconv.ParseInt(s, 10, 16)
		modeU := tmp1 % 10
		modeT := (tmp1 % 100) / 10
		modeH := (tmp1 % 1000) / 100
		modeK := (tmp1 % 10000) / 1000
		modeJ := tmp1 / 10000

		if (modeU & 4) == 4 {
			status |= is_ARMED
		}

		switch modeT {
		case 0:
			md = types.FM_ACRO
		case 1:
			md = types.FM_ANGLE
		case 2:
			md = types.FM_HORIZON
		case 4:
			md = types.FM_MANUAL
		}

		if (modeH & 2) == 2 {
			md = types.FM_AH
		}
		if (modeH & 4) == 4 {
			md = types.FM_PH
		}

		if modeK == 1 {
			md = types.FM_RTH
		} else if modeK == 2 {
			md = types.FM_WP
		} else if modeK == 8 {
			if md == types.FM_AH || md == types.FM_PH {
				md = types.FM_CRUISE3D
			} else {
				md = types.FM_CRUISE2D
			}
		}
		if modeJ == 4 {
			b.Fs = true
		}
	}

	if s, _, ok := get_rec_value(r, "RSSI"); ok {
		rssi, _ := strconv.ParseInt(s, 10, 32)
		b.Rssi = uint8(rssi)
	}

	if s, _, ok := get_rec_value(r, "1RSS"); ok {
		status |= is_CRSF
		rssi, _ := strconv.ParseInt(s, 10, 32)
		b.Rssi = uint8(rssi)

		if s, _, ok = get_rec_value(r, "FM"); ok {
			md = 0
			status |= is_ARMED
			switch s {
			case "0", "OK", "WAIT", "!ERR":
				status &= ^is_ARMED
			case "ACRO", "AIR":
				md = types.FM_ACRO
			case "ANGL", "STAB":
				md = types.FM_ANGLE
			case "HOR":
				md = types.FM_HORIZON
			case "MANU":
				md = types.FM_MANUAL
			case "AH":
				md = types.FM_AH
			case "HOLD":
				md = types.FM_PH
			case "CRS":
				md = types.FM_CRUISE2D
			case "3CRS":
				md = types.FM_CRUISE3D
			case "WP":
				md = types.FM_WP
			case "RTH":
				md = types.FM_RTH
			case "!FS!":
				b.Fs = true
			}

			if s == "0" {
				if s, _, ok := get_rec_value(r, "Thr"); ok {
					thr, _ := strconv.ParseInt(s, 10, 32)
					if thr > -800 {
						status |= is_ARMED
					}
				}
			}
		}

		if s, _, ok := get_rec_value(r, "Sats"); ok {
			ns, _ := strconv.ParseInt(s, 10, 16)
			b.Numsat = uint8(ns)
			if ns > 5 {
				b.Fix = 3
			} else if ns > 0 {
				b.Fix = 1
			} else {
				b.Fix = 0
			}
		}

		if s, _, ok := get_rec_value(r, "Yaw"); ok {
			v1, _ := strconv.ParseFloat(s, 64)
			cse := to_degrees(v1)
			if cse < 0 {
				cse += 360.0
			}
			b.Cse = uint32(cse)
		}
	}
	b.Fmode = md
	b.Fmtext = types.Mnames[md]

	return b, status
}

func to_degrees(rad float64) float64 {
	return (rad * 180.0 / math.Pi)
}

func to_radians(deg float64) float64 {
	return (deg * math.Pi / 180.0)
}

func calc_speed(b types.BBLRec, tdiff time.Duration, llat, llon float64) float64 {
	spd := 0.0
	if tdiff > 0 && llat != 0 && llon != 0 {
		// Flat earth
		x := math.Abs(to_radians(b.Lon-llon) * math.Cos(to_radians(b.Lat)))
		y := math.Abs(to_radians(b.Lat - llat))
		d := math.Sqrt(x*x+y*y) * 6371009.0
		spd = d / tdiff.Seconds()
	}
	return spd
}

func Reader(otxfile string, only_armed bool) bool {
	hlat := 0.0
	hlon := 0.0
	llat := 0.0
	llon := 0.0
	cdist := 0.0

	idx := 0

	var homes []float64
	var recs []types.BBLRec

	fh, err := os.Open(otxfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "log file %s\n", err)
		os.Exit(-1)
	}
	defer fh.Close()

	r := csv.NewReader(fh)
	r.TrimLeadingSpace = true

	//split_sec := 30 // to be parameterised
	//	var armtime time.Time
	var lt, st time.Time

	for i := 0; ; i++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if i == 0 {
			read_headers(record)
			if options.Dump {
				dump_headers()
				return true
			}
		} else {
			b, status := get_otx_line(record)
			if only_armed && (status&is_ARMED) == 0 && b.Alt < 10 && b.Spd < 7 {
				cdist = 0.0
				continue
			}

			if st.IsZero() {
				st = b.Utc
				lt = st
			}

			if b.Utc.Sub(lt).Seconds() > time.Duration(120*time.Second).Seconds() {
				fmt.Fprintf(os.Stderr, "Splitting at %v\n", b.Utc)
				if len(homes) > 0 && len(recs) > 0 {
					if idx == 0 {
						idx = 1
					}
					outfn := kmlgen.GenKmlName(otxfile, idx)
					kmlgen.GenerateKML(homes, recs, outfn, types.BBLSummary{}, types.BBLStats{})
					homes = nil
					recs = nil
					st = b.Utc
					lt = st
					cdist = 0.0
					hlat = 0
					hlon = 0
					llat = 0
					llon = 0
					idx += 1
				}
			}

			if hlat == 0 && hlon == 0 {
				if b.Fix > 1 && b.Numsat > 5 {
					hlat = b.Lat
					hlon = b.Lon
					llat = b.Lat
					llon = b.Lon
					homes = append(homes, hlat, hlon)
				}
			}

			tdiff := b.Utc.Sub(lt)
			if (status & is_CRSF) == is_CRSF {
				b.Spd = calc_speed(b, tdiff, llat, llon)
			}

			var c, d float64
			if hlat != 0 && hlon != 0 {
				c, d = geo.Csedist(hlat, hlon, b.Lat, b.Lon)
				b.Bearing = int32(c)
				b.Vrange = d * 1852.0

				if llat != b.Lat && llon != b.Lon {
					_, d = geo.Csedist(llat, llon, b.Lat, b.Lon)
					cdist += d * 1852.0
				}
			}
			b.Tdist = cdist
			llat = b.Lat
			llon = b.Lon
			lt = b.Utc
			recs = append(recs, b)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "reader %s\n", err)
			os.Exit(-1)
		}
	}

	if len(homes) > 0 && len(recs) > 0 {
		outfn := kmlgen.GenKmlName(otxfile, idx)
		kmlgen.GenerateKML(homes, recs, outfn, types.BBLSummary{}, types.BBLStats{})
		return true
	}
	return false
}
