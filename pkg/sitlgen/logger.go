package sitlgen

import (
	"log"
)

import (
	"options"
)

func Sitl_logger(val int, ofmt string, params ...interface{}) {
	if options.Config.Verbose > val {
		log.Printf(ofmt, params...)
	}
}
