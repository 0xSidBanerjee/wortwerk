// Package cache stores LLM responses on disk for reproducible ingests.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

// Key hashes chunk text, prompt version, and model name.
func Key(chunkText, promptVersion, model string) string {
	h := sha256.Sum256([]byte(promptVersion + "\x00" + model + "\x00" + chunkText))
	return hex.EncodeToString(h[:])
}

// Path returns the cache file path under root/data/llm-cache.
func Path(root, key string) string {
	return filepath.Join(root, "data", "llm-cache", key+".json")
}

// Read returns cached bytes if present.
func Read(root, key string) ([]byte, bool) {
	p := Path(root, key)
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, false
	}
	return b, true
}

// Write stores response bytes.
func Write(root, key string, data []byte) error {
	p := Path(root, key)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}
