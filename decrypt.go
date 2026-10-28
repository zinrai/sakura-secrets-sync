package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Decrypt(cmdline, path string) (string, error) {
	parts := strings.Fields(cmdline)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty -decrypt-cmd")
	}

	cmd := exec.Command(parts[0], append(parts[1:], path)...)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("decrypt failed for %s: %w", path, err)
	}

	return string(out), nil
}
