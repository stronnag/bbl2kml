package bltlog

import (
	"fmt"
	//"io"
	//"math"
	"os"
	"path/filepath"
	//"regexp"
	//"sort"
	"bufio"
	"errors"
	"strconv"
	"strings"
	"time"
)

import (
	"mission"
	"options"
	"types"
)

var (
	homes types.HomeRec
	ms    mission.Mission
	mok   bool
)

var fltmodes = [...]uint8{0, types.FM_MANUAL, types.FM_RTH, types.FM_PH, types.FM_PH, types.FM_CRUISE3D, types.FM_CRUISE3D, types.FM_WP, types.FM_AH, types.FM_ANGLE, types.FM_HORIZON, types.FM_ACRO}

type BLTLOG struct {
	name string
	meta []types.FlightMeta
}

func NewBLTReader(fn string) BLTLOG {
	var l BLTLOG
	l.name = fn
	l.meta = nil
	return l
}

func (o *BLTLOG) LogType() byte {
	return types.LOGBLT
}
func (o *BLTLOG) GetMetas() ([]types.FlightMeta, error) {
	m, err := types.ReadMetaCache(o.name)
	if err != nil {
		m, err = metas(o.name)
		types.WriteMetaCache(o.name, m)
	}
	o.meta = m
	return m, err
}

func (o *BLTLOG) GetDurations() {
}

func (o *BLTLOG) Dump() {
}

func metas(logfile string) ([]types.FlightMeta, error) {
	var metas []types.FlightMeta

	fh, err := os.Open(logfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "log file %s\n", err)
		return metas, err
	}
	defer fh.Close()

	basefile := filepath.Base(logfile)

	scanner := bufio.NewScanner(fh)
	idx := 0
	i := 1
	lasttm := int64(0)
	var parts []string
	for scanner.Scan() {
		line := scanner.Text()
		if parts = strings.Split(line, "|"); len(parts) == 2 {
			lasttm, _ = strconv.ParseInt(parts[0], 10, 64)
		} else {
			continue
		}
		if strings.HasPrefix(parts[1], "Connected") {
			lt := time.Unix(lasttm/1000, 1000*1000*(lasttm%1000))
			if idx > 0 {
				metas[idx-1].End = i - 1
				metas[idx-1].Duration = lt.Sub(metas[idx-1].Date)
			}
			idx += 1
			mt := types.FlightMeta{Logname: basefile, Date: lt, Index: idx, Start: i}
			metas = append(metas, mt)
		} else if idx > 0 && metas[idx-1].Flags == 0 {
			if cstart := strings.Index(parts[1], "cs:"); cstart > -1 {
				cstart += 3
				cend := strings.Index(parts[1][cstart:], ",")
				cend += cstart
				metas[idx-1].Craft = parts[1][cstart:cend]
				metas[idx-1].Flags = types.Has_Craft
			}
		}
		i += 1
	}
	lt := time.Unix(lasttm/1000, 1000*1000*(lasttm%1000))
	if idx > 0 {
		metas[idx-1].End = i - 1
		metas[idx-1].Duration = lt.Sub(metas[idx-1].Date)
	}

	for j, mx := range metas {
		if mx.End-mx.Start > 64 {
			metas[j].Flags |= types.Has_Start | types.Is_Valid
		}
	}
	if len(metas) == 0 {
		err = errors.New("No records in Bullet log")
	}
	return metas, err
}

func parse_bullet(line string, b *types.LogItem) uint16 {
	cap := uint16(0)
	if parts := strings.Split(line, "|"); len(parts) == 2 {
		lasttm, _ := strconv.ParseInt(parts[0], 10, 64)
		b.Utc = time.Unix(lasttm/1000, 1000*1000*(lasttm%1000))
		vals := strings.Split(parts[1], ",")
		for _, kvs := range vals {
			kv := strings.Split(kvs, ":")
			if len(kv) == 2 {
				tmp, _ := strconv.Atoi(kv[1])
				switch kv[0] {
				case "ran":
					b.Roll = int16(tmp / 10)
				case "pan":
					b.Pitch = int16(tmp / 10)
				case "hea":
					b.Cse = uint32(tmp)
					b.Cog = b.Cse
				case "alt":
					b.Alt = float64(tmp) / 100.0
					cap |= types.CAP_ALTITUDE
				case "asl":
					b.GAlt = float64(tmp)
				case "gsp":
					b.Spd = float64(tmp) / 100.0
					cap |= types.CAP_SPEED
				case "bpv":
					b.Volts = float64(tmp) / 100.0
					cap |= types.CAP_VOLTS
				case "cad":
					b.Energy = float64(tmp)
					cap |= types.CAP_ENERGY
				case "rsi":
					b.Rssi = uint8(tmp)
					cap |= types.CAP_RSSI_VALID
				case "ghp":
					b.Hdop = uint16(tmp)
				case "fs":
					if tmp == 1 {
						b.Status |= 2
					} else {
						b.Status &= ^uint8(2)
					}
				case "ftm":
					b.Fmode = parse_flight_mode(uint8(tmp))
					b.Fmtext = types.Mnames[b.Fmode]
				case "hdr":
					b.Bearing = int32(tmp)
				case "hds":
					b.Vrange = float64(tmp) / 1852.0
				case "gla":
					b.Lat = float64(tmp) / 1e7
				case "glo":
					b.Lon = float64(tmp) / 1e7
				case "gsc":
					b.Numsat = uint8(tmp)
				case "3df":
					b.Fix = 2 * uint8(tmp)
				case "arm":
					if tmp == 1 {
						b.Status |= 1
					} else {
						b.Status &= ^uint8(1)
					}
				case "trp":
					b.Throttle = tmp
				case "nvs":
					b.Navmode = byte(tmp)
				case "hla":
					homes.HomeLat = float64(tmp) / 1e7
					homes.Flags |= types.HOME_ARM
				case "hlo":
					homes.HomeLon = float64(tmp) / 1e7
					homes.Flags |= types.HOME_ARM
				case "hal":
					homes.HomeAlt = float64(tmp) / 100.0
					homes.Flags |= types.HOME_ALT
				case "cud":
					b.Amps = float64(tmp) / 100.0
					cap |= types.CAP_AMPS
				case "wpno":
					if mok == false {
						parse_mission(vals)
					}

					// not used (here, for now)
				case "cs": // in metas
				case "mfr":
				case "ont":
				case "flt":
				case "bcc":
				case "ggc":
				case "att":
				case "wpv":
				case "wpc":
				case "id":
				case "vsp":
				case "acv":
				case "bfp":
				case "css":
				case "hwh":
				case "cwn":
				case "whd":
					break
				}
			}
		}
	}
	if b.Lat == 0.0 && b.Lon == 0 {
		b.Fix = 0
	}
	return cap
}

