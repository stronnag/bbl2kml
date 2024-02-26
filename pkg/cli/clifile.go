package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type SafeHome struct {
	Lat   float64
	Lon   float64
	Index uint8
}

type Point struct {
	Lat float64
	Lon float64
}

type GeoZone struct {
	Zid    int
	Shape  int
	Gtype  int
	Minalt int
	Maxalt int
	Action int
	Points []Point
}

const (
	SHAPE_CIRCLE = 0
	SHAPE_POLY   = 1
)

const (
	TYPE_EXC = 0
	TYPE_INC = 1
)

var (
	Safehome_distance float64 = (200.0 / 1852.0)
	Fwapproach_length float64 = (350.0 / 1852.0)
)

func (g *GeoZone) To_string() string {
	var s1 string
	s := fmt.Sprintf("geozone %d %d %d %d %d %d\n", g.Zid, g.Gtype, g.Shape, g.Minalt, g.Maxalt, g.Action)
	if g.Shape == SHAPE_CIRCLE {
		s1 = fmt.Sprintf("%f,%f radius %.2fm", g.Points[0].Lat, g.Points[0].Lon, g.Points[1].Lat)
	} else {
		s1 = fmt.Sprintf("%d points ", len(g.Points))
	}
	return s + s1
}

func Read_clifile(fn string) ([]SafeHome, []FWApproach, []GeoZone) {
	var sha []SafeHome
	var fwa []FWApproach
	var gzone = make([]GeoZone, 0)

	r, err := os.Open(fn)
	if err == nil {
		defer r.Close()
		zid := -1
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			l := scanner.Text()
			l = strings.TrimSpace(l)
			if !(len(l) == 0 || strings.HasPrefix(l, "#") || strings.HasPrefix(l, ";")) {
				parts := strings.Split(l, " ")
				if len(parts) > 0 {
					switch parts[0] {
					case "geozone":
						switch parts[1] {
						case "vertex":
							vid := -1
							ilat := -1
							ilon := -1
							zid, err = strconv.Atoi(parts[2])
							if err == nil {
								vid, err = strconv.Atoi(parts[3])
								if err == nil {
									if zid < len(gzone) && vid == len(gzone[zid].Points) {
										ilat, err = strconv.Atoi(parts[4])
										if err == nil {
											ilon, err = strconv.Atoi(parts[5])
											if err == nil {
												if ilon == 0 {
													gzone[zid].Points = append(gzone[zid].Points, Point{float64(ilat) / 100.0, 0.0})
												} else {
													gzone[zid].Points = append(gzone[zid].Points, Point{float64(ilat) / 1e7, float64(ilon) / 1e7})
												}
											}
										}
									}
								}
							}
						default:
							zid, err = strconv.Atoi(parts[1])
							if zid == len(gzone) {
								var gz = GeoZone{}
								gz.Zid = zid
								gz.Shape, err = strconv.Atoi(parts[2])
								if err == nil {
									gz.Gtype, err = strconv.Atoi(parts[3])
									if err == nil {
										gz.Minalt, err = strconv.Atoi(parts[4])
										if err == nil {
											gz.Maxalt, err = strconv.Atoi(parts[5])
											if err == nil {
												gz.Action, err = strconv.Atoi(parts[6])
												gzone = append(gzone, gz)
											}
										}
									}
								}
							}
						}
					case "safehome":
						if len(parts) == 5 {
							if parts[2] == "1" {
								sh := SafeHome{}
								idx, _ := strconv.Atoi(parts[1])
								sh.Index = uint8(idx)
								v, _ := strconv.Atoi(parts[3])
								sh.Lat = float64(v) / 1e7
								v, _ = strconv.Atoi(parts[4])
								sh.Lon = float64(v) / 1e7
								sha = append(sha, sh)
							}
						}

					case "fwapproach":
						if len(parts) == 8 {
							idx, _ := strconv.Atoi(parts[1])
							if idx < 8 {
								fw := FWApproach{}
								fw.No = int8(idx)
								iv, _ := strconv.Atoi(parts[2])
								fw.Appalt = int32(iv)
								iv, _ = strconv.Atoi(parts[3])
								fw.Landalt = int32(iv)
								if parts[4] == "1" {
									fw.Dref = "right"
								} else {
									fw.Dref = "left"
								}
								iv, _ = strconv.Atoi(parts[5])
								fw.Dirn1 = int16(iv)
								iv, _ = strconv.Atoi(parts[6])
								fw.Dirn2 = int16(iv)
								fw.Aref = parts[7] == "1"
								fwa = append(fwa, fw)
							}
						}
					case "set":
						// nav_fw_land_approach_length safehome_max_distancee
						if len(parts) == 4 {
							switch parts[1] {
							case "nav_fw_land_approach_length":
								fv, _ := strconv.ParseFloat(parts[3], 64)
								if fv != 0 {
									Fwapproach_length = fv / 100.0 / 1852.0
								}
							case "safehome_max_distance":
								fv, _ := strconv.ParseFloat(parts[3], 64)
								if fv != 0 {
									Safehome_distance = fv / 100.0 / 1852.0
								}
							}
						}
					}
				}
			}
		}
	}
	return sha, fwa, gzone
}
