package lib

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func Run(podManifest, networks string, interactive, verbose bool, extra []string) error {
	args := createRunArgList(podManifest, networks, interactive, verbose, extra)
	cmd := exec.Command("rkt", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	log.Print("starting pod...")
	return cmd.Run()
}

func createRunArgList(podManifest, networks string, interactive, verbose bool, extra []string) []string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	uuidSaveFile := fmt.Sprintf("--uuid-file-save=%v/.pod-uuid", pwd)
	manifest := fmt.Sprintf("--pod-manifest=%v/%v", pwd, podManifest)
	parts := []string{"run", manifest, "--net=" + networks, uuidSaveFile}
	if interactive {
		parts = append(parts, "--interactive")
	}
	if verbose {
		parts = append(parts, "--debug")
	}
	if len(extra) > 0 {
		parts = append(parts, extra...)
	}
	return parts
}
