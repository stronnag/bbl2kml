package mission

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

import (
	"options"
	"types"
)

import (
	kml "github.com/twpayne/go-kml"
)

type QGCrec struct {
	jindex  int
	command int
	altmode int
	lat     float64
	lon     float64
	alt     float64
	params  [4]float64
}

type qgc_plan struct {
	Filetype string `json:"fileType"`
	Mission  struct {
		Items []struct {
			Typ          string    `json:"type"`
			Altitude     int       `json:"Altitude"`
			Altitudemode int       `json:"AltitudeMode"`
			Command      int       `json:"command"`
			Jumpid       int       `json:"doJumpId"`
			Frame        int       `json:"frame"`
			Params       []float64 `json:"params"`
			Transect     struct {
				Items []struct {
					Typ         string    `json:"type"`
					Altitude    int       `json:"Altitude"`
					Alitudemode int       `json:"AltitudeMode"`
					Command     int       `json:"command"`
					Jumpid      int       `json:"doJumpId"`
					Frame       int       `json:"frame"`
					Params      []float64 `json:"params"`
				} `json:"items"`
			} `json:"TransectStyleComplexItem,omitempty"`
		} `json:"items"`
	} `json:"mission"`
}

type MissionItem struct {
	No     int     `xml:"no,attr" json:"no"`
	Action string  `xml:"action,attr" json:"action"`
	Lat    float64 `xml:"lat,attr" json:"lat"`
	Lon    float64 `xml:"lon,attr" json:"lon"`
	Alt    int32   `xml:"alt,attr" json:"alt"`
	P1     int16   `xml:"parameter1,attr" json:"p1"`
	P2     int16   `xml:"parameter2,attr" json:"p2"`
	P3     int16   `xml:"parameter3,attr" json:"p3"`
	Flag   uint8   `xml:"flag,attr,omitempty" json:"flag,omitempty"`
}

type MissionMWP struct {
	Zoom      int     `xml:"zoom,attr" json:"zoom"`
	Cx        float64 `xml:"cx,attr" json:"cx"`
	Cy        float64 `xml:"cy,attr" json:"cy"`
	Homex     float64 `xml:"home-x,attr" json:"home-x"`
	Homey     float64 `xml:"home-y,attr" json:"home-y"`
	Stamp     string  `xml:"save-date,attr" json:"save-date"`
	Generator string  `xml:"generator,attr" json:"generator"`
}

type Version struct {
	Value string `xml:"value,attr"`
}

type MissionDetail struct {
	Distance struct {
		Units string `xml:"units,attr,omitempty" json:"units,omitempty"`
		Value int    `xml:"value,attr,omitempty" json:"value,omitempty"`
	} `xml:"distance,omitempty" json:"distance,omitempty"`
}

type MissionSegment struct {
	Metadata     MissionMWP    `xml:"mwp" json:"meta"`
	MissionItems []MissionItem `xml:"missionitem" json:"mission"`
}

type MultiMission struct {
	XMLName xml.Name         `xml:"mission"  json:"-"`
	Version Version          `xml:"version" json:"-"`
	Comment string           `xml:",comment" json:"-"`
	Segment []MissionSegment `json:"missions"`
}

type Mission struct {
	XMLName      xml.Name      `xml:"mission"  json:"-"`
	Version      Version       `xml:"version" json:"-"`
	Comment      string        `xml:",comment" json:"-"`
	Metadata     MissionMWP    `xml:"mwp" json:"meta"`
	MissionItems []MissionItem `xml:"missionitem" json:"mission"`
	mission_file string        `xml:"-" json:"-"`
}

type PlaceMark struct {
	LineString struct {
		AltitudeMode string `xml:"altitudeMode"`
		Coordinates  string `xml:"coordinates"`
	} `xml:"LineString"`
}

type Gpx struct {
	XMLName xml.Name `xml:"gpx"`
	Wpts    []Pts    `xml:"wpt"`
	Rpts    []Pts    `xml:"rte>rtept"`
	Tpts    []Pts    `xml:"trk>trkseg>trkpt"`
}

