package aplog

import (
	"bufio"
	"fmt"
	"encoding/json"
	"os"
	"os/exec"
	"time"
	"math"
	"regexp"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	options "github.com/stronnag/bbl2kml/pkg/options"
	geo "github.com/stronnag/bbl2kml/pkg/geo"
)

type MavMeta struct {
	Type      string  `json:"type"`
	Timestamp float64 `json:"timestamp"`
}

type MavLog struct {
	Meta MavMeta         `json:"meta"`
	Data json.RawMessage `json:"data"`
}

type MavAtt struct {
	DesPitch int64   `json:"DesPitch"`
	DesRoll  int64   `json:"DesRoll"`
	DesYaw   float64 `json:"DesYaw"`
	ErrRP    float64 `json:"ErrRP"`
	ErrYaw   float64 `json:"ErrYaw"`
	Pitch    float64 `json:"Pitch"`
	Roll     float64 `json:"Roll"`
	TimeUS   int64   `json:"TimeUS"`
	Yaw      float64 `json:"Yaw"`
}

type MavGPS struct {
	Alt    float64 `json:"Alt"`
	GCrs   float64 `json:"GCrs"`
	Gms    int64   `json:"GMS"`
	GWk    int64   `json:"GWk"`
	HDop   float64 `json:"HDop"`
	Lat    float64 `json:"Lat"`
	Lng    float64 `json:"Lng"`
	NSats  int64   `json:"NSats"`
	Spd    float64 `json:"Spd"`
	Status int64   `json:"Status"`
	TimeUS int64   `json:"TimeUS"`
	U      int64   `json:"U"`
	Vz     float64 `json:"VZ"`
	Yaw    int64   `json:"Yaw"`
}

type MavMode struct {
	Mode    int64 `json:"Mode"`
	ModeNum int64 `json:"ModeNum"`
	Rsn     int64 `json:"Rsn"`
	TimeUS  int64 `json:"TimeUS"`
}

type MavBatt struct {
	Curr    float64 `json:"Curr"`
	CurrTot float64 `json:"CurrTot"`
	EnrgTot float64 `json:"EnrgTot"`
	Res     float64 `json:"Res"`
	Temp    int64   `json:"Temp"`
	TimeUS  int64   `json:"TimeUS"`
	Volt    float64 `json:"Volt"`
	VoltR   float64 `json:"VoltR"`
}

type MavOrigin struct {
	Alt    float64 `json:"Alt"`
	Lat    float64 `json:"Lat"`
	Lng    float64 `json:"Lng"`
	TimeUS int64   `json:"TimeUS"`
	Type   int64   `json:"Type"`
}

type MavCTUN struct {
	ABst   int64   `json:"ABst"`
	Alt    float64 `json:"Alt"`
	BAlt   float64 `json:"BAlt"`
	CRt    int64   `json:"CRt"`
	DAlt   int64   `json:"DAlt"`
	DCRt   int64   `json:"DCRt"`
	DSAlt  int64   `json:"DSAlt"`
	N      int64   `json:"N"`
	SAlt   int64   `json:"SAlt"`
	TAlt   float64 `json:"TAlt"`
	ThH    float64 `json:"ThH"`
	ThI    int64   `json:"ThI"`
	ThO    int64   `json:"ThO"`
	TimeUS int64   `json:"TimeUS"`
}

type MavRadio struct {
	Fixed    int64 `json:"Fixed"`
	Noise    int64 `json:"Noise"`
	Rssi     int64 `json:"RSSI"`
	RemNoise int64 `json:"RemNoise"`
	RemRSSI  int64 `json:"RemRSSI"`
	RxErrors int64 `json:"RxErrors"`
	TimeUS   int64 `json:"TimeUS"`
	TxBuf    int64 `json:"TxBuf"`
}

type MavRec struct {
	stamp time.Time
	a     MavAtt
	b     MavBatt
	c     MavCTUN
	g     MavGPS
	m     MavMode
	o     MavOrigin
	r     MavRadio
	err   MavErr
	ev    MavEvent
}

