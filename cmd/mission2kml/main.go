package main

import (
	"os"
	"log"
	"strconv"
	mission "github.com/stronnag/bbl2kml/pkg/mission"
)

func main() {
	if len(os.Args) > 1 {
		inf := os.Args[1]
		dms := len(os.Args) > 2
		_, m, err := mission.Read_Mission_File(inf)
		if m != nil && err == nil {
			m.Dump(dms)
		}
		if err != nil {
			log.Fatal(err)
		}
	}
}
