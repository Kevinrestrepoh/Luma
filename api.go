package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

// No global timeout: streaming bodies (SSE, chunked) stay open until the server closes or the user cancels.
var streamingHTTPClient = &http.Client{
	Transport: http.DefaultTransport,
}

func FetchApi(ctx context.Context, streamID int64, url, method, body string, headers []*ApiHeaders) tea.Cmd {
	return func() tea.Msg {
		url = strings.TrimSpace(url)
		if url == "" {
			return ApiResponse{err: fmt.Errorf("empty url")}
		}
		if ProgramSend == nil {
			return ApiResponse{err: fmt.Errorf("program not ready")}
		}
		if !hasScheme(url) {
			if isLocalhost(url) {
				url = "http://" + url
			} else {
				url = "https://" + url
			}
		}
		go runRequestStream(ctx, streamID, url, method, body, headers)
		return nil
	}
}

func isLocalhost(url string) bool {
	host := strings.ToLower(url)
	if strings.HasPrefix(host, "localhost") {
		return true
	}
	// Strip port if present
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	// IPv6 localhost
	if host == "[::1]" || host == "::1" {
		return true
	}
	// Private IPs: 10.x, 172.16-31.x, 192.168.x, 127.x, 0.0.0.0
	parts := strings.Split(host, ".")
	if len(parts) == 4 {
		if parts[0] == "10" || parts[0] == "0" || parts[0] == "127" {
			return true
		}
		if parts[0] == "192" && parts[1] == "168" {
			return true
		}
		if parts[0] == "172" {
			if second, err := strconv.Atoi(parts[1]); err == nil && second >= 16 && second <= 31 {
				return true
			}
		}
	}
	return false
}

func hasScheme(url string) bool {
	lower := strings.ToLower(url)
	return strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "ws://") ||
		strings.HasPrefix(lower, "wss://")
}

func runRequestStream(ctx context.Context, streamID int64, url, method, body string, headers []*ApiHeaders) {
	u := strings.ToLower(url)
	if strings.HasPrefix(u, "ws://") || strings.HasPrefix(u, "wss://") {
		runWebSocketStream(ctx, streamID, url, body, headers)
		return
	}
	runHTTPStream(ctx, streamID, url, method, body, headers)
}

func sendIfReady(msg tea.Msg) {
	if ProgramSend != nil {
		ProgramSend(msg)
	}
}

func runHTTPStream(ctx context.Context, streamID int64, url, method, body string, headers []*ApiHeaders) {
	start := time.Now()
	sendIfReady(streamResetMsg{id: streamID})

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBufferString(body))
	if err != nil {
		sendIfReady(streamDoneMsg{id: streamID, duration: formatDuration(time.Since(start)), err: err})
		return
	}
	for _, h := range headers {
		if h.Key != "" {
			req.Header.Set(h.Key, h.Value)
		}
	}

	resp, err := streamingHTTPClient.Do(req)
	if err != nil {
		sendIfReady(streamDoneMsg{id: streamID, duration: formatDuration(time.Since(start)), err: err})
		return
	}
	defer resp.Body.Close()

	ttfb := time.Since(start)
	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	showStreamUI := strings.Contains(ct, "text/event-stream")
	sendIfReady(streamHeaderMsg{
		id:                 streamID,
		statusCode:         resp.StatusCode,
		status:             resp.Status,
		ttfb:               formatDuration(ttfb),
		showStreamControls: showStreamUI,
	})

	buf := make([]byte, 8192)
	for {
		if ctx.Err() != nil {
			sendIfReady(streamDoneMsg{id: streamID, duration: formatDuration(time.Since(start)), err: ctx.Err()})
			return
		}
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			sendIfReady(streamDataMsg{id: streamID, chunk: string(buf[:n])})
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				sendIfReady(streamDoneMsg{id: streamID, duration: formatDuration(time.Since(start)), err: nil})
				return
			}
			sendIfReady(streamDoneMsg{id: streamID, duration: formatDuration(time.Since(start)), err: readErr})
			return
		}
	}
}

func runWebSocketStream(ctx context.Context, streamID int64, url, body string, headers []*ApiHeaders) {
	start := time.Now()
	sendIfReady(streamResetMsg{id: streamID})

	dialer := websocket.Dialer{
		HandshakeTimeout: 45 * time.Second,
		Proxy:            http.ProxyFromEnvironment,
	}
	hdr := http.Header{}
	for _, h := range headers {
		if h.Key == "" {
			continue
		}
		k := strings.ToLower(strings.TrimSpace(h.Key))
		switch k {
		case "connection", "upgrade", "sec-websocket-key", "sec-websocket-version",
			"sec-websocket-extensions", "sec-websocket-protocol":
			continue
		}
		hdr.Add(h.Key, h.Value)
	}

	conn, httpResp, err := dialer.DialContext(ctx, url, hdr)
	if err != nil {
		sendIfReady(streamDoneMsg{id: streamID, duration: formatDuration(time.Since(start)), err: err})
		return
	}
	defer conn.Close()

	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()

	status := "101 Switching Protocols"
	code := http.StatusSwitchingProtocols
	if httpResp != nil {
		status = httpResp.Status
		code = httpResp.StatusCode
	}
	sendIfReady(streamHeaderMsg{
		id:                 streamID,
		statusCode:         code,
		status:             status,
		ttfb:               formatDuration(time.Since(start)),
		showStreamControls: true,
	})

	if strings.TrimSpace(body) != "" {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(body))
	}

	for {
		_, msg, readErr := conn.ReadMessage()
		if readErr != nil {
			if ctx.Err() != nil {
				sendIfReady(streamDoneMsg{id: streamID, duration: formatDuration(time.Since(start)), err: ctx.Err()})
				return
			}
			// Normal close is still an error from ReadMessage; treat as clean end of stream.
			var closeErr *websocket.CloseError
			if errors.As(readErr, &closeErr) {
				sendIfReady(streamDoneMsg{id: streamID, duration: formatDuration(time.Since(start)), err: nil})
				return
			}
			sendIfReady(streamDoneMsg{id: streamID, duration: formatDuration(time.Since(start)), err: readErr})
			return
		}
		sendIfReady(streamDataMsg{id: streamID, chunk: string(msg)})
	}
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
