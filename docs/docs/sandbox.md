# Sandbox Playground

::: tip START HERE
For a hassle-free experience with INTERCEPT, utilize our sandbox environment to analyze an insecure nginx.conf in under 20 milliseconds.
:::

::: danger 
Establish a **robust** software lifecycle SecConfig compliance workflow in **less than 10 steps.**
:::

<br><br>

## 1. Our friends at Gitpod will host you 

Gitpod offers a convenient way to explore INTERCEPT in a cloud environment

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/xfhg/intercept)

## 2. Build & Experiment

Once in the Gitpod environment, you can build INTERCEPT and start experimenting:

```
make build && cp release/intercept playground/intercept
```

## 3. Explore the playground folder

```sh{3,7,11}
playground/
│
├─ policies/
│  ├─ ..yaml
│  ├─ ..yaml        example INTERCEPT policies for all policy types
│  └─ ..yaml
├─ runtime/
│  ├─ ..yaml
│  ├─ ..yaml        example RUNTIME partial policies 
│  └─ ..yaml
└─ targets/
   ├─ ..json
   ├─ ..toml        example target configuration files
   └─ ..ini
```
::: info
The **playground** folder contains examples for all policy types with simplified configurations.
:::

## 🛰️ INTERCEPT AUDIT

## 4. Check for an insecure nginx.conf in less than 20 miliseconds

::: info
Locate **playground/policies/nginx_insecure.yaml** policy to understand the policy implementation.
:::

#### playground/policies/nginx_insecure.yaml

```yaml
Policies:
  - id: "SCAN-001 Server Tokens"
    type: "scan"
    filepattern: "nginx.conf"
    enforcement:
      - environment: "all"
        fatal: "true"
        exceptions: "false"
        confidence: "high"
    metadata:
      name: "Detect server_tokens on"
      description: "Ensure that server_tokens is not enabled to avoid revealing version info."
      msg_solution: "Set server_tokens to off in the nginx configuration."
      msg_error: "server_tokens is enabled, exposing Nginx version."
      tags:
        - "security"
        - "nginx"
      score: "7"
    _regex:
      - "server_tokens\\s+on;"

  - id: "SCAN-002 SSL Protocols"
    type: "scan"
    filepattern: "nginx.conf"
    enforcement:
      - environment: "all"
        fatal: "true"
        exceptions: "false"
        confidence: "high"
    metadata:
      name: "Detect insecure SSL/TLS protocols"
      description: "Detect the use of insecure SSLv3 or TLSv1/TLSv1.1 protocols."
      msg_solution: "Use secure TLS versions (TLSv1.2 or higher)."
      msg_error: "Insecure SSL/TLS protocols detected."
      tags:
        - "security"
        - "encryption"
        - "tls"
      score: "9"
    _regex:
      - "ssl_protocols.*(SSLv3|TLSv1(\\.1)?);"
      - "ssl_protocols.*SSLv2;"

...REDACTED
```

We will use this policy against two target nginx.conf files:

### our playground example

```sh{5}
cd playground

./intercept audit \
  --policy policies/nginx_insecure.yaml \
  --target targets/ \
  -vvvv \
  -o _playground_nginx
```

 <img alt="insecure nginx" src="/insecure.png" style="border-radius:6px">

### the gitpod nginx.conf

```sh{3}
./intercept audit \
  --policy policies/nginx_insecure.yaml \
  --target /etc/nginx/ \
  -vvvv \
  -o _gitpod_nginx
```


## 5. Check the compliance report

```md
Locate it inside the respective output folders 
- _gitpod_nginx 
- _playground_nginx
```

```sh{11}
_gitpod_nginx/      # output folder // [!code focus]
│
├─ _debug/
│  ├─ ..             # multiple debug state (--debug)
│  └─ ..
├─ _sarif/
│  ├─ ..sarif
│  ├─ ..sarif        # individual policy results sarif 
│  └─ ..sarif
└─ _status/          # // [!code focus]
   ├─ ..json
   ├─ ..json         # COMPLIANCE REPORTS // [!code focus]
   └─ ..json
```

