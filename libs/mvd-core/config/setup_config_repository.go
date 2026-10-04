package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// LoadOrCreate reads the config at path on top of the defaults. When the file
// does not exist it creates it (importing a legacy ./setup.conf if present) and
// reports created=true so the caller can open the config screen.
func LoadOrCreate(path, appDir, downloadsDir string) (cfg Config, created bool, err error) {
	base := Default(appDir, downloadsDir)

	if _, statErr := os.Stat(path); statErr == nil {
		cfg, err = parseInto(path, base)
		return cfg, false, err
	}

	cfg = base
	if legacy := "setup.conf"; fileExists(legacy) {
		if c, perr := parseInto(legacy, base); perr == nil {
			cfg = c
		}
	}
	if err = Save(cfg, path); err != nil {
		return cfg, false, err
	}
	return cfg, true, nil
}

// parseInto overlays the key=value file at path onto base.
func parseInto(path string, base Config) (Config, error) {
	cfg := base
	file, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		applyKey(&cfg, key, val, path)
	}
	return cfg, scanner.Err()
}

// Save writes the config to path as a commented key=value file.
func Save(cfg Config, path string) error {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	return os.WriteFile(path, []byte(configText(cfg)), 0o644)
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}
