package otx

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

import (
	"geo"
	"options"
	"types"
)

const LOGTIMEPARSE = "2006-01-02 15:04:05.000"
const TIMEDATE = "2006-01-02 15:04:05"

type OTXLOG struct {
	name string
	meta []types.FlightMeta
}

func NewOTXReader(fn string) OTXLOG {
	var l OTXLOG
	l.name = fn
	l.meta = nil
	return l
}

func (o *OTXLOG) LogType() byte {
	return types.LOGOTX
}

func (o *OTXLOG) GetMetas() ([]types.FlightMeta, error) {
	m, err := types.ReadMetaCache(o.name)
	if err != nil {
		m, err = metas(o.name)
		types.WriteMetaCache(o.name, m)
	}
	o.meta = m
	return m, err
}

func (o *OTXLOG) GetDurations() {
}

func (o *OTXLOG) Dump() {
	if hdrs == nil || len(hdrs) == 0 {
		fh, err := os.Open(o.name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "log file %s\n", err)
			return
		}
		defer fh.Close()
		r := csv.NewReader(fh)
		r.TrimLeadingSpace = true
		record, err := r.Read()
		read_headers(record) // for future usage
	}
	dump_headers()
}

type hdrrec struct {
	i int
	u string
}

var hdrs map[string]hdrrec

func read_headers(r []string) {
	hdrs = make(map[string]hdrrec)
	rx := regexp.MustCompile(`(\w+)\(([A-Za-z/@%°]*)\)`)
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

func metas(otxfile string) ([]types.FlightMeta, error) {
	var metas []types.FlightMeta

	fh, err := os.Open(otxfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "log file %s\n", err)
		return metas, err
	}
	defer fh.Close()

	basefile := filepath.Base(otxfile)
	r := csv.NewReader(fh)
	r.TrimLeadingSpace = true

	var lasttm time.Time
	dindex := -1
	tindex := -1

	idx := 0
	for i := 1; ; i++ {
		record, err := r.Read()
		if err == io.EOF {
			metas[idx-1].End = (i - 1)
			metas[idx-1].Duration = lasttm.Sub(metas[idx-1].Date)
			break
		}
		if i == 1 {
			read_headers(record) // for future usage
			for j, s := range record {
				switch s {
				case "Date":
					dindex = j
				case "Time":
					tindex = j
				}
				if dindex != -1 && tindex != -1 {
					break
				}
			}
		} else {
			var sb strings.Builder
			sb.WriteString(record[dindex])
			sb.WriteByte(' ')
			sb.WriteString(record[tindex])
			t_utc, _ := time.Parse(LOGTIMEPARSE, sb.String())
			if i == 2 || (options.Config.SplitTime > 0 && t_utc.Sub(lasttm).Seconds() > (time.Duration(options.Config.SplitTime)*time.Second).Seconds()) {
				if idx > 0 {
					metas[idx-1].End = i - 1
					metas[idx-1].Duration = lasttm.Sub(metas[idx-1].Date)
				}
				idx += 1
				mt := types.FlightMeta{Logname: basefile, Date: t_utc, Index: idx, Start: i}
				metas = append(metas, mt)
			}
			lasttm = t_utc
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "reader %s\n", err)
			return metas, err
		}
	}

	for j, mx := range metas {
		if mx.End-mx.Start > 64 {
			metas[j].Flags = types.Has_Start | types.Is_Valid
		}
	}
	if len(metas) == 0 {
		err = errors.New("No records in OTX file")
	}
	return metas, err
}

