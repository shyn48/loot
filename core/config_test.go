package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfigOverridesAndBounds(t *testing.T) {
	base := Config{DownloadDir: "/base", MaxActive: 3, SectionSizeMB: 2}
	data := []byte(`
download_dir = "/tmp/dl"
max_active = 5
section_size_mb = 8
`)
	cfg, err := parseConfig(data, base)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DownloadDir != "/tmp/dl" || cfg.MaxActive != 5 || cfg.SectionSizeMB != 8 {
		t.Fatalf("override failed: %+v", cfg)
	}

	// Out-of-range values are clamped to sane minimums.
	clamped, _ := parseConfig([]byte("max_active = 0\nsection_size_mb = 0\n"), base)
	if clamped.MaxActive != 1 || clamped.SectionSizeMB != 1 {
		t.Fatalf("bounds not applied: %+v", clamped)
	}

	// Empty file → base defaults preserved.
	empty, _ := parseConfig(nil, base)
	if empty.DownloadDir != "/base" || empty.MaxActive != 3 {
		t.Fatalf("empty config didn't preserve base: %+v", empty)
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	if got := expandHome("~/Downloads"); got != filepath.Join(home, "Downloads") {
		t.Fatalf("expandHome = %q", got)
	}
	if got := expandHome("/abs/path"); got != "/abs/path" {
		t.Fatalf("absolute path changed: %q", got)
	}
}
