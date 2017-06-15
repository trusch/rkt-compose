package lib

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func Run(podManifest, networks string, interactive, verbose bool) error {
	args := createRunArgList(podManifest, networks, interactive, verbose)
	cmd := exec.Command("rkt", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	log.Print("starting pod...")
	return cmd.Run()
}

func createRunArgList(podManifest, networks string, interactive, verbose bool) []string {
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
	return parts
}
