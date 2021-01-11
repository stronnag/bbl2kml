package kmlgen

import (
	"image/color"
	"log"
	"os"
	"fmt"
	"strings"
	"github.com/bmizerany/perks/quantile"
	kml "github.com/twpayne/go-kml"
	kmz "github.com/twpayne/go-kmz"
	"github.com/twpayne/go-kml/icon"
	geo "github.com/stronnag/bbl2kml/pkg/geo"
	mission "github.com/stronnag/bbl2kml/pkg/mission"
	options "github.com/stronnag/bbl2kml/pkg/options"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
)

const (
	BS_NAME_DESC = iota
	BS_DESC_ONLY
	COL_STYLE_MODE
	COL_STYLE_RSSI
	COL_STYLE_EFFIC
)

func getflightColour(mode uint8) color.Color {
	var c color.Color
	switch mode {
	case types.FM_LAUNCH:
		c = color.RGBA{R: 0, G: 160, B: 160, A: 0xa0}
	case types.FM_RTH:
		c = color.RGBA{R: 0xff, G: 0xff, B: 0, A: 0xa0}
	case types.FM_WP:
		c = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xa0}
	case types.FM_CRUISE3D, types.FM_CRUISE2D:
		c = color.RGBA{R: 0xbf, G: 0x88, B: 0xf0, A: 0xa0}
	case types.FM_PH:
		c = color.RGBA{R: 0xce, G: 0xff, B: 0x9d, A: 0xa0}
	case types.FM_AH:
		c = color.RGBA{R: 0x3, G: 0xc0, B: 0xf0, A: 0xa0}
	case types.FM_FS:
		c = color.RGBA{R: 0xff, G: 0, B: 0, A: 0xa0}
	case types.FM_EMERG:
		c = color.RGBA{R: 0xff, G: 0x80, B: 0, A: 0xa0}
	default:
		c = color.RGBA{R: 0, G: 0xff, B: 0xff, A: 0xa0}
	}
	return c
}

func getStyleURL(r types.LogItem, colmode uint8) string {
	var s string
	if colmode == COL_STYLE_RSSI {
		s = fmt.Sprintf("#styleGrad%03d", 5*(r.Rssi/5))
	} else if colmode == COL_STYLE_EFFIC {
		s= fmt.Sprintf("#styleGrad%03d", 5*(int(r.Qval)/5))
	} else {
		switch r.Fmode {
		case types.FM_LAUNCH:
			s = "#styleLaunch"
		case types.FM_RTH:
			s = "#styleRTH"
		case types.FM_WP:
			s = "#styleWP"
		case types.FM_CRUISE3D, types.FM_CRUISE2D:
			s = "#styleCRS"
		case types.FM_PH:
		s = "#stylePH"
		case types.FM_EMERG:
			s = "#styleEMERG"
		default:
			s = "#styleNormal"
		}
	}
	return s
}

