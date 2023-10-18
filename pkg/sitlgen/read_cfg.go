package sitlgen

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

import (
	"types"
)

type SimMeta struct {
	sitl     string
	ip       string
	port     string
	path     string
	eeprom   string
	mintime  int
	failmode uint16
}

func read_cfg(cfgfile string) SimMeta {
	sitl := SimMeta{}
	var fn string
	if cfgfile == "" {
		cdir := types.GetConfigDir()
		fn = filepath.Join(cdir, "fl2sitl.conf")
	} else {
		fn = cfgfile
	}
	r, err := os.Open(fn)
	if err == nil {
		defer r.Close()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			l := scanner.Text()
			l = strings.TrimSpace(l)
			if !(len(l) == 0 || strings.HasPrefix(l, "#") || strings.HasPrefix(l, ";")) {
				parts := strings.SplitN(l, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					switch key {
					case "sitl":
						sitl.sitl = val
					case "simip":
						sitl.ip = val
					case "simport":
						sitl.port = val
					case "eeprom-path":
						sitl.path = val
					case "default-eeprom":
						sitl.eeprom = val
					case "min-time":
						sitl.mintime, _ = strconv.Atoi(val)
					case "failmode":
						if val[0] == 'i' {
							sitl.failmode = 0
						} else if val[0] == 'n' {
							sitl.failmode = 0xd0d0
						} else {
							tmp, _ := strconv.Atoi(val)
							sitl.failmode = uint16(tmp)
						}
					}
				}
			}
		}
	} else {
		sitl.sitl = "inav_SITL"
		sitl.ip = "localhost"
		r, err := os.Create(fn)
		if err == nil {
			defer r.Close()
			fmt.Fprintln(r, "# SITL-sim settings")
			fmt.Fprintln(r, "sitl = inav_SITL")
			fmt.Fprintln(r, "simip = localhost")
			fmt.Fprintln(r, "# simport = 49000")
			fmt.Fprintln(r, "# eeprom-path = $HOME/sitl-eeproms")
			fmt.Fprintln(r, "# default-eeprom = test-eeprom.bin")
			fmt.Fprintln(r, "# Options are nopulse, ignore or a throttle value (e.g. 800)")
			fmt.Fprintln(r, "# failmode = 800")
			fmt.Fprintln(r, "# failmode = nopulse")
			fmt.Fprintln(r, "# min-time = 50")
		} else {
			log.Fatalf("%s : %v\n", fn, err)
		}
	}
	return sitl
}
