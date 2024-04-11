# API-ASSURE Type Policies

API-type policies build upon the foundation established by ASSURE policies, incorporating the significant advantage of directly specifying within the policy all the necessary details for retrieving target information from API endpoints—typically those related to configuration. 

This information, once gathered, is subsequently analyzed using ASSURE patterns to affirm compliance and ensure that the output values from these API calls align with the expected standards and configurations.

This enhancement effectively bridges the gap between static policy enforcement and dynamic data retrieval, allowing for a seamless integration of real-time API data into the compliance verification process. By leveraging API-type policies, organizations can extend their compliance checks beyond static codebase analysis to include live configuration data, ensuring a comprehensive coverage of compliance and security standards across both code and configuration environments.

Such policies empower organizations to automate the monitoring and validation of configurations across various services and platforms, ensuring that the infrastructure not only remains compliant at the time of deployment but also continues to adhere to required standards as configurations evolve and change over time.

::: tip
This Policy type gets even better with the addition of CUE lang schema validation or REGO policies.
:::




## Example

### The setup

```sh{4-5}
intercept config -r 
intercept config -a /app/examples/policy/api.yaml

export INTERCEPT_BAUTH=user:pass
intercept api 

cat intercept.api.full.sarif.json
```

::: info
All rule types can be filtered by a combination of TAGS, ENVIRONMENT name and their own ENFORCEMENT levels. Make sure to explore it.
:::


### The Policy

```yaml{7,12-21}

 - name: API value check
    id: 105
    description: Sandbox API check
    error: Misconfiguration or omission
    tags: KEY
    type: api
    fatal: false
    enforcement: true
    environment: all
    confidence: high
    api_endpoint: https://httpbin.org/post
    api_insecure: false
    api_request: POST
    api_body: |
      {"employee":{ "name":"Emma", "age":28, "city":"Boston" }} 
    api_auth: basic
    api_auth_basic: BAUTH
    api_auth_token:
    patterns:
    - \s*\"url\"\s*:\s*\"https://httpbin.org/post\"\s*


```



### SARIF Output 

```json{8,24}
{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    ... REDACTED
      "results": [
        {
          "ruleId": "intercept.cc.api.policy.105: API VALUE CHECK",
          "ruleIndex": 0,
          "level": "note",
          "message": {
            "text": "Sandbox API check"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "output_105"
                },
                "region": {
                  "startLine": 23,
                  "endLine": 23,
                  "snippet": {
                    "text": " \n  \"url\": \"https://httpbin.org/post\""
                  }
                }
              }
            }
          ]
```

### Console Output

```sh{9,11,14}
│ 
│ 
├ API Rule # 105
│ Rule name :  API value check
│ Rule description :  Sandbox API check
│ Impacted Env :  all
│ Tags :  KEY
│ 
│ API ENDPOINT :  https://httpbin.org/post
│ 
│ ✓ 105 API Gathering Data : OK
│ 
│ 
│ API Response Status: 200 OK
│ 
  23: 
  24:  "url": "https://httpbin.org/post"
│ 
│ Compliant
│ 
│ 

```


# Run it

```sh
docker pull ghcr.io/xfhg/intercept:latest

docker run -v --rm -w $PWD -v $PWD:$PWD -e TERM=xterm-256color ghcr.io/xfhg/intercept intercept config -a examples/policy/api.yaml

docker run -v --rm -w $PWD -v $PWD:$PWD -e INTERCEPT_BAUTH=user:pass -e TERM=xterm-256color ghcr.io/xfhg/intercept intercept api 
```