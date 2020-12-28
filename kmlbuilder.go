package main

import (
	"image/color"
	"log"
	"os"
	kml "github.com/twpayne/go-kml"
	"github.com/twpayne/go-kml/icon"
	kmz "github.com/twpayne/go-kmz"
	"time"
	"fmt"
	"io"
	"strings"
)


func getflightColour(mode uint8) color.Color {
	var c color.Color;
	switch mode {
	case FM_LAUNCH:
		c = color.RGBA{R: 0, G: 160, B: 160, A: 0xa0}
	case FM_RTH:
		c = color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xa0}
	case FM_WP:
		c = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xa0}
	case FM_CRUISE3D, FM_CRUISE2D:
		c = color.RGBA{R: 0xbf, G: 0x88, B: 0xf0, A: 0xa0}
	case FM_PH:
		c = color.RGBA{R: 0xce, G: 0xff, B: 0x9d, A: 0xa0}
	case FM_AH:
		c = color.RGBA{R: 0x3, G: 0xc0, B: 0xf0, A: 0xa0}
	case FM_FS:
		c = color.RGBA{R: 0xff, G: 0, B: 0, A: 0xa0}
	default:
		c = color.RGBA{R: 0, G: 0xff, B: 0xff, A: 0xa0}
	}
	return c
}

func getStyleURL(r BBLRec) string {
	var s string
	if r.fs {
		return "#styleFS"
	}
	switch r.fmode {
	case FM_LAUNCH:
		s = "#styleLaunch"
	case FM_RTH:
		s = "#styleRTH"
	case FM_WP:
		s = "#styleWP"
	case FM_CRUISE3D, FM_CRUISE2D:
		s = "#styleCRS"
	case FM_PH:
		s = "#stylePH"
	default:
		s = "#styleNormal"
	}
	return s
}

func getPoints(recs []BBLRec) []kml.Element {
	var pt []kml.Element
	for _, r := range recs {
		ts,_ := time.Parse(time.RFC3339Nano,r.utc)
		tfmt := ts.Format("2006-01-02T15:04:05.99MST")
		fmtxt :=  r.fmtext
		if r.fs {
			fmtxt = fmtxt + " FAILSAFE"
		}
		str := fmt.Sprintf("Time: %s<br/>Position: %.7f %.7f %.0fm<br/>Course: %d°<br/>Speed: %.1fm/s<br/>Satellites: %d<br/>Range: %.0fm<br/>Bearing: %d°<br/>RSSI: %d%%<br/>Mode: %s<br/>Distance: %.0fm<br/>", tfmt, r.lat, r.lon, r.alt, r.cse, r.spd, r.numsat, r.vrange, r.bearing, r.rssi, fmtxt, r.tdist);
		k := kml.Placemark(
			kml.Description(str),
			kml.TimeStamp(kml.When(ts)),
			kml.StyleURL(getStyleURL(r)),
			kml.Point(
				kml.AltitudeMode("relativeToGround"),
				kml.Coordinates(kml.Coordinate{Lon: r.lon, Lat: r.lat, Alt: r.alt}),
			),
		)
		pt = append(pt, k)
	}
	return pt;
}


func getHomes(hpos []float64) []kml.Element {
	var hp []kml.Element
	k := kml.Placemark(
		kml.Name("Armed"),
		kml.Point(
			kml.Coordinates(kml.Coordinate{Lon: hpos[1], Lat: hpos[0]}),
		),
		kml.Style(
			kml.IconStyle(
				kml.Icon(
					kml.Href(icon.PaletteHref(4,29)),
				),
			),
		),
	)
	hp = append(hp, k)
	if len(hpos) > 2 {
		k = kml.Placemark(
			kml.Name("SafeHome"),
			kml.Point(
				kml.Coordinates(kml.Coordinate{Lon: hpos[3], Lat: hpos[2]}),
			),
			kml.Style(
				kml.IconStyle(
					kml.Icon(
						kml.Href(icon.PaletteHref(3, 56)),
					),
				),
			),
		)
		hp = append(hp, k)
	}
	return hp
}

func openStdoutOrFile(path string) (io.WriteCloser, error) {
	var err error
	var w io.WriteCloser

	if len(path) == 0 || path == "-" {
		w = os.Stdout
	} else {
		w, err = os.Create(path)
	}
	return w, err
}

func GenerateKML(hpos []float64, recs []BBLRec, outfn string) {

	a1 := getHomes(hpos)
	a1 = append(a1, getPoints(recs)...)

	f:= kml.Folder(
			append([]kml.Element{
				kml.Name("inav flight"),
				kml.SharedStyle(
					"styleNormal",
					kml.IconStyle(
						kml.Scale(0.5),
						kml.Color(getflightColour(FM_ACRO)),
						kml.Icon(
							kml.Href(icon.PaletteHref(2, 18)),
						),
					),
				),
				kml.SharedStyle(
					"styleLaunch",
					kml.IconStyle(
						kml.Scale(0.5),
						kml.Color(getflightColour(FM_LAUNCH)),
						kml.Icon(
							kml.Href(icon.PaletteHref(2, 18)),
						),
					),
				),
				kml.SharedStyle(
					"styleWP",
					kml.IconStyle(
						kml.Scale(0.5),
						kml.Color(getflightColour(FM_WP)),
						kml.Icon(
							kml.Href(icon.PaletteHref(2, 18)),
						),
					),
				),
				kml.SharedStyle(
					"styleRTH",
					kml.IconStyle(
						kml.Scale(0.5),
						kml.Color(getflightColour(FM_RTH)),
						kml.Icon(
							kml.Href(icon.PaletteHref(2, 18)),
						),
					),
				),
				kml.SharedStyle(
					"styleCRS",
					kml.IconStyle(
						kml.Scale(0.5),
						kml.Color(getflightColour(FM_CRUISE3D)),
						kml.Icon(
							kml.Href(icon.PaletteHref(2, 18)),
						),
					),
				),
				kml.SharedStyle(
					"stylePH",
					kml.IconStyle(
						kml.Scale(0.5),
						kml.Color(getflightColour(FM_PH)),
						kml.Icon(
							kml.Href(icon.PaletteHref(2, 18)),
						),
					),
				),
				kml.SharedStyle(
					"styleAH",
					kml.IconStyle(
						kml.Scale(0.5),
						kml.Color(getflightColour(FM_AH)),
						kml.Icon(
							kml.Href(icon.PaletteHref(2, 18)),
						),
					),
				),
				kml.SharedStyle(
					"styleFS",
					kml.IconStyle(
						kml.Scale(0.5),
						kml.Color(getflightColour(FM_FS)),
						kml.Icon(
							kml.Href(icon.PaletteHref(2, 18)),
						),
					),
				),
			},
				a1...,
			)...,
	)
	var err error
	if strings.HasSuffix(outfn, ".kmz") {
		z := kmz.NewKMZ(f)
		w,err := os.Create(outfn)
		if err == nil {
			err = z.WriteIndent(w, "", "  ")
		}
	} else {
		k := kml.KML(f)
		output,err := openStdoutOrFile(outfn)
		if err == nil {
			err = k.WriteIndent(output, "", "  ")
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}
