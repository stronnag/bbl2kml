package options


import (
	"flag"
	"strings"
	"path/filepath"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Configuration struct {
	Dms             bool   `json:"dms"`
	Dump            bool   `json:"-"`
	Efficiency      bool   `json:"efficiency"`
	Extrude         bool   `json:"extrude"`
	Fast            bool   `json:"-"`
	Kml             bool   `json:"kml"`
	Metas           bool   `json:"-"`
	Rssi            bool   `json:"rssi"`
	Bulletvers      int    `json:"blt-vers"`
	Intvl           int    `json:"-"`
	Idx             int    `json:"-"`
	HomeAlt         int    `json:"home-alt"`
	SplitTime       int    `json:"split-time"`
	Type            int    `json:"type"`
	Blackbox_decode string `json:"blackbox-decode"`
	Gradset         string `json:"gradient"`
	LTMdev          string `json:"-"`
	Mission         string `json:"-"`
	Mqttopts        string `json:"-"`
	Outdir          string `json:"outdir"`
	Rebase          string `json:"-"`
}

var Config Configuration = Configuration{Intvl: 1000, Blackbox_decode: "blackbox_decode", Bulletvers: 2, SplitTime: 120}

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

func parse_confile_file() {
	def := os.Getenv("APPDATA")
	if def == "" {
		def = os.Getenv("HOME")
		if def != "" {
			def = filepath.Join(def, ".config")
		} else {
			def = "./"
		}
	}
	fn := filepath.Join(def, "fl2x", "config.json")
	data, err := ioutil.ReadFile(fn)
	if err == nil {
		err = json.Unmarshal(data, &Config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "JSON Config: %v\n", err)
		}
	}
}

func ParseCLI(gv func() string) ([]string, string) {
	app := filepath.Base(os.Args[0])
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s [options] file...\n", app)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintln(os.Stderr, gv())
	}

	parse_confile_file()

	defs := os.Getenv("BBL2KML_OPTS")
	if defs != "" {
		_parts := strings.Split(defs, " ")
		var parts []string
		for _, p := range _parts {
			if p != "" {
				parts = append(parts, p)
			}
		}

		envflags := flag.NewFlagSet("$BBL2KML_OPTS", flag.ExitOnError)
		kml := envflags.Bool("kml", Config.Kml, "kml")
		rssi := envflags.Bool("rssi", Config.Rssi, "rssi")
		extrude := envflags.Bool("extrude", Config.Extrude, "extrude")
		dms := envflags.Bool("dms", Config.Dms, "dms")
		grad := envflags.String("gradient", Config.Gradset, "gradient")
		bbldec := envflags.String("decoder", Config.Blackbox_decode, "decoder")
		effic := envflags.Bool("efficiency", Config.Efficiency, "efficiency")
		envflags.Parse(parts)
		Config.Dms = *dms
		Config.Extrude = *extrude
		Config.Rssi = *rssi
		Config.Kml = *kml
		Config.Gradset = *grad
		Config.Efficiency = *effic
		if *bbldec != "" {
			Config.Blackbox_decode = *bbldec
		}
	}

	flag.IntVar(&Config.Idx, "index", 0, "Log index")
	flag.BoolVar(&Config.Dump, "dump", false, "Dump log headers and exit")
	flag.StringVar(&Config.Mission, "mission", "", "Optional mission file name")
	flag.IntVar(&Config.SplitTime, "split-time", Config.SplitTime, "[OTX] Time(s) determining log split, 0 disables")
	flag.IntVar(&Config.HomeAlt, "home-alt", Config.HomeAlt, "[OTX] home altitude")
	flag.StringVar(&Config.Rebase, "rebase", "", "rebase all positions on lat,lon[,alt]")
	if app == "fl2mqtt" {
		flag.StringVar(&Config.Mqttopts, "broker", "", "Mqtt URI (mqtt://[user[:pass]@]broker[:port]/topic[?cafile=file]")
		flag.IntVar(&Config.Bulletvers, "blt-vers", Config.Bulletvers, "[MQTT] BulletGCSS version")
		flag.StringVar(&Config.Outdir, "logfile", Config.Outdir, "Log file for browser replay")
	} else if app == "fl2ltm" {
		flag.StringVar(&Config.LTMdev, "device", "", "LTM device")
		flag.BoolVar(&Config.Metas, "metas", false, "list metadata and exit")
		flag.BoolVar(&Config.Fast, "fast", false, "faster replay")
		flag.IntVar(&Config.Type, "type", Config.Type, "model type")
	} else {
		flag.BoolVar(&Config.Kml, "kml", Config.Kml, "Generate KML (vice default KMZ)")
		flag.BoolVar(&Config.Rssi, "rssi", Config.Rssi, "Set RSSI view as default")
		flag.BoolVar(&Config.Extrude, "extrude", Config.Extrude, "Extends track points to ground")
		flag.BoolVar(&Config.Efficiency, "efficiency", Config.Efficiency, "Include efficiency layer in KML/Z")
		flag.StringVar(&Config.Gradset, "gradient", Config.Gradset, "Specific colour gradient [red,rdgn,yor]")
		flag.BoolVar(&Config.Dms, "dms", Config.Dms, "Show positions as DD:MM:SS.s (vice decimal degrees)")
		flag.StringVar(&Config.Outdir, "outdir", Config.Outdir, "Output directory for generated KML")
	}
	flag.IntVar(&Config.Intvl, "interval", Config.Intvl, "Sampling Interval (ms)")
	flag.Parse()

	if !isFlagSet("home-alt") {
		Config.HomeAlt = -999999 // sentinel
	}

	files := flag.Args()
	return files, app
}
