package main

import "github.com/kevinrst/Luma/config"

func (m *model) load() {
	m.tuiVars = make([]EnvVar, 0, len(m.cfg.Env))
	for _, e := range m.cfg.Env {
		m.tuiVars = append(m.tuiVars, EnvVar{Key: e.Key, Value: e.Value})
	}

	if len(m.cfg.Windows) == 0 {
		return
	}

	m.windows = make([]*RequestWindow, 0, len(m.cfg.Windows))
	for _, w := range m.cfg.Windows {
		rw := &RequestWindow{
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
			rw.Headers = append(rw.Headers, &RequestHeader{Key: h.Key, Value: h.Value})
		}
		for _, p := range w.Params {
			rw.Params = append(rw.Params, &RequestParam{Key: p.Key, Value: p.Value})
		}
		m.windows = append(m.windows, rw)
	}

	if m.cfg.Current >= 0 && m.cfg.Current < len(m.windows) {
		m.currentWindow = m.cfg.Current
	}
	m.loadWindow(m.currentWindow)
}

func (m *model) save() {
	m.saveCurrentWindow()

	m.cfg.Env = make([]config.EnvVar, 0, len(m.tuiVars))
	for _, v := range m.tuiVars {
		m.cfg.Env = append(m.cfg.Env, config.EnvVar{Key: v.Key, Value: v.Value})
	}

	m.cfg.Windows = make([]config.Window, 0, len(m.windows))
	for _, w := range m.windows {
		cw := config.Window{
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
			cw.Headers = append(cw.Headers, config.Header{Key: h.Key, Value: h.Value})
		}
		for _, p := range w.Params {
			cw.Params = append(cw.Params, config.Param{Key: p.Key, Value: p.Value})
		}
		m.cfg.Windows = append(m.cfg.Windows, cw)
	}
	m.cfg.Current = m.currentWindow

	config.Save(m.cfg)
}
