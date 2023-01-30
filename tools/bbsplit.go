package main

/*
 * Somewhat resource hungry (in that it loads the whole file into  memory) BBL splitter
 * (c) Jonathan Hudson 2022
 * 0BSD licence <https://opensource.org/licenses/0BSD>
 *
 * go build -ldflags "-s -w" bbsplit.go
 * cross compile Linux to Windows (or other via GOARCH/GOOS)
 * GOOS=windows go build -ldflags "-s -w" bbsplit.go
 *
 * # Multiple files may be given, disassebled into unique names (partno_fileN.TXT).
 * $ bbsplit file1.TXT ... fileN.TXT
 */

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type BBLElement struct {
	Start  int
	Length int
}

func main() {
	if len(os.Args) > 1 {
		for _, fn := range os.Args[1:] {
			dat, err := os.ReadFile(fn)
			if err == nil {
				base := filepath.Base(fn)
				i := 0
				needle := []byte("H Product:Blackbox")
				done := false
				var parts []BBLElement

				for done != true {
					p := bytes.Index(dat[i:], needle)
					done = p == -1
					if p != 0 {
						sp := i - len(needle)
						sz := 0
						if p == -1 {
							sz = len(dat) - sp
						} else {
							sz = p + len(needle)
						}
						parts = append(parts, BBLElement{Start: sp, Length: sz})
						i += sz
					} else {
						i = len(needle)
					}
				}
				if len(parts) > 1 {
					for n, p := range parts {
						fname := fmt.Sprintf("%03d_%s", n+1, base)
						if fh, err := os.Create(fname); err == nil {
							fmt.Printf("==> %v\n", fname)
							fh.Write(dat[p.Start : p.Start+p.Length])
							fh.Close()
						}
					}
				} else {
					fmt.Printf("Single part BBL %s\n", base)
				}
			} else {
				log.Fatalln(fn, err)
			}
		}
	} else {
		fmt.Println("No file(s) given")
	}
}
