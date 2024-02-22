package geo

import (
	"fmt"
)

type DEMMgr struct {
	dem *hgtDb
}

func InitDem(demdir string) (d *DEMMgr) {
	d = &DEMMgr{}
	d.dem = NewHgtDb(demdir)
	return d
}

func (d *DEMMgr) Get_Elevation(lat, lon float64) (float64, error) {
	return d.lookup_and_check(lat, lon)
}

func (d *DEMMgr) lookup_and_check(lat, lon float64) (float64, error) {
	var e float64
	for j := 0; ; j++ {
		e = d.dem.lookup(lat, lon)
		if e == DEM_NODATA {
			if j == 0 {
				fname, _, _ := get_file_name(lat, lon)
				download(fname, d.dem.dir)
			} else {
				return e, fmt.Errorf("DEM: No data for %f %f", lat, lon)
				break
			}
		} else {
			break
		}
	}
	return e, nil
}
