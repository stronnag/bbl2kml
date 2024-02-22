package inav

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

import (
	"mission"
)

type SafeHome struct {
	Lat   float64
	Lon   float64
	Index uint8
}

func Read_safehome(fn string) ([]SafeHome, []mission.FWApproach) {
	var sha []SafeHome
	var fwa []mission.FWApproach
	r, err := os.Open(fn)
	if err == nil {
		defer r.Close()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			l := scanner.Text()
			l = strings.TrimSpace(l)
			if !(len(l) == 0 || strings.HasPrefix(l, "#") || strings.HasPrefix(l, ";")) {
				parts := strings.Split(l, " ")
				if len(parts) > 0 {
					switch parts[0] {
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
								fw := mission.FWApproach{}
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
					}
				}
			}
		}
	}
	return sha, fwa
}
