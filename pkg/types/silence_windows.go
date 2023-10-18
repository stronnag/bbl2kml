//go:build windows
// +build windows

package types

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func SetSilentProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000} // CREATE_NO_WINDOW
}

func copydir(src string, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}

	file, err := f.Stat()
	if err != nil {
		return err
	}
	if !file.IsDir() {
		return fmt.Errorf("Source " + file.Name() + " is not a directory!")
	}

	err = os.Mkdir(dest, 0755)
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			err = copydir(filepath.Join(src, f.Name()), filepath.Join(dest, f.Name()))
			if err != nil {
				return err
			}
		}

		if !f.IsDir() {
			content, err := ioutil.ReadFile(filepath.Join(src, f.Name()))
			if err != nil {
				return err

			}
			err = ioutil.WriteFile(filepath.Join(dest, f.Name()), content, 0755)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func checkdirs(p string) string {
	def := os.Getenv("LOCALAPPDATA")
	nfp := filepath.Join(def, p)
	if _, err := os.Stat(nfp); os.IsNotExist(err) {
		odef := os.Getenv("APPDATA")
		ofp := filepath.Join(odef, p)
		if _, err := os.Stat(ofp); err == nil {
			fmt.Fprintf(os.Stderr, "** Migrating %s %s\n", ofp, nfp)
			copydir(ofp, nfp)
			os.RemoveAll(ofp)
		} else {
			os.MkdirAll(nfp, 0755)
		}
	}
	return nfp
}

func GetConfigDir() string {
	return checkdirs("fl2x")
}

func GetCacheDir() string {
	checkdirs("fl2x")
	return checkdirs(filepath.Join("fl2x", ".cache"))
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
