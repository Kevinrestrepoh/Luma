package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	width          int
	height         int
	methods        []*Method
	selectedMethod int

	focus string
	mode  string

	url          textinput.Model
	body         CustomTextarea
	statusCode   int
	status       string
	output       viewport.Model
	responseTime string

	methodStyles  *Styles
	urlStyles     *Styles
	outputStyles  *Styles
	requestStyles *Styles

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

func initModel() *model {
	methods := []*Method{
		{Name: "GET", Color: lipgloss.Color("#b5e48c")},
		{Name: "POST", Color: lipgloss.Color("#ffe566")},
		{Name: "PUT", Color: lipgloss.Color("#8ecae6")},
		{Name: "DELETE", Color: lipgloss.Color("#ef233c")},
	}

	s := InitStyles()

	// Initialize with default dimensions
	body := NewCustomTextarea()
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

	return &model{
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
	}
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
