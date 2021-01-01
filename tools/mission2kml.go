package main

import (
	"os"
	"log"
)

func main() {
	if len(os.Args) > 1 {
		inf := os.Args[1]
		_, m, err := Read_Mission_File(inf)
		if m != nil && err == nil {
			m.Dump()
		}
		if err != nil {
			log.Fatal(err)
		}
	}
}
