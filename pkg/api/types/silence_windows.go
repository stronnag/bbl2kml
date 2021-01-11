// +build windows

package api

import (
	"syscall"
	"os/exec"
)

func SetSilentProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000} // CREATE_NO_WINDOW
}