func getPoints(rec types.LogRec, hpos types.HomeRec, colmode uint8, viz bool) []kml.Element {
	var pt []kml.Element
	var qval0,qval1 float64
	if colmode == COL_STYLE_EFFIC {
		q := quantile.NewTargeted(0.05, 0.95)
		for _, r := range rec.Items {
			q.Insert(r.Effic)
		}
		qval0 = q.Query(0.05)
		qval1 = q.Query(0.95)
	}

	tpts := len(rec.Items)
	for np, r := range rec.Items {
		tfmt := r.Utc.Format("2006‑01‑02T15:04:05.99MST")
		fmtxt := r.Fmtext
		if r.Fs {
			fmtxt = fmtxt + " FAILSAFE"
		}
		var alt float64
		var altmode kml.AltitudeModeEnum
		if (hpos.Flags & types.HOME_ALT) == types.HOME_ALT {
			alt = hpos.HomeAlt + r.Alt
			altmode = kml.AltitudeModeAbsolute
		} else {
			alt = r.Alt
			altmode = kml.AltitudeModeRelativeToGround
		}
		if colmode == COL_STYLE_EFFIC {
			if r.Effic > qval1 {
				r.Qval = 100
			} else if r.Effic < qval0 {
				r.Qval = 0
			} else {
				r.Qval = 100*(1-(r.Effic-qval0)/(qval1-qval0))
			}
		}

		var sb strings.Builder
		sb.Write([]byte(fmt.Sprintf("<h3>Track Point %d of %d</h3>", np, tpts)))

		sb.Write([]byte(`<table style="border="1px" silver; border="1" silver; rules="all";;">`))

		sb.Write([]byte(fmt.Sprintf("<tr><td><b>%s</b></td><td>%s</td></tr>","Time", tfmt)))
		sb.Write([]byte(fmt.Sprintf("<tr><td><b>%s</b></td><td>%s</td></tr>", "Position", geo.PositionFormat(r.Lat, r.Lon, options.Dms))))
		sb.Write([]byte(fmt.Sprintf("<tr><td><b>%s</b></td><td>%.0f m</td></tr>", "Elevation", r.Alt)))
		sb.Write([]byte(fmt.Sprintf("<tr><td><b>%s</b></td><td>%.0f m</td></tr>", "GPS Altitude", alt)))
		sb.Write([]byte(fmt.Sprintf("<tr><td><b>%s</b></td><td>%d°</td></tr>", "Course", r.Cse)))
		sb.Write([]byte(fmt.Sprintf("<tr><td><b>%s</b></td><td>%.1f m/s</td></tr>", "Speed", r.Spd)))
		sb.Write([]byte(fmt.Sprintf("<tr><td><b>%s</b></td><td>%d</td></tr>", "Satellites", r.Numsat)))

		sb.Write([]byte(fmt.Sprintf("<tr><td><b>%s</b></td><td>%.0f m</td></tr>", "Range", r.Vrange)))
		sb.Write([]byte(fmt.Sprintf("<tr><td><b>%s</b></td><td>%d°</td></tr>", "Bearing", r.Bearing)))
		sb.Write([]byte(fmt.Sprintf("<tr><td><b>%s</b></td><td>%d %%</td></tr>", "RSSI", r.Rssi)))
		sb.Write([]byte(fmt.Sprintf("<tr><td><b>%s</b></td><td>%s</td></tr>", "Mode", fmtxt)))
		sb.Write([]byte(fmt.Sprintf("<tr><td><b>%s</b></td><td>%.0f m</td></tr>", "Cumulative Distance", r.Tdist)))
		if r.Volts > 0 {
			sb.Write([]byte(fmt.Sprintf("<tr><td><b>%s</b></td><td>%.1f V</br>", "Voltage",  r.Volts)))
		}
		if (rec.Cap & types.CAP_AMPS) == types.CAP_AMPS {
			sb.Write([]byte(fmt.Sprintf("<tr><td><b>%s</b></td><td>%.1f A</td></tr>", "Current", r.Amps)))
			if (rec.Cap & types.CAP_ENERGY) == types.CAP_ENERGY {
				sb.Write([]byte(fmt.Sprintf("<tr><td><b>%s</b></td><td>%.1f mah</td></tr>", "Total Energy", r.Energy)))
				ceav := r.Energy*1000/r.Tdist
				sb.Write([]byte(fmt.Sprintf("<tr><td><b>%s</b></td><td>%.1f mah / km</td></tr>", "Efficiency",r.Effic)))
				sb.Write([]byte(fmt.Sprintf("<tr><td><b>%s</b></td><td>%.1f mah / km</td></tr>", "Average Efficiency",ceav)))
			}
		}
		sb.Write([]byte("</table>"))

		po := kml.Point(
				kml.AltitudeMode(altmode),
				kml.Coordinates(kml.Coordinate{Lon: r.Lon, Lat: r.Lat, Alt: alt}),
			)

		k := kml.Placemark(
			kml.Visibility(viz),
			kml.Description(sb.String()),
			kml.TimeStamp(kml.When(r.Utc)),
			kml.StyleURL(getStyleURL(r, colmode)),
		)

		se:= kml.Style()

		if options.Extrude {
			po.Add (
				kml.Extrude(true),
				kml.Tessellate(false),
			)
			se.Add(
				kml.LineStyle(
					kml.Width(2),
					kml.Color(color.RGBA{R: 0xc0, G: 0xc0, B: 0xc0, A: 0x66}),
				),
			)
		}
		if r.Fs {
			se.Add(
				kml.IconStyle(
					kml.Icon(
						kml.Href(icon.PaddleHref("wht-circle-lv")),
					),
				),
			)
		}

		if options.Extrude || r.Fs {
			k.Add(se)
		}
		k.Add(po)
		pt = append(pt, k)
	}
	return pt
}

