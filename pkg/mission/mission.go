package mission

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"os"
	"github.com/beevik/etree"
	"encoding/json"
	kml "github.com/twpayne/go-kml"
	"github.com/twpayne/go-kml/icon"
	"image/color"
	geo "github.com/stronnag/bbl2kml/pkg/geo"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	"path/filepath"
)

type QGCrec struct {
	jindex  int
	command int
	lat     float64
	lon     float64
	alt     float64
	params  [4]float64
}

type MissionItem struct {
	No     int
	Action string
	Lat    float64
	Lon    float64
	Alt    int32
	P1     int16
	P2     int16
	P3     uint16
}

const (
	wp_WAYPOINT = 1 + iota
	wp_POSHOLD_UNLIM
	wp_POSHOLD_TIME
	wp_RTH
	wp_SET_POI
	wp_JUMP
	wp_SET_HEAD
	wp_LAND
)

type Mission struct {
	mission_file string
	Version      string
	MissionItems []MissionItem
}


func read_gpx(dat []byte) *Mission {
	items := []MissionItem{}
	mission := &Mission{"", "0.0", items}
	stypes := []string{"//trkpt", "//rtept", "//wpt"}

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(dat); err == nil {
		root := doc.SelectElement("gpx")
		for _, stype := range stypes {
			for k, pts := range root.FindElements(stype) {
				alt := 0.0
				lat, _ := strconv.ParseFloat(pts.SelectAttrValue("lat", "0"), 64)
				lon, _ := strconv.ParseFloat(pts.SelectAttrValue("lon", "0"), 64)
				item := MissionItem{No: k + 1, Lat: lat, Lon: lon, Alt: int32(alt), Action: "WAYPOINT"}
				mission.MissionItems = append(mission.MissionItems, item)
			}
			break
		}
	}
	return mission
}

func decode_action(b byte) string {
	var a string
	switch b {
	case wp_WAYPOINT:
		a = "WAYPOINT"
	case wp_POSHOLD_UNLIM:
		a = "POSHOLD_UNLIM"
	case wp_POSHOLD_TIME:
		a = "POSHOLD_TIME"
	case wp_RTH:
		a = "RTH"
	case wp_SET_POI:
		a = "SET_POI"
	case wp_JUMP:
		a = "JUMP"
	case wp_SET_HEAD:
		a = "SET_HEAD"
	case wp_LAND:
		a = "LAND"
	default:
		a = "UNKNOWN"
	}
	return a
}

func (m *Mission) is_valid() bool {
	force := os.Getenv("IMPLOAD_NO_VERIFY")
	if len(force) > 0 {
		return true
	}
	mlen := int16(len(m.MissionItems))
	if mlen > 60 {
		return false
	}
	// Urg, Urg array index v. WP Nos ......
	for i := int16(0); i < mlen; i++ {
		var target = m.MissionItems[i].P1 - 1
		if m.MissionItems[i].Action == "JUMP" {
			if (i == 0) || ((target > (i - 2)) && (target < (i + 2))) || (target >= mlen) || (m.MissionItems[i].P2 < -1) {
				return false
			}
			if !(m.MissionItems[target].Action == "WAYPOINT" || m.MissionItems[target].Action == "POSHOLD_TIME" || m.MissionItems[target].Action == "LAND") {
				return false
			}
		}
	}
	return true
}

func (mi *MissionItem) Is_GeoPoint() bool {
	a := mi.Action
	return !(a == "RTH" || a == "SET_HEAD" || a == "JUMP")
}

