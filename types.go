package main

import (
	"bytes"
	"context"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kevinrst/Luma/config"
)

// ProgramSend posts messages from background network goroutines into the active Bubble Tea program.
// main sets this to tea.Program.Send before Run.
var ProgramSend func(tea.Msg)

type model struct {
	width          int
	height         int
	methods        []*Method
	selectedMethod int

	focus string
	mode  string

	url          textinput.Model
	body         Textarea
	statusCode   int
	status       string
	output       viewport.Model
	responseTime string

	methodStyles  *Styles
	urlStyles     *Styles
	outputStyles  *Styles
	requestStyles *Styles

	streamID     int64
	cancelStream context.CancelFunc
	streamBuf    bytes.Buffer
	streamFollow bool // when true, new stream chunks scroll the response viewport to the bottom

	// showStreamControls: live dot + Stop (SSE / WebSocket only); avoids flash on normal HTTP
	showStreamControls bool

	// outputInteractMode: toggle with i on the result panel; enables scrollbar/mouse scroll and white border
	outputInteractMode bool

	// output cursor position (absolute line/col in full document)
	outputCursorLine int
	outputCursorCol  int

	// visual selection state: "none", "char", "line"
	outputSelectMode       string
	outputSelectAnchorLine int
	outputSelectAnchorCol  int

	// pending key for multi-key sequences (e.g. "gg")
	outputPendingKey string

	// outputLines stores the full sanitized content lines for cursor bounds and access
	outputLines []string
	// outputRaw stores the unsanitized response body for formatting toggles
	outputRaw string
	// jsonPretty toggles JSON pretty-print in output
	jsonPretty bool

	// New request section
	requestSection struct {
		selectedTab   int
		tabs          []string
		params        []*RequestParam
		headers       []*RequestHeader
		editingParam  int
		editingHeader int
		paramsView    viewport.Model
		headersView   viewport.Model
	}

	// Workspace: multiple request windows
	windows       []*RequestWindow
	currentWindow int

	// Modal state
	showModal     bool
	modalSelected int

	// Menu modal state
	showMenuModal bool
	menuSelected  int

	// Env modal state
	showEnvModal bool
	envSelected  int
	envVars      []EnvVar // read-only from .env files
	tuiVars      []EnvVar // user-created, editable

	// Env creation state
	creatingEnv    bool
	envKeyInput    textinput.Model
	envValueInput  textinput.Model
	envCreatingKey bool // true = editing key, false = editing value

	// Env editing state
	editingEnv    bool
	editingEnvIdx int
	editingKey    bool // true = editing key, false = editing value

	// Config
	cfg config.Config
}

type Method struct {
	Name  string
	Color lipgloss.Color
}

type ApiResponse struct {
	statusCode int
	status     string
	body       string
	duration   string
	err        error
}

type streamResetMsg struct {
	id int64
}

type streamHeaderMsg struct {
	id                 int64
	statusCode         int
	status             string
	ttfb               string
	showStreamControls bool
}

type streamDataMsg struct {
	id    int64
	chunk string
}

type streamDoneMsg struct {
	id       int64
	duration string
	err      error
}

type ApiHeaders struct {
	Key   string
	Value string
}

type RequestParam struct {
	Key    string
	Value  string
	Inputs textinput.Model
}

type RequestHeader struct {
	Key    string
	Value  string
	Inputs textinput.Model
}

type RequestWindow struct {
	Method        int
	URL           string
	Body          string
	Headers       []*RequestHeader
	Params        []*RequestParam
	SelectedTab   int
	StatusCode    int
	Status        string
	ResponseTime  string
	OutputContent string
	OutputRaw     string
	JsonPretty    bool
}