func parse_mission(vals []string) {
	mi := mission.MissionItem{}
	for _, kvs := range vals {
		kv := strings.Split(kvs, ":")
		if len(kv) == 2 {
			tmp, _ := strconv.Atoi(kv[1])
			switch kv[0] {
			case "wpno":
				if tmp > 0 {
					mi.No = tmp
				}
			case "la":
				mi.Lat = float64(tmp) / 1e7
			case "lo":
				mi.Lon = float64(tmp) / 1e7
			case "al":
				mi.Alt = int32(tmp) / 100
			case "ac":
				mi.Action = ms.Decode_action(byte(tmp))
			case "p1":
				mi.P1 = int16(tmp)
			case "p2":
				mi.P2 = int16(tmp)
			case "p3":
				mi.P3 = int16(tmp)
			case "f":
				if mi.No != 0 {
					mok = true
				}
			}
		}
	}
	if mi.No != 0 {
		ms.MissionItems = append(ms.MissionItems, mi)
	}
}

func parse_flight_mode(fmode uint8) uint8 {

	if fmode < uint8(len(fltmodes)) {
		return fltmodes[fmode]
	} else {
		return 0
	}
}

func (lg *BLTLOG) Reader(m types.FlightMeta, ch chan interface{}) (types.LogSegment, bool) {
	var stats types.LogStats
	ls := types.LogSegment{}
	var lt, st time.Time

	fh, err := os.Open(lg.name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "log file %s\n", err)
		return ls, false
	}
	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	i := 1
	rec := types.LogRec{}
	b := types.LogItem{}
	hseen := false
	leffic := 0.0
	lwhkm := 0.0
	whacc := 0.0
	for scanner.Scan() {
		line := scanner.Text()
		if i >= m.Start && i <= m.End {
			cap := parse_bullet(line, &b)
			rec.Cap |= cap
			if ch != nil {
				if !hseen && (homes.Flags&types.HOME_ARM) != 0 {
					hseen = true
					ch <- homes
				}
			}
			if b.Utc != lt && b.Fix != 0 {
				tdiff := b.Utc.Sub(lt)
				if tdiff.Nanoseconds()/(1000*1000) >= int64(options.Config.Intvl) {
					if st.IsZero() {
						st = b.Utc
						lt = st
					} else {
						mdiff := b.Utc.Sub(st).Microseconds()
						if mdiff > 0 {
							b.Stamp = uint64(mdiff)
						}
					}

					if b.Vrange > stats.Max_range {
						stats.Max_range = b.Vrange
						stats.Max_range_time = uint64(b.Utc.Sub(st).Nanoseconds() / 1000)
					}

					if b.Alt > stats.Max_alt {
						stats.Max_alt = b.Alt
						stats.Max_alt_time = uint64(b.Utc.Sub(st).Nanoseconds() / 1000)
					}

					deltat := b.Utc.Sub(lt).Seconds()
					if b.Spd > 0 && b.Spd < 400 {
						if b.Spd > stats.Max_speed {
							stats.Max_speed = b.Spd
							stats.Max_speed_time = uint64(b.Utc.Sub(st).Nanoseconds() / 1000)
						}

						if deltat > 0 {
							deltad := b.Spd / deltat
							b.Tdist += deltad
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

					if b.Amps > stats.Max_current {
						stats.Max_current = b.Amps
						stats.Max_current_time = uint64(b.Utc.Sub(st).Nanoseconds() / 1000)
					}

					lt = b.Utc
					if ch != nil {
						ch <- b
					} else {
						rec.Items = append(rec.Items, b)
					}
					stats.Distance = b.Tdist / 1852.0
				}
			}
		}
		i += 1
	}

	srec := stats.Summary(uint64(lt.Sub(st).Nanoseconds() / 1000))
	if mok {
		options.Config.Mission = filepath.Join(options.Config.Tmpdir, "tmpmission.xml")
		ms.To_MWXML(options.Config.Mission)
	}

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
