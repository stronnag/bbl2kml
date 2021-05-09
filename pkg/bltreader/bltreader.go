package bltlog

import (
	"fmt"
	//"io"
	//"math"
	"os"
	"path/filepath"
	//"regexp"
	//"sort"
	"strconv"
	"strings"
	"errors"
	"time"
	"bufio"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	options "github.com/stronnag/bbl2kml/pkg/options"
	mission "github.com/stronnag/bbl2kml/pkg/mission"
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
	return 'G'
}
func (o *BLTLOG) GetMetas() ([]types.FlightMeta, error) {
	m, err := metas(o.name)
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

func parse_bullet(line string, b *types.LogItem) {
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
				case "alt":
					b.Alt = float64(tmp) / 100.0
				case "asl":
					b.GAlt = float64(tmp)
				case "gsp":
					b.Spd = float64(tmp) / 100.0
				case "bpv":
					b.Volts = float64(tmp) / 100.0
				case "cad":
					b.Energy = float64(tmp)
				case "rsi":
					b.Rssi = uint8(tmp)
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
					b.Fix = 3 * uint8(tmp)
				case "arm":
					if tmp == 1 {
						b.Status |= 1
					} else {
						b.Status &= ^uint8(1)
					}
				case "trp":
					b.Throttle = tmp
				case "nvs":
					b.NavMode = byte(tmp)
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
				case "whd":
					b.Tdist = float64(tmp)
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
					break
				}
			}
		}
	}
	if b.Lat == 0.0 && b.Lon == 0 {
		b.Fix = 0
	}
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

func (lg *BLTLOG) Reader(m types.FlightMeta) (types.LogSegment, bool) {
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

	for scanner.Scan() {
		line := scanner.Text()
		if i >= m.Start && i <= m.End {
			parse_bullet(line, &b)
			if b.Utc != lt && b.Fix != 0 {
				tdiff := b.Utc.Sub(lt)
				if tdiff.Milliseconds() >= int64(options.Config.Intvl) {
					if st.IsZero() {
						st = b.Utc
						lt = st
					}

					if b.Vrange > stats.Max_range {
						stats.Max_range = b.Vrange
						stats.Max_range_time = uint64(b.Utc.Sub(st).Microseconds())
					}

					if b.Alt > stats.Max_alt {
						stats.Max_alt = b.Alt
						stats.Max_alt_time = uint64(b.Utc.Sub(st).Microseconds())
					}

					if b.Spd < 400 && b.Spd > stats.Max_speed {
						stats.Max_speed = b.Spd
						stats.Max_speed_time = uint64(b.Utc.Sub(st).Microseconds())
					}

					if b.Amps > stats.Max_current {
						stats.Max_current = b.Amps
						stats.Max_current_time = uint64(b.Utc.Sub(st).Microseconds())
					}

					lt = b.Utc
					rec.Items = append(rec.Items, b)
					stats.Distance = b.Tdist / 1852.0
				}
			}
		}
		i += 1
	}

	srec := stats.Summary(uint64(lt.Sub(st).Microseconds()))
	ok := homes.Flags != 0 && len(rec.Items) > 0
	if ok {
		ls.L = rec
		ls.H = homes
		ls.M = srec
		if mok {
			options.Config.Mission = filepath.Join(options.Config.Tmpdir, "tmpmission.xml")
			ms.To_MWXML(options.Config.Mission)
		}
	}
	return ls, ok
}
