# ASSURE API Policies

API-type policies build upon the foundation established by ASSURE policies, incorporating the significant advantage of directly specifying within the policy all the necessary details for retrieving target information from API endpoints—typically those related to configuration. 

This information, once gathered, is subsequently analyzed using ASSURE patterns to affirm compliance and ensure that the output values from these API calls align with the expected standards and configurations.

This enhancement effectively bridges the gap between static policy enforcement and dynamic data retrieval, allowing for a seamless integration of real-time API data into the compliance verification process. By leveraging API-type policies, organizations can extend their compliance checks beyond static codebase analysis to include live configuration data, ensuring a comprehensive coverage of compliance and security standards across both code and configuration environments.

Such policies empower organizations to automate the monitoring and validation of configurations across various services and platforms, ensuring that the infrastructure not only remains compliant at the time of deployment but also continues to adhere to required standards as configurations evolve and change over time.

::: info
API Policy type benefits from both **REGEX** and **SCHEMA** capabilities from ASSURE policies
:::


## Examples

```yaml{3,19-27,53-55}
Policies:
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



::: tip
To pass the secrets for the API auth (in this case TOKEN), INTERCEPT will read them directly from the environment variables
:::



## Struct and API Auth

```go{5-6,11,15,18}
type APIConfig struct {
	Endpoint     string            `yaml:"endpoint"`
	Insecure     bool              `yaml:"insecure"`
	ResponseType string            `yaml:"response_type"`
	Method       string            `yaml:"method"`
	Body         string            `yaml:"body"`
	Auth         map[string]string `yaml:"auth"`
}

switch authType {
	case "basic":
		username := os.Getenv(auth["username_env"])
		password := os.Getenv(auth["password_env"])
		req.SetHeader("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
	case "bearer":
		token := os.Getenv(auth["token_env"])
		req.SetHeader("Authorization", "Bearer "+token)
	case "api_key":
		key := os.Getenv(auth["key_env"])
		req.SetHeader(auth["header"], key)
	default:
		return fmt.Errorf("unsupported authentication type: %s", authType)
	}


```