```json
 "invocations": [
        {
          "executionSuccessful": true,
          "commandLine": "./intercept audit --policy policies/nginx_insecure.yaml --target /etc/nginx/ -vvvv -o _gitpod_nginx",
          "properties": {
            "end-time": "2024-09-18T06:13:53Z",
            "execution-time-ms": "15", // [!code focus]
            "report-compliant": "false",
            "report-status": "non-compliant", // [!code focus]
            "report-timestamp": "2024-09-18T06:13:53Z",
            "start-time": "2024-09-18T06:13:53Z"
          }
        }
]
```

::: info

With these results, we can now automate the process and distribute reports to stakeholders.

:::

## 🛰️ INTERCEPT OBSERVE

## 6.  Implement a Scheduled Audit

While ad-hoc scans are useful for pre-deployment CI/CD, let's set up a scheduler and path monitor for our playground target to run scans on a defined schedule.

```sh
./intercept observe \
  --policy policies/nginx_observe.yaml \
  -vvvv \
  -o _nginx_observe
```

::: tip

Refer to **playground/policies/nginx_observe.yaml** for the policy configuration. We cut the policies short to be easier to understand the changes later.

For the observe command, we define the **target** in the policy file. We can also set an optional global **policy_schedule** that applies to all policies without their own schedule (e.g., SCAN-001).

The **report_schedule** is crucial for generating a merged report of all policy results (Compliance Report). Without it, a compliance report is only generated when the process exits.

:::

#### playground/policies/nginx_observe.yaml

```yaml
Config: 
  Flags: // [!code focus]
    policy_schedule: "*/30 * * * * *" // [!code focus]
    report_schedule: "*/50 * * * * *" // [!code focus]
    target: "targets/scan/" // [!code focus]
 
Policies:
  - id: "SCAN-001 Server Tokens"
    type: "scan"
    filepattern: "nginx.conf" // [!code focus]
    enforcement:
      - environment: "all"
        fatal: "true"
        exceptions: "false"
        confidence: "high"
    metadata:
      name: "Detect server_tokens on"
      description: "Ensure that server_tokens is not enabled to avoid revealing version info."
      msg_solution: "Set server_tokens to off in the nginx configuration."
      msg_error: "server_tokens is enabled, exposing Nginx version."
      tags:
        - "security"
        - "nginx"
      score: "7"
    _regex:
      - "server_tokens\\s+on;"


  - id: "SCAN-007 Access Logs Disabled"
    type: "scan"
    filepattern: "nginx.conf"
    schedule: "*/30 * * * * *" // [!code focus]
    enforcement:
      - environment: "all"
        fatal: "true"
        exceptions: "false"
        confidence: "medium"
    metadata:
      name: "Detect access_log off"
      description: "Ensure access logs are enabled."
      msg_solution: "Remove 'access_log off' to enable logging."
      msg_error: "Access logs are disabled, which may hide important access data."
      tags:
        - "security"
        - "logging"
      score: "6"
    _regex:
      - "access_log\\s+off;"
```

```sh
 INF Added policy to Scheduler policy=SCAN-001-SERVER-TOKENS schedule="*/30 * * * * *"
 INF Added policy to Scheduler policy=SCAN-007-ACCESS-LOGS-DISABLED schedule="*/30 * * * * *"
 INF Added policy to Scheduler policy=SCAN-008-AUTOINDEX-ENABLED schedule="*/30 * * * * *"
```

::: info SCHEDULE SETUP
As configured, an AUDIT is triggered every 30 seconds, with results saved as individual policy reports. Every 50 seconds, a Compliance Report is generated, containing all individual results captured in the last 50 seconds (delta). Individual result files are cleaned up upon creation of each compliance report.
:::