func (m *Mission) Dump(dms bool, homep...float64)  {
	var hpos types.HomeRec
	hpos.HomeLat = homep[0]
	hpos.HomeLon = homep[1]
	hpos.Flags = types.HOME_ARM
	if len(homep) > 2 {
		hpos.HomeAlt = homep[2]
		hpos.Flags |= types.HOME_ALT
	}
	k := kml.KML(m.To_kml(hpos, dms, true))
	k.WriteIndent(os.Stdout, "", "  ")
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
			alt = mi.Alt + addAlt
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

		if  mi.Action != "SET_POI" && mi.Action != "JUMP" && mi.Action != "SET_HEAD" &&
			mi.Action != "RTH" {
			pt := kml.Coordinate{Lon: mi.Lon, Lat: mi.Lat, Alt: float64(alt)}
			points = append(points, pt)
		}
	}

	var desc string
	if (hpos.Flags & types.HOME_ALT) == types.HOME_ALT {
		desc = fmt.Sprintf("Created from %s with elevations adjusted for home location %s",
			m.mission_file, geo.PositionFormat(hpos.HomeLat, hpos.HomeLon, dms))
		points = append(points, kml.Coordinate{Lon: hpos.HomeLon, Lat: hpos.HomeLat,Alt: float64(addAlt)})
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

func read_simple(dat []byte) *Mission {
	r := csv.NewReader(strings.NewReader(string(dat)))

	items := []MissionItem{}
	mission := &Mission{"", "0.0", items}

	n := 1
	has_no := false

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if record[0] == "no" {
			has_no = true
			continue
		}

		if record[0] == "wp" {
			continue
		}

		var lat, lon float64

		j := 0
		no := n
		if has_no {
			no, _ = strconv.Atoi(record[0])
			j = 1
		}

		p1 := int16(0)
		p2 := int16(0)
		fp2 := 0.0
		lat, _ = strconv.ParseFloat(record[j+1], 64)
		lon, _ = strconv.ParseFloat(record[j+2], 64)
		alt, _ := strconv.ParseFloat(record[j+3], 64)
		fp1, _ := strconv.ParseFloat(record[j+4], 64)
		if len(record) > j+5 {
			fp2, _ = strconv.ParseFloat(record[j+5], 64)
		}
		var action string

		iaction, err := strconv.Atoi(record[j])
		if err == nil {
			action = decode_action(byte(iaction))
		} else {
			action = record[j]
		}
		switch action {
		case "RTH":
			lat = 0.0
			lon = 0.0
			alt = 0
			if fp1 != 0 {
				p1 = 1
			}
		case "WAYPOINT", "WP":
			action = "WAYPOINT"
			if fp1 > 0 {
				p1 = int16(fp1 * 100)
			}
		case "POSHOLD_TIME":
			if fp2 > 0 {
				p2 = int16(fp2 * 100)
			}
			p1 = int16(fp1)
		case "JUMP":
			lat = 0.0
			lon = 0.0
			p1 = int16(fp1)
			p2 = int16(fp2)
		case "LAND":
			if fp1 > 0 {
				p1 = int16(fp1 * 100)
			}
		default:
			continue
		}
		item := MissionItem{No: no, Lat: lat, Lon: lon, Alt: int32(alt), Action: action, P1: p1, P2: p2}
		mission.MissionItems = append(mission.MissionItems, item)
		n++
	}
	return mission
}

func read_qgc_json(dat []byte) []QGCrec {
	qgcs := []QGCrec{}
	var result map[string]interface{}

	json.Unmarshal(dat, &result)
	mi := result["mission"].(interface{})
	mid := mi.(map[string]interface{})
	it := mid["items"].([]interface{})

	for _, l := range it {
		ll := l.(map[string]interface{})
		ps := ll["params"].([]interface{})
		qg := QGCrec{}
		qg.jindex = int(ll["doJumpId"].(float64))
		qg.command = int(ll["command"].(float64))
		qg.lat = ps[4].(float64)
		qg.lon = ps[5].(float64)
		qg.alt = ps[6].(float64)
		for j := 0; j < 4; j++ {
			if ps[j] != nil {
				qg.params[j] = ps[j].(float64)
			}
		}
		qgcs = append(qgcs, qg)
	}
	return qgcs
}

func read_qgc_text(dat []byte) []QGCrec {
	qgcs := []QGCrec{}

	r := csv.NewReader(strings.NewReader(string(dat)))
	r.Comma = '\t'
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err == nil {
		for _, record := range records {
			if len(record) == 12 {
				no, err := strconv.Atoi(record[0])
				if err == nil && no > 0 {
					qg := QGCrec{}
					qg.jindex = no
					qg.command, _ = strconv.Atoi(record[3])
					qg.alt, _ = strconv.ParseFloat(record[10], 64)
					qg.lat, _ = strconv.ParseFloat(record[8], 64)
					qg.lon, _ = strconv.ParseFloat(record[9], 64)
					for j := 0; j < 4; j++ {
						qg.params[j], _ = strconv.ParseFloat(record[4+j], 64)
					}
					qgcs = append(qgcs, qg)
				}
			}
		}
	} else {
		log.Fatal(err)
	}
	return qgcs
}

