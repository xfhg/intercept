
# A starting point

create your first policy file (mypolicy.yaml) and add the following :

```yaml

Banner: |

  | Starting point 1 SCAN and 1 COLLECT RULE

Rules:
  - name: Passwords being used in URIs
    id: 100
    description: Detecting the pattern "protocol://username:password@host"
    error: This violation immediately blocks your code deployment
    tags: URI
    type: scan
    fatal: true
    enforcement: true
    environment: all
    confidence: high
    patterns:
      - \s*^(.*):\/\/([^:]*):([^@]*)@(.*)$

  - name: Collect proxy modifications on your bootstrap
    id: 800
    description: The following proxy modifications were collected
    type: collect
    tags: AWS,AZURE
    patterns:
      - \b(?:http_proxy|https_proxy|ftp_proxy|socks_proxy|no_proxy|HTTP_PROXY|HTTPS_PROXY|FTP_PROXY|SOCKS_PROXY|NO_PROXY)\s*=\s*['"]?(https?|socks[45])://(?:[^\s'"]+)


ExitCritical: "Critical irregularities found in your code"
ExitWarning: "Irregularities found in your code"
ExitClean: "Clean report"

```



## Policy Struct Schema


```go
type Rule struct {
	ID               int      `yaml:"id"`
	Name             string   `yaml:"name"`
	Description      string   `yaml:"description"`
	Solution         string   `yaml:"solution"`
	Error            string   `yaml:"error"`
	Type             string   `yaml:"type"`
	Environment      string   `yaml:"environment"`
	Enforcement      bool     `yaml:"enforcement"`
	Fatal            bool     `yaml:"fatal"`
    
	Tags             string   `yaml:"tags,omitempty"`
	Impact           string   `yaml:"impact,omitempty"`
	Confidence       string   `yaml:"confidence,omitempty"`
	
    Api_Endpoint     string   `yaml:"api_endpoint,omitempty"`
	Api_Request      string   `yaml:"api_request,omitempty"`
	Api_Insecure     bool     `yaml:"api_insecure"`
	Api_Body         string   `yaml:"api_body,omitempty"`
	Api_Auth         string   `yaml:"api_auth,omitempty"`
	Api_Auth_Basic   *string  `yaml:"api_auth_basic,omitempty"`
	Api_Auth_Token   *string  `yaml:"api_auth_token,omitempty"`
	Api_Trace        bool     `yaml:"api_trace,omitempty"`
	
    Filepattern      string   `yaml:"filepattern,omitempty"`
	
    Yml_Filepattern  string   `yaml:"yml_filepattern,omitempty"`
	Yml_Structure    string   `yaml:"yml_structure,omitempty"`

	Ini_Filepattern string   `yaml:"ini_filepattern,omitempty"`
	Ini_Structure   string   `yaml:"ini_structure,omitempty"`
	
    Toml_Filepattern string   `yaml:"toml_filepattern,omitempty"`
	Toml_Structure   string   `yaml:"toml_structure,omitempty"`
	
    Json_Filepattern string   `yaml:"json_filepattern,omitempty"`
	Json_Structure   string   `yaml:"json_structure,omitempty"`

	Rego_Filepattern      string   `yaml:"rego_filepattern,omitempty"`
	Rego_Policy_File      string   `yaml:"rego_policy_file,omitempty"`
	Rego_Policy_Data      string   `yaml:"rego_policy_data,omitempty"`
	Rego_Policy_Query     string   `yaml:"rego_policy_query,omitempty"`

	Patterns              []string `yaml:"patterns,omitempty"`
}
```




## Run it

```sh
docker pull ghcr.io/xfhg/intercept:latest
docker run -v --rm -w $PWD -v $PWD:$PWD -e TERM=xterm-256color ghcr.io/xfhg/intercept intercept config -r
docker run -v --rm -w $PWD -v $PWD:$PWD -e TERM=xterm-256color ghcr.io/xfhg/intercept intercept config -a mypolicy.yaml
docker run -v --rm -w $PWD -v $PWD:$PWD -e TERM=xterm-256color ghcr.io/xfhg/intercept intercept audit -t yourtargetfolder/
```