func getHomes(hpos types.HomeRec) []kml.Element {
	var htext, hdesc string

	if (hpos.Flags & types.HOME_SAFE) == types.HOME_SAFE {
		htext = "Armed"
	} else {
		htext = "Home"
	}
	hdesc = fmt.Sprintf("Location %s<br/>",
		geo.PositionFormat(hpos.HomeLat, hpos.HomeLon, options.Dms))
	if (hpos.Flags & types.HOME_ALT) == types.HOME_ALT {
		hdesc = hdesc + fmt.Sprintf("GPS Altitude: %.0fm<br/>", hpos.HomeAlt)
	}
	var hp []kml.Element
	k := kml.Placemark(
		kml.Name(htext),
		kml.Description(hdesc),
		kml.Point(
			kml.Coordinates(kml.Coordinate{Lon: hpos.HomeLon, Lat: hpos.HomeLat}),
		),
		kml.Style(
			kml.IconStyle(
				kml.Icon(
					kml.Href(icon.PaletteHref(4, 29)),
				),
			),
		).Add(balloon_style(BS_NAME_DESC)),
	)
	hp = append(hp, k)
	if (hpos.Flags & types.HOME_SAFE) == types.HOME_SAFE {
		k = kml.Placemark(
			kml.Name("Home"),
			kml.Description(fmt.Sprintf("Location %s<br/>",
				geo.PositionFormat(hpos.SafeLat, hpos.SafeLon, options.Dms))),
			kml.Point(
				kml.Coordinates(kml.Coordinate{Lon: hpos.SafeLon, Lat: hpos.SafeLat}),
			),
			kml.Style(
				kml.IconStyle(
					kml.Icon(
						kml.Href(icon.PaletteHref(3, 56)),
					),
				),
			).Add(balloon_style(BS_NAME_DESC)),
		)
		hp = append(hp, k)
	}
	return hp
}

