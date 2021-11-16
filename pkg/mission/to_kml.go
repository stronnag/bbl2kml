package mission

import (
	"fmt"
	kml "github.com/twpayne/go-kml"
	"github.com/twpayne/go-kml/icon"
	"image/color"
	geo "github.com/stronnag/bbl2kml/pkg/geo"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
)

func (m *Mission) get_fly_points (addAlt int32) ([]kml.Coordinate, bool) {
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

func (m *Mission) To_kml(hpos types.HomeRec, dms bool, fake bool) kml.Element {
	var points []kml.Coordinate
	var wps  []kml.Element
	llat := 0.0
	llon := 0.0
	lalt := int32(0)
	var altmode kml.AltitudeModeEnum

	addAlt := int32(0)

	if (hpos.Flags & types.HOME_ARM) != 0 {
		if (hpos.Flags & types.HOME_ALT) == types.HOME_ALT {
			addAlt = int32(hpos.HomeAlt)
		} else {
			bingelev, err :=  geo.GetElevation(hpos.HomeLat, hpos.HomeLon)
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

	for _, mi := range m.MissionItems {
		if  mi.Action == "JUMP" || mi.Action == "SET_HEAD" || mi.Action == "RTH" {
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
		name:= fmt.Sprintf("%s %d", bname, mi.No)
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

	if rth 	&& (hpos.Flags & types.HOME_ALT) == types.HOME_ALT {
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

	return kml.Folder(kml.Name("Mission File")).Add(kml.Description(desc)).
		Add(kml.Visibility(true)).Add(mission_styles()...).Add(track).Add(wps...)
}

func mission_styles() []kml.Element {
	return []kml.Element{
		kml.SharedStyle(
			"styleSET_POI",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("ylw-diamond"),
					),
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
					kml.Href(icon.PaddleHref("red-diamond"),
					),
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
					kml.Href(icon.PaddleHref("ylw-diamond"),
					),
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
					kml.Href(icon.PaddleHref("ltblu-circle"),
					),
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
					kml.Href(icon.PaddleHref("grn-diamond"),
					),
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
					kml.Href(icon.PaddleHref("grn-circle"),
					),
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
					kml.Href(icon.PaddleHref("purple-circle"),
					),
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
					kml.Href(icon.PaddleHref("pink-stars"),
					),
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
					kml.Href(icon.PaddleHref("orange-stars"),
					),
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
		kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
			kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
	}
}
