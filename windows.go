package main

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type windowHeader struct {
	Key   string `toml:"key"`
	Value string `toml:"value"`
}

type windowParam struct {
	Key   string `toml:"key"`
	Value string `toml:"value"`
}

type windowData struct {
	Method        int            `toml:"method"`
	URL           string         `toml:"url"`
	Body          string         `toml:"body"`
	Headers       []windowHeader `toml:"headers"`
	Params        []windowParam  `toml:"params"`
	SelectedTab   int            `toml:"selected_tab"`
	StatusCode    int            `toml:"status_code"`
	Status        string         `toml:"status"`
	ResponseTime  string         `toml:"response_time"`
	OutputContent string         `toml:"output_content"`
}

type windowsFile struct {
	Windows       []windowData `toml:"windows"`
	CurrentWindow int          `toml:"current_window"`
}

func windowsPath() string {
	return filepath.Join(configDir(), "windows.toml")
}

func loadWindowsFile() windowsFile {
	var wf windowsFile
	data, err := os.ReadFile(windowsPath())
	if err != nil {
		return wf
	}
	_ = toml.Unmarshal(data, &wf)
	return wf
}

func (m *model) saveWindows() {
	_ = os.MkdirAll(configDir(), 0o755)
	f, err := os.Create(windowsPath())
	if err != nil {
		return
	}
	defer f.Close()

	m.saveCurrentWindow()

	wf := windowsFile{
		CurrentWindow: m.currentWindow,
	}
	for _, w := range m.windows {
		wd := windowData{
			Method:        w.Method,
			URL:           w.URL,
			Body:          w.Body,
			SelectedTab:   w.SelectedTab,
			StatusCode:    w.StatusCode,
			Status:        w.Status,
			ResponseTime:  w.ResponseTime,
			OutputContent: w.OutputContent,
		}
		for _, h := range w.Headers {
			wd.Headers = append(wd.Headers, windowHeader{Key: h.Key, Value: h.Value})
		}
		for _, p := range w.Params {
			wd.Params = append(wd.Params, windowParam{Key: p.Key, Value: p.Value})
		}
		wf.Windows = append(wf.Windows, wd)
	}
	_ = toml.NewEncoder(f).Encode(wf)
}

func (m *model) loadWindows() {
	wf := loadWindowsFile()
	if len(wf.Windows) == 0 {
		return
	}

	m.windows = make([]*RequestWindow, 0, len(wf.Windows))
	for _, wd := range wf.Windows {
		w := &RequestWindow{
			Method:        wd.Method,
			URL:           wd.URL,
			Body:          wd.Body,
			SelectedTab:   wd.SelectedTab,
			StatusCode:    wd.StatusCode,
			Status:        wd.Status,
			ResponseTime:  wd.ResponseTime,
			OutputContent: wd.OutputContent,
		}
		for _, h := range wd.Headers {
			w.Headers = append(w.Headers, &RequestHeader{Key: h.Key, Value: h.Value})
		}
		for _, p := range wd.Params {
			w.Params = append(w.Params, &RequestParam{Key: p.Key, Value: p.Value})
		}
		m.windows = append(m.windows, w)
	}

	if wf.CurrentWindow >= 0 && wf.CurrentWindow < len(m.windows) {
		m.currentWindow = wf.CurrentWindow
	}
	m.loadWindow(m.currentWindow)
}
