
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
  Observe     string        `yaml:"observe"`
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

## Example policies

::: code-group

```yaml [SCAN.yaml]
Policies:

  - id: "SCAN-001 Private Keys"
    type: "scan"
    enforcement:
      - environment: "production"
        fatal: "true"
        exceptions: "false"
        confidence: "high"
      - environment: "development"
        fatal: "true"
        exceptions: "false"
        confidence: "high"
    metadata:
      name: "Detect private keys"
      description: "Scan for potential private key leaks in the codebase"
      msg_solution: "Remove the private key and use secure key management practices."
      msg_error: "Private key detected in the codebase."
      tags:
        - "security"
        - "encryption"
      score: "9"
    _regex:
      - \s*(-----BEGIN PRIVATE KEY-----)
      - \s*(-----BEGIN RSA PRIVATE KEY-----)
      - \s*(-----BEGIN DSA PRIVATE KEY-----)
      - \s*(-----BEGIN EC PRIVATE KEY-----)
      - \s*(-----BEGIN OPENSSH PRIVATE KEY-----)
      - \s*(-----BEGIN PGP PRIVATE KEY BLOCK-----)

  - id: "SCAN-002 API Keys"
    type: "scan"
    enforcement:
      - environment: "all"
        fatal: "true"
        exceptions: "false"
        confidence: "medium"
    metadata:
      name: "Detect API keys"
      description: "Scan for potential API key leaks in the codebase"
      msg_solution: "Remove the API key and use secure key management practices."
      msg_error: "Potential API key detected in the codebase."
      tags:
        - "security"
        - "api"
      score: "8"
    _regex:
      - \b[A-Za-z0-9]{20,}\b
      - api[_-]?key[_-]?=\s*['"]?\w+['"]?
```

```yaml [ASSURE-REGEX.yaml]
Policies:

  - id: "ASSURE-001 Required Security Settings"
    type: "assure"
    filepattern: "config.*\\.(json|yaml|ini)$"
    enforcement:
      - environment: "production"
        fatal: "true"
        exceptions: "false"
        confidence: "high"
      - environment: "development"
        fatal: "false"
        exceptions: "false"
        confidence: "high"
    metadata:
      name: "Ensure required security settings"
      description: "Verify that configuration files contain required security settings"
      msg_solution: "Add the missing security settings to the configuration file."
      msg_error: "Configuration file is missing required security settings."
      tags:
        - "security"
        - "config"
      score: "8"
    _regex:
      - "ssl_enabled:\\s*true"
      - "use_encryption:\\s*true"
      - "min_password_length:\\s*12"
      - "enable_2fa:\\s*true"

  - id: "ASSURE-002 Logging Configuration"
    type: "assure"
    filepattern: "log.*\\.(json|yaml|ini)$"
    enforcement:
      - environment: "all"
        fatal: "true"
        exceptions: "false"
        confidence: "high"
    metadata:
      name: "Ensure proper logging configuration"
      description: "Verify that logging configuration files contain required settings"
      msg_solution: "Add the missing logging settings to the configuration file."
      msg_error: "Logging configuration file is missing required settings."
      tags:
        - "logging"
        - "config"
      score: "7"
    _regex:
      - "log_level:\\s*(info|debug|warn|error)"
      - "log_format:\\s*json"
      - "log_retention_days:\\s*\\d+"
      - "enable_audit_logs:\\s*true"

  - id: "ASSURE-003 Required Environment Variables"
    type: "assure"
    filepattern: ".*\\.env"
    enforcement:
      - environment: "all"
        fatal: "true"
        exceptions: "false"
        confidence: "high"
    metadata:
      name: "Ensure required environment variables"
      description: "Verify that .env files contain all required environment variables"
      msg_solution: "Add the missing environment variables to the .env file."
      msg_error: ".env file is missing required environment variables."
      tags:
        - "env"
        - "config"
      score: "7"
    _regex:
      - "^DATABASE_URL="
      - "^API_KEY="
      - "^NODE_ENV="
      - "^PORT="
```