type MavEvent struct {
	ID     int64 `json:"Id"`
	TimeUS int64 `json:"TimeUS"`
}

type MavErr struct {
	Subsys int64 `json:"Subsys"`
	Ecode  int64 `json:"Ecode"`
	TimeUS int64 `json:"TimeUS"`
}

type APLOG struct {
	name string
	meta []types.FlightMeta
}

func NewAPReader(fn string) APLOG {
	var l APLOG
	l.name = fn
	l.meta = nil
	return l
}

func (o *APLOG) LogType() byte {
	return 'A'
}

func (o *APLOG) GetMetas() ([]types.FlightMeta, error) {
	m, err := metas(o.name)
	o.meta = m
	return m, err
}

func (o *APLOG) GetDurations() {
}

func (o *APLOG) Dump() {
}

func metas(logfile string) ([]types.FlightMeta, error) {
	var metas []types.FlightMeta
	fi, err := os.Stat(logfile)
	if err != nil {
		return nil, err
	}
	size := fi.Size()
	cmd := exec.Command("mavlogdump.py", "--format", "json", logfile)
	types.SetSilentProcess(cmd)
	out, err := cmd.StdoutPipe()
	defer cmd.Wait()
	defer out.Close()
	err = cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start err=%v", err)
		os.Exit(1)
	}

	mt := types.FlightMeta{Logname: logfile, Size: size, Start: 0}

	var st time.Time
	var mlog MavLog
	nl := 0
	scanner := bufio.NewScanner(out)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text := scanner.Text()
		if err := json.Unmarshal([]byte(text), &mlog); err == nil {
			st = time_from_log_stamp(mlog.Meta.Timestamp)
			if mt.Date.IsZero() {
				mt.Date = st
			}
			if mlog.Meta.Type == "GPS" {
				nl += 1
			}
		}
	}

	mt.Duration = st.Sub(mt.Date)
	mt.End = nl
	mt.Flags = types.Is_Valid | types.Has_Size | types.Has_Start
	metas = append(metas, mt)
	return metas, err
}

func time_from_log_stamp(s float64) time.Time {
	isec, fract := math.Modf(s)
	nsec := int64(fract * 1e9)
	return time.Unix(int64(isec), nsec)
}

func create_record(m MavRec, have_origin bool) (types.LogItem, bool) {
	b := types.LogItem{}
	b.Numsat = uint8(m.g.NSats)
	b.Hdop = uint16(m.g.HDop * 100)
	b.Volts = m.b.Volt
	b.Alt = m.c.Alt

	switch m.g.Status {
	case 0, 1:
		b.Fix = 0
	case 2:
		b.Fix = 1
	default:
		b.Fix = 2
	}
	b.Lat = m.g.Lat
	b.Lon = m.g.Lng
	b.GAlt = m.g.Alt
	b.Spd = m.g.Spd
	b.Stamp = uint64(m.g.TimeUS)
	b.Cog = uint32(m.g.GCrs)

	b.Amps = m.b.Curr
	b.Energy = m.b.CurrTot

	b.Roll = int16(m.a.Roll)
	b.Pitch = int16(m.a.Pitch)
	b.Cse = uint32(m.a.Yaw)
	b.Rssi = uint8(m.r.Rssi * 100 / 255)
	b.Throttle = int(m.c.ThI)

	b.Fmode = apmode_normalise(m.m.Mode)
	b.Fmtext = types.Mnames[b.Fmode]

	b.Status = types.Is_ARMED

	if m.ev.ID == 11 {
		b.Status = 0
	}

	switch m.err.Subsys {
	case 2, 3, 5, 6, 8, 9, 11, 12, 17, 18, 19, 20, 26:
		if m.err.Ecode != 0 {
			b.Status |= types.Is_FAIL
			b.HWfail = true // FIXME
		}
	}

	if !have_origin {
		if m.o.Type == 1 {
			have_origin = true
			b.Hlat = m.o.Lat
			b.Hlon = m.o.Lng
		}
	}
	b.Utc = m.stamp
	return b, have_origin
}

