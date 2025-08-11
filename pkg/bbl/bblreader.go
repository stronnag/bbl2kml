package bbl

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

import (
	"geo"
	"inav"
	"mission"
	"options"
	"types"
)

var inav_vers int
var hdrs map[string]int

type BBLOG struct {
	name string
	meta []types.FlightMeta
}

func NewBBLReader(fn string) BBLOG {
	var l BBLOG
	l.name = fn
	l.meta = nil
	return l
}

func (o *BBLOG) GetMetas() ([]types.FlightMeta, error) {
	m, err := types.ReadMetaCache(o.name)
	if err != nil || options.Config.Nocache {
		m, err = metas(o.name)
		types.WriteMetaCache(o.name, m)
	}
	o.meta = m
	return m, err
}

func (o *BBLOG) GetDurations() {
	get_durations(o.name, o.meta)
}

func (o *BBLOG) LogType() byte {
	return types.LOGBBL
}

func (o *BBLOG) Dump() {
	get_headers(o.name)
	dump_headers()
}

func get_headers(fn string) {
	cmd := exec.Command(options.Config.Blackbox_decode,
		"--datetime", "--merge-gps", "--stdout", "--index", "1", fn)
	types.SetSilentProcess(cmd)
	out, err := cmd.StdoutPipe()

	defer cmd.Wait()
	defer out.Close()
	r := csv.NewReader(out)
	r.TrimLeadingSpace = true
	err = cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start err=%v", err)
		os.Exit(1)
	}
	record, err := r.Read()
	build_headers(record)
}

func build_headers(record []string) {
	hdrs = make(map[string]int)
	for i, s := range record {
		hdrs[s] = i
	}
	if _, ok := hdrs["dateTime"]; !ok {
		fmt.Fprintln(os.Stderr, "No \"datetime\" header, probably blackbox_decode too old or broken")
	}
}

