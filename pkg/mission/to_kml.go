package mission

import (
	"fmt"
	kml "github.com/twpayne/go-kml"
	"github.com/twpayne/go-kml/icon"
	"image/color"
)

import (
	"geo"
	"types"
)

const LAYLEN = (350.0 / 1852.0)

func (m *Mission) get_fly_points(addAlt int32) ([]kml.Coordinate, bool) {
	var points []kml.Coordinate
	nsize := len(m.MissionItems)
	ret := false

	jumpC := make([]int16, nsize)

	for j, mi := range m.MissionItems {
		if mi.Action == "JUMP" {
			jumpC[j] = mi.P2
		}
	}
	n := 0
	var alt int32
	for {
		if n >= nsize {
			break
		}
		var typ = m.MissionItems[n].Action
		if typ == "SET_POI" || typ == "SET_HEAD" {
			n += 1
			continue
		}
		if typ == "JUMP" {
			if jumpC[n] == -1 {
				n += 1
			} else {
				if jumpC[n] == 0 {
					jumpC[n] = m.MissionItems[n].P2
					n += 1
				} else {
					jumpC[n] -= 1
					n = int(m.MissionItems[n].P1) - 1
				}
			}
			continue
		}

		if typ == "RTH" {
			ret = true
			break
		}
		mi := m.MissionItems[n]
		if mi.P3 == 0 {
			alt = mi.Alt + addAlt
		} else {
			alt = mi.Alt
		}
		pt := kml.Coordinate{Lon: mi.Lon, Lat: mi.Lat, Alt: float64(alt)}
		points = append(points, pt)
		n += 1
	}
	return points, ret
}

