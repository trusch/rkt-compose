package lib

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
)

func Logs(args []string) error {
	bs, err := ioutil.ReadFile(".pod-uuid")
	if err != nil {
		return errors.New("can not open .pod-uuid file: " + err.Error())
	}
	args = append([]string{"-M", "rkt-" + string(bs)}, args...)
	cmd := exec.Command("journalctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
