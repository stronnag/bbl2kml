package kmlgen

import (
	"fmt"
	"path/filepath"
	"os"
	options "github.com/stronnag/bbl2kml/pkg/options"
)

func GenKmlName(inp string, idx int) string {
	outfn := filepath.Base(inp)
	ext := filepath.Ext(outfn)
	if len(ext) < len(outfn) {
		outfn = outfn[0 : len(outfn)-len(ext)]
	}
	if options.Kml {
		ext = ".kml"
	} else {
		ext = ".kmz"
	}
	if idx > 0 {
		ext = fmt.Sprintf(".%d%s", idx, ext)
	}
	outfn = outfn + ext
	if len(options.Outdir) > 0 {
		os.MkdirAll(options.Outdir, os.ModePerm)
		stat, err := os.Stat(options.Outdir)
		if err == nil && stat.IsDir() {
			outfn = filepath.Join(options.Outdir, outfn)
		}
	}
	return outfn
}
