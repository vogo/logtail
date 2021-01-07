// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package logtail

import (
	"os/exec"
	"syscall"
)

func setCmdSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func killCmd(cmd *exec.Cmd) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