func dump_headers() {
	var s string
	n := map[int][]string{}
	var a []int
	for k, v := range hdrs {
		if v.u == "" {
			s = k
		} else {
			s = fmt.Sprintf("%s units=(%s)", k, v.u)
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

func dataCapability() uint16 {
	var ret uint16 = 0
	var ok bool
	if _, ok = hdrs["Curr"]; ok {
		ret |= types.CAP_AMPS
	}
	if _, ok = hdrs["VFAS"]; ok {
		ret |= types.CAP_VOLTS
	} else if _, ok = hdrs["RxBt"]; ok {
		ret |= types.CAP_VOLTS
	}

	var v hdrrec
	if v, ok = hdrs["Fuel"]; ok {
		ret |= types.CAP_ENERGY
	} else if v, ok = hdrs["Capa"]; ok {
		ret |= types.CAP_ENERGY
	}

	if ret&types.CAP_VOLTS|types.CAP_AMPS == types.CAP_VOLTS|types.CAP_AMPS && !(v.u == "mwh" || v.u == "mWh") {
		ret |= (types.CAP_ENERGY | types.CAP_ENERGYC)
	}

	if _, ok = hdrs["GSpd"]; ok {
		ret |= types.CAP_SPEED
	}

	if _, ok = hdrs["Alt"]; ok {
		ret |= types.CAP_ALTITUDE
	}

	return ret
}

func normalise_units(v float64, u string) float64 {
	switch u {
	case "kmh":
		v /= 3.6
	case "mph":
		v *= 0.44704
	case "kts":
		v *= 0.51444444
	case "ft":
		v *= 0.3048
	}
	return v
}

func get_otx_line(r []string) types.LogItem {
	b := types.LogItem{}
	status := uint8(0)
	if s, _, ok := get_rec_value(r, "Tmp2"); ok {
		tmp2, _ := strconv.ParseInt(s, 10, 32)
		b.Numsat = uint8(tmp2 % 100)
		gfix := tmp2 / 1000
		if (gfix & 1) == 1 {
			b.Fix = 2
		} else if b.Numsat > 0 {
			b.Fix = 1
		} else {
			b.Fix = 0
		}
		hdp := uint16((tmp2 % 1000) / 100)
		b.Hdop = uint16(550 - (hdp * 50))
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
			b.Utc, _ = time.Parse(LOGTIMEPARSE, sb.String())
		}
	}

	if s, u, ok := get_rec_value(r, "Alt"); ok {
		b.Alt, _ = strconv.ParseFloat(s, 64)
		b.Alt = normalise_units(b.Alt, u)
	}

	if s, u, ok := get_rec_value(r, "GAlt"); ok {
		b.GAlt, _ = strconv.ParseFloat(s, 64)
		b.GAlt = normalise_units(b.GAlt, u)
	} else {
		b.GAlt = -999999.9
	}

	if s, units, ok := get_rec_value(r, "GSpd"); ok {
		spd, _ := strconv.ParseFloat(s, 64)
		spd = normalise_units(spd, units)
		if spd > 255 || spd < 0 {
			spd = 0
		}
		b.Spd = spd
	}

	if s, _, ok := get_rec_value(r, "Hdg"); ok {
		v, _ := strconv.ParseFloat(s, 64)
		if v < 0 {
			v += 360.0
		}
		b.Cse = uint32(v)
		b.Cog = b.Cse
	}

	if s, _, ok := get_rec_value(r, "AccX"); ok {
		ax, _ := strconv.ParseFloat(s, 64)
		if s, _, ok := get_rec_value(r, "AccY"); ok {
			ay, _ := strconv.ParseFloat(s, 64)
			if s, _, ok = get_rec_value(r, "AccZ"); ok {
				az, _ := strconv.ParseFloat(s, 64)
				b.Pitch, b.Roll = acc_to_ah(ax, ay, az)
			}
		}
	}

	if s, _, ok := get_rec_value(r, "Thr"); ok {
		v, _ := strconv.ParseInt(s, 10, 32)
		b.Throttle = int(v)
	}

	md := uint8(0)

	if s, _, ok := get_rec_value(r, "Tmp1"); ok {
		tmp1, _ := strconv.ParseInt(s, 10, 32)
		modeE := tmp1 % 10
		modeD := (tmp1 % 100) / 10
		modeC := (tmp1 % 1000) / 100
		modeB := (tmp1 % 10000) / 1000
		modeA := tmp1 / 10000

		if (modeE & 4) == 4 {
			status |= types.Is_ARMED
		}

		switch modeD {
		case 0:
			md = types.FM_ACRO
		case 1:
			md = types.FM_ANGLE
		case 2:
			md = types.FM_HORIZON
		case 4:
			md = types.FM_MANUAL
		}

		if (modeC & 2) == 2 {
			md = types.FM_AH
		}
		if (modeC & 4) == 4 {
			md = types.FM_PH
		}

		if modeB == 1 {
			md = types.FM_RTH
		} else if modeB == 2 {
			md = types.FM_WP
		} else if modeB == 8 {
			if md == types.FM_AH || md == types.FM_PH {
				md = types.FM_CRUISE3D
			} else {
				md = types.FM_CRUISE2D
			}
		}
		if modeA == 4 {
			status |= types.Is_FAIL
		}
	}

	if s, _, ok := get_rec_value(r, "ARM"); ok {
		as, _ := strconv.ParseInt(s, 10, 32)
		if as == 100 {
			status |= (types.Is_ARMED | types.Is_ARDU)
			md = types.FM_ACRO
			b.Fix = 2
			b.Numsat = 13

		} else {
			status = types.Is_ARDU
			b.Fix = 0
		}
	}

	if s, _, ok := get_rec_value(r, "RSSI"); ok {
		rssi, _ := strconv.ParseInt(s, 10, 32)
		b.Rssi = uint8(rssi)
	}

	if s, _, ok := get_rec_value(r, "VFAS"); ok {
		b.Volts, _ = strconv.ParseFloat(s, 64)
	}

	if s, _, ok := get_rec_value(r, "1RSS"); ok {
		status |= types.Is_CRSF
		rssi, _ := strconv.ParseInt(s, 10, 32)
		b.Rssi = uint8(rssi)

		if s, _, ok = get_rec_value(r, "RxBt"); ok {
			b.Volts, _ = strconv.ParseFloat(s, 64)
		}

		if s, _, ok = get_rec_value(r, "FM"); ok {
			md = 0
			status |= types.Is_ARMED
			switch s {
			case "0", "OK", "WAIT", "!ERR":
				status &= ^types.Is_ARMED
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
			case "CRS", "CRSH":
				md = types.FM_CRUISE2D
			case "3CRS", "CRUZ":
				md = types.FM_CRUISE3D
			case "WP":
				md = types.FM_WP
			case "RTH":
				md = types.FM_RTH
			case "!FS!":
				status |= types.Is_FAIL
			}

			if s == "0" {
				if s, _, ok := get_rec_value(r, "Thr"); ok {
					thr, _ := strconv.ParseInt(s, 10, 32)
					if thr > -1024 {
						status |= types.Is_ARMED
					}
				}
			}
		}

		if s, _, ok := get_rec_value(r, "Sats"); ok {
			ns, _ := strconv.ParseInt(s, 10, 16)
			b.Numsat = uint8(ns)
			if ns > 5 {
				b.Fix = 2
				b.Hdop = uint16((3.3 - float64(ns)/12.0) * 100)
				if b.Hdop < 50 {
					b.Hdop = 50
				}
			} else if ns > 0 {
				b.Fix = 1
				b.Hdop = 800
			} else {
				b.Fix = 0
				b.Hdop = 999
			}
		}

		if s, u, ok := get_rec_value(r, "Yaw"); ok {
			v1, _ := strconv.ParseFloat(s, 64)
			cse := 0.0
			if u == "rad" {
				cse = to_degrees(v1)
			} else {
				cse = v1
			}
			if cse < 0 {
				cse += 360.0
			}
			b.Cse = uint32(cse)
		}

		if s, u, ok := get_rec_value(r, "Ptch"); ok {
			v1, _ := strconv.ParseFloat(s, 64)
			if u == "rad" {
				b.Pitch = int16(to_degrees(v1))
			} else {
				b.Pitch = int16(v1)
			}
		}
		if s, u, ok := get_rec_value(r, "Roll"); ok {
			v1, _ := strconv.ParseFloat(s, 64)
			if u == "rad" {
				b.Roll = int16(to_degrees(v1))
			} else {
				b.Roll = int16(v1)
			}
		}
	}
	b.Fmode = md
	b.Fmtext = types.Mnames[md]

	if s, u, ok := get_rec_value(r, "Curr"); ok {
		b.Amps, _ = strconv.ParseFloat(s, 64)
		if u == "mA" {
			b.Amps /= 1000
		}
		if s, u, ok = get_rec_value(r, "Fuel"); ok {
			b.Energy, _ = strconv.ParseFloat(s, 64)
		} else if s, u, ok = get_rec_value(r, "Capa"); ok {
			b.Energy, _ = strconv.ParseFloat(s, 64)
		}
		if b.Energy > 0 {
			switch u {
			case "mwh", "mWh":
				if b.Volts > 0 {
					b.Energy /= b.Volts
				} else {
					b.Energy = 0
				}
			case "pct", "%", "PCT":
				b.Energy = 0
			}
		}
	}
	b.Throttle = 100 * (b.Throttle + 1024) / 2048
	b.Status = status

	if s, _, ok := get_rec_value(r, "Ail"); ok {
		i64, _ := strconv.Atoi(s)
		b.Ail = normalise_stick(i64)
		if s, _, ok = get_rec_value(r, "Ele"); ok {
			i64, _ := strconv.Atoi(s)
			b.Ele = normalise_stick(i64)
		}
		if s, _, ok = get_rec_value(r, "Rud"); ok {
			i64, _ := strconv.Atoi(s)
			b.Rud = normalise_stick(i64)
		}
		if s, _, ok = get_rec_value(r, "Thr"); ok {
			i64, _ := strconv.Atoi(s)
			b.Thr = normalise_stick(i64)
		}
	}
	return b
}

func normalise_stick(v int) int16 {
	return int16((v+1024)*1000/2048 + 1000)
}

func to_degrees(rad float64) float64 {
	return (rad * 180.0 / math.Pi)
}

func to_radians(deg float64) float64 {
	return (deg * math.Pi / 180.0)
}

func calc_speed(b types.LogItem, tdiff time.Duration, llat, llon float64) float64 {
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

func acc_to_ah(ax, ay, az float64) (pitch int16, roll int16) {
	pitch = -int16((180.0 * math.Atan2(ax, math.Sqrt(ay*ay+az*az)) / math.Pi))
	roll = int16((180.0 * math.Atan2(ay, math.Sqrt(ax*ax+az*az)) / math.Pi))
	return pitch, roll
}

func (lg *OTXLOG) Reader(m types.FlightMeta, ch chan interface{}) (types.LogSegment, bool) {
	var stats types.LogStats

	llat := 0.0
	llon := 0.0

	var homes types.HomeRec
	rec := types.LogRec{}
	var froboff time.Duration

	fb := geo.Getfrobnication()

	fh, err := os.Open(lg.name)
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

	leffic := 0.0
	lwhkm := 0.0
	whacc := 0.0
	accEnergy := 0.0
	for i := 1; ; i++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if i == 1 {
			read_headers(record)
			rec.Cap = dataCapability()
			continue
		}
		if i >= m.Start && i <= m.End {
			b := get_otx_line(record)
			if (b.Status&types.Is_ARMED) == 0 && b.Alt < 10 && b.Spd < 7 {
				continue
			}

			tdiff := b.Utc.Sub(lt)
			if tdiff.Nanoseconds()/(1000*1000) >= int64(options.Config.Intvl) {
				if st.IsZero() {
					st = b.Utc
					lt = st
				} else {
					adiff := b.Utc.Sub(st)
					b.Stamp = uint64(adiff.Microseconds())
				}

				if homes.Flags == 0 {
					if b.Fix > 1 && b.Numsat > 5 {
						homes.HomeLat = b.Lat
						homes.HomeLon = b.Lon
						homes.Flags = types.HOME_ARM
						if options.Config.HomeAlt != -999999 {
							homes.HomeAlt = float64(options.Config.HomeAlt)
							homes.Flags |= types.HOME_ALT
						} else if b.GAlt > -999999 {
							homes.HomeAlt = b.GAlt
							homes.Flags |= types.HOME_ALT
						} else {
							if options.Config.UseTopo {
								d := geo.InitDem("")
								elev, err := d.Get_Elevation(homes.HomeLat, homes.HomeLon)
								if err == nil {
									homes.HomeAlt = elev
									homes.Flags |= types.HOME_ALT
								}
							}
						}

						if fb != nil {
							fb.Set_origin(homes.HomeLat, homes.HomeLon, b.GAlt)
							homes.HomeLat, homes.HomeLon, homes.HomeAlt = fb.Relocate(homes.HomeLat, homes.HomeLon, homes.HomeAlt)
							ttmp := time.Now().Add(time.Hour * 24 * 42)
							froboff = ttmp.Sub(b.Utc)
							b.Utc = ttmp
						}
						llat = b.Lat
						llon = b.Lon
						if ch != nil {
							ch <- homes
						}
					}
				} else {
					if fb != nil {
						b.Utc = b.Utc.Add(froboff)
						b.Lat, b.Lon, _ = fb.Relocate(b.Lat, b.Lon, 0)
					}
				}

				b.Hlat = homes.HomeLat
				b.Hlon = homes.HomeLon

				if (rec.Cap & types.CAP_SPEED) == 0 {
					if (b.Status & (types.Is_CRSF | types.Is_ARDU)) != 0 {
						b.Spd = calc_speed(b, tdiff, llat, llon)
						rec.Cap |= types.CAP_SPEED
					}
				}

				if b.Spd > 200 {
					continue // sanity check, 200m/s == 720kph, 388 knots
				}

				var c, d float64
				if homes.Flags != 0 {
					c, d = geo.Csedist(homes.HomeLat, homes.HomeLon, b.Lat, b.Lon)

					b.Bearing = int32(c)
					b.Vrange = d * 1852.0

					if d > stats.Max_range {
						stats.Max_range = d
						stats.Max_range_time = uint64(b.Utc.Sub(st).Nanoseconds() / 1000)
					}

					if b.Alt > stats.Max_alt {
						stats.Max_alt = b.Alt
						stats.Max_alt_time = uint64(b.Utc.Sub(st).Nanoseconds() / 1000)
					}

					if b.Spd < 400 && b.Spd > stats.Max_speed {
						stats.Max_speed = b.Spd
						stats.Max_speed_time = uint64(b.Utc.Sub(st).Nanoseconds() / 1000)
					}

					if b.Amps > stats.Max_current {
						stats.Max_current = b.Amps
						stats.Max_current_time = uint64(b.Utc.Sub(st).Nanoseconds() / 1000)
					}

					if llat != b.Lat || llon != b.Lon {
						_, d = geo.Csedist(llat, llon, b.Lat, b.Lon)
						stats.Distance += d
					}
				}

				b.Tdist = stats.Distance * 1852.0
				if (rec.Cap & types.CAP_AMPS) == types.CAP_AMPS {
					if d > 0 {
						deltat := tdiff.Seconds()
						aspd := d * 1852 / deltat              // m/s
						b.Effic = b.Amps * 1000 / (3.6 * aspd) // efficiency
						leffic = b.Effic
						b.Whkm = b.Amps * b.Volts / (3.6 * aspd)
						whacc += b.Amps * b.Volts * deltat / 3600
						b.WhAcc = whacc
						lwhkm = b.Whkm
						if rec.Cap&types.CAP_ENERGYC == types.CAP_ENERGYC {
							accEnergy += (b.Amps * deltat / 3.6)
							b.Energy = accEnergy
						}
					} else {
						b.Effic = leffic
						b.Whkm = lwhkm
					}
				}

				if b.Rssi > 0 {
					rec.Cap |= types.CAP_RSSI_VALID
				}

				if ch != nil {
					ch <- b
				} else {
					rec.Items = append(rec.Items, b)
				}
				llat = b.Lat
				llon = b.Lon
				lt = b.Utc
			}
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "reader %s\n", err)
			os.Exit(-1)
		}
	}
	srec := stats.Summary(uint64(lt.Sub(st).Nanoseconds() / 1000))
	ls := types.LogSegment{}
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
