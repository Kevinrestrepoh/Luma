package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type EnvVar struct {
	Key   string `toml:"key"`
	Value string `toml:"value"`
}

type Config struct {
	Env     []EnvVar `toml:"env"`
	Windows []Window `toml:"windows"`
	Current int      `toml:"current_window"`
}

func Dir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ".luma"
	}
	return filepath.Join(dir, "luma")
}

func path() string {
	return filepath.Join(Dir(), "config.toml")
}

func Load() Config {
	var cfg Config
	data, err := os.ReadFile(path())
	if err != nil {
		return cfg
	}
	_ = toml.Unmarshal(data, &cfg)
	return cfg
}

func Save(cfg Config) {
	_ = os.MkdirAll(Dir(), 0o755)
	f, err := os.Create(path())
	if err != nil {
		return
	}
	defer f.Close()
	_ = toml.NewEncoder(f).Encode(cfg)
}
