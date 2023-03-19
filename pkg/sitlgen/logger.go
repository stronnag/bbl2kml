package sitlgen

import (
	options "github.com/stronnag/bbl2kml/pkg/options"
	"log"
)

func Sitl_logger(val int, ofmt string, params ...interface{}) {
	if options.Config.Verbose > val {
		log.Printf(ofmt, params...)
	}
}
