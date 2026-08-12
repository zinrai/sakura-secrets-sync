package main

import (
	"flag"
	"fmt"
	"os"
)

// Injected at build time by goreleaser via -ldflags -X
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var (
		dir         string
		dryRun      bool
		zone        string
		decryptCmd  string
		showVersion bool
	)

	flag.StringVar(&dir, "dir", "", "Environment directory to sync (required)")
	flag.BoolVar(&dryRun, "dry-run", false, "Report what would be done without writing to the store")
	flag.StringVar(&zone, "zone", "is1a", "Zone name (default: is1a)")
	flag.StringVar(&decryptCmd, "decrypt-cmd", "sops -d", "Command that decrypts a file to stdout, the file path is appended")
	flag.BoolVar(&showVersion, "version", false, "Print version")
	flag.Parse()

	if showVersion {
		fmt.Printf("sakura-secrets-sync version %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built: %s\n", date)
		return
	}

	if dir == "" {
		fmt.Fprintln(os.Stderr, "Error: -dir is required")
		flag.Usage()
		os.Exit(1)
	}

	if err := run(dir, zone, decryptCmd, dryRun); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
