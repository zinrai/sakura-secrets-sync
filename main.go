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
		filesPath   string
		all         bool
		zone        string
		decryptCmd  string
		showVersion bool
	)

	flag.StringVar(&dir, "dir", "", "Environment directory to sync (required)")
	flag.StringVar(&filesPath, "files", "", "File listing candidate paths, one per line")
	flag.BoolVar(&all, "all", false, "Sync every file under -dir and report orphans")
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

	if all == (filesPath != "") {
		fmt.Fprintln(os.Stderr, "Error: exactly one of -files or -all is required")
		flag.Usage()
		os.Exit(1)
	}

	if err := run(dir, filesPath, all, zone, decryptCmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