func dump_headers() {
	n := map[int][]string{}
	var a []int
	for k, v := range hdrs {
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

func metas(fn string) ([]types.FlightMeta, error) {
	var bes []types.FlightMeta
	get_headers(fn)
	r, err := os.Open(fn)
	if err == nil {
		var nbes int
		var loffset int64
		var has_fbat bool
		var has_vbat bool
		var has_intp bool

		base := filepath.Base(fn)
		scanner := bufio.NewScanner(r)

		zero_or_nl := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}
			for i, b := range data {
				if b == '\n' || b == 0 || b == 0xff {
					return i + 1, data[0:i], nil
				}
			}

			if atEOF {
				return len(data), data, nil
			}
			return
		}

		scanner.Split(zero_or_nl)
		for scanner.Scan() {
			l := scanner.Text()
			switch {
			case strings.Contains(string(l), "H Product:"):
				offset, _ := r.Seek(0, io.SeekCurrent)
				if loffset != 0 {
					bes[nbes].Size = offset - loffset
					if bes[nbes].Size > 4096 {
						bes[nbes].Flags |= types.Is_Valid
					}
					if !has_intp || (has_fbat && !has_vbat) {
						bes[nbes].Flags |= types.Is_Suspect
					}
				}
				loffset = offset
				be := types.FlightMeta{Disarm: 0, Size: 0,
					Fwdate: "<no date>",
					Flags:  types.Has_Disarm | types.Has_Size}
				bes = append(bes, be)
				nbes = len(bes) - 1
				bes[nbes].Logname = base
				bes[nbes].Index = nbes + 1
				has_fbat = false
				has_vbat = false
				has_intp = false
			case strings.HasPrefix(string(l), "H Firmware revision:"):
				if n := strings.Index(string(l), ":"); n != -1 {
					fw := string(l)[n+1:]
					bes[nbes].Firmware = fw
					bes[nbes].Flags |= types.Has_Firmware
				}

			case strings.HasPrefix(string(l), "H Firmware date:"):
				if n := strings.Index(string(l), ":"); n != -1 {
					fw := string(l)[n+1:]
					bes[nbes].Fwdate = fw
				}

			case strings.HasPrefix(string(l), "H Log start datetime:"):
				if n := strings.Index(string(l), ":"); n != -1 {
					date := string(l)[n+1:]
					if len(date) > 0 {
						bes[nbes].Date, _ = time.Parse(time.RFC3339, date)
					}
				}

			case strings.HasPrefix(string(l), "H Craft name:"):
				if n := strings.Index(string(l), ":"); n != -1 {
					cname := string(l)[n+1:]
					if len(cname) > 0 {
						bes[nbes].Craft = cname
					}
					bes[nbes].Flags |= types.Has_Craft
				}

			case strings.HasPrefix(string(l), "H Field I name:"):
				// check for motors and servos
				if strings.Contains(l, "motor[7]") {
					bes[nbes].Motors = 8
				} else if strings.Contains(l, "motor[5]") {
					bes[nbes].Motors = 6
				} else if strings.Contains(l, "motor[3]") {
					bes[nbes].Motors = 4
				} else if strings.Contains(l, "motor[2]") {
					bes[nbes].Motors = 3
				} else if strings.Contains(l, "motor[1]") {
					bes[nbes].Motors = 2
				} else if strings.Contains(l, "motor[0]") {
					bes[nbes].Motors = 1
				}

				if strings.Contains(l, "servo[7]") {
					bes[nbes].Servos = 1
				}

			case strings.HasPrefix(string(l), "H acc_1G:"):
				if n := strings.Index(string(l), ":"); n != -1 {
					fstr := string(l)[n+1:]
					if len(fstr) > 0 {
						fs, _ := strconv.Atoi(fstr)
						if fs != 0 {
							bes[nbes].Acc1G = uint16(fs)
						}
					}
				}

			case strings.HasPrefix(string(l), "H acc_hardware:"):
				if n := strings.Index(string(l), ":"); n != -1 {
					fstr := string(l)[n+1:]
					if len(fstr) > 0 {
						fs, _ := strconv.Atoi(fstr)
						if fs != 0 {
							bes[nbes].Sensors |= types.Has_Acc
						}
					}
				}
			case strings.HasPrefix(string(l), "H baro_hardware:"):
				if n := strings.Index(string(l), ":"); n != -1 {
					fstr := string(l)[n+1:]
					if len(fstr) > 0 {
						fs, _ := strconv.Atoi(fstr)
						if fs != 0 {
							bes[nbes].Sensors |= types.Has_Baro
						}
					}
				}
			case strings.HasPrefix(string(l), "H mag_hardware:"):
				if n := strings.Index(string(l), ":"); n != -1 {
					fstr := string(l)[n+1:]
					if len(fstr) > 0 {
						fs, _ := strconv.Atoi(fstr)
						if fs != 0 {
							bes[nbes].Sensors |= types.Has_Mag
						}
					}
				}

			case strings.HasPrefix(string(l), "H features:"):
				if n := strings.Index(string(l), ":"); n != -1 {
					fstr := string(l)[n+1:]
					if len(fstr) > 0 {
						features, _ := strconv.Atoi(fstr)
						bes[nbes].Features = uint32(features)
						if (features & types.Feature_GPS) != 0 {
							bes[nbes].Sensors |= types.Has_GPS
						}
						if (features & types.Feature_VBAT) != 0 {
							has_fbat = true
						}
					}
				}

			case strings.Contains(string(l), "H vbatref:"):
				has_vbat = true

			case strings.Contains(string(l), "H P interval:"):
				has_intp = true

			case strings.Contains(string(l), "reason:"):
				if n := strings.Index(string(l), ":"); n != -1 {
					dindx, _ := strconv.Atoi(string(l)[n+1 : n+2])
					bes[nbes].Disarm = types.Reason(dindx)
				}
			}
			if err = scanner.Err(); err != nil {
				return bes, err
			}
		}
		if len(bes) > 0 {
			if bes[nbes].Size == 0 {
				offset, _ := r.Seek(0, io.SeekCurrent)
				if loffset != 0 {
					bes[nbes].Size = offset - loffset
					if bes[nbes].Size > 4096 {
						bes[nbes].Flags |= types.Is_Valid
					}
				}
				if !has_intp || (has_fbat && !has_vbat) {
					bes[nbes].Flags |= types.Is_Suspect
				}
			}
			r.Close()
			/*
				for i := 0; i < len(bes); i++ {
					if bes[i].Flags&types.Is_Suspect != 0 {
						fmt.Fprintf(os.Stderr, " * Log entry %d may be corrupt\n", i+1)
					}
				}
			*/
		}
	} else {
		err = errors.New("No records in BBL")
	}
	return bes, err
}

