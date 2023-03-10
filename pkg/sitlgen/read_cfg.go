package sitlgen

import (
	"bufio"
	"fmt"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func read_cfg(eeprom string) []string {
	cdir := types.GetConfigDir()
	fn := filepath.Join(cdir, "fl2sitl.conf")
	var args []string
	r, err := os.Open(fn)
	if err == nil {
		defer r.Close()
		h := make(map[string]string)
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			l := scanner.Text()
			l = strings.TrimSpace(l)
			if !(len(l) == 0 || strings.HasPrefix(l, "#") || strings.HasPrefix(l, ";")) {
				parts := strings.SplitN(l, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					h[key] = val
				}
			}
		}

		if v, ok := h["sitl"]; ok {
			args = append(args, v)
			args = append(args, "--sim=xp")
			if v, ok = h["simip"]; ok {
				args = append(args, fmt.Sprintf("--simip=%s", v))
			}
			if v, ok = h["simport"]; ok {
				args = append(args, fmt.Sprintf("--simport=%s", v))
			}

			if v, ok = h["eeprom-path"]; ok {
				ep := os.ExpandEnv(v)
				if len(eeprom) == 0 {
					if v, ok = h["default-eeprom"]; ok {
						eeprom = v
					} else {
						eeprom = "eeprom.bin"
					}
				}
				ep = filepath.Join(ep, eeprom)
				args = append(args, fmt.Sprintf("--path=%s", ep))
			}
		}
	} else {
		log.Fatal("%s : %v\n", fn, err)
	}
	return args
}
