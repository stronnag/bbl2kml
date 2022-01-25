//go:build windows
// +build windows

package api

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func SetSilentProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000} // CREATE_NO_WINDOW
}

func GetConfigDir() string {
	return os.Getenv("APPDATA")
}

func GetCacheDir() string {
	def := os.Getenv("APPDATA")
	if def == "" {
		def = "./"
	}
	return filepath.Join(def, "fl2x", ".cache")
}