func get_durations(fn string, meta []types.FlightMeta) {
	for i := 0; i < len(meta); i++ {
		meta[i].Duration = get_bb_duration(fn, fmt.Sprintf("%d", i+1))
	}
}

func get_bb_duration(bbfile string, idx string) time.Duration {
	cmd := exec.Command(options.Config.Blackbox_decode, "--stdout", "--index", idx, bbfile)
	out, err := cmd.StdoutPipe()
	defer cmd.Wait()
	defer out.Close()
	err = cmd.Start()
	scanner := bufio.NewScanner(out)
	i := 0
	var ssec string
	var lsec string
	var line string
	for scanner.Scan() {
		line = scanner.Text()
		if i == 1 {
			parts := strings.SplitN(line, ",", 3)
			if len(parts) > 2 {
				lsec = strings.TrimLeft(parts[1], " ")
				ssec = lsec
			}
		}
		i += 1
	}
	parts := strings.SplitN(line, ",", 3)
	if len(parts) > 2 {
		lsec = strings.TrimLeft(parts[1], " ")
	}

	ilsec, _ := strconv.ParseInt(lsec, 10, 64)
	issec, _ := strconv.ParseInt(ssec, 10, 64)
	sdiff := ilsec - issec
	diff := time.Duration(sdiff) * time.Microsecond
	if err = scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	return diff
}

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

func dataCapability() uint16 {
	var ret uint16 = 0
	if _, ok := hdrs["amperage (A)"]; ok {
		ret |= types.CAP_AMPS
	} else if _, ok := hdrs["amperageLatest (A)"]; ok {
		ret |= types.CAP_AMPS
	}
	if _, ok := hdrs["vbat (V)"]; ok {
		ret |= types.CAP_VOLTS
	} else if _, ok = hdrs["vbatLatest (V)"]; ok {
		ret |= types.CAP_VOLTS
	}

	if _, ok := hdrs["energyCumulative (mAh)"]; ok {
		ret |= types.CAP_ENERGY
	}

	if _, ok := hdrs["GPS_speed (m/s)"]; ok {
		ret |= types.CAP_SPEED
	}

	if _, ok := hdrs["navPos[2]"]; ok {
		ret |= types.CAP_ALTITUDE
	}

	if _, ok := hdrs["activeWpNumber"]; ok {
		ret |= types.CAP_WPNO
	}
	if _, ok := hdrs["wind[0]"]; ok {
		ret |= types.CAP_WIND
	}
	return ret
}

