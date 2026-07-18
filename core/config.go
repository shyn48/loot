package core

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds user-editable settings loaded from
// ~/.config/godownloader/config.toml.
type Config struct {
	DownloadDir   string `toml:"download_dir"`
	MaxActive     int    `toml:"max_active"`
	SectionSizeMB int    `toml:"section_size_mb"`
}

func defaultConfig() Config {
	home, _ := os.UserHomeDir()
	return Config{
		DownloadDir:   filepath.Join(home, "Downloads"),
		MaxActive:     defaultMaxActive,
		SectionSizeMB: 2,
	}
}

func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "godownloader", "config.toml")
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return p
}

// parseConfig overlays TOML data onto base and normalizes the result. Factored
// out from LoadConfig so it can be unit-tested without touching the filesystem.
func parseConfig(data []byte, base Config) (Config, error) {
	cfg := base
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return base, err
	}
	cfg.DownloadDir = expandHome(cfg.DownloadDir)
	if cfg.DownloadDir == "" {
		cfg.DownloadDir = base.DownloadDir
	}
	if cfg.MaxActive < 1 {
		cfg.MaxActive = 1
	}
	if cfg.SectionSizeMB < 1 {
		cfg.SectionSizeMB = 1
	}
	return cfg, nil
}

// LoadConfig reads the config file, writing a default one on first run. Missing
// or unreadable config falls back to defaults rather than failing.
func LoadConfig() Config {
	def := defaultConfig()
	path := configPath()
	if path == "" {
		return def
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		writeDefaultConfig(path, def)
		return def
	}
	if err != nil {
		return def
	}
	cfg, err := parseConfig(data, def)
	if err != nil {
		return def
	}
	return cfg
}

func writeDefaultConfig(path string, cfg Config) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	content := "# godownloader configuration\n\n" +
		"download_dir = \"" + cfg.DownloadDir + "\"\n" +
		"max_active = " + strconv.Itoa(cfg.MaxActive) + "        # simultaneous downloads\n" +
		"section_size_mb = " + strconv.Itoa(cfg.SectionSizeMB) + "   # ~1 parallel section per N MB (max 20 sections)\n"
	os.WriteFile(path, []byte(content), 0o644)
}
