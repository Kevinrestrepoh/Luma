package config

type Header struct {
	Key   string `toml:"key"`
	Value string `toml:"value"`
}

type Param struct {
	Key   string `toml:"key"`
	Value string `toml:"value"`
}

type Window struct {
	Method        int      `toml:"method"`
	URL           string   `toml:"url"`
	Body          string   `toml:"body"`
	Headers       []Header `toml:"headers"`
	Params        []Param  `toml:"params"`
	SelectedTab   int      `toml:"selected_tab"`
	StatusCode    int      `toml:"status_code"`
	Status        string   `toml:"status"`
	ResponseTime  string   `toml:"response_time"`
	OutputContent string   `toml:"output_content"`
	OutputRaw     string   `toml:"output_raw"`
	JsonPretty    bool     `toml:"json_pretty"`
}
