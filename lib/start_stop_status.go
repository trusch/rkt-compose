package lib

import (
	"os"
	"os/exec"
)

func Start(name string, podManifest, networks string, verbose bool) error {
	args := createRunArgList(podManifest, networks, false, verbose)
	args = append([]string{"--unit=" + name, "rkt"}, args...)
	cmd := exec.Command("systemd-run", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func Stop(name string) error {
	cmd := exec.Command("systemctl", "stop", name+".service")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
	cmd = exec.Command("systemctl", "reset-failed", name+".service")
	cmd.Run()
	return nil
}

func Status(name string) error {
	cmd := exec.Command("systemctl", "status", name+".service")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func Restart(name string) error {
	cmd := exec.Command("systemctl", "restart", name+".service")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
