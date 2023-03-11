package sitlgen

import (
	"bufio"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type SimMeta struct {
	sitl   string
	ip     string
	port   string
	path   string
	eeprom string
}

func read_cfg() SimMeta {
	sitl := SimMeta{}
	cdir := types.GetConfigDir()
	fn := filepath.Join(cdir, "fl2sitl.conf")
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
					}
				}
			}
		}
	} else {
		log.Fatal("%s : %v\n", fn, err)
	}
	return sitl
}
