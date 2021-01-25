package options


import (
	"flag"
	"strings"
	"path/filepath"
	"fmt"
	"os"
)

var (
	Dms             bool   = false
	Dump            bool   = false
	Extrude         bool   = false
	Kml             bool   = false
	Rssi            bool   = false
	Efficiency      bool   = false
	Intvl           int    = 1000
	Idx             int    = 0
	SplitTime       int    = 0
	HomeAlt         int    = -999999
	Blackbox_decode string = "blackbox_decode"
	Mission         string
	Gradset         string
	Outdir          string
	Mqttopts        string
	Rebase          string
)

func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func Usage() {
	flag.Usage()
}

func ParseCLI(gv func() string) []string {
	app := filepath.Base(os.Args[0])

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s [options] file...\n", app)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintln(os.Stderr, gv())
	}

	defs := os.Getenv("BBL2KML_OPTS")
	_parts := strings.Split(defs, " ")
	var parts []string
	for _, p := range _parts {
		if p != "" {
			parts = append(parts, p)
		}
	}

	envflags := flag.NewFlagSet("$BBL2KML_OPTS", flag.ExitOnError)
	kml := envflags.Bool("kml", false, "kml")
	rssi := envflags.Bool("rssi", false, "rssi")
	extrude := envflags.Bool("extrude", false, "extrude")
	dms := envflags.Bool("dms", false, "dms")
	grad := envflags.String("gradient", "", "gradient")
	bbldec := envflags.String("decoder", Blackbox_decode, "decoder")
	effic := envflags.Bool("efficiency", false, "efficiency")
	envflags.Parse(parts)
	Dms = *dms
	Extrude = *extrude
	Rssi = *rssi
	Kml = *kml
	Gradset = *grad
	Efficiency = *effic

	if *bbldec != "" {
		Blackbox_decode = *bbldec
	}

	flag.IntVar(&Idx, "index", 0, "Log index")
	flag.IntVar(&Intvl, "interval", 1000, "Sampling Interval (ms)")
	flag.BoolVar(&Dump, "dump", false, "Dump log headers and exit")
	flag.StringVar(&Mission, "mission", "", "Optional mission file name")
	flag.IntVar(&SplitTime, "split-time", 120, "[OTX] Time(s) determining log split, 0 disables")
	flag.IntVar(&HomeAlt, "home-alt", 0, "[OTX] home altitude")
	flag.StringVar(&Rebase, "rebase", "", "rebase all positions on lat,lon[,alt]")
	if app == "fl2mqtt" {
		flag.StringVar(&Mqttopts, "broker", "", "Mqtt URI (mqtt://[user[:pass]@]broker[:port]/topic[?cafile=file]")
	} else {
		flag.BoolVar(&Kml, "kml", Kml, "Generate KML (vice default KMZ)")
		flag.BoolVar(&Rssi, "rssi", Rssi, "Set RSSI view as default")
		flag.BoolVar(&Extrude, "extrude", Extrude, "Extends track points to ground")
		flag.BoolVar(&Efficiency, "efficiency", Efficiency, "Include efficiency layer in KML/Z")
		flag.StringVar(&Gradset, "gradient", Gradset, "Specific colour gradient [red,rdgn,yor]")
		flag.BoolVar(&Dms, "dms", Dms, "Show positions as DD:MM:SS.s (vice decimal degrees)")
		flag.StringVar(&Outdir, "outdir", "", "Output directory for generated KML")
	}

	flag.Parse()

	if !isFlagSet("home-alt") {
		HomeAlt = -999999 // sentinel
	}

	files := flag.Args()
	return files
}
