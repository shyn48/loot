package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// meta is the on-disk record for a download, written once when the download is
// added. Everything mutable (bytes downloaded) is derived from temp file sizes,
// so this file never needs rewriting — which is what makes restart recovery
// cheap: reload the meta, re-measure the temp files, and continue.
type meta struct {
	ID           string     `json:"id"`
	URL          string     `json:"url"`
	Filename     string     `json:"filename"`
	TargetPath   string     `json:"targetPath"`
	Size         int64      `json:"size"`
	TotalSection int        `json:"totalSection"`
	Sections     [][2]int64 `json:"sections"`
	AcceptRanges bool       `json:"acceptRanges"`
	CreatedAt    time.Time  `json:"createdAt"`
}

func metaPath(dir, id string) string {
	return filepath.Join(dir, id+".meta.json")
}

func writeMeta(dir string, m meta) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath(dir, m.ID), b, 0o644)
}

func readMeta(path string) (meta, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return meta{}, err
	}
	var m meta
	if err := json.Unmarshal(b, &m); err != nil {
		return meta{}, err
	}
	return m, nil
}

func listMetaFiles(dir string) ([]string, error) {
	return filepath.Glob(filepath.Join(dir, "*.meta.json"))
}
