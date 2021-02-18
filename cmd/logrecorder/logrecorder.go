package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type recorder struct{}

func (r *recorder) Write(p []byte) (int, error) {
	_, _ = fmt.Fprintf(os.Stdout, "\n----------------\n")

	return os.Stdout.Write(p)
}

func main() {
	c := flag.String("c", "", "command")

	flag.Parse()

	if *c == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	cmd := exec.Command("/bin/sh", "-c", *c)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	cmd.Stdout = &recorder{}
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "command error: %v", err)
	}
}
