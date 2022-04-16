package api

import (
	"fmt"
	"strings"
	"time"
)

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

const (
	CAP_AMPS = 1 << iota
	CAP_VOLTS
	CAP_ENERGY
	CAP_RSSI_VALID
)

const (
	Is_ARMED uint8 = 1
	Is_FAIL  uint8 = 2
	Is_CRSF  uint8 = 4
	Is_ARDU  uint8 = 8
)

type Reason int

func (r Reason) String() string {
	var reasons = [...]string{"None", "Timeout", "Sticks", "Switch_3d", "Switch", "Killswitch", "Failsafe", "Navigation"}
	if r < 0 || int(r) >= len(reasons) {
		r = 0
	}
	return reasons[r]
}

type LogItem struct {
	Stamp    uint64
	Lat      float64
	Lon      float64
	Alt      float64
	GAlt     float64
	Cse      uint32
	Cog      uint32
	Spd      float64
	Amps     float64
	Volts    float64
	Fix      uint8
	Numsat   uint8
	Fmode    uint8
	Rssi     uint8
	Fmtext   string
	Utc      time.Time
	Hlat     float64
	Hlon     float64
	Vrange   float64
	Bearing  int32 // -ve => not defined
	Tdist    float64
	Effic    float64
	Energy   float64
	Whkm     float64
	WhAcc    float64
	Qval     float64 // scaled efficiency
	Sval     float64 // scaled speed
	Aval     float64 // scaled Altitude
	Bval     float64 // scaled Battery
	Throttle int
	Roll     int16
	Pitch    int16
	Status   uint8
	Hdop     uint16
	HWfail   bool
	NavMode  byte
}

type LogRec struct {
	Cap   uint8
	Items []LogItem
}

var Mnames = []string{"Acro", "Manual", "Horizon", "Angle", "Launch", "RTH", "WP",
	"3CRS", "CRS", "PH", "AH", "EMERGENCY", "F/S"}

const (
	HOME_ARM  = 1
	HOME_SAFE = 2
	HOME_ALT  = 4
)

type HomeRec struct {
	Flags   uint8
	HomeLat float64
	HomeLon float64
	HomeAlt float64
	SafeLat float64
	SafeLon float64
}

type MetaLog interface {
	LogName() string
	MetaData() map[string]string
	Valid() bool
}

type LogStats struct {
	Max_alt          float64
	Max_alt_time     uint64
	Max_range        float64
	Max_range_time   uint64
	Max_speed        float64
	Max_speed_time   uint64
	Max_current      float64
	Max_current_time uint64
	Distance         float64
	Duration         uint64
}

func (b *LogStats) Show_time(t uint64) string {
	secs := t / 1000000
	m := secs / 60
	s := secs % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func (b *LogStats) Summary(t uint64) MapRec {
	var m MapRec
	m = make(MapRec)
	b.Duration = t
	b.Max_range *= 1852.0
	b.Distance *= 1852.0
	m["Altitude"] = fmt.Sprintf("%.1f m at %s", b.Max_alt, b.Show_time(b.Max_alt_time))
	m["Speed"] = fmt.Sprintf("%.1f m/s at %s", b.Max_speed, b.Show_time(b.Max_speed_time))
	m["Range"] = fmt.Sprintf("%.0f m at %s", b.Max_range, b.Show_time(b.Max_range_time))
	if b.Max_current > 0 {
		m["Current"] = fmt.Sprintf("%.1f A at %s", b.Max_current, b.Show_time(b.Max_current_time))
	}
	m["Distance"] = fmt.Sprintf("%.0f m", b.Distance)
	m["Duration"] = fmt.Sprintf("%s", b.Show_time(b.Duration))
	return m
}

type MapRec map[string]string

type LogSegment struct {
	L LogRec
	H HomeRec
	M MapRec
}

type FlightLog interface {
	Reader(FlightMeta, chan interface{}) (LogSegment, bool)
	GetMetas() ([]FlightMeta, error)
	GetDurations()
	Dump()
	LogType() byte
}

const (
	Is_Valid = 1 << iota
	Has_Craft
	Has_Firmware
	Has_Disarm
	Has_Size
	Has_Start
	Is_Suspect = (1 << 7)
)

const (
	Has_Acc = 1 << iota
	Has_Baro
	Has_Mag
	Has_GPS
	Has_Sonar
	Has_Opflow
	Has_Pitot
)

const (
	Feature_GPS     = (1 << 7)
	Feature_VBAT    = (1 << 1)
	Feature_CURRENT = (1 << 11)
)

const (
	AFlags_EFFIC = 1 << iota
	AFlags_SPEED
	AFlags_ALTITUDE
	AFlags_BATTERY
)

type FlightMeta struct {
	Logname  string
	Date     time.Time
	Duration time.Duration
	Craft    string
	Firmware string
	Fwdate   string
	Disarm   Reason
	Size     int64
	Index    int
	Start    int
	End      int
	Flags    uint8
	Motors   uint8
	Servos   uint8
	Sensors  uint16
}

func (b *FlightMeta) LogName() string {
	name := b.Logname
	if b.Index > 0 {
		name = name + fmt.Sprintf(" / %d", b.Index)
	}
	return name
}

func (b *FlightMeta) ShowSize() (string, bool) {
	if b.Flags&Has_Size == 0 {
		return "", false
	} else {
		var s string
		switch {
		case b.Size > 1024*1024:
			s = fmt.Sprintf("%.2f MB", float64(b.Size)/(1024*1024))
		case b.Size > 10*1024:
			s = fmt.Sprintf("%.1f KB", float64(b.Size)/1024)
		default:
			s = fmt.Sprintf("%d B", b.Size)
		}
		return s, true
	}
}

func (b *FlightMeta) ShowDisarm() (string, bool) {
	if b.Flags&Has_Disarm == 0 {
		return "", false
	} else {
		return b.Disarm.String(), true
	}
}

func (b *FlightMeta) ShowFirmware() (string, bool) {
	if b.Flags&Has_Firmware == 0 {
		return "", false
	} else {
		return fmt.Sprintf("%s of %s", b.Firmware, b.Fwdate), true
	}
}

func (b *FlightMeta) Flight() string {
	var sb strings.Builder
	if b.Flags&Has_Craft != 0 {
		sb.WriteString(b.Craft)
		sb.WriteString(" on ")
	}
	sb.WriteString(b.Date.Format("2006-01-02 15:04:05"))
	return sb.String()
}

func (b *FlightMeta) Summary() MapRec {
	var m MapRec
	m = make(MapRec)
	m["Log"] = b.LogName()
	m["Flight"] = b.Flight()
	if s, ok := b.ShowFirmware(); ok {
		m["Firmware"] = s
	}
	if s, ok := b.ShowSize(); ok {
		m["Size"] = s
	}
	return m
}
