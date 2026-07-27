package main

import "strings"

// PathToName derives the secret name from a file path under dir.
// The path relative to dir is the name as-is. Secret Manager accepts '/'
// in names, so the repository layout carries over unmangled.
func PathToName(dir, path string) string {
	return strings.TrimPrefix(path, dir+"/")
}

// Normalize strips trailing newlines from a decrypted value. The store
// removes them on write, so keeping them would make every comparison differ.
func Normalize(value string) string {
	return strings.TrimRight(value, "\n")
}
