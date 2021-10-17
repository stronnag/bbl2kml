package api

import (
	"path/filepath"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	b64 "encoding/base64"
)


func get_cache_dir() string {
	def := os.Getenv("APPDATA")
	if def == "" {
		def = os.Getenv("HOME")
		if def != "" {
			def = filepath.Join(def, ".cache")
		} else {
			def = "./"
		}
	}
	return filepath.Join(def, "fl2x")
}

func Init() {
	dir := get_cache_dir()
	os.Mkdir(dir, 0755)
}

func get_cache_name(lname string) (string, error) {
	fi, err := os.Stat(lname)
	if err != nil {
		return "", err
	}
	sz := fi.Size()
	mt := fi.ModTime().UTC().UnixNano() / 1000 // for Ubuntu 20.04 et al
	base := filepath.Base(lname)
	uenc := b64.URLEncoding.EncodeToString([]byte(base))
	fn := fmt.Sprintf("%s.%x.%x", uenc, sz, mt)
	dn := get_cache_dir()
	return filepath.Join(dn, fn), nil
}

func ReadMetaCache(lname string) ([]FlightMeta, error) {
	var mt []FlightMeta
	fn, _ := get_cache_name(lname)
	data, err := ioutil.ReadFile(fn)
	if err == nil {
		err = json.Unmarshal(data, &mt)
	}
	return mt, err
}

func WriteMetaCache(lname string, m []FlightMeta) {
	fn, _ := get_cache_name(lname)
	e, err := json.Marshal(m)
	if err == nil {
		ioutil.WriteFile(fn, e, 0644)
	}
}
