
# 🧬 Basic Concepts


As a forward-thinking CISO/CIO, you understand the critical importance of maintaining a strong security posture across your entire software lifecycle. INTERCEPT offers you:

1. **Lightning-Fast SecConf Management**: Identify potential security risks in configuration files within milliseconds. And patch them immediatly.
2. **Streamlined Compliance Workflow**: Implement a full-cycle compliance process with minimal setup time.
3. **Proactive Risk Management**: Stay ahead of threats, leaks and drift by continuously monitoring and auditing your infrastructure.


<br><br>

::: tip 
To gain a deeper understanding of INTERCEPT's architecture and how it fits into your security strategy, please refer to our [Architecture page](/docs/architecture). This page provides a comprehensive overview of INTERCEPT's capabilities, operational modes, and how it can transform your security operations.
:::


<br><br>



---


<img alt="Intercept Arch" src="/focus.png">

The POLICY file (also known as the configuration file) is the primary component of the INTERCEPT workflow. It configures the policy engine and provides all necessary policies, triggers, and hooks to generate a comprehensive compliance report and distribute it to the appropriate recipients.
<br><br>

## The Policy File Structure

A policy file is a YAML document comprising:



```yaml

Config: # (optional) used to configure INTERCEPT

Version: # (optional) version of the policy schema

Policies: # (mandatory) List of all the policies to be loaded // [!code focus]
  - id:
    type:
  - id:
    type:
  - id:
    type:

```
::: tip POLICY FILE
**Also known as INTERCEPT config file** 
:::

A minimal policy file might look like this:

```yaml

Policies: // [!code focus]
  - id: "SCAN-001 Private Keys" // [!code focus]
    type: "scan" // [!code focus]
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
      name: "Detect private keys, certs, etc"
      description: "Generic long description for (metadata) policy"
      msg_solution: "Generic solution message to production issue."
      msg_error: "Generic error message for production issue"
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
```

<br><br>

## The Policy Definition

The policy file can contain all types of policy definitions or be segregated by type, depending on your team's preference. Each policy has the following structure:

::: tip POLICY
**An individual policy structure** 
:::



```yaml{24,32,38,44}
- id:   # short name of your policy 
  type: # "scan","assure","api","json","yaml","toml","ini","runtime","rego"

  filepattern: # file name pattern filter to narrow down the target of the policy
  schedule: # (only for observe daemon) set the policy audit accodringly to a cron
  observe: # (only for observe daemon) set up path monitoring to trigger this policy

  enforcement: # check the Enforcement section for the value matrix.
     - environment:
       fatal:
       exceptions:
       confidence:

   metadata: # All the metadata that will populate your compliance report
     name:
     description:
     tags: # when running an audit you can filter the policies by tag
       - # the ID of the policy is always an automatic tag 
       -
     msg_solution:
     msg_error:
     score:

    # ASSURE Filetype policies
   _schema:
      patch: # (defaults false) if true and your CUE Lang schema has values instead of types 
             #                  the non compliant will have a patch created
      strict: # (defaults false) if true the target file needs to adhere to the full schema below
      structure: # here goes your CUE Lang schema to be 
                 # applied/verified against the target files

    # REGO TYPE Policies
   _rego:
      policy_file:  # your REGO policy file
      policy_data:  # (optional) additional data needed
      policy_query: # the query to access compliance of the policy

    # SCAN & ASSURE (&api) REGEX Policies
   _regex:
      - "regex_here" # a list of REGEX patterns
      -
      -

    # RUNTIME Policies
   _runtime:
      config:   # the goss configuration file 
      observe:  # (only for observe daemon) file or file path to 
                # be observed for changes and trigger this policy

```


The common area to all policies :

