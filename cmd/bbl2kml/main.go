package main

import (
	"os"
	"fmt"
	"log"
	"path/filepath"
	bbl "github.com/stronnag/bbl2kml/pkg/bbl"
	options "github.com/stronnag/bbl2kml/pkg/options"
)

var GitCommit = "local"
var GitTag = "0.0.0"

func getVersion() string {
	return fmt.Sprintf("%s %s, commit: %s", filepath.Base(os.Args[0]), GitTag, GitCommit)
}

func main() {
	files := options.ParseCLI(getVersion, false)

	if options.Dump {
		bbl.Reader(files[0], bbl.BBLMeta{Index: 1})
		os.Exit(1)
	}

	for _, fn := range files {
		bmeta, err := bbl.Meta(fn)
		if err == nil {
			for _, b := range bmeta {
				if (options.Idx == 0 || options.Idx == b.Index) && b.Size > 4096 {
					m := b.MetaData()
					for _, k := range []string{"Log", "Flight", "Firmware", "Size"} {
						if v, ok := m[k]; ok {
							fmt.Printf("%-8.8s : %s\n", k, v)
						}
					}
					res := bbl.Reader(fn, b)
					fmt.Printf("%-8.8s : %s\n", "Disarm", m["Disarm"])
					if !res {
						fmt.Fprintf(os.Stderr, "*** skipping KML/Z for log  with no valid geospatial data\n")
					}
					fmt.Println()
				}
			}
		} else {
			log.Fatal(err)
		}
	}
}
