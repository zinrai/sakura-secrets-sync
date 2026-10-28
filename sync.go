package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	secretmanager "github.com/sacloud/sacloud-sdk-go/api/secretmanager"
)

func run(dir, zone, decryptCmd string, dryRun bool) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}

	files, err := walkFiles(dir)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "walked %d file(s) under %s\n", len(files), dir)

	// An empty directory would mean deleting every secret in the vault.
	// That is never a sync, so refuse instead of mirroring emptiness.
	if len(files) == 0 {
		return fmt.Errorf("no files under %s, refusing to sync", dir)
	}

	vaultID := os.Getenv("SAKURA_SECRETS_ID")
	if vaultID == "" {
		return fmt.Errorf("required environment variable not set: SAKURA_SECRETS_ID")
	}

	client, err := NewClient(zone)
	if err != nil {
		return err
	}
	op := secretmanager.NewSecretOp(client, vaultID)

	existing, err := ListNames(client, vaultID)
	if err != nil {
		return err
	}

	synced := make(map[string]bool)
	for _, f := range files {
		name, err := PathToName(dir, f)
		if err != nil {
			return err
		}
		synced[name] = true
		if err := syncFile(op, decryptCmd, f, name, existing[name], dryRun); err != nil {
			return err
		}
	}

	if err := deleteOrphans(op, existing, synced, dryRun); err != nil {
		return err
	}

	if dryRun {
		fmt.Fprintln(os.Stderr, "dry-run: no changes applied")
	}

	return nil
}

func syncFile(op secretmanager.SecretAPI, decryptCmd, path, name string, exists bool, dryRun bool) error {
	value, err := Decrypt(decryptCmd, path)
	if err != nil {
		return err
	}
	value = Normalize(value)
	if value == "" {
		return fmt.Errorf("empty value after decrypt: %s", path)
	}

	if !exists {
		if err := putUnlessDryRun(op, name, value, dryRun); err != nil {
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

	if err := putUnlessDryRun(op, name, value, dryRun); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "[update] %s -> %s\n", path, name)
	return nil
}

func deleteOrphans(op secretmanager.SecretAPI, existing, synced map[string]bool, dryRun bool) error {
	names := make([]string, 0, len(existing))
	for name := range existing {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if synced[name] {
			continue
		}
		if err := deleteUnlessDryRun(op, name, dryRun); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "[delete] %s\n", name)
	}

	return nil
}

func putUnlessDryRun(op secretmanager.SecretAPI, name, value string, dryRun bool) error {
	if dryRun {
		return nil
	}
	return PutSecret(op, name, value)
}

func deleteUnlessDryRun(op secretmanager.SecretAPI, name string, dryRun bool) error {
	if dryRun {
		return nil
	}
	return DeleteSecret(op, name)
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
