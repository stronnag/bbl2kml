package api

import (
	"time"
	"fmt"
)

type BBLSummary struct {
	Valid    bool
	Logname  string
	Craft    string
	Cdate    string
	Firmware string
	Fwdate   string
	Disarm   string
	Index    int
	Size     int64
}

func (b *BBLSummary) Show_size(sz int64) string {
	var s string
	switch {
	case sz > 1024*1024:
		s = fmt.Sprintf("%.2f MB", float64(sz)/(1024*1024))
	case sz > 10*1024:
		s = fmt.Sprintf("%.1f KB", float64(sz)/1024)
	default:
		s = fmt.Sprintf("%d B", sz)
	}
	return s
}

type BBLStats struct {
	Valid            bool
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

func (b *BBLStats) Show_time(t uint64) string {
	secs := t / 1000000
	m := secs / 60
	s := secs % 60
	return fmt.Sprintf("%02d:%02d", m, s)
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
	Stamp   uint64
	Lat     float64
	Lon     float64
	Alt     float64
	GAlt    float64
	Cse     uint32
	Spd     float64
	Amps    float64
	Fix     uint8
	Numsat  uint8
	Fmode   uint8
	Rssi    uint8
	Fmtext  string
	Utc     time.Time
	Fs      bool
	Hlat    float64
	Hlon    float64
	Vrange  float64
	Bearing int32 // -ve => not defined
	Tdist   float64
}

type HomeRec struct {
	Flags   uint8
	HomeLat float64
	HomeLon float64
	HomeAlt float64
	SafeLat float64
	SafeLon float64
}

const (
	HOME_ARM  = 1
	HOME_SAFE = 2
	HOME_ALT  = 4
)

var Mnames = []string{"Acro", "Manual", "Horizon", "Angle", "Launch", "RTH", "WP",
	"3CRS", "CRS", "PH", "AH", "EMERG", "F/S"}
