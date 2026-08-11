package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	secretmanager "github.com/sacloud/sacloud-sdk-go/api/secretmanager"
)

func run(dir, filesPath string, all bool, zone, decryptCmd string) error {
	dir = strings.TrimSuffix(dir, "/")
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}

	files, err := collectFiles(dir, filesPath, all)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}

	vaultID := os.Getenv("SAKURA_SECRETS_ID")
	if vaultID == "" {
		return fmt.Errorf("required environment variable not set: SAKURA_SECRETS_ID")
	}

	op, err := NewSecretOp(zone, vaultID)
	if err != nil {
		return err
	}

	existing, err := ListNames(op)
	if err != nil {
		return err
	}

	synced := make(map[string]bool)
	for _, f := range files {
		name := PathToName(dir, f)
		synced[name] = true
		if err := syncFile(op, decryptCmd, f, name, existing[name]); err != nil {
			return err
		}
	}

	if all {
		reportOrphans(existing, synced, dir)
	}

	return nil
}

// collectFiles determines the sync candidates, either every file under dir
// or the paths listed in filesPath.
func collectFiles(dir, filesPath string, all bool) ([]string, error) {
	if all {
		files, err := walkFiles(dir)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "walked %d file(s) under %s\n", len(files), dir)
		return files, nil
	}

	files, err := loadCandidates(filesPath, dir)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "loaded %d candidate(s) from %s\n", len(files), filesPath)
	return files, nil
}

// syncFile decrypts one file and writes it to the store only when needed:
// create when the secret does not exist, update when the value differs.
func syncFile(op secretmanager.SecretAPI, decryptCmd, path, name string, exists bool) error {
	value, err := Decrypt(decryptCmd, path)
	if err != nil {
		return err
	}
	value = Normalize(value)
	if value == "" {
		return fmt.Errorf("empty value after decrypt: %s", path)
	}

	if !exists {
		if err := PutSecret(op, name, value); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "[create] %s -> %s\n", path, name)
		return nil
	}

	remote, err := GetSecret(op, name)
	if err != nil {
		return err
	}
	if remote == value {
		fmt.Fprintf(os.Stderr, "[unchanged] %s -> %s\n", path, name)
		return nil
	}

	if err := PutSecret(op, name, value); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "[update] %s -> %s\n", path, name)
	return nil
}

// reportOrphans lists store entries that no synced file corresponds to.
// They are only reported, never deleted.
func reportOrphans(existing, synced map[string]bool, dir string) {
	names := make([]string, 0, len(existing))
	for name := range existing {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if !synced[name] {
			fmt.Fprintf(os.Stderr, "[orphan] %s has no file under %s, delete manually\n", name, dir)
		}
	}
}

func walkFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func loadCandidates(path, dir string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read candidates: %w", err)
	}

	var files []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, dir+"/") {
			return nil, fmt.Errorf("candidate not under %s: %s", dir, line)
		}
		files = append(files, line)
	}
	sort.Strings(files)
	return files, nil
}