func initModel() *model {
	methods := []*Method{
		{Name: "GET", Color: lipgloss.Color("#b5e48c")},
		{Name: "POST", Color: lipgloss.Color("#ffe566")},
		{Name: "PUT", Color: lipgloss.Color("#8ecae6")},
		{Name: "PATCH", Color: lipgloss.Color("#cdb4db")},
		{Name: "DELETE", Color: lipgloss.Color("#ef233c")},
	}

	s := InitStyles()

	// Initialize with default dimensions
	body := NewTextarea()
	body.SetWidth(80)  // Default width
	body.SetHeight(20) // Default height

	paramsView := viewport.New(0, 0)
	headersView := viewport.New(0, 0)

	// Initialize default headers
	defaultHeaders := []*RequestHeader{
		{
			Key:   "Content-Type",
			Value: "application/json",
			Inputs: func() textinput.Model {
				input := textinput.New()
				input.Placeholder = "Key: Value"
				input.Prompt = "> "
				input.SetValue("Content-Type: application/json")
				return input
			}(),
		},
	}

	m := &model{
		focus:          "url",
		mode:           "normal",
		methods:        methods,
		selectedMethod: 0,
		url:            textinput.New(),
		body:           body,
		output:         viewport.New(0, 0),
		methodStyles:   s,
		urlStyles:      s,
		outputStyles:   s,
		requestStyles:  s,
		requestSection: struct {
			selectedTab   int
			tabs          []string
			params        []*RequestParam
			headers       []*RequestHeader
			editingParam  int
			editingHeader int
			paramsView    viewport.Model
			headersView   viewport.Model
		}{
			selectedTab:   0,
			tabs:          []string{"Body", "Headers", "Params"},
			params:        []*RequestParam{},
			headers:       defaultHeaders,
			editingParam:  -1,
			editingHeader: -1,
			paramsView:    paramsView,
			headersView:   headersView,
		},
		windows: []*RequestWindow{
			{
				Method:      0,
				URL:         "",
				Body:        "",
				Headers:     copyHeaders(defaultHeaders),
				Params:      []*RequestParam{},
				SelectedTab: 0,
			},
		},
		currentWindow:  0,
		showModal:      false,
		modalSelected:  0,
		showEnvModal:   false,
		envSelected:    0,
		envVars:        loadEnvFiles(),
		cfg:            config.Load(),
		creatingEnv:    false,
		envKeyInput:    textinput.New(),
		envValueInput:  textinput.New(),
		envCreatingKey: true,
	}

	m.load()
	return m
}

func newRequestParam() *RequestParam {
	inputs := textinput.New()
	inputs.Placeholder = "Key=Value"
	inputs.Prompt = "> "
	return &RequestParam{
		Inputs: inputs,
	}
}

func newRequestHeader() *RequestHeader {
	inputs := textinput.New()
	inputs.Placeholder = "Key: Value"
	inputs.Prompt = "> "
	return &RequestHeader{
		Inputs: inputs,
	}
}

func copyHeaders(src []*RequestHeader) []*RequestHeader {
	dst := make([]*RequestHeader, len(src))
	for i, h := range src {
		dst[i] = &RequestHeader{Key: h.Key, Value: h.Value}
	}
	return dst
}

func (m *model) saveCurrentWindow() {
	if m.currentWindow < 0 || m.currentWindow >= len(m.windows) {
		return
	}
	w := m.windows[m.currentWindow]
	w.Method = m.selectedMethod
	w.URL = m.url.Value()
	w.Body = m.body.Value()
	w.SelectedTab = m.requestSection.selectedTab
	w.StatusCode = m.statusCode
	w.Status = m.status
	w.ResponseTime = m.responseTime
	w.OutputContent = m.output.View()
	w.OutputRaw = m.outputRaw
	w.JsonPretty = m.jsonPretty

	w.Headers = make([]*RequestHeader, len(m.requestSection.headers))
	for i, h := range m.requestSection.headers {
		w.Headers[i] = &RequestHeader{Key: h.Key, Value: h.Value}
	}

	w.Params = make([]*RequestParam, len(m.requestSection.params))
	for i, p := range m.requestSection.params {
		w.Params[i] = &RequestParam{Key: p.Key, Value: p.Value}
	}
}

func (m *model) loadWindow(idx int) {
	if idx < 0 || idx >= len(m.windows) {
		return
	}
	m.currentWindow = idx
	w := m.windows[idx]

	m.selectedMethod = w.Method
	m.url.SetValue(w.URL)
	m.body.SetValue(w.Body)
	m.requestSection.selectedTab = w.SelectedTab
	m.statusCode = w.StatusCode
	m.status = w.Status
	m.responseTime = w.ResponseTime
	m.outputRaw = w.OutputRaw
	m.jsonPretty = w.JsonPretty
	if m.outputRaw != "" {
		m.setOutput(m.outputRaw)
	} else {
		m.setOutput(w.OutputContent)
	}

	m.requestSection.headers = make([]*RequestHeader, len(w.Headers))
	for i, h := range w.Headers {
		m.requestSection.headers[i] = &RequestHeader{Key: h.Key, Value: h.Value}
	}

	m.requestSection.params = make([]*RequestParam, len(w.Params))
	for i, p := range w.Params {
		m.requestSection.params[i] = &RequestParam{Key: p.Key, Value: p.Value}
	}
}

func (m *model) newWindow() {
	m.saveCurrentWindow()
	w := &RequestWindow{
		Method:      0,
		URL:         "",
		Body:        "",
		Headers:     []*RequestHeader{},
		Params:      []*RequestParam{},
		SelectedTab: 0,
	}
	m.windows = append(m.windows, w)
	m.loadWindow(len(m.windows) - 1)
	m.save()
}

func (m *model) deleteWindow(idx int) {
	if idx < 0 || idx >= len(m.windows) || len(m.windows) <= 1 {
		return
	}
	m.windows = append(m.windows[:idx], m.windows[idx+1:]...)
	if m.currentWindow >= len(m.windows) {
		m.currentWindow = len(m.windows) - 1
	}
	m.loadWindow(m.currentWindow)
	m.save()
}