```yaml{24,32,38,44}
- id:   # short name of your policy // [!code focus]
  type: # "scan","assure","api","json","yaml","toml","ini","runtime","rego" // [!code focus]

  filepattern: # file name pattern filter to narrow down the target of the policy
  schedule: # (only for observe daemon) set the policy audit accodringly to a cron

  enforcement: # check the Enforcement section for the value matrix. // [!code focus]
     - environment: // [!code focus]
       fatal: // [!code focus]
       exceptions: // [!code focus]
       confidence: // [!code focus]

   metadata: # All the metadata that will populate your compliance report // [!code focus]
     name: // [!code focus]
     description: // [!code focus]
     tags: # when running an audit you can filter the policies by tag // [!code focus]
       - # the ID of the policy is always an automatic tag  // [!code focus]
       - // [!code focus]
     msg_solution: // [!code focus]
     msg_error: // [!code focus]
     score: // [!code focus]

    # ASSURE Filetype policies
   _schema:
      patch: # (defaults false) if true and your CUE Lang schema has values instead of types 
             #                  the non compliant will have a patch created
      strict: # (defaults false) if true the target file needs to adhere to the full schema below
      structure: # here goes your CUE Lang schema to be 
                 # applied/verified against the target files

    # REGO TYPE Policies
   _rego:
      policy_file:  # your REGO policy file
      policy_data:  # (optional) additional data needed
      policy_query: # the query to access compliance of the policy

    # SCAN & ASSURE (&api) REGEX Policies
   _regex:
      - "regex_here" # a list of REGEX patterns
      -
      -

    # RUNTIME Policies
   _runtime:
      config:   # the goss configuration file 
      observe:  # (only for observe daemon) file or file path to 
                # be observed for changes and trigger this policy

```


<br><br>


## The CLI

INTERCEPT offers two primary operating modes:


::: tip AUDIT

Performs a full audit run, cycling through all loaded policies, generating individual reports per policy, and a final compliance SARIF report.
:::


::: tip OBSERVE

Runs a daemon that monitors file paths, mounts, services, and configs, reacting to drifts or triggering scheduled policy audits.
:::




```
Usage:
  intercept [command] // [!code focus]

Available Commands:
  audit       Run an optimized audit through all loaded policies // [!code focus]
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  observe     Observe and trigger realtime policies based on schedules or active path monitoring // [!code focus]
  sys         Test intercept embedded core binaries
  version     Print the build info of intercept

Flags:
      --experimental        Enables unreleased experimental features
  -h, --help                help for intercept
      --nolog               Disables all loggging
  -o, --output-dir string   directory to write output files
      --silent              Enables log to file intercept.log
  -v, --verbose count       increase verbosity level

Use "intercept [command] --help" for more information about a command.

```



<br><br>


## The Compliance Report

INTERCEPT's AUDIT output is a SARIF-compliant report containing essential metadata for data-driven decision-making. The report includes:

- Individual Policy Attestation result details
- Individual Policy Attestation result summaries
- Overall compliance status based on configured environment enforcement levels

<br>

```json
{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "Intercept",
          "version": "v1.0.4"
        }
      },
      "results": [ // [!code focus]
        {
          "ruleId": "SCAN-001-PRIVATE-KEYS", // [!code focus]
          "level": "error", // [!code focus]
          "message": { // [!code focus]
            "text": "Policy violation: Detect private keys Matched text: \n-----BEGIN PGP PRIVATE KEY BLOCK-----" // [!code focus]
          },
          "locations": [ // [!code focus]
            {
              "physicalLocation": { // [!code focus]
                "artifactLocation": { // [!code focus]
                  "uri": "targets/scan/long.code" // [!code focus]
                }, 
                "region": { // [!code focus]
                  "startLine": 10927, // [!code focus]
                  "startColumn": 1103, // [!code focus]
                  "endColumn": 1141 // [!code focus] 
                }
              }
            }
          ],
          "properties": { // [!code focus]
            "description": "Generic long description for (metadata) policy", // [!code focus]
            "error": "true", // [!code focus]
            "msg-error": "Generic error message for production issue", // [!code focus]
            "msg-solution": "Generic solution message to production issue.", // [!code focus]
            "name": "Detect private keys",  // [!code focus]
            "observe-run-id": "",
            "result-timestamp": "2024-09-11T15:01:00+08:00",
            "result-type": "detail" // [!code focus]
          }
        }
      ],
      "invocations": [ // [!code focus]
        {
          "executionSuccessful": true, // [!code focus]
          "commandLine": "./intercept audit --policy policies/test_scan.yaml --target targets -vvvv -o _my_first_run", // [!code focus]
          "properties": {
            "debug": "false",
            "end_time": "2024-09-11T15:01:00+08:00",
            "environment": "",
            "execution_time_ms": "364",
            "report-compliant": "false", // [!code focus]
            "report-status": "non-compliant", // [!code focus]
            "report-timestamp": "2024-09-11T15:01:00+08:00", // [!code focus]
            "run_id": "2lulu0kvIoO5xkZ5Te4VgkqxEVH", // [!code focus]
            "start_time": "2024-09-11T15:01:00+08:00"
          }
        }
      ]
    }
  ]
}
```
::: tip COMPLIANCE RESULTS

This comprehensive report enables organizations to make informed decisions about their security and compliance posture.
:::


<br><br><br>

