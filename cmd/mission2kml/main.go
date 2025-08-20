package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

import (
	"kmlgen"
	"mission"
	"types"
)

import (
	kml "github.com/twpayne/go-kml"
)

var GitCommit = "local"
var GitTag = "0.0.0"

var (
	dms     bool
	homepos string
	idx     int
	outfile string
)

func GetVersion() string {
	return fmt.Sprintf("%s %s commit:%s", filepath.Base(os.Args[0]), GitTag, GitCommit)
}

func split(s string, separators []rune) []string {
	f := func(r rune) bool {
		for _, s := range separators {
			if r == s {
				return true
			}
		}
		return false
	}
	return strings.FieldsFunc(s, f)
}

func main() {
	defer types.RemoveTmpDir()

	flag.Usage = func() {
		extra := `The home location is given as decimal degrees latitude and
longitude and optional altitude. The values should be separated by a single
separator, one of "/:; ,". If space is used, then the values must be enclosed
in quotes.

In locales where comma is used as decimal "point", then comma should not be
used as a separator.

If a syntactically valid home position is given, without altitude, an online
elevation service is used to adjust mission elevations in the KML.

Examples:
    -home 54.353974/-4.5236
    --home 48,9975:2,5789/104
    -home 54.353974;-4.5236
    --home "48,9975 2,5789"
    -home 54.353974,-4.5236,24
`
		fmt.Fprintf(os.Stderr, "Usage of %s [options] files...\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, extra)
		fmt.Fprintln(os.Stderr, GetVersion())
	}

	defs := os.Getenv("BBL2KML_OPTS")
	dms = strings.Contains(defs, "-dms")

	outfile = "-"

	flag.BoolVar(&dms, "dms", dms, "Show positions as DMS (vice decimal degrees)")
	flag.StringVar(&homepos, "home", homepos, "Use home location")
	flag.StringVar(&outfile, "out", outfile, "Output file")
	flag.IntVar(&idx, "mission-index", 0, "Mission Index")
	flag.Parse()
	files := flag.Args()
	if len(files) == 0 {
		flag.Usage()
		os.Exit(-1)
	}

	var home []float64
	var v float64

	if len(homepos) > 0 {
		parts := split(homepos, []rune{'/', ':', ';', ' ', ','})
		if len(parts) >= 2 {
			var err error
			v, err = strconv.ParseFloat(parts[0], 64)
			if err == nil {
				home = append(home, v)
				v, err = strconv.ParseFloat(parts[1], 64)
				if err == nil {
					home = append(home, v)
					if len(parts) == 3 {
						v, err = strconv.ParseFloat(parts[2], 64)
						if err == nil {
							home = append(home, v)
						}
					}
				}
			}
		}
	}

	var mfile, cfile string
	for _, fn := range files {
		f, err := os.Open(fn)
		if err == nil {
			defer f.Close()
			dat, _ := io.ReadAll(f)
			switch {
			case bytes.HasPrefix(dat, []byte("<?xml")):
				if bytes.Contains(dat, []byte("<MISSION")) || bytes.Contains(dat, []byte("<mission")) {
					mfile = fn
				}
			case bytes.Contains(dat, []byte("safehome")), bytes.Contains(dat, []byte("fwapproach")),
				bytes.Contains(dat, []byte("geozone")):
				cfile = fn
			}
		}
	}

	err := generateKML(mfile, idx, dms, home, cfile)
	if err != nil {
		log.Fatalf("mission2kmk: %+v\n", err)
	}
}

func generateKML(mfile string, idx int, dms bool, homep []float64, clifile string) error {
	var sb strings.Builder
	kname := ""
	sb.Write([]byte(fmt.Sprintf("Generator: %s", GetVersion())))
	if mfile != "" {
		sb.Write([]byte(fmt.Sprintf(" mission: %s", filepath.Base(mfile))))
		kname = filepath.Base(mfile)
	}
	if clifile != "" {
		sb.Write([]byte(fmt.Sprintf(" cli: %s", filepath.Base(clifile))))
		if mfile == "" {
			kname = filepath.Base(clifile)
		}
	}
	d := kml.Folder(kml.Name(kname)).Add(kml.Description(sb.String())).Add(kml.Open(true))
	k := kml.KML(d)
	var err error

	if mfile != "" {
		inithp := len(homep)
		_, mm, err := mission.Read_Mission_File(mfile)
		if err == nil {
			isviz := true
			for nm, _ := range mm.Segment {
				nmx := nm + 1
				if idx == 0 || nmx == idx {
					ms := mm.To_mission(nmx)

					if len(homep) == 0 {
						if ms.Metadata.Homey != 0 && ms.Metadata.Homex != 0 {
							homep = append(homep, ms.Metadata.Homey, ms.Metadata.Homex)
						}
					}

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

					mf := ms.To_kml(hpos, dms, false, nmx, isviz)
					d.Add(mf)
					isviz = false
				}
				homep = homep[:inithp]
			}
		}
	}

	if clifile != "" {
		sfx := kmlgen.Generate_cli_kml(clifile, nil)
		for _, s := range sfx {
			d.Add(s)
		}
	}

	var w io.WriteCloser
	if outfile == "-" || outfile == "" {
		w = os.Stdout
	} else {
		w, err = os.Create(outfile)
		defer w.Close()
	}
	k.WriteIndent(w, "", "  ")
	return err
}