func fixup_qgc_mission(mission *Mission, have_jump bool) (*Mission, bool) {
	ok := true
	if have_jump {
		for i := 0; i < len(mission.MissionItems); i++ {
			if mission.MissionItems[i].Action == "JUMP" {
				jumptgt := mission.MissionItems[i].P1
				ajump := int16(0)
				for j := 0; j < len(mission.MissionItems); j++ {
					if mission.MissionItems[j].P3 == uint16(jumptgt) {
						ajump = int16(j + 1)
						break
					}
				}
				if ajump == 0 {
					ok = false
				} else {
					mission.MissionItems[i].P1 = ajump
				}
				no := int16(i + 1) // item index
				if mission.MissionItems[i].P1 < 1 || ((mission.MissionItems[i].P1 > no-2) &&
					(mission.MissionItems[i].P1 < no+2)) {
					ok = false
				}
			}
		}
	}
	if ok {
		for i := 0; i < len(mission.MissionItems); i++ {
			mission.MissionItems[i].P3 = 0
		}
		return mission, ok
	} else {
		return nil, false
	}
}

func process_qgc(dat []byte, mtype string) *Mission {
	var qs []QGCrec
	items := []MissionItem{}
	mission := &Mission{"", "0.0", items}

	if mtype == "qgc-text" {
		qs = read_qgc_text(dat)
	} else {
		qs = read_qgc_json(dat)
	}
	last_alt := 0.0
	last_lat := 0.0
	last_lon := 0.0

	have_land := false
	lastj := -1

	for j, rq := range qs {
		if rq.command == 20 {
			lastj = j
		} else if rq.command == 21 && j == lastj+1 {
			have_land = true
		}
	}

	last := false
	have_jump := false

	no := 0
	for _, q := range qs {
		ok := true
		var action string
		var p1, p2 int16

		switch q.command {
		case 16:
			if q.params[0] == 0 {
				action = "WAYPOINT"
				p1 = 0
			} else {
				action = "POSHOLD_TIME"
				p1 = int16(q.params[0])
			}

		case 19:
			action = "POSHOLD_TIME"
			p1 = int16(q.params[0])
			if q.alt == 0 {
				q.alt = last_alt
			}
			if q.lat == 0.0 {
				q.lat = last_lat
			}
			if q.lon == 0.0 {
				q.lon = last_lon
			}
		case 20:
			action = "RTH"
			q.lat = 0.0
			q.lon = 0.0
			if q.alt == 0 || have_land {
				p1 = 1
			}
			q.alt = 0
			last = true

		case 21:
			action = "LAND"
			p1 = 0
			if q.alt == 0 {
				q.alt = last_alt
			}
			if q.lat == 0.0 {
				q.lat = last_lat
			}
			if q.lon == 0.0 {
				q.lon = last_lon
			}
		case 177:
			p1 = int16(q.params[0])
			action = "JUMP"
			p2 = int16(q.params[1])
			q.lat = 0.0
			q.lon = 0.0
			have_jump = true

		case 195, 201:
			action = "SET_POI"

		case 115:
			p1 = int16(q.params[0])
			act := int(q.params[3])
			if p1 == 0 && act == 0 {
				p1 = -1
			}
			action = "SET_HEAD"
			q.lat = 0
			q.lon = 0
			q.alt = 0

		case 197:
			p1 = -1
			action = "SET_HEAD"
			q.lat = 0
			q.lon = 0
			q.alt = 0

		default:
			ok = false
		}
		if ok {
			last_alt = q.alt
			last_lat = q.lat
			last_lon = q.lon
			p3 := uint16(q.jindex)
			no += 1
			item := MissionItem{No: no, Lat: q.lat, Lon: q.lon, Alt: int32(q.alt), Action: action, P1: p1, P2: p2, P3: p3}
			mission.MissionItems = append(mission.MissionItems, item)
			if last {
				break
			}
		}
	}

	mission, ok := fixup_qgc_mission(mission, have_jump)
	if !ok {
		log.Fatalf("Unsupported QGC file\n")
	}
	return mission
}