func (m *Mission) To_kml(hpos types.HomeRec, dms bool, fake bool, mmidx int, isvis bool) kml.Element {
	var points []kml.Coordinate
	var wps []kml.Element
	llat := 0.0
	llon := 0.0
	lalt := int32(0)
	landid := -1
	var altmode kml.AltitudeModeEnum

	addAlt := int32(0)

	if (hpos.Flags & types.HOME_ARM) != 0 {
		if (hpos.Flags & types.HOME_ALT) == types.HOME_ALT {
			addAlt = int32(hpos.HomeAlt)
		} else {
			bingelev, err := geo.GetElevation(hpos.HomeLat, hpos.HomeLon)
			if err == nil {
				addAlt = int32(bingelev)
				hpos.Flags |= types.HOME_ALT
			}
		}

		if (hpos.Flags & types.HOME_ALT) == types.HOME_ALT {
			altmode = kml.AltitudeModeAbsolute
		} else {
			altmode = kml.AltitudeModeRelativeToGround
		}
		points = append(points, kml.Coordinate{Lon: hpos.HomeLon, Lat: hpos.HomeLat, Alt: float64(addAlt)})
	}

	var lat, lon float64
	var alt int32

	for mn, mi := range m.MissionItems {
		if mi.Action == "JUMP" || mi.Action == "SET_HEAD" || mi.Action == "RTH" {
			lat = llat
			lon = llon
			alt = lalt
		} else {
			lat = mi.Lat
			lon = mi.Lon
			if mi.P3 == 0 {
				alt = mi.Alt + addAlt
			} else {
				alt = mi.Alt
			}
			llat = lat
			llon = lon
			lalt = alt
			if mi.Action == "LAND" {
				landid = mn
			}
		}
		var bname string

		switch mi.Action {
		case "WAYPOINT":
			bname = "WP"
		case "POSHOLD_UNLIM", "POSHOLD_TIME":
			bname = "PH"
		case "SET_POI":
			bname = "POI"
		case "SET_HEAD":
			bname = "HD"
		default:
			bname = mi.Action
		}
		name := fmt.Sprintf("%s %d", bname, mi.No)
		p := kml.Placemark(
			kml.Name(name),
			kml.Description(fmt.Sprintf("Action: %s<br/>Position: %s<br/>Elevation: %dm<br/>GPS Altitude: %dm<br/>",
				mi.Action, geo.PositionFormat(lat, lon, dms), mi.Alt, alt)),
			kml.StyleURL(fmt.Sprintf("#style%s", mi.Action)),
			kml.Point(
				kml.AltitudeMode(altmode),
				kml.Coordinates(kml.Coordinate{Lon: lon, Lat: lat, Alt: float64(alt)}),
			),
		)
		p.Add(kml.Visibility(isvis))
		wps = append(wps, p)
	}

	var desc string
	if (hpos.Flags & types.HOME_ALT) == types.HOME_ALT {
		desc = fmt.Sprintf("Created from %s with elevations adjusted for home location %s",
			m.mission_file, geo.PositionFormat(hpos.HomeLat, hpos.HomeLon, dms))
		if fake {
			p := kml.Placemark(
				kml.Name("Home"),
				kml.Description(fmt.Sprintf("Assumed Home<br/>Position: %s<br/>GPS Altitude: %dm<br/>",
					geo.PositionFormat(hpos.HomeLat, hpos.HomeLon, dms), addAlt)),
				kml.StyleURL("#styleFakeHome"),
				kml.Point(
					kml.Coordinates(kml.Coordinate{Lon: hpos.HomeLon, Lat: hpos.HomeLat}),
				),
			)
			wps = append(wps, p)
		} else {
			desc = fmt.Sprintf("Created from %s", m.mission_file)
		}
	}

	pts, rth := m.get_fly_points(addAlt)
	points = append(points, pts...)

	if rth && (hpos.Flags&types.HOME_ALT) == types.HOME_ALT {
		points = append(points, kml.Coordinate{Lon: hpos.HomeLon, Lat: hpos.HomeLat, Alt: float64(addAlt)})
	}

	track := kml.Placemark(
		kml.Description("inav mission"),
		kml.StyleURL("#styleWPTrack"),
		kml.LineString(
			kml.AltitudeMode(altmode),
			kml.Extrude(true),
			kml.Tessellate(false),
			kml.Coordinates(points...),
		),
	)

	track.Add(kml.Visibility(isvis))
	fldnam := fmt.Sprintf("Mission #%d", mmidx)
	kelem := kml.Folder(kml.Name(fldnam)).Add(kml.Description(desc)).
		Add(kml.Visibility(isvis)).Add(mission_styles()...).Add(track).Add(wps...)

	if landid != -1 && m.FWApproach.No == int8(mmidx+7) && m.FWApproach.Dirn1 != 0 && m.FWApproach.Dirn2 != 0 {
		lpath1, lpath2, apath1, apath2 := update_laylines(m.MissionItems[landid].Lat, m.MissionItems[landid].Lon, addAlt, m.FWApproach)

		if len(lpath1) > 0 {
			track := kml.Placemark(
				kml.Description("land path1"),
				kml.StyleURL("#styleFWLand"),
				kml.LineString(
					kml.AltitudeMode(altmode),
					kml.Extrude(false),
					kml.Tessellate(false),
					kml.Coordinates(lpath1...),
				),
			)
			track.Add(kml.Visibility(isvis))
			kelem.Add(track)
		}

		if len(lpath2) > 0 {
			track := kml.Placemark(
				kml.Description("land path2"),
				kml.StyleURL("#styleFWLand"),
				kml.LineString(
					kml.AltitudeMode(altmode),
					kml.Extrude(false),
					kml.Tessellate(false),
					kml.Coordinates(lpath2...),
				),
			)
			track.Add(kml.Visibility(isvis))
			kelem.Add(track)
		}

		if len(apath1) > 0 {
			track := kml.Placemark(
				kml.Description("approach path1"),
				kml.StyleURL("#styleFWApproach"),
				kml.LineString(
					kml.AltitudeMode(altmode),
					kml.Extrude(false),
					kml.Tessellate(false),
					kml.Coordinates(apath1...),
				),
			)
			track.Add(kml.Visibility(isvis))
			kelem.Add(track)
		}

		if len(apath2) > 0 {
			track := kml.Placemark(
				kml.Description("approach path"),
				kml.StyleURL("#styleFWApproach"),
				kml.LineString(
					kml.AltitudeMode(altmode),
					kml.Extrude(false),
					kml.Tessellate(false),
					kml.Coordinates(apath2...),
				),
			)
			track.Add(kml.Visibility(isvis))
			kelem.Add(track)
		}
	}
	return kelem
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
			la, lo = geo.Posit(lat, lon, float64(lnd.Dirn1), LAYLEN)
			p0 = kml.Coordinate{Lon: lo, Lat: la, Alt: float64(lnd.Appalt)}
			lpath1 = append(lpath1, p0)
		}
		p0 = kml.Coordinate{Lon: lon, Lat: lat, Alt: float64(lnd.Landalt)}
		lpath1 = append(lpath1, p0)
		adir := (lnd.Dirn1 + 180) % 360
		la, lo = geo.Posit(lat, lon, float64(adir), LAYLEN)
		p0 = kml.Coordinate{Lon: lo, Lat: la, Alt: float64(lnd.Appalt)}
		lpath1 = append(lpath1, p0)
		apath1 = add_approach(lnd.Dref, int(lnd.Dirn1), lpath1)
	}

	if lnd.Dirn2 != 0 {
		if lnd.Dirn2 < 0 {
			lnd.Dirn2 = -lnd.Dirn2
		} else {
			la, lo = geo.Posit(lat, lon, float64(lnd.Dirn2), LAYLEN)
			p0 = kml.Coordinate{Lon: lo, Lat: la, Alt: float64(lnd.Appalt)}
			lpath2 = append(lpath2, p0)
		}
		p0 = kml.Coordinate{Lon: lon, Lat: lat, Alt: float64(lnd.Landalt)}
		lpath2 = append(lpath2, p0)
		adir := (lnd.Dirn2 + 180) % 360
		la, lo = geo.Posit(lat, lon, float64(adir), LAYLEN)
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

	lax, lox := geo.Posit(lpath[iap].Lat, lpath[iap].Lon, float64(xdir), LAYLEN/3.0)
	apath = append(apath, lpath[iap])
	apath = append(apath, kml.Coordinate{Lon: lox, Lat: lax, Alt: lpath[iap].Alt})
	apath = append(apath, lpath[ilp])
	if len(lpath) == 3 {
		lax, lox = geo.Posit(lpath[0].Lat, lpath[0].Lon, float64(xdir), LAYLEN/3.0)
		apath = append(apath, kml.Coordinate{Lon: lox, Lat: lax, Alt: lpath[iap].Alt})
		apath = append(apath, lpath[0])
	}
	return apath
}

