package options

import (
	"encoding/json"
	"flag"
	"fmt"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Configuration struct {
	Dms             bool    `json:"dms"`
	Dump            bool    `json:"-"`
	Efficiency      bool    `json:"efficiency"`
	Extrude         bool    `json:"extrude"`
	Fast            bool    `json:"-"`
	Kml             bool    `json:"kml"`
	Metas           bool    `json:"-"`
	Rssi            bool    `json:"rssi"`
	Summary         bool    `json:"-"`
	Bulletvers      int     `json:"blt-vers"`
	Intvl           int     `json:"-"`
	Idx             int     `json:"-"`
	HomeAlt         int     `json:"home-alt"`
	SplitTime       int     `json:"split-time"`
	Type            int     `json:"type"`
	Blackbox_decode string  `json:"blackbox-decode"`
	Gradset         string  `json:"gradient"`
	Engunit         string  `json:"energy-unit"`
	LTMdev          string  `json:"-"`
	Mission         string  `json:"-"`
	MissionIndex    int     `json:"-"`
	MaxWP           int     `json:"max-wp"`
	Mqttopts        string  `json:"-"`
	Outdir          string  `json:"outdir"`
	Rebase          string  `json:"-"`
	Visibility      int     `json:"visibility"`
	Tmpdir          string  `json:"-"`
	Epsilon         float64 `json:"-"`
	StartOff        int     `json:"start-offset"`
	EndOff          int     `json:"end-offset"`
	Modefilter      string  `json:"-"`
	UseTopo         bool    `json:"-"`
	Attribs         string  `json:"attributes"`
	Aflags          int     `json:"-"`
	RedIsFast       bool    `json:"fast-is-red"`
	RedIsLow        bool    `json:"low-is-red"`
}

var Config Configuration = Configuration{Intvl: 1000, Blackbox_decode: "blackbox_decode", Bulletvers: 2, SplitTime: 120, Epsilon: 0.015, StartOff: 30, EndOff: -30, Engunit: "mah", MaxWP: 120}

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

func parse_config_file(cfgfile string) error {
	var err error
	var fn string
	if cfgfile != "" {
		fn = cfgfile
	} else {
		def := types.GetConfigDir()
		fn = filepath.Join(def, "config.json")
	}
	data, oerr := ioutil.ReadFile(fn)
	if oerr == nil {
		err = json.Unmarshal(data, &Config)
	} else {
		res, xerr := json.MarshalIndent(Config, "", "  ")
		if xerr == nil {
			ioutil.WriteFile(fn, res, 0644)
		}
	}
	return err
}

func ParseCLI(gv func() string) ([]string, string) {
	types.Init()
	app := filepath.Base(os.Args[0])
	var err error

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s [options] file...\n", app)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintln(os.Stderr, gv())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: config file ignored due to error: %v\n", err)
		}
	}

	cfgfile := os.Getenv("FL2X_CONFIG_FILE")
	needcf := false
	for i := 0; i < len(os.Args); i++ {
		if needcf {
			if !strings.HasPrefix(os.Args[i], "-") {
				cfgfile = os.Args[i]
			}
			break
		}
		if strings.HasPrefix(os.Args[i], "-config") || strings.HasPrefix(os.Args[i], "--config") {
			parts := strings.Split(os.Args[i], "=")
			if len(parts) == 2 {
				cfgfile = parts[i]
				break
			} else {
				needcf = true
			}
		}
	}

	err = parse_config_file(cfgfile)
	Config.Blackbox_decode = types.SetBBLFallback(Config.Blackbox_decode)

	/**
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
	**/

	showversion := false
	flag.IntVar(&Config.Idx, "index", 0, "Log index")
	flag.IntVar(&Config.SplitTime, "split-time", Config.SplitTime, "[OTX] Time(s) determining log split, 0 disables")
	flag.StringVar(&Config.Rebase, "rebase", "", "rebase all positions on lat,lon[,alt]")

	if app != "log2mission" {
		flag.IntVar(&Config.HomeAlt, "home-alt", Config.HomeAlt, "[OTX] home altitude")
		flag.BoolVar(&Config.Dump, "dump", false, "Dump log headers and exit")
		flag.StringVar(&Config.Mission, "mission", "", "Optional mission file name")
		flag.IntVar(&Config.MissionIndex, "mission-index", 1, "Optional mission file index")
	}
	if strings.HasPrefix(app, "fl2mqtt") {
		flag.StringVar(&Config.Mqttopts, "broker", "", "Mqtt URI (mqtt://[user[:pass]@]broker[:port]/topic[?cafile=file]")
		flag.IntVar(&Config.Bulletvers, "blt-vers", Config.Bulletvers, "[MQTT] BulletGCSS version")
		flag.StringVar(&Config.Outdir, "logfile", Config.Outdir, "Log file for browser replay")
	} else if strings.HasPrefix(app, "fl2ltm") {
		flag.StringVar(&Config.LTMdev, "device", "", "LTM device")
		flag.BoolVar(&Config.Metas, "metas", false, "list metadata and exit")
		flag.BoolVar(&Config.Fast, "fast", false, "faster replay")
		flag.IntVar(&Config.Type, "type", Config.Type, "model type")
	} else if strings.HasPrefix(app, "log2mission") {
		flag.Float64Var(&Config.Epsilon, "epsilon", Config.Epsilon, "Epsilon")
		flag.IntVar(&Config.StartOff, "start-offset", Config.StartOff, "Start Offset (seconds)")
		flag.IntVar(&Config.EndOff, "end-offset", Config.EndOff, "End Offset (seconds)")
		flag.StringVar(&Config.Modefilter, "mode-filter", Config.Modefilter, "Mode filter (cruise,wp)")
		flag.IntVar(&Config.MaxWP, "max-wp", Config.MaxWP, "Maximum WPs in mission")
	} else {
		flag.BoolVar(&Config.Kml, "kml", Config.Kml, "Generate KML (vice default KMZ)")
		flag.BoolVar(&Config.Rssi, "rssi", Config.Rssi, "Set RSSI view as default")
		flag.BoolVar(&Config.Extrude, "extrude", Config.Extrude, "Extends track points to ground")
		flag.BoolVar(&Config.Efficiency, "efficiency", Config.Efficiency, "Include efficiency layer in KML/Z")
		flag.StringVar(&Config.Engunit, "energy-unit", Config.Engunit, "Energy unit [mah, wh]")
		flag.StringVar(&Config.Gradset, "gradient", Config.Gradset, "Specific colour gradient [red,rdgn,yor]")
		flag.BoolVar(&Config.Dms, "dms", Config.Dms, "Show positions as DD:MM:SS.s (vice decimal degrees)")
		flag.StringVar(&Config.Outdir, "outdir", Config.Outdir, "Output directory for generated KML")
		flag.IntVar(&Config.Visibility, "visibility", Config.Visibility, "0=folder value,-1=don't set,1=all on")
		flag.BoolVar(&Config.Summary, "summary", Config.Summary, "Just show summary")
		flag.StringVar(&Config.Attribs, "attributes", Config.Attribs, "Attributes to plot (effic,speed,altitude)")
	}
	flag.IntVar(&Config.Intvl, "interval", Config.Intvl, "Sampling Interval (ms)")
	flag.BoolVar(&showversion, "version", false, "Just show version")
	flag.StringVar(&cfgfile, "config", "", "alternate file")

	flag.Parse()

	if showversion {
		fmt.Println(gv())
		os.Exit(0)
	}

	if strings.HasPrefix(app, "flightlog2kml") {
		Config.UseTopo = true
	}
	if strings.HasPrefix(app, "bbsummary") {
		Config.Summary = true
	}
	if !isFlagSet("home-alt") {
		Config.HomeAlt = -999999 // sentinel
	}
	/*
		if Config.Idx == 0 {
			Config.Idx = 1
		}
	*/

	if Config.MissionIndex == 0 {
		Config.MissionIndex = 1
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: config file ignored due to error: %v\n", err)
	}

	if os.Getenv("DUMP_CONFIG") != "" {
		fmt.Fprintf(os.Stderr, "%+v\n", Config)
	}

	if Config.Efficiency {
		Config.Aflags |= types.AFlags_EFFIC
	}

	if Config.Attribs != "" {
		if strings.Contains(Config.Attribs, "effic") {
			Config.Aflags |= types.AFlags_EFFIC
		}
		if strings.Contains(Config.Attribs, "speed") {
			Config.Aflags |= types.AFlags_SPEED
		}
		if strings.Contains(Config.Attribs, "altitude") {
			Config.Aflags |= types.AFlags_ALTITUDE
		}
		if strings.Contains(Config.Attribs, "battery") {
			Config.Aflags |= types.AFlags_BATTERY
		}
	}

	files := flag.Args()
	return files, app
}
