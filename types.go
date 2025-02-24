package main

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	width          int
	height         int
	methods        []*Method
	selectedMethod int

	focus string
	mode  string

	url    textinput.Model
	body   textarea.Model
	output string

	methodStyles *Styles
	urlStyles    *Styles
	bodyStyles   *Styles
	outputStyles *Styles
}

type Method struct {
	Name  string
	Color lipgloss.Color
}

func initModel() *model {
	// methods := []string{"GET", "POST", "PUT", "DELETE"}
	methods := []*Method{
		{Name: "GET", Color: lipgloss.Color("#b5e48c")},
		{Name: "POST", Color: lipgloss.Color("#ffe566")},
		{Name: "PUT", Color: lipgloss.Color("#8ecae6")},
		{Name: "DELETE", Color: lipgloss.Color("#ef233c")},
	}

	s := InitStyles()

	return &model{
		focus:          "url",
		mode:           "normal",
		methods:        methods,
		selectedMethod: 0,
		url:            textinput.New(),
		body:           textarea.New(),
		methodStyles:   s,
		urlStyles:      s,
		bodyStyles:     s,
		outputStyles:   s,
	}
}
