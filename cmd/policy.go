package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-resty/resty/v2"
	"gopkg.in/yaml.v3"
)

type PolicyFile struct {
	Config     Config      `yaml:"Config"`
	Version    string      `yaml:"Version"`
	Namespace  string      `yaml:"Namespace"`
	Policies   []Policy    `yaml:"Policies"`
	SARIFRules []SARIFRule `json:"sarif_rules,omitempty"`
}

type Config struct {
	System struct {
		RGVersion        string `yaml:"RGVersion,omitempty"`
		GossVersion      string `yaml:"GossVersion,omitempty"`
		InterceptVersion string `yaml:"InterceptVersion,omitempty"`
	} `yaml:"System,omitempty"`
	Flags struct {
		OutputType     []string `yaml:"output_type,omitempty"`
		Target         string   `yaml:"target,omitempty"`
		Index          string   `yaml:"index,omitempty"`
		Ignore         []string `yaml:"ignore,omitempty"`
		Tags           []string `yaml:"tags,omitempty"`
		PolicySchedule string   `yaml:"policy_schedule,omitempty"`
		ReportSchedule string   `yaml:"report_schedule,omitempty"`
		WebhookSecret  string   `yaml:"webhook_secret_env,omitempty"`
		RemoteAuth     []string `yaml:"remote_auth,omitempty"`
	} `yaml:"Flags,omitempty"`
	Metadata struct {
		HostOS          string `yaml:"host_os,omitempty"`
		HostMAC         string `yaml:"host_mac,omitempty"`
		HostARCH        string `yaml:"host_arch,omitempty"`
		HostNAME        string `yaml:"host_name,omitempty"`
		HostFingerprint string `yaml:"host_fingerprint,omitempty"`
		HostInfo        string `yaml:"host_info,omitempty"`
		MsgExitClean    string `yaml:"MsgExitClean,omitempty"`
		MsgExitWarning  string `yaml:"MsgExitWarning,omitempty"`
		MsgExitCritical string `yaml:"MsgExitCritical,omitempty"`
	} `yaml:"Metadata,omitempty"`
	Hooks []HookConfig `yaml:"Hooks"`
}

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
	HelpURL     string   `yaml:"help_url"`
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

type PolicySourceType int

const (
	LocalFile PolicySourceType = iota
	RemoteURL
)

func LoadPolicyFile(filename string) (*PolicyFile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var policyFile PolicyFile
	err = yaml.Unmarshal(data, &policyFile)
	if err != nil {
		return nil, err
	}

	// log.Debug().Interface("raw config", policyFile.Config).Msg("Raw Config data")

	rules := make([]SARIFRule, 0, len(policyFile.Policies))

	// Generate intercept_id for each policy, add its own ID as a tag for easy filtering with tags flag
	for i := range policyFile.Policies {
		policyFile.Policies[i].ID = NormalizePolicyName(policyFile.Policies[i].ID)
		policyFile.Policies[i].InterceptID = intercept_run_id + "-" + NormalizeFilename(policyFile.Policies[i].ID)
		policyFile.Policies[i].Metadata.Tags = append(policyFile.Policies[i].Metadata.Tags, policyFile.Policies[i].ID)

		// Generate rule entry for SARIF
		rule := SARIFRule{
			ID: policyFile.Policies[i].ID,
			ShortDescription: ShortDescription{
				Text: policyFile.Policies[i].Metadata.Description,
			},
			FullDescription: &FullDescription{
				Text: policyFile.Policies[i].Metadata.MsgError,
			},
			HelpURI: policyFile.Policies[i].Metadata.HelpURL,
			Help: &Help{
				Text: policyFile.Policies[i].Metadata.MsgSolution,
			},
			Properties: Properties{
				Category: policyFile.Policies[i].Metadata.Tags[0],
				Tags:     policyFile.Policies[i].Metadata.Tags,
			},
		}
		rules = append(rules, rule)

	}

	// Add rules to policyFile
	policyFile.SARIFRules = rules

	return &policyFile, nil
}

// Load Remote

// LoadRemotePolicy loads a policy file from a remote HTTPS endpoint
func LoadRemotePolicy(url string, expectedChecksum string) (*PolicyFile, error) {
	// Create a temporary directory to store the downloaded file
	remoteDir := filepath.Join(outputDir, "_remote")
	err := os.MkdirAll(remoteDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(remoteDir) // Clean up the temporary directory when done

	// Generate a temporary file name
	tempFile := filepath.Join(remoteDir, "remote_policy.yaml")

	// Create a resty client
	client := resty.New()

	// Download the file
	resp, err := client.R().SetOutput(tempFile).Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download policy file: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to download policy file: HTTP status %d", resp.StatusCode())
	}

	// If a checksum is provided, validate it
	if expectedChecksum != "" {
		actualChecksum, err := calculateSHA256(tempFile)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to calculate policy checksum")
		}

		if actualChecksum != expectedChecksum {
			log.Fatal().Msgf("Policy checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)

		}
	}

	// Load the policy file
	policyFile, err := LoadPolicyFile(tempFile)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load policy file")
	}

	return policyFile, nil
}

func DeterminePolicySource(input string) (PolicySourceType, string, error) {
	// First, check if it's a valid URL
	if isURL(input) {
		return RemoteURL, input, nil
	}

	// If not a URL, treat it as a file path
	absPath, err := filepath.Abs(input)
	if err != nil {
		return LocalFile, "", err
	}

	// Check if the file exists
	_, err = os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return LocalFile, "", fmt.Errorf("file does not exist: %s", absPath)
		}
		return LocalFile, "", err
	}

	return LocalFile, absPath, nil
}

// Deprecated: the real cache implementation lives in cmd/cache.go
// These forwarders maintain API compatibility if other files call them with the old signatures.

// StorePolicyInCacheLegacy saves a Policy struct in the global store with the given key
// Deprecated: Use StorePolicyInCache(path, policy, isDirectory) from cmd/cache.go instead
func StorePolicyInCacheLegacy(key string, policy Policy) {
	// default: treat as file. callers that know it's a directory should call StorePolicyInCacheDir
	StorePolicyInCacheDir(key, policy)
}

// StorePolicyInCacheDir stores policy for a directory key
// Deprecated: Use StorePolicyInCache(path, policy, true) from cmd/cache.go instead
func StorePolicyInCacheDir(key string, policy Policy) {
	StorePolicyInCache(key, policy, true)
}

// StorePolicyInCacheFile stores policy for a file key
// Deprecated: Use StorePolicyInCache(path, policy, false) from cmd/cache.go instead
func StorePolicyInCacheFile(key string, policy Policy) {
	StorePolicyInCache(key, policy, false)
}

// GetPolicyCacheCount returns the number of policies in the store
func GetPolicyCacheCount() int {
	return len(ListAllPolicyCacheKeys())
}
