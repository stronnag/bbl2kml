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

func get_style(t int) string {
	var st string
	if t == cli.TYPE_EXC {
		st = "#styleEXC"
	} else {
		st = "#styleINC"
	}
	return st
}

func add_poly(g cli.GeoZone, fb *geo.Frob) kml.Element {
	var points []kml.Coordinate
	st := get_style(g.Gtype)
	for _, pt := range g.Points {
		if fb != nil {
			pt.Lat, pt.Lon, _ = fb.Relocate(pt.Lat, pt.Lon, 0)
		}

		points = append(points, kml.Coordinate{Lon: pt.Lon, Lat: pt.Lat, Alt: float64(g.Maxalt / 100.0)})
	}
	points = append(points, points[0])
	track := kml.Placemark(
		kml.Name(fmt.Sprintf("Track %d", g.Zid)),
		kml.Description(fmt.Sprintf("Polyline Track %d", g.Zid)),
		kml.StyleURL(st))
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
	name := fmt.Sprintf("Poly %d", g.Zid)
	desc := fmt.Sprintf(g.To_string())
	kml := kml.Folder(kml.Name(name)).Add(kml.Description(desc)).Add(kml.Visibility(true)).Add(track)
	return kml
}

func add_circle(g cli.GeoZone, fb *geo.Frob) kml.Element {
	var points []kml.Coordinate
	st := get_style(g.Gtype)

	if fb != nil {
		g.Points[0].Lat, g.Points[0].Lon, _ = fb.Relocate(g.Points[0].Lat, g.Points[0].Lon, 0)
	}

	for j := 0; j < 360; j += 5 {
		lat, lon := geo.Posit(g.Points[0].Lat, g.Points[0].Lon, float64(j), g.Points[1].Lat/1852.0)
		points = append(points, kml.Coordinate{Lon: lon, Lat: lat, Alt: float64(g.Maxalt / 100.0)})
	}
	points = append(points, points[0])
	track := kml.Placemark(
		kml.Name(fmt.Sprintf("Circle %d", g.Zid)),
		kml.Description(fmt.Sprintf("Circle Zone %d", g.Zid)),
		kml.StyleURL(st))

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
	name := fmt.Sprintf("Circle %d", g.Zid)
	desc := fmt.Sprintf(g.To_string())
	kml := kml.Folder(kml.Name(name)).Add(kml.Description(desc)).Add(kml.Visibility(true)).Add(track)
	return kml
}

func Gen_geozones(gzones []cli.GeoZone, fb *geo.Frob) kml.Element {
	d := kml.Folder(kml.Name("Geozone")).Add(kml.Open(true))
	d.Add(styles.Get_zone_styles()...)
	for _, g := range gzones {
		switch g.Shape {
		case cli.SHAPE_CIRCLE:
			d.Add(add_circle(g, fb))
		case cli.SHAPE_POLY:
			d.Add(add_poly(g, fb))
		}
	}
	return d
}