type Pts struct {
	Lat  float64 `xml:"lat,attr"`
	Lon  float64 `xml:"lon,attr"`
	Elev float64 `xml:"ele"`
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

var ActionMap = map[string]int{
	"WAYPOINT":      wp_WAYPOINT,
	"POSHOLD_UNLIM": wp_POSHOLD_UNLIM,
	"POSHOLD_TIME":  wp_POSHOLD_TIME,
	"RTH":           wp_RTH,
	"SET_POI":       wp_SET_POI,
	"JUMP":          wp_JUMP,
	"SET_HEAD":      wp_SET_HEAD,
	"LAND":          wp_LAND,
}

func (m *Mission) Decode_action(b byte) string {
	return decode_action(b)
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
	mlen := len(m.MissionItems)
	if mlen > options.Config.MaxWP {
		return false
	}
	// Urg, Urg array index v. WP Nos ......
	for i := 0; i < mlen; i++ {
		var target = int(m.MissionItems[i].P1 - 1)
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

func (m *Mission) Dump(dms bool, homep ...float64) {
	var hpos types.HomeRec
	if len(homep) == 2 {
		hpos.HomeLat = homep[0]
		hpos.HomeLon = homep[1]
		hpos.Flags = types.HOME_ARM
	}
	if len(homep) > 2 {
		hpos.HomeAlt = homep[2]
		hpos.Flags |= types.HOME_ALT
	}
	k := kml.KML(m.To_kml(hpos, dms, true, 1, true))
	k.WriteIndent(os.Stdout, "", "  ")
}

func (m *Mission) To_MWXML(fname string) {
	m.Comment = "bbl2kml"
	m.Metadata.Generator = "impload"
	m.Metadata.Stamp = time.Now().Format(time.RFC3339)
	w, err := os.Create(fname)
	if err == nil {
		defer w.Close()
		xs, _ := xml.MarshalIndent(m, "", " ")
		fmt.Fprint(w, xml.Header)
		fmt.Fprintln(w, string(xs))
	}
}

/****************************************************************************
 * Generic, shared with impload
 *****************************************************************************/

func (ml *MissionSegment) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if err := e.EncodeElement(ml.Metadata, xml.StartElement{Name: xml.Name{Local: "mwp"}}); err != nil {
		return err
	}
	for _, mi := range ml.MissionItems {
		if err := e.EncodeElement(mi, xml.StartElement{Name: xml.Name{Local: "missionitem"}}); err != nil {
			return err
		}
	}
	return nil
}

func find_kml_coords(dat []byte) *PlaceMark {
	buf := bytes.NewBuffer(dat)
	dec := xml.NewDecoder(buf)
	for {
		t, _ := dec.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "Placemark" {
				var p PlaceMark
				dec.DecodeElement(&p, &se)
				if len(p.LineString.Coordinates) > 0 {
					return &p
				}
			}
		default:
		}
	}
	return nil
}

func NewMultiMission(mis []MissionItem) *MultiMission {
	mm := &MultiMission{Segment: []MissionSegment{{}}}
	if mis != nil {
		segno := 0
		no := 1
		for j := range mis {
			mis[j].No = no
			no++
			mm.Segment[segno].MissionItems = append(mm.Segment[segno].MissionItems, mis[j])
			if mis[j].Flag == 0xa5 {
				if j != len(mis)-1 {
					mm.Segment = append(mm.Segment, MissionSegment{})
					segno++
					no = 1
				}
			}
		}
		if no > 1 {
			mm.Segment[segno].MissionItems[no-2].Flag = 0xa5
		}
	}
	return mm
}

func read_kml(dat []byte) *MultiMission {
	mis := []MissionItem{}
	pm := find_kml_coords(dat)
	if pm != nil {
		p3 := int16(0)
		if pm.LineString.AltitudeMode == "absolute" {
			p3 = 1
		}
		st := strings.Trim(pm.LineString.Coordinates, "\n\r\t ")
		ss := strings.Split(st, " ")
		n := 1
		for _, val := range ss {
			coords := strings.Split(val, ",")
			if len(coords) > 1 {
				for i, c := range coords {
					coords[i] = strings.Trim(c, "\n\r\t ")
				}
				lon, _ := strconv.ParseFloat(coords[0], 64)
				lat, _ := strconv.ParseFloat(coords[1], 64)
				alt := 0.0
				if len(coords) > 2 {
					alt, _ = strconv.ParseFloat(coords[2], 64)
				}
				item := MissionItem{No: n, Lat: lat, Lon: lon, Alt: int32(alt), P3: p3,
					Action: "WAYPOINT"}
				n++
				mis = append(mis, item)
			}
		}
	}
	return NewMultiMission(mis)
}

func read_gpx(dat []byte) *MultiMission {
	mis := []MissionItem{}
	var pts []Pts
	var g Gpx
	err := xml.Unmarshal(dat, &g)
	if err == nil {
		if len(g.Wpts) > 0 {
			pts = g.Wpts
		} else if len(g.Rpts) > 0 {
			pts = g.Rpts
		} else if len(g.Tpts) > 0 {
			pts = g.Tpts
		}
		if pts != nil {
			for k, p := range pts {
				item := MissionItem{No: k + 1, Lat: p.Lat, Lon: p.Lon,
					Alt: int32(p.Elev), P3: 1, Action: "WAYPOINT"}
				mis = append(mis, item)
			}
		}
	} else {
		fmt.Fprintf(os.Stderr, "GPX error: %v", err)
	}
	return NewMultiMission(mis)
}

func (mi *MissionItem) is_GeoPoint() bool {
	a := mi.Action
	return !(a == "RTH" || a == "SET_HEAD" || a == "JUMP")
}

func (mm *MultiMission) is_valid() bool {
	force := os.Getenv("IMPLOAD_NO_VERIFY")
	if len(force) > 0 {
		return true
	}
	// Urg, Urg array index v. WP Nos ......
	xmlen := 0
	for _, m := range mm.Segment {
		mlen := len(m.MissionItems)
		xmlen += mlen
		for i := 0; i < mlen; i++ {
			var target = int(m.MissionItems[i].P1 - 1)
			if m.MissionItems[i].Action == "JUMP" {
				if (i == 0) || ((target > (i - 2)) && (target < (i + 2))) || (target >= mlen) || (m.MissionItems[i].P2 < -1) {
					return false
				}
				if !(m.MissionItems[target].Action == "WAYPOINT" || m.MissionItems[target].Action == "POSHOLD_TIME" || m.MissionItems[target].Action == "LAND") {
					return false
				}
			}
		}
	}
	if xmlen > options.Config.MaxWP {
		return false
	}
	return true
}

/*
*

	func (m *MissionSegment) Add_rtl(land bool) {
		k := len(m.MissionItems)
		p1 := int16(0)
		if land {
			p1 = 1
		}
		if k > 0 {
			if m.MissionItems[k-1].Flag == 0xa5 {
				m.MissionItems[k-1].Flag = 0
			}
		}
		item := MissionItem{No: k + 1, Lat: 0.0, Lon: 0.0, Alt: 0, Action: "RTH", P1: p1}
		m.MissionItems = append(m.MissionItems, item)
	}

	func (m *MultiMission) Dump(outfmt string, params ...string) {
		switch outfmt {
		case "cli":
			m.To_cli(params[0])
		case "json":
			m.To_json(params[0])
		default:
			m.To_xml(params...)
		}
	}

*
*/
func read_simple(dat []byte) *MultiMission {
	var mis = []MissionItem{}
	r := csv.NewReader(strings.NewReader(string(dat)))
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
		if has_no {
			j = 1
		}

		p1 := int16(0)
		p2 := int16(0)
		fp2 := 0.0
		p3 := 0
		flag := 0
		lat, _ = strconv.ParseFloat(record[j+1], 64)
		lon, _ = strconv.ParseFloat(record[j+2], 64)
		alt, _ := strconv.ParseFloat(record[j+3], 64)
		fp1, _ := strconv.ParseFloat(record[j+4], 64)
		if len(record) > j+5 {
			fp2, _ = strconv.ParseFloat(record[j+5], 64)
		}
		if len(record) > j+6 {
			p3, _ = strconv.Atoi(record[j+6])
		}
		if len(record) > j+7 {
			flag, _ = strconv.Atoi(record[j+7])
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
		case "SET_POI":
		case "SET_HEAD":
			p1 = int16(fp1)
		default:
			continue
		}
		item := MissionItem{No: n, Lat: lat, Lon: lon, Alt: int32(alt), Action: action, P1: p1, P2: p2, P3: int16(p3), Flag: uint8(flag)}
		mis = append(mis, item)
		n++
	}
	return NewMultiMission(mis)
}

