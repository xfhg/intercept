package cmd

import "github.com/gookit/color"

var (
	scanPath  string
	scanTags  string
	scanBreak string
	scanTurbo string
	fatal     = false
	warning   = false
)

var line = "├────────────────────────────────────────────────────────────"

type allStats struct {
	Total int `json:"Total"`
	Clean int `json:"Clean"`
	Dirty int `json:"Dirty"`
	Fatal int `json:"Fatal"`
}

type ScannedFile struct {
	Path   string `json:"path"`
	Sha256 string `json:"sha256"`
}

type Rule struct {
	ID                string   `yaml:"id"`
	Name              string   `yaml:"name"`
	Description       string   `yaml:"description"`
	Solution          string   `yaml:"solution"`
	Error             string   `yaml:"error"`
	Type              string   `yaml:"type"`
	Environment       string   `yaml:"environment"`
	Enforcement       bool     `yaml:"enforcement"`
	Fatal             bool     `yaml:"fatal"`
	Tags              string   `yaml:"tags,omitempty"`
	Impact            string   `yaml:"impact,omitempty"`
	Confidence        string   `yaml:"confidence,omitempty"`
	Api_Endpoint      string   `yaml:"api_endpoint,omitempty"`
	Api_Request       string   `yaml:"api_request,omitempty"`
	Api_Insecure      bool     `yaml:"api_insecure"`
	Api_Body          string   `yaml:"api_body,omitempty"`
	Api_Auth          string   `yaml:"api_auth,omitempty"`
	Api_Auth_Basic    *string  `yaml:"api_auth_basic,omitempty"`
	Api_Auth_Token    *string  `yaml:"api_auth_token,omitempty"`
	Api_Trace         bool     `yaml:"api_trace,omitempty"`
	Filepattern       string   `yaml:"filepattern,omitempty"`
	Yml_Filepattern   string   `yaml:"yml_filepattern,omitempty"`
	Yml_Structure     string   `yaml:"yml_structure,omitempty"`
	Ini_Filepattern   string   `yaml:"ini_filepattern,omitempty"`
	Ini_Structure     string   `yaml:"ini_structure,omitempty"`
	Toml_Filepattern  string   `yaml:"toml_filepattern,omitempty"`
	Toml_Structure    string   `yaml:"toml_structure,omitempty"`
	Json_Filepattern  string   `yaml:"json_filepattern,omitempty"`
	Json_Structure    string   `yaml:"json_structure,omitempty"`
	Rego_Filepattern  string   `yaml:"rego_filepattern,omitempty"`
	Rego_Policy_File  string   `yaml:"rego_policy_file,omitempty"`
	Rego_Policy_Data  string   `yaml:"rego_policy_data,omitempty"`
	Rego_Policy_Query string   `yaml:"rego_policy_query,omitempty"`
	Patterns          []string `yaml:"patterns,omitempty"`
}

type allRules struct {
	Banner           string `yaml:"banner"`
	Rules            []Rule `yaml:"Rules"`
	ExitCritical     string `yaml:"exitcritical"`
	ExitWarning      string `yaml:"exitwarning"`
	ExitClean        string `yaml:"exitclean"`
	Exceptions       []int  `yaml:"exceptions"`
	ExceptionMessage string `yaml:"exceptionmessage"`
}

var (
	stats           allStats
	rules           *allRules
	colorRedBold    = color.New(color.Red, color.OpBold)
	colorGreenBold  = color.New(color.Green, color.OpBold)
	colorYellowBold = color.New(color.Yellow, color.OpBold)
	colorBlueBold   = color.New(color.Blue, color.OpBold)
	colorBold       = color.New(color.OpBold)
	colorYellow     = color.New(color.Yellow)
)
