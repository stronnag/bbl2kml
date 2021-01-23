package geo

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"os"
)

type TopoRes struct {
	lat  float64
	lon  float64
	elev float64
}

func topo_parse_response(js []byte) []TopoRes {
	var res []TopoRes
	var result map[string]interface{}
	json.Unmarshal(js, &result)
	if result["status"] == "OK" {
		m0 := result["results"].([]interface{})
		for _, m00 := range m0 {
			m1 := m00.(map[string]interface{})
			elv := m1["elevation"].(float64)
			m2 := m1["location"].(map[string]interface{})
			la := m2["lat"].(float64)
			lo := m2["lng"].(float64)
			res = append(res, TopoRes{lat: la, lon: lo, elev: elv})
		}
	} else {
		fmt.Fprintln(os.Stderr, string(js))
	}
	return res
}

func getopentopo(lat, lon float64) ([]TopoRes, error) {
	var res []TopoRes
	// alternatives to mapzen are etopo1 and aster30m
	req := fmt.Sprintf("https://api.opentopodata.org/v1/mapzen?locations=%.7f,%.7f", lat, lon)
	response, err := http.Get(req)
	if err == nil {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err == nil {
			res = topo_parse_response(contents)
		}
	}
	return res, err
}

func GetElevation(lat, lon float64) (float64, error) {
	elev, err := getopentopo(lat, lon)
	return elev[0].elev, err
}