func read_qgc_json(dat []byte) []QGCrec {
	qgcs := []QGCrec{}
	var qm qgc_plan
	json.Unmarshal(dat, &qm)
	if qm.Filetype == "Plan" {
		for _, qmi := range qm.Mission.Items {
			if qmi.Typ == "SimpleItem" {
				if len(qmi.Params) == 7 {
					qg := QGCrec{}
					qg.jindex = qmi.Jumpid
					qg.altmode = qmi.Altitudemode
					qg.command = qmi.Command
					qg.lat = qmi.Params[4]
					qg.lon = qmi.Params[5]
					qg.alt = qmi.Params[6]
					for j := 0; j < 4; j++ {
						qg.params[j] = qmi.Params[j]
					}
					qgcs = append(qgcs, qg)
				}
			} else if qmi.Typ == "ComplexItem" {
				for _, qmii := range qmi.Transect.Items {
					if len(qmii.Params) == 7 {
						qg := QGCrec{}
						qg.jindex = qmii.Jumpid
						qg.altmode = qmi.Altitudemode
						qg.command = qmii.Command
						qg.lat = qmii.Params[4]
						qg.lon = qmii.Params[5]
						qg.alt = qmii.Params[6]
						for j := 0; j < 4; j++ {
							qg.params[j] = qmii.Params[j]
						}
						qgcs = append(qgcs, qg)
					}
				}
			}
		}
	} else {
		fmt.Fprintln(os.Stderr, "Skipping non-Plan file")
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

func fixup_qgc_mission(mis []MissionItem, have_jump bool) ([]MissionItem, bool) {
	ok := true
	if have_jump {
		for i := range mis {
			if mis[i].Action == "JUMP" {
				jumptgt := mis[i].P1
				ajump := int16(0)
				for j := range mis {
					p3abs := mis[j].P3 // -ve indicate amsl
					if p3abs < 0 {
						p3abs *= -1
					}
					if p3abs == int16(jumptgt) {
						ajump = int16(j + 1)
						break
					}
				}
				if ajump == 0 {
					ok = false
				} else {
					mis[i].P1 = ajump
				}
				no := int16(i + 1) // item index
				if mis[i].P1 < 1 || ((mis[i].P1 > no-2) &&
					(mis[i].P1 < no+2)) {
					ok = false
				}
			}
		}
	}
	if ok {
		for i := range mis {
			if mis[i].P3 < 0 {
				mis[i].P3 = 1
			} else {
				mis[i].P3 = 0
			}
		}
		return mis, ok
	} else {
		return nil, false
	}
}

func process_qgc(dat []byte, mtype string) *MultiMission {
	var qs []QGCrec
	var mis = []MissionItem{}
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
			if /*q.alt == 0 ||*/ have_land {
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
			// P3 stores the original ID, which may not match No
			p3 := int16(q.jindex)
			no++
			item := MissionItem{No: no, Lat: q.lat, Lon: q.lon, Alt: int32(q.alt), Action: action, P1: p1, P2: p2, P3: p3}
			if item.is_GeoPoint() && q.altmode == 2 { // AMSL
				item.P3 *= -1 // -ve P3 indicates amsl
			}
			mis = append(mis, item)
			if last {
				break
			}
		}
	}

	mis, ok := fixup_qgc_mission(mis, have_jump)
	if !ok {
		log.Fatalf("Unsupported QGC file\n")
	}
	return NewMultiMission(mis)
}

func read_xml_mission(dat []byte) *MultiMission {
	v := Version{}
	mwps := []MissionMWP{}
	mis := []MissionItem{}
	buf := bytes.NewBuffer(dat)
	dec := xml.NewDecoder(buf)
	for {
		t, _ := dec.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			switch strings.ToLower(se.Name.Local) {
			case "mission":
			case "version":
				dec.DecodeElement(&v, &se)
			case "mwp", "meta":
				var mwp MissionMWP
				dec.DecodeElement(&mwp, &se)
				mwps = append(mwps, mwp)
			case "missionitem":
				var mi MissionItem
				dec.DecodeElement(&mi, &se)
				mis = append(mis, mi)
			default:
				fmt.Printf("Unknown MWXML tag %s\n", se.Name.Local)
			}
		}
	}
	mm := NewMultiMission(mis)
	mm.Version = v
	for j := range mm.Segment {
		if j < len(mwps) {
			mm.Segment[j].Metadata = mwps[j]
		}
	}
	return mm
}