func balloon_style(bs uint8) *kml.CompoundElement {
	if bs == BS_NAME_DESC {
		return kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
		kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`))
	} else {
		return kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
			kml.Text(`$[description]`))
	}
}

func generate_shared_styles(style uint8) []kml.Element {
	switch style {
	default:
		return []kml.Element{
			kml.SharedStyle(
				"styleNormal",
				kml.IconStyle(
					kml.Scale(0.5),
					kml.Color(getflightColour(types.FM_ACRO)),
					kml.Icon(
						kml.Href(icon.PaletteHref(2, 18)),
					),
				),
			).Add(balloon_style(BS_DESC_ONLY)),
			kml.SharedStyle(
				"styleLaunch",
				kml.IconStyle(
					kml.Scale(0.5),
					kml.Color(getflightColour(types.FM_LAUNCH)),
					kml.Icon(
						kml.Href(icon.PaletteHref(2, 18)),
					),
				),
			).Add(balloon_style(BS_DESC_ONLY)),
			kml.SharedStyle(
				"styleWP",
				kml.IconStyle(
					kml.Scale(0.5),
					kml.Color(getflightColour(types.FM_WP)),
					kml.Icon(
						kml.Href(icon.PaletteHref(2, 18)),
					),
				),
			).Add(balloon_style(BS_DESC_ONLY)),
			kml.SharedStyle(
				"styleRTH",
				kml.IconStyle(
					kml.Scale(0.5),
					kml.Color(getflightColour(types.FM_RTH)),
					kml.Icon(
						kml.Href(icon.PaletteHref(2, 18)),
					),
				),
			).Add(balloon_style(BS_DESC_ONLY)),
			kml.SharedStyle(
				"styleCRS",
				kml.IconStyle(
					kml.Scale(0.5),
					kml.Color(getflightColour(types.FM_CRUISE3D)),
					kml.Icon(
						kml.Href(icon.PaletteHref(2, 18)),
					),
				),
			).Add(balloon_style(BS_DESC_ONLY)),
			kml.SharedStyle(
				"stylePH",
				kml.IconStyle(
					kml.Scale(0.5),
					kml.Color(getflightColour(types.FM_PH)),
					kml.Icon(
						kml.Href(icon.PaletteHref(2, 18)),
					),
				),
			).Add(balloon_style(BS_DESC_ONLY)),
			kml.SharedStyle(
				"styleAH",
				kml.IconStyle(
					kml.Scale(0.5),
					kml.Color(getflightColour(types.FM_AH)),
					kml.Icon(
						kml.Href(icon.PaletteHref(2, 18)),
					),
				),
			).Add(balloon_style(BS_DESC_ONLY)),
			kml.SharedStyle(
				"styleFS",
				kml.IconStyle(
					kml.Scale(0.5),
					kml.Color(getflightColour(types.FM_FS)),
					kml.Icon(
						kml.Href(icon.PaletteHref(2, 18)),
					),
				),
			).Add(balloon_style(BS_DESC_ONLY)),
			kml.SharedStyle(
				"styleEMERG",
				kml.IconStyle(
					kml.Scale(0.5),
					kml.Color(getflightColour(types.FM_EMERG)),
					kml.Icon(
						kml.Href(icon.PaletteHref(2, 18)),
					),
				),
			).Add(balloon_style(BS_DESC_ONLY)),
		}
	case COL_STYLE_RSSI:
		{
			gidx := 0

			switch options.Gradset {
			case "rdgn":
				gidx = GRAD_RGN
			case "yor":
				gidx = GRAD_YOR
			default:
				gidx = GRAD_RED
			}
			gcols := Get_gradset(gidx)
			icons := []kml.Element{}
			for j,c := range gcols {
				sname := fmt.Sprintf("styleGrad%03d", j*5)
				el := kml.SharedStyle(
					sname,
					kml.IconStyle(
						kml.Scale(0.5),
						kml.Color(color.RGBA{R: c.R, G: c.G, B: c.B, A: c.A}),
						kml.Icon(
							kml.Href(icon.PaletteHref(2, 18)),
						),
					),
				).Add(balloon_style(BS_DESC_ONLY))
				icons = append(icons, el)
			}
			return icons
		}
	}
}

func add_ground_track (rec types.LogRec) kml.Element {
	f := kml.Folder(kml.Name("Ground Track")).Add(kml.Visibility(true))
	var points []kml.Coordinate

	for _, r := range rec.Items {
		points = append(points, kml.Coordinate{Lon: r.Lon, Lat: r.Lat})
	}

	tk := kml.Placemark(
		kml.Style(
			kml.LineStyle(
				kml.Width(4.0),
				kml.Color(color.RGBA{R: 0xd0, G: 0xd0, B: 0xd0, A: 0x66}),
			),
		),
		kml.LineString(kml.Coordinates(points...),),
	)
	f.Add(tk)
	return f
}

func GenerateKML(hpos types.HomeRec, rec types.LogRec, outfn string,
	meta types.FlightMeta, smap types.MapRec) {

	defviz := !(options.Rssi && rec.Items[0].Rssi > 0)
	ts0 := rec.Items[0].Utc
	ts1 := rec.Items[len(rec.Items)-1].Utc

	f0 := kml.Folder(kml.Name("Flight modes")).Add(kml.Visibility(defviz)).
		Add(generate_shared_styles(0)...).
		Add(getPoints(rec, hpos, COL_STYLE_MODE, defviz)...)

	d := kml.Folder(kml.Name(meta.LogName())).Add(kml.Open(true))
	d.Add(add_ground_track(rec))

	if len(options.Mission) > 0 {
		 _, ms, err := mission.Read_Mission_File(options.Mission)
		if err == nil {
			if geo.Getfrobnication() {
				for k,mi := range ms.MissionItems {
					if mi.Is_GeoPoint() {
						ms.MissionItems[k].Lat, ms.MissionItems[k].Lon, _ = geo.Frobnicate_move(ms.MissionItems[k].Lat, ms.MissionItems[k].Lon, 0)
					}
				}
			}
			mf := ms.To_kml(hpos, options.Dms, false)
			d.Add(mf)
		} else {
			fmt.Fprintf(os.Stderr,"* Failed to read mission file %s\n", options.Mission)
		}
	}

	e := kml.ExtendedData(kml.Data(kml.Name("Log"), kml.Value(meta.LogName())))

	for k,v := range meta.Summary() {
		e.Add(kml.Data(kml.Name(k), kml.Value(v)))
	}
	for k,v := range smap {
		e.Add(kml.Data(kml.Name(k), kml.Value(v)))
	}
	if s,ok := meta.ShowDisarm(); ok {
		e.Add(kml.Data(kml.Name("Disarm"), kml.Value(s)))
	}
	d.Add(e)

	d.Add(kml.TimeSpan(kml.Begin(ts0), kml.End(ts1)))
	d.Add(getHomes(hpos)...)
	d.Add(f0)
	if rec.Cap&types.CAP_RSSI_VALID !=0 || options.Efficiency {
		d.Add(generate_shared_styles(COL_STYLE_RSSI)...)
	}

	if rec.Cap&types.CAP_RSSI_VALID !=0  {
		f1 := kml.Folder(kml.Name("RSSI")).Add(kml.Visibility(!defviz)).
			Add(getPoints(rec,hpos,COL_STYLE_RSSI,!defviz)...)
		d.Add(f1)
	}
	if options.Efficiency {
		f1 := kml.Folder(kml.Name("Efficiency")).Add(kml.Visibility(false)).
			Add(getPoints(rec,hpos,COL_STYLE_EFFIC,!defviz)...)
		d.Add(f1)
	}

	var err error

	if strings.HasSuffix(outfn, ".kmz") {
		z := kmz.NewKMZ(d)
		w, err := os.Create(outfn)
		if err == nil {
			err = z.WriteIndent(w, "", "  ")
		}
	} else {
		k := kml.KML(d)
		output, err := os.Create(outfn)
		if err == nil {
			err = k.WriteIndent(output, "", "  ")
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}
