// Command wortwerk is the Wortwerk CLI (serve, ingest, validate, audio, dict).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/0xSidBanerjee/wortwerk/internal/ingest/validate"
	"github.com/0xSidBanerjee/wortwerk/internal/story"
)

const usage = `wortwerk — German reading app tooling

Usage:
  wortwerk <command> [args]

Commands:
  serve      Run the embedded static server (not implemented yet)
  ingest     Compile stories from markdown (not implemented yet)
  validate   Validate compiled JSON files under content/compiled/
  audio      Generate pronunciation audio (not implemented yet)
  dict       Build local dictionary index (not implemented yet)

Examples:
  wortwerk validate
  wortwerk validate fixture
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	cmd := os.Args[1]
	switch cmd {
	case "validate":
		if err := runValidate(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "serve", "ingest", "audio", "dict":
		fmt.Fprintf(os.Stderr, "wortwerk %s: not implemented yet (see project milestones)\n", cmd)
		os.Exit(1)
	case "-h", "--help", "help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n%s", cmd, usage)
		os.Exit(2)
	}
}

func runValidate(args []string) error {
	root, err := repoRoot()
	if err != nil {
		return err
	}
	dir := filepath.Join(root, "content", "compiled")

	var files []string
	if len(args) == 0 {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if filepath.Ext(name) != ".json" || name == "index.json" {
				continue
			}
			files = append(files, filepath.Join(dir, name))
		}
	} else {
		for _, slug := range args {
			name := slug
			if filepath.Ext(name) != ".json" {
				name += ".json"
			}
			files = append(files, filepath.Join(dir, name))
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("no compiled stories to validate in %s", dir)
	}

	var failed bool
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var compiled story.CompiledStory
		if err := json.Unmarshal(data, &compiled); err != nil {
			fmt.Fprintf(os.Stderr, "%s: JSON decode: %v\n", filepath.Base(path), err)
			failed = true
			continue
		}
		res := validate.Story(&compiled)
		for _, iss := range res.Issues {
			if iss.Level == "warn" {
				fmt.Fprintf(os.Stderr, "%s: warn %s: %s\n", filepath.Base(path), iss.Where, iss.Message)
			}
		}
		if !res.OK() {
			failed = true
			for _, iss := range res.Issues {
				if iss.Level == "error" {
					fmt.Fprintf(os.Stderr, "%s: error %s: %s\n", filepath.Base(path), iss.Where, iss.Message)
				}
			}
		} else {
			fmt.Printf("%s: ok\n", filepath.Base(path))
		}
	}
	if failed {
		return fmt.Errorf("validation failed")
	}
	return nil
}

func repoRoot() (string, error) {
	if root := os.Getenv("WORTWERK_ROOT"); root != "" {
		return root, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("go.mod not found from %s", wd)
		}
		wd = parent
	}
}
