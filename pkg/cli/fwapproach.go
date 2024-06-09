package cli

import (
	kml "github.com/twpayne/go-kml"
)

import (
	"geo"
)

type FWApproach struct {
	No      int8   `xml:"no,attr" json:"no"`
	Index   int8   `xml:"index,attr" json:"index"`
	Appalt  int32  `xml:"approachalt,attr" json:"appalt"`
	Landalt int32  `xml:"landalt,attr" json:"landalt"`
	Dirn1   int16  `xml:"landheading1,attr" json:"dirn1"`
	Dirn2   int16  `xml:"landheading2,attr" json:"dirn2"`
	Dref    string `xml:"approachdirection,attr" json:"dref"`
	Aref    bool   `xml:"sealevelref,attr" json:"aref"`
}

func AddLaylines(lat, lon float64, addAlt int32, lnd FWApproach, isvis bool) []kml.Element {
	ll := []kml.Element{}
	var altmode kml.AltitudeModeEnum

	lpath1, lpath2, apath1, apath2 := update_laylines(lat, lon, addAlt, lnd)

	altmode = kml.AltitudeModeAbsolute
	if !lnd.Aref {
		if addAlt == 0 {
			altmode = kml.AltitudeModeRelativeToGround
		}
	}

	if len(lpath1) > 0 {
		track := kml.Placemark(
			kml.Description("land path1"),
			kml.Name("land1"),
			kml.StyleURL("#styleFWLand"),
			kml.LineString(
				kml.AltitudeMode(altmode),
				kml.Extrude(false),
				kml.Tessellate(false),
				kml.Coordinates(lpath1...),
			),
		)
		track.Add(kml.Visibility(isvis))
		ll = append(ll, track)
	}

	if len(lpath2) > 0 {
		track := kml.Placemark(
			kml.Description("land path2"),
			kml.Name("land2"),
			kml.StyleURL("#styleFWLand"),
			kml.LineString(
				kml.AltitudeMode(altmode),
				kml.Extrude(false),
				kml.Tessellate(false),
				kml.Coordinates(lpath2...),
			),
		)
		track.Add(kml.Visibility(isvis))
		ll = append(ll, track)
	}

	if len(apath1) > 0 {
		track := kml.Placemark(
			kml.Description("approach path1"),
			kml.Name("approach1"),
			kml.StyleURL("#styleFWApproach"),
			kml.LineString(
				kml.AltitudeMode(altmode),
				kml.Extrude(false),
				kml.Tessellate(false),
				kml.Coordinates(apath1...),
			),
		)
		track.Add(kml.Visibility(isvis))
		ll = append(ll, track)
	}

	if len(apath2) > 0 {
		track := kml.Placemark(
			kml.Description("approach path"),
			kml.Name("approach2"),
			kml.StyleURL("#styleFWApproach"),
			kml.LineString(
				kml.AltitudeMode(altmode),
				kml.Extrude(false),
				kml.Tessellate(false),
				kml.Coordinates(apath2...),
			),
		)
		track.Add(kml.Visibility(isvis))
		ll = append(ll, track)
	}
	return ll
}

func update_laylines(lat, lon float64, addAlt int32, lnd FWApproach) ([]kml.Coordinate, []kml.Coordinate, []kml.Coordinate, []kml.Coordinate) {
	var apath1 []kml.Coordinate
	var apath2 []kml.Coordinate
	var lpath1 []kml.Coordinate
	var lpath2 []kml.Coordinate
	var p0 kml.Coordinate
	la := 0.0
	lo := 0.0

	lnd.Appalt = lnd.Appalt / 100
	lnd.Landalt = lnd.Landalt / 100
	if !lnd.Aref {
		lnd.Appalt += int32(addAlt)
		lnd.Landalt += int32(addAlt)
	}

	if lnd.Dirn1 != 0 {
		if lnd.Dirn1 < 0 {
			lnd.Dirn1 = -lnd.Dirn1
		} else {
			la, lo = geo.Posit(lat, lon, float64(lnd.Dirn1), Fwapproach_length)
			p0 = kml.Coordinate{Lon: lo, Lat: la, Alt: float64(lnd.Appalt)}
			lpath1 = append(lpath1, p0)
		}
		p0 = kml.Coordinate{Lon: lon, Lat: lat, Alt: float64(lnd.Landalt)}
		lpath1 = append(lpath1, p0)
		adir := (lnd.Dirn1 + 180) % 360
		la, lo = geo.Posit(lat, lon, float64(adir), Fwapproach_length)
		p0 = kml.Coordinate{Lon: lo, Lat: la, Alt: float64(lnd.Appalt)}
		lpath1 = append(lpath1, p0)
		apath1 = add_approach(lnd.Dref, int(lnd.Dirn1), lpath1)
	}

	if lnd.Dirn2 != 0 {
		if lnd.Dirn2 < 0 {
			lnd.Dirn2 = -lnd.Dirn2
		} else {
			la, lo = geo.Posit(lat, lon, float64(lnd.Dirn2), Fwapproach_length)
			p0 = kml.Coordinate{Lon: lo, Lat: la, Alt: float64(lnd.Appalt)}
			lpath2 = append(lpath2, p0)
		}
		p0 = kml.Coordinate{Lon: lon, Lat: lat, Alt: float64(lnd.Landalt)}
		lpath2 = append(lpath2, p0)
		adir := (lnd.Dirn2 + 180) % 360
		la, lo = geo.Posit(lat, lon, float64(adir), Fwapproach_length)
		p0 = kml.Coordinate{Lon: lo, Lat: la, Alt: float64(lnd.Appalt)}
		lpath2 = append(lpath2, p0)
		apath2 = add_approach(lnd.Dref, int(lnd.Dirn2), lpath2)
	}
	return lpath1, lpath2, apath1, apath2
}

func add_approach(dref string, dirn int, lpath []kml.Coordinate) []kml.Coordinate {
	var apath []kml.Coordinate
	xdir := dirn
	if dref == "right" {
		xdir += 90
	} else {
		xdir -= 90
	}
	ilp := 0
	iap := 1

	if len(lpath) == 3 {
		ilp = 1
		iap = 2
	}

	fwax := Fwapproach_length / 2.0
	fwlr := Fwloiter_radius * 4.0

	if fwax < fwlr {
		fwax = fwlr
	}

	lax, lox := geo.Posit(lpath[iap].Lat, lpath[iap].Lon, float64(xdir), fwax)
	apath = append(apath, lpath[iap])
	apath = append(apath, kml.Coordinate{Lon: lox, Lat: lax, Alt: lpath[iap].Alt})
	apath = append(apath, lpath[ilp])
	if len(lpath) == 3 {
		lax, lox = geo.Posit(lpath[0].Lat, lpath[0].Lon, float64(xdir), fwax)
		apath = append(apath, kml.Coordinate{Lon: lox, Lat: lax, Alt: lpath[iap].Alt})
		apath = append(apath, lpath[0])
	}
	return apath
}
