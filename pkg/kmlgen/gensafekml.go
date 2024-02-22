package kmlgen

import (
	"fmt"
	kml "github.com/twpayne/go-kml"
)

import (
	"inav"
	"mission"
)

func Generate_safekml(fn string) kml.Element {
	sf := kml.Folder(kml.Name("Safehomes")).Add(kml.Open(true))
	sha, fwa := inav.Read_safehome(fn)
	if len(sha) > 0 {
		var wps []kml.Element
		for i, sh := range sha {
			sname := fmt.Sprintf("Safehome %d", i)
			p := kml.Placemark(
				kml.Name(sname),
				kml.StyleURL("#styleSAFEHOME"),
				kml.Point(
					kml.AltitudeMode(kml.AltitudeModeRelativeToGround),
					kml.Coordinates(kml.Coordinate{Lon: sh.Lon, Lat: sh.Lat, Alt: 0.0}),
				),
			)
			p.Add(kml.Visibility(true))
			wps = append(wps, p)
			fidx := -1
			for fi, fw := range fwa {
				if int(fw.No) == i {
					fidx = fi
					break
				}
			}
			if fidx != -1 {
				for _, lf := range mission.AddLaylines(sh.Lat, sh.Lon, 0, fwa[fidx], true) {
					sf.Add(lf)
				}
			}
		}
		sf.Add(wps...)
	}
	return sf
}