::: tip REPORT STRUCTURE
Individual reports are generated based on specified policy schedules. Compliance Reports are compiled from all interval delta (report_schedule) policy results. A final report is created upon termination of the observe command, containing only the delta between the last compliance report and the quit cleanup.
:::

```sh{3,11}
_nginx_observe/      # output folder // [!code focus]
│
├─ _debug/
│  ├─ ..             # multiple debug state (--debug)
│  └─ ..
├─ _sarif/
│  ├─ ..sarif
│  ├─ ..sarif        # individual policy results sarif 
│  └─ ..sarif
└─ _status/          # // [!code focus]
   ├─ ..json
   ├─ ..json         # COMPLIANCE REPORTS // [!code focus]
   └─ ..json
```

## 7. Implement Path Monitoring for Real-time Audits

To enable real-time reactive features, we can set up a path watcher (file or folder) per policy. This allows us to trigger policy audits for any policy type when observed files are modified. In the following example, we'll reuse a "SCAN" policy to be triggered if the observed file (nginx.conf) is modified.


#### playground/policies/nginx_observe_path.yaml

```yaml

Config: 
  Flags:
    policy_schedule: "*/35 * * * * *" // [!code --]
    report_schedule: "*/50 * * * * *"
    target: "targets/scan/" // [!code focus]
 
Policies:
  - id: "SCAN-001 Server Tokens"
    type: "scan" // [!code focus]
    schedule: "*/30 * * * * *" // [!code --]
    observe: /workspace/intercept/playground/targets/scan/nginx.conf // [!code focus]
    enforcement:
      - environment: "all"
        fatal: "true"
        exceptions: "false"
        confidence: "high"
    metadata:
      ...redacted
    _regex:
      - "server_tokens\\s+on;"
```

::: info REPORT STRUCTURE
While **target:** and **observe:** don't need to be related, it's recommended for consistency.
:::

```sh
./intercept observe \
  --policy policies/nginx_observe_path.yaml \
  -vvvv \
  -o _nginx_observe_path
```


::: tip Mixing Schedulers and Path Monitors
While possible is not reccomended to have multiple policies on a tight schedule and tied to realtime monitoring. Discretion advised.
:::

## 8. Leveraging Serverspec RUNTIME Policies

INTERCEPT (+goss) provides access to the following resources on the hosts:

 <img alt="insecure nginx" src="/serverspec.png" style="border-radius:6px">

Let's create an alternative policy file to demonstrate how to leverage these features using INTERCEPT RUNTIME policies.

> [!IMPORTANT]
> A RUNTIME POLICY always consists of two components: a goss definition file and its respective INTERCEPT policy (type=runtime)

#### playground/policies/runtime/nginx.yaml 

```yaml
service: // [!code focus]
  nginx: // [!code focus]
    enabled: true // [!code focus]
    running: true // [!code focus]

package: // [!code focus]
  nginx: // [!code focus]
    installed: true // [!code focus]

# Other examples :

# process:
#   nginx:
#     running: true
#     count: 1

# file:
#   /etc/nginx/nginx.conf:
#     exists: true
#     mode: "644"
#     owner: root
#     group: root
#   /var/log/nginx/access.log:
#     exists: true
#     mode: "644"
#     owner: root
#     group: adm
#   /var/log/nginx/error.log:
#     exists: true
#     mode: "644"
#     owner: root
#     group: adm

# command:
#   nginx -t:
#     exit-status: 0
#     stdout:
#       - "syntax is ok"
#       - "test is successful"

```
::: info 
This is a goss definition file, not an INTERCEPT policy file. INTERCEPT policy files import these partials. See below:
:::

Connect it to an INTERCEPT policy file :

#### playground/policies/nginx_observe_runtime.yaml 