func read_kmz(path string) (string, *MultiMission) {
	r, err := zip.OpenReader(path)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	for _, f := range r.File {
		rc, err := f.Open()
		defer rc.Close()
		if err == nil {
			dat, err := ioutil.ReadAll(rc)
			if err == nil {
				mtype, m := handle_mission_data(dat, path)
				if m != nil {
					return mtype, m
				}
			}
		}
	}
	return "", nil
}

func read_json(dat []byte, flg int) *MultiMission {
	switch flg {
	case 0:
		m := &Mission{}
		json.Unmarshal(dat, m)
		mm := NewMultiMission(m.MissionItems)
		return mm
	case 1:
		mm := &MultiMission{}
		json.Unmarshal(dat, mm)
		return mm
	default:
		return nil
	}
}

func read_inav_cli(dat []byte) *MultiMission {
	mis := []MissionItem{}
	for _, ln := range strings.Split(string(dat), "\n") {
		if strings.HasPrefix(ln, "wp ") {
			parts := strings.Split(ln, " ")
			if len(parts) == 10 {
				no, _ := strconv.Atoi(parts[1])
				iact, _ := strconv.Atoi(parts[2])
				ilat, _ := strconv.Atoi(parts[3])
				ilon, _ := strconv.Atoi(parts[4])
				alt, _ := strconv.Atoi(parts[5])
				p1, _ := strconv.Atoi(parts[6])
				p2, _ := strconv.Atoi(parts[7])
				p3, _ := strconv.Atoi(parts[8])
				flg, _ := strconv.Atoi(parts[9])
				lat := float64(ilat) / 1.0e7
				lon := float64(ilon) / 1.0e7
				action := decode_action(byte(iact))
				if iact == 6 {
					p1++
				}
				alt /= 100
				item := MissionItem{no, action, lat, lon, int32(alt), int16(p1), int16(p2), int16(p3), uint8(flg)}
				mis = append(mis, item)
			}
		}
	}
	return NewMultiMission(mis)
}

func handle_mission_data(dat []byte, path string) (string, *MultiMission) {
	var m *MultiMission
	mtype := "unknown"
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
		case bytes.Contains(dat, []byte("<kml ")):
			m = read_kml(dat)
			mtype = "kml"
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
	case bytes.HasPrefix(dat, []byte("PK\003\004")):
		mtype, m = read_kmz(path)
	case bytes.HasPrefix(dat, []byte(`{"meta":{`)):
		mtype = "mwp-json-s"
		m = read_json(dat, 0)
	case bytes.HasPrefix(dat, []byte(`{"missions":[`)):
		mtype = "mwp-json-m"
		m = read_json(dat, 1)
	case bytes.Contains(dat[0:100], []byte(`"fileType": "Plan"`)):
		mtype = "qgc-json"
		m = process_qgc(dat, mtype)
	case bytes.HasPrefix(dat, []byte("# wp")), bytes.HasPrefix(dat, []byte("#wp")), bytes.HasPrefix(dat, []byte("wp 0")):
		mtype = "inav cli"
		m = read_inav_cli(dat)
	default:
		m = nil
	}
	return mtype, m
}
