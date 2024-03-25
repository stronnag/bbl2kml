package geo

import (
	"strconv"
	"strings"
)

import (
	"options"
)

type Point struct {
	lat float64
	lon float64
}

type Frob struct {
	orig  Point
	reloc Point
	ralt  float64
}

var (
	fb *Frob
)

func Msplit(s string, separators []rune) []string {
	f := func(r rune) bool {
		for _, s := range separators {
			if r == s {
				return true
			}
		}
		return false
	}
	return strings.FieldsFunc(s, f)
}

func Frobnicate_init() *Frob {
	fb = nil
	if len(options.Config.Rebase) != 0 {
		parts := Msplit(options.Config.Rebase, []rune{'/', ':', ';', ' ', ','})
		if len(parts) > 1 {
			jmp_up := 0.0
			jlat, _ := strconv.ParseFloat(parts[0], 64)
			jlon, _ := strconv.ParseFloat(parts[1], 64)
			if len(parts) == 3 {
				jmp_up, _ = strconv.ParseFloat(parts[2], 64)
			} else {
				d := InitDem("")
				jmp_up, _ = d.Get_Elevation(jlat, jlon)
			}
			fb = &Frob{
				Point{0.0, 0.0},
				Point{jlat, jlon},
				jmp_up,
			}
			return fb
		}
		return nil
	}
	return nil
}

func (f *Frob) Get_rebase() (float64, float64, float64) {
	return f.reloc.lat, f.reloc.lon, f.ralt
}

func (f *Frob) Set_origin(olat, olon, oalt float64) {
	f.orig.lat = olat
	f.orig.lon = olon
	if oalt != 0 {
		f.ralt -= oalt
	}
}

func (f *Frob) Relocate(lat, lon, alt float64) (float64, float64, float64) {
	c, d := Csedist(f.orig.lat, f.orig.lon, lat, lon)
	xlat, xlon := Posit(f.reloc.lat, f.reloc.lon, c, d)
	xalt := alt + f.ralt
	return xlat, xlon, xalt
}

func Getfrobnication() *Frob {
	return fb
}
