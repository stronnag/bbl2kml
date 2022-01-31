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
	return getcommondir(".config")
}

func GetCacheDir() string {
	return getcommondir(".cache")
}

func SetBBLFallback(bblname string) string {
	return bblname
}
