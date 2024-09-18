# Sandbox Playground

::: tip START HERE
Free to play, no hassle intercept sandbox :

Check an insecure nginx.conf in less than 20 miliseconds
:::


<br><br>

## 1. Our friends at Gitpod will host you 

Gitpod offers a convenient way to explore Intercept in a cloud environment

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/xfhg/intercept)

## 2. Build & Experiment :

Once in the Gitpod environment, you can build Intercept and start experimenting:

```
make build && cp release/intercept playground/intercept
```

## 3. Explore the playground folder

```sh{3,11}
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
## 🛰️ INTERCEPT AUDIT

## 4. Check for an insecure nginx.conf in less than 20 miliseconds

::: info
open **playground/policies/nginx_insecure.yaml** policy to understand what's happening under the hood
:::

### our playground example

```sh{5}
cd playground

./intercept audit \
  --policy policies/nginx_insecure.yaml \
  --target targets/ \
  -vvvv \
  -o _playground_nginx
```

### the gitpod nginx.conf

```sh{3}
./intercept audit \
  --policy policies/nginx_insecure.yaml \
  --target /etc/nginx/ \
  -vvvv \
  -o _gitpod_nginx
```


## 5. Check the compliance report

```
find it inside the output folders (_gitpod_nginx and _playground_nginx)
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
Great, let's automate it and send the report to the CIO
:::

## 🛰️ INTERCEPT OBSERVE

## 6. Setup a scheduler to run the audit

Running adhoc scans (audits) is just useful in your pre-deployment CI/CD, so let's setup a scheduler and later a path monitor into our playground target that will run scans on a schedule and also whenever a file change is detected (in this case nginx.conf)


```sh
./intercept observe --policy policies/nginx_observe.yaml  -vvvv -o _nginx_observe
```

::: tip
open **playground/policies/nginx_observe.yaml** policy , we cut the policies short to be easier to understand the changes later.

For the observe command we define the **target** through the policy file and we can also set an optional global **policy_schedule** that applies to all policies that don't have their own schedule defined (SCAN-001).

The **report_schedule** is essential to generate a merged report of all the policies.
:::


```yaml
Config: 
  Flags: // [!code focus]
    policy_schedule: "*/30 * * * * *" // [!code focus]
    report_schedule: "*/50 * * * * *" // [!code focus]
    target: "targets/" // [!code focus]
 
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

::: tip
after 50seconds a SARIF report with all the results from all the policies is generated and a minute later you can CTRL-C and a final report will be generated on the output folder.

each scheduled report will have all the results of it's defined interval delta. The final report has the delta from the previous report to the time of quitting OBSERVE.
:::

## 6. Setup a path monitor to run audits on file changes

::: warning WIP
TBC
:::

## 7. Setup integration webhooks to receive AUDIT results and compliance reports

::: warning WIP
TBC
:::

## 8. EXTRA Validate CUE Lang Schemas and REGO Policies

To validate your CUE Lang Schemas or REGO policies, you can use these online tools:


- [CUE Sandbox](https://cuelang.org/play/#cue@export@cue)
    - Use this sandbox to test and refine your CUE Lang schemas.

- [REGO Sandbox](https://play.openpolicyagent.org/p/ZWGVA8oCSE)
    - This sandbox allows you to experiment with and validate your REGO policies.


These tools provide an excellent way to ensure your schemas and policies are correct before adding them in your Intercept Policies.

## 10. Run all these examples with INTERCEPT container

Check the [Docker Quickstart](docker-quickstart) for the command syntax