func get_bbl_line(r []string, have_origin bool) types.LogItem {
	status := types.Is_ARMED
	b := types.LogItem{}

	s, ok := get_rec_value(r, "GPS_numSat")
	if ok {
		i64, _ := strconv.Atoi(s)
		b.Numsat = uint8(i64)
	}

	if s, ok = get_rec_value(r, "GPS_hdop"); ok {
		i64, _ := strconv.Atoi(s)
		b.Hdop = uint16(i64)
	}

	if s, ok = get_rec_value(r, "vbat (V)"); ok {
		b.Volts, _ = strconv.ParseFloat(s, 64)
	} else if s, ok = get_rec_value(r, "vbatLatest (V)"); ok {
		b.Volts, _ = strconv.ParseFloat(s, 64)
	}

	if s, ok = get_rec_value(r, "navPos[2]"); ok {
		b.Alt, _ = strconv.ParseFloat(s, 64)
		b.Alt = b.Alt / 100.0
	} else if s, ok = get_rec_value(r, "BaroAlt (cm)"); ok {
		b.Alt, _ = strconv.ParseFloat(s, 64)
		b.Alt = b.Alt / 100.0
	}

	if s, ok = get_rec_value(r, "GPS_fixType"); ok {
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

	if s, ok = get_rec_value(r, "GPS_coord[0]"); ok {
		b.Lat, _ = strconv.ParseFloat(s, 64)
	}

	if s, ok = get_rec_value(r, "GPS_coord[1]"); ok {
		b.Lon, _ = strconv.ParseFloat(s, 64)
	}

	if s, ok = get_rec_value(r, "GPS_altitude"); ok {
		b.GAlt, _ = strconv.ParseFloat(s, 64)
	}

	if s, ok = get_rec_value(r, "GPS_speed (m/s)"); ok {
		b.Spd, _ = strconv.ParseFloat(s, 64)
	}

	if s, ok = get_rec_value(r, "time (us)"); ok {
		i64, _ := strconv.ParseInt(s, 10, 64)
		b.Stamp = uint64(i64)
	}

	if s, ok = get_rec_value(r, "activeWpNumber"); ok {
		i64, _ := strconv.ParseInt(s, 10, 64)
		b.ActiveWP = uint8(i64)
	}

	md := uint8(0)
	s0, sok := get_rec_value(r, "flightModeFlags (flags)")
	if s, ok = get_rec_value(r, "navState"); ok {
		i64, _ := strconv.ParseInt(s, 10, 64)
		if inav.IsCruise3d(inav_vers, int(i64)) {
			md = types.FM_CRUISE3D
		} else if inav.IsCruise2d(inav_vers, int(i64)) {
			md = types.FM_CRUISE2D
		} else if inav.IsRTH(inav_vers, int(i64)) {
			md = types.FM_RTH
		} else if inav.IsWP(inav_vers, int(i64)) {
			md = types.FM_WP
		} else if inav.IsLand(inav_vers, int(i64)) {
			md = types.FM_LAND
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
		b.Navmode = inav.Navmode(inav_vers, int(i64))
	}

	// Ancient INAV (pre 1.very-early)
	if s, ok = get_rec_value(r, "navMode"); ok {
		i64, _ := strconv.ParseInt(s, 10, 64)
		switch i64 {
		case 3:
			md = types.FM_PH
			b.Navmode = 1
		case 5:
			md = types.FM_RTH
			b.Navmode = 2
		}
	}

	// fallback for old inav bug
	if sok && strings.Contains(s0, "NAVRTH") {
		md = types.FM_RTH
	}

	b.Fmode = md
	b.Fmtext = types.Mnames[md]

	if s, ok = get_rec_value(r, "failsafePhase (flags)"); ok {
		if !strings.Contains(s, "IDLE") {
			status |= types.Is_FAIL
		}
	}

	b.Status = uint8(status)

	if !have_origin {
		b.Hlat = 0
		b.Hlon = 0
		b.Vrange = -1
		b.Bearing = -1
		if s, ok = get_rec_value(r, "GPS_home_lat"); ok {
			b.Hlat, _ = strconv.ParseFloat(s, 64)
		}
		if s, ok = get_rec_value(r, "GPS_home_lon"); ok {
			b.Hlon, _ = strconv.ParseFloat(s, 64)
			b.Bearing = -2
		} else {
			if s, ok = get_rec_value(r, "homeDirection"); ok {
				i64, _ := strconv.Atoi(s)
				b.Bearing = int32(i64)
			} else {
				if s, ok = get_rec_value(r, "Azimuth"); ok {
					i64, _ := strconv.Atoi(s)
					b.Bearing = int32((i64 + 180) % 360)
				}
			}

			if b.Bearing != -1 {
				if s, ok = get_rec_value(r, "Distance (m)"); ok {
					b.Vrange, _ = strconv.ParseFloat(s, 64)
				}
			}
		}
	} else {
		if s, ok = get_rec_value(r, "GPS_home_lat"); ok {
			b.Hlat, _ = strconv.ParseFloat(s, 64)
		}
		if s, ok = get_rec_value(r, "GPS_home_lon"); ok {
			b.Hlon, _ = strconv.ParseFloat(s, 64)
		}
	}

	if s, ok = get_rec_value(r, "rcData[0]"); ok {
		i64, _ := strconv.Atoi(s)
		b.Ail = int16(i64)
		if s, ok = get_rec_value(r, "rcData[1]"); ok {
			i64, _ := strconv.Atoi(s)
			b.Ele = int16(i64)
		}
		if s, ok = get_rec_value(r, "rcData[2]"); ok {
			i64, _ := strconv.Atoi(s)
			b.Rud = int16(i64)
		}
		if s, ok = get_rec_value(r, "rcData[3]"); ok {
			i64, _ := strconv.Atoi(s)
			b.Thr = int16(i64)
		}
	} else if s, ok = get_rec_value(r, "rcCommand[0]"); ok {
		i64, _ := strconv.Atoi(s)
		b.Ail = int16(i64) + 1500
		if s, ok = get_rec_value(r, "rcCommand[1]"); ok {
			i64, _ := strconv.Atoi(s)
			b.Ele = int16(i64) + 1500
		}
		if s, ok = get_rec_value(r, "rcCommand[2]"); ok {
			i64, _ := strconv.Atoi(s)
			b.Rud = -1*int16(i64) + 1500
		}
		if s, ok = get_rec_value(r, "rcCommand[3]"); ok {
			i64, _ := strconv.Atoi(s)
			b.Thr = int16(i64)
		}
	}

	if s, ok = get_rec_value(r, "attitude[0]"); ok {
		i64, _ := strconv.Atoi(s)
		b.Roll = int16(i64 / 10)
	}

	if s, ok = get_rec_value(r, "attitude[1]"); ok {
		i64, _ := strconv.Atoi(s)
		b.Pitch = int16(i64 / 10)
	}

	if s, ok = get_rec_value(r, "attitude[2]"); ok {
		i64, _ := strconv.Atoi(s)
		b.Cse = uint32(i64 / 10)
	} else if s, ok = get_rec_value(r, "navHeading"); ok {
		i64, _ := strconv.Atoi(s)
		b.Cse = uint32(i64 / 100)
	}

	if s, ok = get_rec_value(r, "GPS_ground_course"); ok {
		v, _ := strconv.ParseFloat(s, 64)
		b.Cog = uint32(v)
	}

	if s, ok = get_rec_value(r, "rssi"); ok {
		i64, _ := strconv.Atoi(s)
		b.Rssi = uint8(i64 * 100 / 1023)
	}

	if s, ok = get_rec_value(r, "dateTime"); ok {
		b.Utc, _ = time.Parse(time.RFC3339Nano, s)
	}

	if s, ok = get_rec_value(r, "amperage (A)"); ok {
		b.Amps, _ = strconv.ParseFloat(s, 64)
	} else if s, ok = get_rec_value(r, "amperageLatest (A)"); ok {
		b.Amps, _ = strconv.ParseFloat(s, 64)
	}

	if s, ok = get_rec_value(r, "energyCumulative (mAh)"); ok {
		b.Energy, _ = strconv.ParseFloat(s, 64)
		if b.Energy < 0 {
			b.Energy = 0
		}
	}

	if s, ok = get_rec_value(r, "rcData[3]"); ok {
		i64, _ := strconv.Atoi(s)
		b.Throttle = int(i64)
		b.Throttle = (b.Throttle - 1000) / 10
	}

	if s, ok = get_rec_value(r, "gyroADC[0]"); ok {
		i64, _ := strconv.Atoi(s)
		b.Gyro_x = int16(i64)
	}
	if s, ok = get_rec_value(r, "gyroADC[1]"); ok {
		i64, _ := strconv.Atoi(s)
		b.Gyro_y = int16(i64)
	}
	if s, ok = get_rec_value(r, "gyroADC[2]"); ok {
		i64, _ := strconv.Atoi(s)
		b.Gyro_z = int16(i64)
	}

	if s, ok = get_rec_value(r, "accSmooth[0]"); ok {
		i64, _ := strconv.Atoi(s)
		b.Acc_x = int16(i64)
	}
	if s, ok = get_rec_value(r, "accSmooth[1]"); ok {
		i64, _ := strconv.Atoi(s)
		b.Acc_y = int16(i64)
	}
	if s, ok = get_rec_value(r, "accSmooth[2]"); ok {
		i64, _ := strconv.Atoi(s)
		b.Acc_z = int16(i64)
	}

	if s, ok = get_rec_value(r, "hwHealthStatus"); ok {
		b.HWfail = false
		val, _ := strconv.Atoi(s)
		for n := 0; n < 7; n++ {
			sv := val & 3
			if sv > 1 || ((n < 2 || n == 4) && sv != 1) {
				b.HWfail = true
				break
			}
			val = (val >> 2)
		}
	}

	if s, ok = get_rec_value(r, "wind[0]"); ok {
		i64, _ := strconv.Atoi(s)
		b.Wind[0] = int16(i64)
		if s, ok = get_rec_value(r, "wind[1]"); ok {
			i64, _ = strconv.Atoi(s)
			b.Wind[1] = int16(i64)
			if s, ok = get_rec_value(r, "wind[2]"); ok {
				i64, _ = strconv.Atoi(s)
				b.Wind[2] = int16(i64)
			}
		}
	} /*
		else {
			b.Wind[0] = -32768
			b.Wind[1] = -32768
			b.Wind[2] = -32768
		}
	*/
	return b
}

func proc_start(w1 *os.File, args ...string) (p *os.Process, err error) {
	if args[0], err = exec.LookPath(args[0]); err == nil {
		var procAttr os.ProcAttr
		procAttr.Files = []*os.File{nil, w1, nil}
		p, err := os.StartProcess(args[0], args, &procAttr)
		if err == nil {
			return p, nil
		}
	}
	return nil, err
}

func read_mission(fb *geo.Frob) *mission.Mission {
	var ms *mission.Mission
	ms = nil
	if len(options.Config.Mission) > 0 {
		var err error
		_, ms, err = mission.Read_Mission_File_Index(options.Config.Mission, options.Config.MissionIndex)
		if err == nil {
			if fb != nil {
				if ms.Metadata.Homey != 0 && ms.Metadata.Homex != 0 {
					fb.Set_origin(ms.Metadata.Homey, ms.Metadata.Homex, 0)
					ms.Metadata.Homey, ms.Metadata.Homex, _ = fb.Get_rebase()
					ms.Metadata.Cy, ms.Metadata.Cx, _ = fb.Relocate(ms.Metadata.Cy, ms.Metadata.Cx, 0)
				}
			}

			for k, mi := range ms.MissionItems {
				if mi.Is_GeoPoint() && fb != nil {
					ms.MissionItems[k].Lat, ms.MissionItems[k].Lon, _ = fb.Relocate(ms.MissionItems[k].Lat, ms.MissionItems[k].Lon, 0)
				}
				if mi.Action == "JUMP" {
					ms.MissionItems[k].P3 = ms.MissionItems[k].P2
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "* Failed to read mission file %s\n", options.Config.Mission)
		}
	}
	return ms
}

func (lg *BBLOG) Reader(meta types.FlightMeta, ch chan interface{}) (types.LogSegment, bool) {
	cmd := exec.Command(options.Config.Blackbox_decode,
		"--datetime", "--merge-gps", "--stdout", "--index",
		strconv.Itoa(meta.Index), lg.name)
	types.SetSilentProcess(cmd)
	out, _ := cmd.StdoutPipe()
	serr, _ := cmd.StderrPipe()

	defer cmd.Wait()
	defer out.Close()

	var homes types.HomeRec
	var rec types.LogRec
	var froboff time.Duration

	fb := geo.Getfrobnication()

	ms := read_mission(fb)

	r := csv.NewReader(out)
	r.TrimLeadingSpace = true

	err := cmd.Start()
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

	ndelay := 1000 * uint64(options.Config.Intvl)
	tgt := 0
	laststat := uint8(255)
	leffic := 0.0
	lwhkm := 0.0
	whacc := 0.0
	var skiptime uint64 = 0

	if options.Config.SkipTime > 0 {
		skiptime = uint64(options.Config.SkipTime) * 1000
	}

	for i := 0; ; i++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if i == 0 {
			build_headers(record)
			rec.Cap = dataCapability()
			if ch != nil {
				ch <- rec.Cap
			}
			continue
		}

		b := get_bbl_line(record, have_origin)

		if !have_origin {
			if b.Fix > 1 && b.Numsat > 5 {
				have_origin = true
				if fb != nil {
					fb.Set_origin(b.Lat, b.Lon, b.GAlt)
					b.Lat, b.Lon, b.GAlt = fb.Relocate(b.Lat, b.Lon, b.GAlt)
					ttmp := time.Now().Add(time.Hour * 24 * 42)
					froboff = ttmp.Sub(b.Utc)
					b.Utc = ttmp
				}
				llat = b.Lat
				llon = b.Lon
				st = b.Stamp
				homes.HomeLat = b.Lat
				homes.HomeLon = b.Lon
				homes.HomeAlt = b.GAlt
				homes.Flags = types.HOME_ARM | types.HOME_ALT
				if b.Bearing == -2 {
					if b.Hlat != 0.0 && b.Hlon != 0.0 {
						_, dh := geo.Csedist(b.Hlat, b.Hlon, b.Lat, b.Lon)
						if dh > 2.0/1852.0 {
							homes.SafeLat = b.Hlat
							homes.SafeLon = b.Hlon
							homes.Flags |= types.HOME_SAFE
						}
					}
				} else if b.Bearing > -1 {
					hlat, hlon := geo.Posit(b.Lat, b.Lon, float64(b.Bearing), b.Vrange/1852.0)
					homes.SafeLat = hlat
					homes.SafeLon = hlon
					homes.Flags |= types.HOME_SAFE
				}
				if fb != nil && (homes.Flags&types.HOME_SAFE != 0) {
					homes.SafeLat, homes.SafeLon, _ = fb.Relocate(homes.SafeLat, homes.SafeLon, b.GAlt)
				}
				if ch != nil {
					ch <- homes
				}
			}
			if b.Utc.IsZero() {
				basetime, _ = time.Parse("Jan 2 2006 15:04:05", meta.Fwdate)
			}
		} else {
			us := b.Stamp
			var deltat = us - st
			if skiptime > 0 && deltat < skiptime {
				continue
			}

			if us > st {
				var d float64
				var c float64
				// Do the plot every 100ms
				if (us - dt) >= ndelay {
					if !basetime.IsZero() {
						b.Utc = basetime.Add(time.Duration(us) * time.Microsecond)
					}
					if fb != nil {
						b.Utc = b.Utc.Add(froboff)
						b.Lat, b.Lon, _ = fb.Relocate(b.Lat, b.Lon, 0)
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

					if b.Fmode != laststat {
						if b.Fmode == types.FM_WP && ms != nil {
							tgt = 1
						} else {
							tgt = 0
						}
						laststat = b.Fmode
					}

					if b.Fmode == types.FM_WP && b.ActiveWP == 0 && ms != nil {
						tgt, _ = inav.WP_state(ms, b, tgt)
						b.ActiveWP = uint8(tgt)
					}

					if (rec.Cap & types.CAP_AMPS) == types.CAP_AMPS {
						if d > 0 {
							deltat := float64((us - dt)) / 1000000.0 // seconds
							aspd := d * 1852 / deltat                // m/s
							b.Effic = b.Amps * 1000 / (3.6 * aspd)   // efficiency mAh/km
							leffic = b.Effic
							b.Whkm = b.Amps * b.Volts / (3.6 * aspd)
							whacc += b.Amps * b.Volts * deltat / 3600
							b.WhAcc = whacc
							lwhkm = b.Whkm
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
			log.Printf("bblreader: %+v\n", err)
			break
		}
	}

	sbytes, err := io.ReadAll(serr)
	serr.Close()

	logerrs := parse_errors(string(sbytes))

	srec := stats.Summary(lt - st)
	ls := types.LogSegment{}
	ls.S = logerrs

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

func parse_errors(s string) string {
	var sb strings.Builder
	parts := strings.Split(s, "\n")
	for _, p := range parts {
		if strings.HasPrefix(p, "\tWarning: ") || strings.HasPrefix(p, "\tError: ") {
			fmt.Fprintf(&sb, "%s\n", p[1:])
		}
	}
	return sb.String()
}
