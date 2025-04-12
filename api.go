package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

func (m *model) FetchApi() {
	url := m.url.Value()
	if url == "" {
		return
	}

	method := m.methods[m.selectedMethod].Name
	body := m.body.Value()
	headers := []*ApiHeaders{}

	reqBody := bytes.NewBufferString(body)

	start := time.Now()
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return
	}

	for _, h := range headers {
		req.Header.Set(h.Key, h.Value)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	m.responseTime = formatDuration(time.Since(start))

	m.statusCode = resp.StatusCode
	m.status = resp.Status
	m.output.SetContent(writeJSON(respBody))
}

func writeJSON(data []byte) string {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "  "); err != nil {
		return string(data)
	}
	return pretty.String()
}

func formatDuration(d time.Duration) string {
	switch {
	case d < time.Microsecond:
		return d.Round(time.Nanosecond).String()
	case d < time.Millisecond:
		return d.Round(time.Microsecond).String()
	case d < time.Second:
		return d.Round(time.Millisecond).String()
	default:
		return d.Round(time.Millisecond).String()
	}
}
