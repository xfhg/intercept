
# Policy Structure


```go

type PolicyFile struct {
	Config    Config   `yaml:"Config"`
	Version   string   `yaml:"Version"`
	Namespace string   `yaml:"Namespace"`
	Policies  []Policy `yaml:"Policies"`
}

// Intercept Config

type Config struct {
	System struct {
		RGVersion        string `yaml:"RGVersion"`
		GossVersion      string `yaml:"GossVersion"`
		InterceptVersion string `yaml:"InterceptVersion"`
	} `yaml:"System"`
	Flags struct {
		OutputType     string   `yaml:"output_type"`
		Target         string   `yaml:"target"`
		Ignore         []string `yaml:"ignore"`
		Tags           []string `yaml:"tags"`
		PolicySchedule string   `yaml:"policy_schedule"`
		ReportSchedule string   `yaml:"report_schedule"`
	} `yaml:"Flags"`
	Metadata struct {
    ...
	} `yaml:"Metadata"`
	Hooks []HookConfig `yaml:"Hooks"`
}

// Webhooks

type HookConfig struct {
	Name           string            `yaml:"name"`
	Endpoint       string            `yaml:"endpoint"`
	Insecure       bool              `yaml:"insecure"`
	Auth           map[string]string `yaml:"auth"`
	Method         string            `yaml:"method"`
	Headers        map[string]string `yaml:"headers"`
	RetryAttempts  int               `yaml:"retry_attempts"`
	RetryDelay     string            `yaml:"retry_delay"`
	TimeoutSeconds int               `yaml:"timeout_seconds"`
	EventTypes     []string          `yaml:"event_types"`
}

// Policies

type Policy struct {
	ID          string        `yaml:"id"`
	InterceptID string        `yaml:"intercept_id,omitempty"`
	RunID       string        `yaml:"intercept_run_id,omitempty"`
	Schedule    string        `yaml:"schedule"`
	Type        string        `yaml:"type"`
	Enforcement []Enforcement `yaml:"enforcement"`
	Metadata    Metadata      `yaml:"metadata"`
	FilePattern string        `yaml:"filepattern"`
	Schema      Schema        `yaml:"_schema"`
	Rego        Rego          `yaml:"_rego"`
	Regex       []string      `yaml:"_regex"`
	API         APIConfig     `yaml:"_api"`
	Runtime     Runtime       `yaml:"_runtime"`
}

type Enforcement struct {
	Environment string `yaml:"environment"`
	Fatal       string `yaml:"fatal"`
	Exceptions  string `yaml:"exceptions"`
	Confidence  string `yaml:"confidence"`
}

type Metadata struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
	Score       string   `yaml:"score"`
	MsgSolution string   `yaml:"msg_solution"`
	MsgError    string   `yaml:"msg_error"`
	TargetInfo  []string `yaml:"target_info,omitempty"`
}

type Schema struct {
	Structure string `yaml:"structure"`
	Strict    bool   `yaml:"strict"`
	Patch     bool   `yaml:"patch"`
}

type Rego struct {
	PolicyFile  string `yaml:"policy_file"`
	PolicyData  string `yaml:"policy_data"`
	PolicyQuery string `yaml:"policy_query"`
}

type APIConfig struct {
	Endpoint     string            `yaml:"endpoint"`
	Insecure     bool              `yaml:"insecure"`
	ResponseType string            `yaml:"response_type"`
	Method       string            `yaml:"method"`
	Body         string            `yaml:"body"`
	Auth         map[string]string `yaml:"auth"`
}

type Runtime struct {
	Config  string `yaml:"config"`
	Observe string `yaml:"observe"`
}

```












