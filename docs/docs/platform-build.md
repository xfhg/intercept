# Platform Build Quick Start

<br><br>

### 1. Clone the Intercept repository

```sh
git clone https://github.com/xfhg/intercept.git
cd intercept
```

::: tip
Ensure you have Go 1.22 or later and common build essentials installed.
:::

### 2. Build **INTERCEPT**  for your platform

```sh
make build
```

### 3. Copy the binary to the playground folder

```sh
cp release/intercept playground/intercept
```

### 4. Playground folder structure

```sh{3,11}
playground/
│
├─ policies/
│  ├─ ..yaml
│  ├─ ..yaml        example INTERCEPT policies for all policy types
│  └─ ..yaml
├─ runtime/
│  ├─ ..yaml
│  ├─ ..yaml        example RUNTIME specific policies 
│  └─ ..yaml
└─ targets/
   ├─ ..json
   ├─ ..toml        example target configuration files
   └─ ..ini
```

### 5. Check for leaked private keys or SSL certificates

Let's check for private keys or ssl certs leaked on our target folder using a scan policy 

::: info
snippet of policies/test_scan.yaml
:::

```yaml{3,18}
Policies:
  - id: "SCAN-001 Private Keys"
    type: "scan" // [!code focus]
    enforcement:
      - environment: "development"
        fatal: "true"
        exceptions: "false"
        confidence: "high"
    metadata:
      name: "Detect private keys"
      description: "Generic long description for (metadata) policy"
      msg_solution: "Generic solution message to production issue."
      msg_error: "Generic error message for production issue"
      tags:
        - "security"
        - "encryption"
      score: "9"
    _regex: // [!code focus]
      - \s*(-----BEGIN PRIVATE KEY-----)
      - \s*(-----BEGIN RSA PRIVATE KEY-----)
      - \s*(-----BEGIN DSA PRIVATE KEY-----)
      - \s*(-----BEGIN EC PRIVATE KEY-----)
      - \s*(-----BEGIN OPENSSH PRIVATE KEY-----)
      - \s*(-----BEGIN PGP PRIVATE KEY BLOCK-----)
```

### 6. Run Intercept with the policy

Load it into intercept and point it to our target path:

```sh 
./intercept audit --policy policies/test_scan.yaml --target targets -vvv -o _my_first_run
```

Example output:

```log{4}
2024-09-10T21:14:26+08:00 INF Output directory set to: ../intercept/playground/_my_first_run
2024-09-10T21:14:26+08:00 INF Total Policies: 1
2024-09-10T21:14:26+08:00 INF Total Policies (after filtering): 1
2024-09-10T21:14:26+08:00 INF INTERCEPT Run ID: 2lsgC1awDJ6HC1XRCLRrDYTw45x
2024-09-10T21:14:26+08:00 INF Performance Metrics:
2024-09-10T21:14:26+08:00 INF   Start Time: 2024-09-10T21:14:26+08:00
2024-09-10T21:14:26+08:00 INF   End Time: 2024-09-10T21:14:26+08:00
2024-09-10T21:14:26+08:00 INF   Execution Time: 160 milliseconds
```

### 7. INTERCEPT Command breakdown

```yml{2}
intercept 
audit # the main audit command 
--policy policies/test_scan.yaml # load the policy file
--target targets # set the target path (only applicable to some policy types)
-vvv # set the verbosity level
-o _my_first_run # set the output path for the results/debug/report
```

### 8. INTERCEPT Results

```sh{2}
_my_first_run/ # // [!code focus]
├─ intercept_2lshLG.sarif.json      # The sarif report from our audit run // [!code focus]
├─ _debug/
│  ├─ ..json
│  ├─ ..sarif       # this is a debug folder for internal workflow output validation
│  └─ ..log         #   (not applicable for this example)
├─ _patched/
│  ├─ ..json
│  ├─ ..yaml        # non-compliant files patched into compliance 
│  └─ ..ini         #   (not applicable for this example)
└─ _sarif/
   ├─ ..sarif
   ├─ ..sarif        
   └─ scan-001-private-keys.sarif   # individual policy results
```

::: info
A comprehensive SARIF Report is generated for every audit run (identified by a runID).
:::

```json{15,17,23,26-28,57}
{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "Intercept",
          "version": "1.0.3"
        }
      },
      "results": [
        {
          "ruleId": "SCAN-001-PRIVATE-KEYS", // [!code focus]
          "level": "error", // [!code focus]
          "message": { // [!code focus]
            "text": "Policy violation: Detect private keys Matched text: \n-----BEGIN PGP PRIVATE KEY BLOCK-----" // [!code focus]
          }, // [!code focus]
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "targets/scan/long.code"
                },
                "region": {
                  "startLine": 10927,
                  "startColumn": 1103,
                  "endColumn": 1141
                }
              }
            }
          ],
          "properties": {
            "description": "Generic long description for (metadata) policy",
            "error": "true",
            "msg-error": "Generic error message for production issue",
            "msg-solution": "Generic solution message to production issue.",
            "name": "Detect private keys",
            "observe-run-id": "",
            "result-timestamp": "2024-09-10T21:23:53+08:00",
            "result-type": "detail"
          }
        }
      ],
      "invocations": [
        {
          "executionSuccessful": true,
          "commandLine": "./intercept audit --policy policies/test_scan.yaml --target targets -vvv -o _my_first_run",
          "properties": {
            "debug": "false",
            "end_time": "2024-09-10T21:23:53+08:00",
            "environment": "",
            "execution_time_ms": "160",
            "report-compliant": "false",
            "report-status": "non-compliant",
            "report-timestamp": "2024-09-10T21:23:53+08:00",
            "run_id": "2lsgC1awDJ6HC1XRCLRrDYTw45x",
            "start_time": "2024-09-10T21:23:53+08:00"
          }
        }
      ]
    }
  ]
}
```

This is the minimum intercept report generated that includes 1 policy and 1 audit run.

### 9. Explore the remainding policy files to understand and build more complex workflows.

::: tip
<-- More info for each policy type

POST the sarif reports using the WEBHOOKS workflow

Check the INTERCEPT feature flags and the OBSERVE daemon
:::

## Local Build Options


A simple make build command that builds for the current platform:
```sh
make build
```

Platform-specific build commands:
```sh
make build-linux-amd64
make build-linux-arm
make build-linux-arm64
make build-darwin-amd64
make build-darwin-arm64
make build-windows-amd64
```


A command to build for all platforms:
```shell
make build-all
```


Platform-specific Docker build commands:
```sh
make docker-build-linux-amd64
make docker-build-linux-arm
make docker-build-linux-arm64
```

A command to build Docker images for all platforms:
```sh
make docker-build-all
```