func read_xml_mission(dat []byte) *Mission {
	items := []MissionItem{}
	mission := &Mission{"", "0.0", items}
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(dat); err == nil {
		for _, root := range doc.ChildElements() {
			if strings.EqualFold(root.Tag, "MISSION") {
				for _, el := range root.ChildElements() {
					switch {
					case strings.EqualFold(el.Tag, "VERSION"):
						version := el.SelectAttrValue("value", "")
						if version != "" {
							mission.Version = version
						}
					case strings.EqualFold(el.Tag, "MISSIONITEM"):
						no, _ := strconv.Atoi(el.SelectAttrValue("no", "0"))
						action := el.SelectAttrValue("action", "WAYPOINT")
						lat, _ := strconv.ParseFloat(el.SelectAttrValue("lat", "0"), 64)
						lon, _ := strconv.ParseFloat(el.SelectAttrValue("lon", "0"), 64)
						alt, _ := strconv.Atoi(el.SelectAttrValue("alt", "0"))
						p1, _ := strconv.Atoi(el.SelectAttrValue("parameter1", "0"))
						p2, _ := strconv.Atoi(el.SelectAttrValue("parameter2", "0"))
						p3, _ := strconv.Atoi(el.SelectAttrValue("parameter3", "0"))
						item := MissionItem{no, action, lat, lon, int32(alt), int16(p1), int16(p2), uint16(p3)}
						mission.MissionItems = append(mission.MissionItems, item)
					default:
						// fmt.Printf("ignoring tag %s\n", el.Tag)
					}
				}
			}
		}
	}
	return mission
}

func read_json(dat []byte) *Mission {
	items := []MissionItem{}
	mission := &Mission{"", "0.0", items}
	var result map[string]interface{}
	json.Unmarshal(dat, &result)
	mi := result["mission"].([]interface{})
	for _, l := range mi {
		ll := l.(map[string]interface{})
		item := MissionItem{int(ll["no"].(float64)), ll["action"].(string),
			ll["lat"].(float64), ll["lon"].(float64),
			int32(ll["alt"].(float64)), int16(ll["p1"].(float64)),
			int16(ll["p2"].(float64)), uint16(ll["p3"].(float64))}
		mission.MissionItems = append(mission.MissionItems, item)
	}
	return mission
}

func Read_Mission_File(path string) (string, *Mission, error) {
	var dat []byte
	r, err := os.Open(path)
	if err == nil {
		defer r.Close()
		dat, err = ioutil.ReadAll(r)
	}
	if err != nil {
		return "?", nil, err
	} else {
		mtype, m := handle_mission_data(dat, path)
		m.mission_file = 	filepath.Base(path)
		return mtype, m, nil
	}
}

func handle_mission_data(dat []byte, path string) (string, *Mission) {
	var m *Mission
	mtype := ""
	switch {
	case bytes.HasPrefix(dat, []byte("<?xml")):
		switch {
		case bytes.Contains(dat, []byte("<MISSION")),
			bytes.Contains(dat, []byte("<mission")):
			m = read_xml_mission(dat)
			mtype = "mwx"
		case bytes.Contains(dat, []byte("<gpx ")):
			m = read_gpx(dat)
			mtype = "gpx"
		default:
			m = nil
		}
	case bytes.HasPrefix(dat, []byte("QGC WPL 110")):
		mtype = "qgc-text"
		m = process_qgc(dat, mtype)
	case bytes.HasPrefix(dat, []byte("no,wp,lat,lon,alt,p1")),
		bytes.HasPrefix(dat, []byte("wp,lat,lon,alt,p1")):
		m = read_simple(dat)
		mtype = "csv"
	case bytes.HasPrefix(dat, []byte("{\"meta\":{")):
		mtype = "mwp-json"
		m = read_json(dat)
	case bytes.Contains(dat[0:100], []byte("\"fileType\": \"Plan\"")):
		mtype = "qgc-json"
		m = process_qgc(dat, mtype)
	default:
		m = nil
	}
	return mtype, m
}
