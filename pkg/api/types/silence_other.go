// +build !windows

package api

import (
	"os/exec"
)

func SetSilentProcess(cmd *exec.Cmd) {
	/* No-op
	 * Thank's windows for such stupidity
	 */
}
