package logtail

import "os/exec"

func setCmdSysProcAttr(cmd *exec.Cmd) {
}

func killCmd(cmd *exec.Cmd) error {
	return cmd.Process.Kill()
}
