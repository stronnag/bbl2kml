package types

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"strings"
)

const (
	IS_UNKNOWN = -1
	IS_BBL     = 1
	IS_OTX     = 2
	IS_BLT     = 3
	IS_MWP     = 4
	IS_AP      = 5
	IS_SQL     = 6
)

func EvinceFileType(fn string) int {
	res := IS_UNKNOWN
	file, err := os.Open(fn)
	if err != nil {
		log.Fatalf("filetype: %+v\n", err)
	}
	defer file.Close()
	fh := bufio.NewReader(file)
	sig, err := fh.Peek(128) //read a few bytes without consuming
	if err == nil {
		switch {
		case strings.HasPrefix(string(sig), "H Product:Blackbox"):
			res = IS_BBL
		case strings.HasPrefix(string(sig), "Date,Time,"):
			res = IS_OTX
		case strings.Contains(string(sig), "|Connected to"):
			res = IS_BLT
		case strings.HasPrefix(string(sig), `{"type":`):
			res = IS_MWP
		case bytes.HasPrefix(sig, []byte{0xa3, 0x95, 0x80, 0x80, 0x59, 0x46, 0x4d, 0x54}):
			res = IS_AP
		case strings.HasPrefix(string(sig), "SQLite format 3"):
			res = IS_SQL
		}
	}
	return res
}