func apmode_normalise(mnum int64) uint8 {
	md := uint8(0)
	switch mnum {
	case 0:
		md = types.FM_ANGLE
	case 1:
		md = types.FM_ACRO
	case 2:
		md = types.FM_AH
	case 3:
		md = types.FM_WP
	case 4, 11:
		md = types.FM_CRUISE3D
	case 5, 7, 22:
		md = types.FM_PH
	case 6, 9, 21:
		md = types.FM_RTH
	case 13:
		md = types.FM_HORIZON
	case 18:
		md = types.FM_LAUNCH
	default:
		md = types.FM_ACRO
	}
	return md
}

func (lg *APLOG) Reader(m types.FlightMeta, ch chan interface{}) (types.LogSegment, bool) {
	cmd := exec.Command("mavlogdump.py", "--format", "json", lg.name)
	types.SetSilentProcess(cmd)
	out, err := cmd.StdoutPipe()
	defer cmd.Wait()
	defer out.Close()
	err = cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start err=%v", err)
		os.Exit(1)
	}

	var homes types.HomeRec
	var rec types.LogRec
	ndelay := 1000 * uint64(options.Config.Intvl)

	stats := types.LogStats{}
	var mrec MavRec
	var mlog MavLog
	have_origin := false
	var llat, llon float64
	var dt, st, lt uint64

	leffic := 0.0
	lwhkm := 0.0
	whacc := 0.0

	re := regexp.MustCompile(": NaN,")

	scanner := bufio.NewScanner(out)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text := scanner.Text()
		t := re.ReplaceAllString(text, ": null,")
		if err := json.Unmarshal([]byte(t), &mlog); err != nil {
			continue
		}
		switch mlog.Meta.Type {
		case "ATT":
			json.Unmarshal(mlog.Data, &mrec.a)
		case "ORGN":
			json.Unmarshal(mlog.Data, &mrec.o)
		case "BAT":
			json.Unmarshal(mlog.Data, &mrec.b)
		case "MODE":
			json.Unmarshal(mlog.Data, &mrec.m)
		case "CTUN":
			json.Unmarshal(mlog.Data, &mrec.c)
		case "ERR":
			json.Unmarshal(mlog.Data, &mrec.err)
		case "EV":
			json.Unmarshal(mlog.Data, &mrec.ev)
		case "RAD":
			json.Unmarshal(mlog.Data, &mrec.r)
		case "GPS":
			json.Unmarshal(mlog.Data, &mrec.g)
			mrec.stamp = time_from_log_stamp(mlog.Meta.Timestamp)
			b, xhave_origin := create_record(mrec, have_origin)
			if xhave_origin && have_origin == false {
				homes.HomeLat = b.Hlat
				homes.HomeLon = b.Hlon
				homes.HomeAlt = b.GAlt
				homes.Flags = types.HOME_ARM | types.HOME_ALT
				have_origin = true
				st = b.Stamp
				llat = b.Lat
				llon = b.Lon
				if ch != nil {
					ch <- homes
				}
			} else {
				c, d := geo.Csedist(homes.HomeLat, homes.HomeLon, b.Lat, b.Lon)
				b.Bearing = int32(c)
				b.Vrange = d * 1852.0
				us := b.Stamp
				if us > st {
					// Do the plot every 100ms
					if (us - dt) >= ndelay {
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

						if b.Amps > 0 || b.Energy > 0 {
							rec.Cap = rec.Cap | types.CAP_AMPS
						}
						if b.Volts > 0 {
							rec.Cap |= types.CAP_VOLTS
						}

						if b.Rssi > 0 {
							rec.Cap |= types.CAP_RSSI_VALID
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
		default:

		}
	}
	srec := stats.Summary(lt - st)
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
