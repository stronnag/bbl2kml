package kmlgen

import (
	"fmt"
	kml "github.com/twpayne/go-kml"
)

import (
	"cli"
	"geo"
	"styles"
)

func add_sh_circle(sh cli.SafeHome, i int) kml.Element {
	var points []kml.Coordinate

	for j := 0; j < 360; j += 5 {
		lat, lon := geo.Posit(sh.Lat, sh.Lon, float64(j), cli.Safehome_distance)
		points = append(points, kml.Coordinate{Lon: lon, Lat: lat, Alt: 0})
	}
	points = append(points, points[0])
	track := kml.Placemark(
		//		kml.Name(fmt.Sprintf("Circle %d", g.Zid)),
		//		kml.Description(fmt.Sprintf("Circle Zone %d", g.Zid)),
		kml.StyleURL("#styleSAFEHOME"))

	track.Add(
		kml.Polygon(
			kml.AltitudeMode(kml.AltitudeModeRelativeToGround),
			kml.Extrude(true),
			kml.Tessellate(false),
			kml.OuterBoundaryIs(
				kml.LinearRing(
					kml.Coordinates(points...),
				),
			),
		),
	)
	return track
}

func Generate_cli_kml(fn string, fb *geo.Frob) []kml.Element {
	kmls := []kml.Element{}
	sf := kml.Folder(kml.Name("Safehomes")).Add(kml.Open(true))
	kmls = append(kmls, sf)
	sha, fwa, gzone := cli.Read_clifile(fn)
	if len(sha) > 0 {
		sf.Add(styles.Get_safe_styles()...)
		for i, sh := range sha {
			if fb != nil {
				sh.Lat, sh.Lon, _ = fb.Relocate(sh.Lat, sh.Lon, 0)
			}
			name := fmt.Sprintf("Safehome %d", i)
			shf := kml.Folder(kml.Name(name)).Add(kml.Description(name)).Add(kml.Visibility(true)).Add(add_sh_circle(sh, i))
			p := kml.Placemark(
				kml.Name(name),
				kml.StyleURL("#styleSAFEHOME"),
				kml.Point(
					kml.AltitudeMode(kml.AltitudeModeClampToGround),
					kml.Coordinates(kml.Coordinate{Lon: sh.Lon, Lat: sh.Lat, Alt: 0.0}),
				),
			)
			p.Add(kml.Visibility(true))
			shf.Add(p)
			if len(fwa) > 0 {
				sf.Add(styles.Get_approach_styles()...)
				fidx := -1
				for fi, fw := range fwa {
					if int(fw.No) == i {
						fidx = fi
						break
					}
				}
				if fidx != -1 {
					for _, lf := range cli.AddLaylines(sh.Lat, sh.Lon, 0, fwa[fidx], true) {
						shf.Add(lf)
					}
				}
			}
			sf.Add(shf)
		}
	}
	if len(gzone) > 0 {
		gf := Gen_geozones(gzone)
		kmls = append(kmls, gf)
	}
	return kmls
}
