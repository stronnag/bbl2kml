package api

import (
	"time"
	"fmt"
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

type LogRec struct {
	Stamp   uint64
	Lat     float64
	Lon     float64
	Alt     float64
	GAlt    float64
	Cse     uint32
	Spd     float64
	Amps    float64
	Volts   float64
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

var Mnames = []string{"Acro", "Manual", "Horizon", "Angle", "Launch", "RTH", "WP",
	"3CRS", "CRS", "PH", "AH", "EMERG", "F/S"}

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

func (b *LogStats) ShowSummary(t uint64) {
	b.Duration = t
	b.Max_range *= 1852.0
	b.Distance *= 1852.0
	fmt.Printf("Altitude : %.1f m at %s\n", b.Max_alt, b.Show_time(b.Max_alt_time))
	fmt.Printf("Speed    : %.1f m/s at %s\n", b.Max_speed, b.Show_time(b.Max_speed_time))
	fmt.Printf("Range    : %.0f m at %s\n", b.Max_range, b.Show_time(b.Max_range_time))
	if b.Max_current > 0 {
		fmt.Printf("Current  : %.1f A at %s\n", b.Max_current, b.Show_time(b.Max_current_time))
	}
	fmt.Printf("Distance : %.0f m\n", b.Distance)
	fmt.Printf("Duration : %s\n", b.Show_time(b.Duration))
}
