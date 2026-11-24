// Copyright 2026 Ibrahim Karakira
// SPDX-License-Identifier: Apache-2.0

// Command signoff creates and verifies signed provenance for agent work.
package main

import (
	"fmt"
	"io"
	"os"
)

// version is overridden at release time with -ldflags "-X main.version=...".
var version = "dev"

const usage = "usage: signoff <command>\n\ncommands:\n  version   print the version\n"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[0] {
	case "version":
		fmt.Fprintf(stdout, "signoff %s\n", version)
		return 0
	default:
		fmt.Fprintf(stderr, "signoff: unknown command %q\n\n%s", args[0], usage)
		return 2
	}
}
