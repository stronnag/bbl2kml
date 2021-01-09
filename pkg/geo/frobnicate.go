package geo

import (
	"strings"
	"os"
	"strconv"
)

var (
	jmp_dist  float64 = -1.0
	jmp_angle float64 = 0.0
	jmp_up    float64 = 0.0
	jlat      float64
	jlon      float64
	frobinit  bool
)

func Frobnicate_init() bool {
	fob := os.Getenv("BBL2KML_SHIFT")
	if fob != "" {
		parts := strings.Split(fob, ",")
		if len(parts) > 1 {
			jlat, _ = strconv.ParseFloat(parts[0], 64)
			jlon, _ = strconv.ParseFloat(parts[1], 64)
			if len(parts) == 3 {
				jmp_up, _ = strconv.ParseFloat(parts[2], 64)
			}
			frobinit = true
		}
	}
	return frobinit
}

func Frobnicate_set(lat float64, lon float64, alt float64) (float64, float64) {
	jmp_angle, jmp_dist = Csedist(lat, lon, jlat, jlon)
	if alt != 0 {
		jmp_up -= alt
	}
	return jmp_angle, jmp_dist
}

func Frobnicate_move(lat float64, lon float64, alt float64) (float64, float64, float64) {
	var nlat, nlon, nalt float64
	nlat, nlon = Posit(lat, lon, jmp_angle, jmp_dist, false)
	nalt = alt + jmp_up
	return nlat, nlon, nalt
}

func Getfrobnication() bool {
	return frobinit
}
