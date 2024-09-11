package cmd

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

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
		MsgExitClean    string `yaml:"MsgExitClean"`
		MsgExitWarning  string `yaml:"MsgExitWarning"`
		MsgExitCritical string `yaml:"MsgExitCritical"`
	} `yaml:"Metadata"`
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

type PolicyFile struct {
	Config    Config   `yaml:"Config"`
	Version   string   `yaml:"Version"`
	Namespace string   `yaml:"Namespace"`
	Policies  []Policy `yaml:"Policies"`
}

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

	// Generate intercept_id for each policy, add its own ID as a tag for easy filtering with tags flag
	for i := range policyFile.Policies {
		policyFile.Policies[i].ID = NormalizePolicyName(policyFile.Policies[i].ID)
		policyFile.Policies[i].InterceptID = intercept_run_id + "-" + NormalizeFilename(policyFile.Policies[i].ID)
		policyFile.Policies[i].Metadata.Tags = append(policyFile.Policies[i].Metadata.Tags, policyFile.Policies[i].ID)
	}

	return &policyFile, nil
}

// Policy store

var (
	policyStore = sync.Map{}
)

// StorePolicyInCache saves a Policy struct in the global store with the given key
func StorePolicyInCache(key string, policy Policy) {
	policyStore.Store(key, policy)
}

// LoadPolicyFromCache retrieves a Policy struct from the global store by key
func LoadPolicyFromCache(key string) (Policy, bool) {
	value, ok := policyStore.Load(key)
	if !ok {
		return Policy{}, false
	}
	policy, ok := value.(Policy)
	return policy, ok
}

// DeletePolicyFromCache removes a Policy struct from the global store by key
func DeletePolicyFromCache(key string) {
	policyStore.Delete(key)
}

// ClearAllPoliciesFromCache removes all Policy structs from the global store
func ClearAllPoliciesFromCache() {
	policyStore.Range(func(key, value interface{}) bool {
		policyStore.Delete(key)
		return true
	})
}

// ListAllPolicyCacheKeys returns a slice of all policy keys in the store
func ListAllPolicyCacheKeys() []string {
	var keys []string
	policyStore.Range(func(key, value interface{}) bool {
		if k, ok := key.(string); ok {
			keys = append(keys, k)
		}
		return true
	})
	return keys
}

// LoadAllPoliciesFromCache returns a map of all policies in the store
func LoadAllPoliciesFromCache() map[string]Policy {
	policies := make(map[string]Policy)
	policyStore.Range(func(key, value interface{}) bool {
		if k, ok := key.(string); ok {
			if p, ok := value.(Policy); ok {
				policies[k] = p
			}
		}
		return true
	})
	return policies
}

// UpdatePolicyInCache updates an existing Policy in the store or adds it if it doesn't exist
func UpdatePolicyInCache(key string, policy Policy) {
	policyStore.Store(key, policy)
}

// PolicyExistsInCache checks if a Policy with the given key exists in the store
func PolicyExistsInCache(key string) bool {
	_, exists := policyStore.Load(key)
	return exists
}

// GetPolicyCacheCount returns the number of policies in the store
func GetPolicyCacheCount() int {
	count := 0
	policyStore.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
