package main

import (
	"bufio"
	"os"
	"strings"
)

type EnvVar struct {
	Key   string
	Value string
}

func loadEnvFiles() []EnvVar {
	var vars []EnvVar
	entries, err := os.ReadDir(".")
	if err != nil {
		return vars
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".env") || name == ".env" {
			vars = append(vars, parseEnvFile(name)...)
		}
	}
	return vars
}

func parseEnvFile(path string) []EnvVar {
	var vars []EnvVar
	file, err := os.Open(path)
	if err != nil {
		return vars
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "="); idx != -1 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			value = unquote(value)
			vars = append(vars, EnvVar{Key: key, Value: value})
		}
	}
	return vars
}

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func (m *model) resolveEnvVars(s string) string {
	vars := append(m.envVars, m.tuiVars...)
	for _, v := range vars {
		s = strings.ReplaceAll(s, "$"+v.Key, v.Value)
	}
	return s
}
