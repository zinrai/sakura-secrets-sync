package main

import (
	"path/filepath"
	"strings"
)

func PathToName(dir, path string) (string, error) {
	// Not mangled: Secret Manager accepts '/' in names, so escaping it
	// would only make the name differ from the path it came from.
	return filepath.Rel(dir, path)
}

func Normalize(value string) string {
	// Not kept as decrypted: the store strips trailing newlines on write,
	// so every comparison would differ and every run would update.
	return strings.TrimRight(value, "\n")
}
