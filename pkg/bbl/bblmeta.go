package bbl

import (
	"os"
	"bufio"
	"strings"
	"strconv"
	"path"
	"io"
	"fmt"
	api "github.com/stronnag/bbl2kml/pkg/api"
)

type reason int

func Show_size(sz int64) string {
	var s string
	switch {
	case sz > 1024*1024:
		s = fmt.Sprintf("%.2f MB", float64(sz)/(1024*1024))
	case sz > 10*1024:
		s = fmt.Sprintf("%.1f KB", float64(sz)/1024)
	default:
		s = fmt.Sprintf("%d B", sz)
	}
	return s
}

func (r reason) String() string {
	var reasons = [...]string{"None", "Timeout", "Sticks", "Switch_3d", "Switch", "Killswitch", "Failsafe", "Navigation"}
	if r < 0 || int(r) >= len(reasons) {
		r = 0
	}
	return reasons[r]
}

func Meta(fn string) ([]api.BBLSummary, error) {
	var bes []api.BBLSummary
	r, err := os.Open(fn)
	if err == nil {
		var nbes int
		var loffset int64

		base := path.Base(fn)
		scanner := bufio.NewScanner(r)

		zero_or_nl := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}
			for i, b := range data {
				if b == '\n' || b == 0 || b == 0xff {
					return i + 1, data[0:i], nil
				}
			}

			if atEOF {
				return len(data), data, nil
			}
			return
		}

		scanner.Split(zero_or_nl)
		for scanner.Scan() {
			l := scanner.Text()
			switch {
			case strings.Contains(string(l), "H Product:"):
				offset, _ := r.Seek(0, io.SeekCurrent)

				if loffset != 0 {
					bes[nbes].Size = offset - loffset
				}
				loffset = offset
				be := BBLSummary{Disarm: "NONE", Size: 0}
				bes = append(bes, be)
				nbes = len(bes) - 1
				bes[nbes].Logname = base
				bes[nbes].Index = nbes + 1
				bes[nbes].Cdate = "<no date>"
				bes[nbes].Craft = "<unknown>"
			case strings.HasPrefix(string(l), "H Firmware revision:"):
				if n := strings.Index(string(l), ":"); n != -1 {
					fw := string(l)[n+1:]
					bes[nbes].Firmware = fw
				}

			case strings.HasPrefix(string(l), "H Firmware date:"):
				if n := strings.Index(string(l), ":"); n != -1 {
					fw := string(l)[n+1:]
					bes[nbes].Fwdate = fw
				}

			case strings.HasPrefix(string(l), "H Log start datetime:"):
				if n := strings.Index(string(l), ":"); n != -1 {
					date := string(l)[n+1:]
					bes[nbes].Cdate = date
				}

			case strings.HasPrefix(string(l), "H Craft name:"):
				if n := strings.Index(string(l), ":"); n != -1 {
					cname := string(l)[n+1:]
					bes[nbes].Craft = cname
				}

			case strings.Contains(string(l), "reason:"):
				if n := strings.Index(string(l), ":"); n != -1 {
					dindx, _ := strconv.Atoi(string(l)[n+1 : n+2])
					bes[nbes].Disarm = reason(dindx).String()
				}
			}
			if err = scanner.Err(); err != nil {
				return bes, err
			}
		}
		if bes[nbes].Size == 0 {
			offset, _ := r.Seek(0, io.SeekCurrent)
			if loffset != 0 {
				bes[nbes].Size = offset - loffset
			}
		}
	}
	return bes, err
}
