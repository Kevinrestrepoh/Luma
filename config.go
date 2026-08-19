package main

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	EnvVars map[string]string `toml:"env"`
}

func configDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ".luma"
	}
	return filepath.Join(dir, "luma")
}

func configPath() string {
	return filepath.Join(configDir(), "config.toml")
}

func loadConfig() Config {
	var cfg Config
	data, err := os.ReadFile(configPath())
	if err != nil {
		return Config{EnvVars: make(map[string]string)}
	}
	_ = toml.Unmarshal(data, &cfg)
	if cfg.EnvVars == nil {
		cfg.EnvVars = make(map[string]string)
	}
	return cfg
}

func (m *model) saveConfig() {
	_ = os.MkdirAll(configDir(), 0o755)
	f, err := os.Create(configPath())
	if err != nil {
		return
	}
	defer f.Close()
	_ = toml.NewEncoder(f).Encode(m.config)
}

func (m *model) loadTuiVarsFromConfig() {
	vars := make([]EnvVar, 0, len(m.config.EnvVars))
	for k, v := range m.config.EnvVars {
		vars = append(vars, EnvVar{Key: k, Value: v})
	}
	m.tuiVars = vars
}

func (m *model) syncTuiVarsToConfig() {
	m.config.EnvVars = make(map[string]string)
	for _, v := range m.tuiVars {
		m.config.EnvVars[v.Key] = v.Value
	}
	m.saveConfig()
}