func mission_styles() []kml.Element {
	return []kml.Element{
		kml.SharedStyle(
			"styleSET_POI",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("ylw-diamond")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),
		kml.SharedStyle(
			"styleRTH",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("red-diamond")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),
		kml.SharedStyle(
			"styleSET_HEAD",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("ylw-diamond")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),
		kml.SharedStyle(
			"styleWAYPOINT",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("ltblu-circle")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),
		kml.SharedStyle(
			"stylePOSHOLD_UNLIM",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("grn-diamond")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),
		kml.SharedStyle(
			"stylePOSHOLD_TIME",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("grn-circle")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),
		kml.SharedStyle(
			"styleJUMP",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("purple-circle")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),
		kml.SharedStyle(
			"styleLAND",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("pink-stars")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),
		kml.SharedStyle(
			"styleFakeHome",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("orange-stars")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),
		kml.SharedStyle(
			"styleWPTrack",
			kml.LineStyle(
				kml.Width(2.0),
				kml.Color(color.RGBA{R: 0, G: 0xff, B: 0xff, A: 0x66}),
			),
			kml.PolyStyle(
				kml.Color(color.RGBA{R: 0xc0, G: 0xc0, B: 0xc0, A: 0x66}),
			),
		),
		kml.SharedStyle(
			"styleFWLand",
			kml.LineStyle(
				kml.Width(2.0),
				kml.Color(color.RGBA{R: 0xfc, G: 0xac, B: 0x64, A: 0xa0}),
			),
			kml.PolyStyle(
				kml.Color(color.RGBA{R: 0xfc, G: 0xac, B: 0x64, A: 0}),
			),
		),
		kml.SharedStyle(
			"styleFWApproach",
			kml.LineStyle(
				kml.Width(2.0),
				kml.Color(color.RGBA{R: 0x63, G: 0xa0, B: 0xfc, A: 0xa0}),
			),
			kml.PolyStyle(
				kml.Color(color.RGBA{R: 0x63, G: 0xa0, B: 0xfc, A: 0}),
			),
		),
		kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
			kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
	}
}
