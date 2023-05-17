//go:build !windows
// +build !windows

package api

import (
	"os"
	"os/exec"
	"path/filepath"
)

func SetSilentProcess(cmd *exec.Cmd) {
	/* No-op
	 * Thank's windows for such stupidity
	 */
}

func getcommondir(p string) string {
	def := os.Getenv("HOME")
	if def == "" {
		def = "./"
	}
	return filepath.Join(def, p, "fl2x")
}

func GetConfigDir() string {
	dir := getcommondir(".config")
	os.MkdirAll(dir, 0755)
	return dir
}

func GetCacheDir() string {
	dir := getcommondir(".cache")
	os.MkdirAll(dir, 0755)
	return dir
}

func SetBBLFallback(bblname string) string {
	return bblname
}
