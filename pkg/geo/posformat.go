package geo

import (
	"fmt"
	"strings"
	"math"
)

func LatFormat(lat float64, dms bool) string {
	if dms == false {
		return fmt.Sprintf("%.6f", lat)
	} else {
		return dms_format(lat, "%02d:%02d:%04.1f%c", "NS")
	}
}

func LonFormat(lon float64, dms bool) string {
	if dms == false {
		return fmt.Sprintf("%.6f", lon)
	} else {
		return dms_format(lon, "%03d:%02d:%04.1f%c", "EW")
	}
}

func PositionFormat(lat float64, lon float64, dms bool) string {
	if dms == false {
		return fmt.Sprintf("%.6f %.6f", lat, lon)
	} else {
		var sb strings.Builder
		sb.WriteString(LatFormat(lat, dms))
		sb.WriteByte(' ')
		sb.WriteString(LonFormat(lon, dms))
		return sb.String()
	}
}

func dms_format(coord float64, ofmt string, ind string) string {
	neg := (coord < 0.0)
	ds := math.Abs(coord)
	d := int(ds)
	rem := (ds - float64(d)) * 3600.0
	m := int(rem / 60)
	var s float64 = rem - float64(m*60)
	if int(s*10) == 600 {
		m += 1
		s = 0
	}
	if m == 60 {
		m = 0
		d += 1
	}
	var q byte
	if neg {
		q = ind[1]
	} else {
		q = ind[0]
	}
	return fmt.Sprintf(ofmt, int(d), int(m), s, q)
}
