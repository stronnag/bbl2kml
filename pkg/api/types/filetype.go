package api

import (
	"strings"
	"bufio"
	"os"
	"log"
)

const (
	IS_UNKNOWN = -1
	IS_BBL     = 1
	IS_OTX     = 2
)

func EvinceFileType(fn string) int {
	res := IS_UNKNOWN
	file, err := os.Open(fn)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	fh := bufio.NewReader(file)
	sig, err := fh.Peek(64) //read a few bytes without consuming
	if err == nil {
		if strings.HasPrefix(string(sig), "H Product:Blackbox") {
			res = IS_BBL
		} else if strings.HasPrefix(string(sig), "Date,Time,") {
			res = IS_OTX
		}
	}
	return res
}
