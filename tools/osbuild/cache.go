package main

import (
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

// CacheEntries represents a set of entries in the cache manifest.
type CacheEntries struct {
	Entries []struct {
		Key   string
		Value string
	} `yaml:",flow"`
}

// initCache loads the cached build cache manifest, or initialize an empty
// one if it can't be loaded.
func initCache(ctxt *OsbuildContext) {
	ctxt.CacheManifest = make(map[string]string)
	if data, err := os.ReadFile(path.Join(ctxt.CacheDir, "cache-manifest.yml")); err == nil && data != nil {
		entries := &CacheEntries{}
		yaml.Unmarshal(data, entries)
		for _, entry := range entries.Entries {
			ctxt.CacheManifest[entry.Key] = entry.Value
		}
	}
	ctxt.CleanupFuncs = append(ctxt.CleanupFuncs, func() {
		// Save the cache manifest at exit.
		entries := &CacheEntries{}
		for key, value := range ctxt.CacheManifest {
			entries.Entries = append(entries.Entries, struct {
				Key   string
				Value string
			}{Key: key, Value: value})
		}
		saved := false
		if data, err := yaml.Marshal(entries); err == nil {
			os.MkdirAll(ctxt.CacheDir, 0755)
			if err := os.WriteFile(path.Join(ctxt.CacheDir, "cache-manifest.yml"), data, 0666); err == nil {
				saved = true
			}
		}
		if !saved {
			ctxt.Logger.Warn("failed to save cache manifest")
		}
	})
}