```yaml [SCHEMA-FILETYPE.yaml]
Policies:

  - id: "JSON-001"
    type: "json"
    filepattern: "example\\.json$"
    enforcement:
      - environment: "all"
        fatal: "true"
        exceptions: "false"
        confidence: "high"
    metadata:
      name: "Strict Application Configuration Schema"
      description: "Enforce strict schema compliance on application configuration JSON files"
      msg_solution: "Update the configuration file to exactly match the required schema, removing any extra fields and correcting data types."
      msg_error: "Application configuration JSON file does not comply with the strict required schema."
      tags:
        - "config"
        - "json"
        - "strict"
      score: "9"
    _schema:
      strict: true
      patch: false
      structure: |
        {
          app: {
            name: string
            version: string & =~"^\\d+\\.\\d+\\.\\d+$"
            port: int
          }
          database: {
            host: string
            port: int
            name: string
            user: string
          }
          logging: {
            level: "debug" | "info" | "warn" | "error"
            format: "json" | "text"
          }
          features: {
            featureA: bool
            featureB: bool
            flags: {
              blocker: bool
            }
          }
        }

  - id: "JSON-002"
    type: "json"
    filepattern: "example\\.json$"
    enforcement:
      - environment: "all"
        fatal: "false"
        exceptions: "false"
        confidence: "high"
    metadata:
      name: "Flexible Application Configuration Schema"
      description: "Enforce flexible schema compliance on application configuration JSON files"
      msg_solution: "Consider adjusting the configuration file to better match the recommended structure, but additional fields are allowed."
      msg_error: "Application configuration JSON file has some deviations from the recommended schema."
      tags:
        - "config"
        - "json"
        - "flexible"
      score: "7"
    _schema:
      strict: false
      patch: false
      structure: |
        {
          app: {
            name: string
            version: string
            port: string | int
          }
          database: {
            host: string
            port: int
            name: string
            user: string
          }
          logging: {
            level: string
            format: string
          }
          features: {
            featureA: bool
            featureB: bool
            [string]: bool | {...}
          }
        }
```


```yaml [API.yaml]
Policies:

  - id: "API-001"
    type: "api"
    enforcement:
      - environment: "production"
        fatal: "true"
        exceptions: "false"
        confidence: "high"
      - environment: "development"
        fatal: "true"
        exceptions: "false"
        confidence: "high"
    metadata:
      name: "API Regex Compliance"
      description: "Enforce schema compliance on API configuration files"
      msg_solution: "Generic solution message to development issue."
      msg_error: "Generic error message for development issue"
      tags:
        - "config"
        - "ini"
        - "schema"
      confidence: "high"
      score: "8"
    _api:
      endpoint: "https://httpbin.org/user-agent"
      insecure: false
      request: "GET"
      response_type: "application/json"
      auth: 
        type: bearer
        token_env: TOKEN 
    _regex:
      - \s*user-agent\s*

  - id: "API-002"
    type: "api"
    enforcement:
      - environment: "all"
        fatal: "false"
        exceptions: "development"
        msg_solution: "Ensure all required fields are present and comply with the schema."
        msg_error: "API file does not comply with the required schema."
    metadata:
      name: "API Regex Compliance"
      description: "Enforce schema compliance on API configuration files"
      tags:
        - "config"
        - "ini"
        - "schema"
      confidence: "high"
      score: "8"
    _api:
      endpoint: "https://httpbin.org/ip"
      insecure: false
      request: "GET"
      auth: 
        type: bearer
        token_env: TOKEN 
    _regex:
      - \s*user-agent\s*

  - id: "API-003"
    type: "api"
    enforcement:
      - environment: "all"
        fatal: "false"
        exceptions: "development"
    metadata:
      name: "API Regex Compliance"
      description: "Enforce schema compliance on API configuration files"
      msg_solution: "Ensure all required fields are present and comply with the schema."
      msg_error: "API file does not comply with the required schema."
      tags:
        - "config"
        - "ini"
        - "schema"
      confidence: "high"
      score: "8"
    _api:
      endpoint: "https://httpbin.org/bearer"
      insecure: false
      request: "GET"
      auth: 
        type: bearer
        token_env: TOKEN 
    _regex:
      - \"authenticated\"\s*:\s*true\s*,?

  - id: "API-004"
    type: "api"
    enforcement:
      - environment: "all"
        fatal: "false"
        exceptions: "development"

    metadata:
      name: "API Regex Compliance"
      description: "Enforce schema compliance on API configuration files"
      msg_solution: "Ensure all required fields are present and comply with the schema."
      msg_error: "API file does not comply with the required schema."
      tags:
        - "config"
        - "ini"
        - "schema"
      confidence: "high"
      score: "8"
    _api:
      endpoint: "https://httpbin.org/bearer"
      insecure: false
      request: "GET"
      auth: 
        type: bearer
        token_env: TOKEN 
    _schema:
      structure: |
        { authenticated : true}
```

:::




