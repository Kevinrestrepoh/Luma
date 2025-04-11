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