```yaml
Config: 
  Flags:
    report_schedule:  "*/50 * * * * *"

Policies:

  - id: "RUNTIM-001-NGINX"
    type: "runtime" // [!code focus]
    enforcement:
      - environment: "all"
        fatal: "false"
        exceptions: "development"
        confidence: "high"
    metadata:
      name: "RUNTIME System Compliance"
      description: "Enforce system compliance at runtime"
      msg_solution: ""
      msg_error: ""
      tags:
        - "config"
        - "runtime"
        - "enforce"
      score: "8"
    _runtime: // [!code focus]
      # relative to this policy
      config: runtime/nginx.yaml # // [!code focus]
      # ideally absolute path
      observe: /workspace/intercept/playground/targets/scan/nginx.conf # // [!code focus]
```

```sh
./intercept observe \
  --policy policies/nginx_observe_runtime.yaml \
  -vvvv \
  -o _nginx_observe_runtime
```

::: tip SIMULATE A CHANGE 
To test INTERCEPT OBSERVE's reaction, manually modify the tracked file:

Edit **/workspace/intercept/playground/targets/scan/nginx.conf**
:::


## 9. Configure Integration Webhooks for Audit Results and Compliance Reports

::: tip
Prometheus and OpenTelemetry options will be available.
:::

To distribute compliance reports effectively, we can define a list of WebHooks as targets to receive the latest generated reports. The available **event types** are:

- **"report"**: Every SARIF compliance report (merge of all policy results) is posted to the webhook (usually following report_schedule or OBSERVE exit)
- **"policy"**: Every individual SARIF result is posted to the webhook upon creation (audit, policy run, scheduled or triggered)



#### playground/policies/nginx_observe_runtime_hooks.yaml 

```yaml
Config: // [!code focus]
  Flags:
    policy_schedule: "*/15 * * * * *"
    report_schedule:  "*/50 * * * * *" // [!code focus]
  Hooks: // [!code focus]
    - name: "Test Webhook Report"
      endpoint: "https://webhook.site/ea939de5-d1bc-4078-8f6e-8873103e77bd"
      insecure: false
      method: "POST"
      auth: 
        type: bearer
        token_env: TOKEN 
      headers:
        Content-Type: "application/json"
      retry_attempts: 3
      retry_delay: "5s"
      timeout_seconds: 30
      event_types:
        - "report" // [!code focus]
    - name: "Test Webhook Policy"
      endpoint: "https://webhook-test.com/e42b61438d6f12f0478e253bb6c1dfa1"
      insecure: false
      method: "POST"
      auth: 
        type: bearer
        token_env: TOKEN 
      headers:
        Content-Type: "application/json"
      retry_attempts: 3
      retry_delay: "5s"
      timeout_seconds: 30
      event_types:
        - "policy" // [!code focus]

Policies:

  - id: "RUNTIM-001-NGINX"
    type: "runtime"
    enforcement:
      - environment: "all"
        fatal: "false"
        exceptions: "development"
        confidence: "high"
    metadata:
      name: "RUNTIME System Compliance"
      description: "Enforce system compliance at runtime"
      msg_solution: ""
      msg_error: ""
      tags:
        - "config"
        - "runtime"
        - "enforce"
      score: "8"
    _runtime:
      config: runtime/nginx.yaml
      schedule: "*/30 * * * * *"
```

```sh
./intercept observe \
  --policy policies/nginx_observe_runtime_hooks.yaml \
  -vvvv \
  -o _nginx_observe_runtime
```


## 10. EXTRA Validate CUE Lang Schemas and REGO Policies

To validate your CUE Lang Schemas or REGO policies, you can use these online tools:


- [CUE Sandbox](https://cuelang.org/play/#cue@export@cue)
    - Use this sandbox to test and refine your CUE Lang schemas.

- [REGO Sandbox](https://play.openpolicyagent.org/p/ZWGVA8oCSE)
    - This sandbox allows you to experiment with and validate your REGO policies.


These tools provide an excellent way to ensure your schemas and policies are correct before adding them in your Intercept Policies.

## 11. Run all these examples with INTERCEPT containers

Check the [Docker Quickstart](docker-quickstart) for the command syntax