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

func SetBBLFallback(bblname string) string {
	if bblname == "blackbox_decode" {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		exPath := filepath.Dir(ex)
		bblpath := filepath.Join(exPath, "blackbox_decode.exe")
		if _, err := os.Stat(bblpath); err == nil {
			return bblpath
		}
	}
	return bblname
}
