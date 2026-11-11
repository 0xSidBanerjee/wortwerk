// Command wortwerk is the Wortwerk CLI (serve, ingest, validate, audio, dict).
package main

import (
	"fmt"
	"os"
)

const usage = `wortwerk — German reading app tooling

Usage:
  wortwerk <command> [args]

Commands:
  serve      Run the embedded static server (not implemented yet)
  ingest     Compile stories from markdown (not implemented yet)
  validate   Validate compiled JSON (not implemented yet)
  audio      Generate pronunciation audio (not implemented yet)
  dict       Build local dictionary index (not implemented yet)
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	cmd := os.Args[1]
	switch cmd {
	case "serve", "ingest", "validate", "audio", "dict":
		fmt.Fprintf(os.Stderr, "wortwerk %s: not implemented yet (milestone M3/M6)\n", cmd)
		os.Exit(1)
	case "-h", "--help", "help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n%s", cmd, usage)
		os.Exit(2)
	}
}
