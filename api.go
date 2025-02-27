package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

func FetchApi(url string, method string, headers []*ApiHeaders, body string) (*ApiResponse, error) {
	reqBody := bytes.NewBufferString(body)

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	for _, h := range headers {
		req.Header.Set(h.Key, h.Value)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	apiResponse := &ApiResponse{
		Status: resp.Status,
		Body:   writeJSON(respBody),
	}

	return apiResponse, nil
}

func writeJSON(data []byte) string {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "  "); err != nil {
		return string(data)
	}
	return pretty.String()
}
