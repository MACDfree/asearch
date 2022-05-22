package util

import (
	"os/exec"
	"runtime"
	"syscall"
)

func OpenLocal(path string) {
	if runtime.GOOS == "windows" {
		cmd := exec.Command(`cmd`, `/c`, `start`, path)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		cmd.Start()
	}
}
