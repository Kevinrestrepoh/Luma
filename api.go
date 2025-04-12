package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var sharedClient = &http.Client{
	Timeout: 10 * time.Second,
}

func FetchApi(url, method, body string) tea.Cmd {
	return func() tea.Msg {
		if url == "" {
			return ApiResponse{err: fmt.Errorf("empty url")}
		}

		headers := []*ApiHeaders{}

		start := time.Now()
		req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
		if err != nil {
			return ApiResponse{err: err}
		}

		for _, h := range headers {
			req.Header.Set(h.Key, h.Value)
		}

		resp, err := sharedClient.Do(req)
		if err != nil {
			return ApiResponse{err: err}
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return ApiResponse{err: err}
		}

		return ApiResponse{
			statusCode: resp.StatusCode,
			status:     resp.Status,
			body:       writeJSON(respBody),
			duration:   formatDuration(time.Since(start)),
			err:        nil,
		}
	}
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
