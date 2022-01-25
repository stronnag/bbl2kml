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

func GetConfigDir() string {
	def := os.Getenv("HOME")
	if def != "" {
		def = filepath.Join(def, ".config")
	} else {
		def = "./"
	}
	return def
}

func GetCacheDir() string {
	def := os.Getenv("HOME")
	if def == "" {
		def = "./"
	}
	return filepath.Join(def, ".cache", "fl2x")
}
