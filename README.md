<p align="center">

<img src="static/intercept.png" width="250">

</p>

# INTERCEPT

Easy to integrate, small footprint **Policy as Code** subsecond command-line scanner that leverages the power of the fastest multi-line search tool to scan your codebase. It can be used as a linter, guard rail control or simple data collector and inspector. Consider it a cross-platform weaponized **ripgrep**.

<br>

![Go](https://github.com/xfhg/intercept/workflows/Go/badge.svg?branch=master)

![GitHub release (latest by date)](https://img.shields.io/github/v/release/xfhg/intercept)
![GitHub Release Date](https://img.shields.io/github/release-date/xfhg/intercept)
![GitHub last commit](https://img.shields.io/github/last-commit/xfhg/intercept)
![GitHub commits since latest release (by date)](https://img.shields.io/github/commits-since/xfhg/intercept/latest)

![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/xfhg/intercept)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/xfhg/intercept)
![GitHub issues](https://img.shields.io/github/issues-raw/xfhg/intercept)
![GitHub pull requests](https://img.shields.io/github/issues-pr-raw/xfhg/intercept)

<br>
<br>

**v1 comming soon with SARIF output and Azure DevOps integration**

<br>
<br>

## Features

- Policy as Code
- Fine-grained regex policies
- Multiple enforcement levels
- Static Analysis, no daemons
- Low footprint, self-updatable binary
- Easy to integrate on any CI/CD Pipeline
- Declarative form policies / reduced complexity
- No custom policy language

## Policy as Code

Policy as code is the idea of writing code to manage and automate policies. By representing policies as code in YAML files, proven software development best practices can be adopted such as version control, automated testing, and automated deployment.

## How it works

- **intercept** CLI binary
- policies YAML file

**Intercept** merges environment flags, policies YAML and optional exceptions YAML to generate a global config.
It recursively scans a target path for policy breaches against your code and generates a human-readable detailed output of the findings.

<br>

#### Example Output

<p align="center">
<img src="static/output.png" >
</p>

<br>

### Use cases

- Reduced complexity **Hashicorp Sentinel** drop-in alternative. Policies are just regular expressions, does not use a custom policy language.

- Do you find [Open Policy Agent](https://www.openpolicyagent.org/) **rego** files too much sugar for your pipeline?

- Captures the patterns from [git-secrets](https://github.com/awslabs/git-secrets) and [trufflehog](https://github.com/dxa4481/truffleHog) and can prevent sensitive information to run through your pipeline. ([trufflehog regex](https://github.com/dxa4481/truffleHog/blob/dev/scripts/searchOrg.py))

- Identifies policy breach (file path and line numbers), reports solutions/suggestions to its findings making it a great tool to ease onboarding developer teams to your unified deployment pipeline.

- Can enforce style-guides, coding-standards, best practices and also report on suboptimal configurations.

- Can collect patterns or high entropy data and output it in multiple formats.

- Anything you can crunch on a regular expression can be actioned on.

## Latest [Release](https://github.com/xfhg/intercept/releases)

```sh
# Standard package (intercept + ripgrep) for individual platforms
-- core-intercept-rg-*.zip

# Cross Platform Full package (intercept + ripgrep)
-- x-intercept.zip

# Build package to build on all platforms (Development)
-- setup-buildpack.zip

# Package of the latest compatible release of ripgrep (doesn't include intercept)
-- i-ripgrep-*.zip
```

# Quick Start

Start by downloading **intercept** for your platform

```shell
--- Darwin
curl -fSL https://github.com/xfhg/intercept/releases/latest/download/intercept-darwin_amd64 -o intercept


--- Linux
curl -fSL https://github.com/xfhg/intercept/releases/latest/download/intercept-linux_amd64 -o intercept


--- Windows
curl -fSL https://github.com/xfhg/intercept/releases/latest/download/intercept-windows_amd64 -o intercept.exe
```

Let's grab some quick examples to scan

```
curl -fSLO https://github.com/xfhg/intercept/releases/latest/download/_examples.zip
```

Now we have our intercept binary ready plus an [examples/](https://github.com/xfhg/intercept/tree/master/examples) folder to play around.

Before we start looking in detail on policy files these are the types of policies available :

```
- scan : where we enforce breaking rules on matched patterns
- collect : where we just collect matched patterns
```

On our example intend to :

- **scan** if private keys are present on infra code (rule id 1)

  - we want this policy to be fatal (**fatal:true**) and accept no exceptions (**enforcement:true**)
  - setting **environment: all** guaranteed this policy will be enforced on all environments

- **scan** if modules are being sourced from its compliant source and not locally or from git (rule id 5)

  - we want this policy to be fatal (**fatal:true**) only when the environment is PROD (**environment:prod**)
  - this policy can accept local exceptions (**enforcement:false**)

- **collect** instances of terraform resources detected outside of the module usage

Take a quick glance of what a policy file with 2 **scan** rules and 1 **collect** rule :
([examples/policy/simple.yaml](https://github.com/xfhg/intercept/tree/master/examples/policy/simple.yaml)) (some text redacted for clarity) :

```yaml
# This banner is shown on the start of the scanning report,
# use it to point out important documentation/warnings/contacts
Banner: |

| Banner text here, drop documentation link or quick instructions on how to react to the report

Rules:
# This is the main policy block, all rules will be part of this array

  # This is a rule structure block
  # Each rule can have one or more patterns (regex)
  # The rule is triggered by any of the patterns listed
  #
  # Essential settings :

  # id : ( must be unique )
  # type : ( scan | collect )
  # fatal : ( true | false )
  # enforcement : ( true | false )
  # environment : ( all | anystring)

  # All other settings are free TEXT to complement your final report
  - name: Private key committed in code
    id: 1
    description: Private key committed to code version control
    solution:
    error: This violation immediately blocks your code deployment
    type: scan
    enforcement: true
    environment: all
    fatal: true
    patterns:
      - \s*(-----BEGIN PRIVATE KEY-----)
      - \s*(-----BEGIN RSA PRIVATE KEY-----)
      - \s*(-----BEGIN DSA PRIVATE KEY-----)
      - \s*(-----BEGIN EC PRIVATE KEY-----)
      - \s*(-----BEGIN OPENSSH PRIVATE KEY-----)
      - \s*(-----BEGIN PGP PRIVATE KEY BLOCK-----)

  # Another scan rule
  - name: Compliant module source
    id: 5
    description: Modules should not be sourced locally nor from git
    error: This breach blocks your deployment on production environments
    type: scan
    solution:
    environment: prod
    fatal: true
    enforcement: false
    patterns:
      - source\s*.*\.git"
      - \s+source\s*=\s*"((?!https\:).)

  # A different type of policy rule that just collects findings matched with the patterns listed
  - name: Collect sparse TF resources outside of modules.
    description: The following resources were detected outside of compliant module usage
    type: collect
    patterns:
      - (resource)\s*"(.*)"

# These are the messages displayed at the end of the report
# Clean for no finds
# Warning for at least one non-fatal find
# Critical for at least one fatal find
ExitCritical: "Critical irregularities found in your code"
ExitWarning: "Irregularities found in your code"
ExitClean: "Clean report"
```

## Scan a target repository

Let's take a real-world example and verify how the development teams are using our compliant terraform modules

On the folder [examples/](https://github.com/xfhg/intercept/tree/master/examples) we will scan the imaginary infra repo that contains terraform code at [examples/target/](https://github.com/xfhg/intercept/tree/master/examples/target)

## Integrity validation step: Before start

The following command will update your binary and its core tools to the latest version

```
intercept system --setup --update
```

## 1. Add the config file to intercept

```sh
intercept config -a examples/policy/simple.yaml

# you can also download config from remote endpoints

intercept config -a https://raw.githubusercontent.com/xfhg/intercept/master/examples/policy/simple.yaml
```

intercept will always create a config.yaml from the imported configuration files, at the moment it does not support merging of the same type of items

```

| INTERCEPT
|
| Policy file : config.yaml
|
| Config file updated

```

You can reset the config file with :

```
intercept config -r
```

## 2. Run the scan against target/ directory

This is the simplest call of audit:

```sh
intercept audit -t examples/target/

# you can merge the previous step with the audit by calling :

intercept audit -c examples/policy/simple.yaml -t examples/target/
```

<br>

<p align="center">
<img src="static/step01.png" style="border-radius:10px">
</p>

Exiting with just a warning...

Adding **prod** as environment variable:

```
intercept audit -t examples/target/target/ -e prod
```

<br>
<p align="center">
<img src="static/step02.png" style="border-radius:10px">
</p>

Notice the fatal exception and the exit code 1

## 3. Add more policies ([examples/policy/complex.yaml](https://github.com/xfhg/intercept/tree/master/examples/policy/complex.yaml))

Looks great so far... let's validate that networking resources are not being hardcoded and also intercept any module deployment with suboptimal configuration parameters.

- **scan** if any SUBNET or VPC ids are being hardcoded instead of captured via data lookups (rule id 7)

  - we want this policy to be fatal (**fatal:true**) immediately on DEV environment (**environment:dev**)
  - accept no exceptions (**enforcement:true**)

- **scan** if modules are being set up with suboptimal configuration parameters. (rule id 8)

  - we just want this policy to be a notice warning with fixing recommendation

### Example patterns on file (some text redacted for clarity) :

```yaml
- name: Hardcoded ids on code or variables
  id: 7
  description:
  solution:
  error:
  fatal: true
  environment: dev
  enforcement: true
  type: scan
  patterns:
    - (subnet_ids\s*=\s*\[\s*"\$\{v)
    - (subnet_ids\s*=\s*\[\s*"[s])
    - (subnet_ids\s*=\s*=\s*"\$\{v)
    - (subnet_id\s*=\s*"\s*[s])
    - (subnet_id\s*=\s*"\s*\$\{v)
    ...redacted for clarity

- name: Sub-optimal parameter on Module/Resource
  id: 8
  description:
  solution:
  environment:
  error:
  type: scan
  fatal: false
  patterns:
    - \s+healthcheck_target\s*=\s*"22"
    - \s+healthcheck_target\s*=\s*"3389"
    - \s+protocol\s*=\s*"-1"
    - \s+from_port\s*=\s*"-1"
    - \s+to_port\s*=\s*"-1"
    - ("0\.0\.0\.0)
```

Recompile the config file :

```bash
intercept config -a examples/policy/complex.yaml
```

Let's pretend to run the audit on DEV environment and check the differences on the report :

```
intercept audit -t examples/target/ -e dev
```

Redacted report:

<p align="center">
<img src="static/step03.png" style="border-radius:10px">
</p>

## 4. Add local exceptions ([examples/exception/local_exception.yaml](https://github.com/xfhg/intercept/tree/master/examples/exception/local_exception.yaml))

**Use case :** If you parse the config file from a global location and need local (per repo/per CI/CD job) exceptions you can add a local YAML file and merge it to the main config.

We will try to have an exception on policy rule id 5 (accepts exceptions) and policy rule id 7 (doesn't accept exceptions)

```yaml
Exceptions:
  - 5
  - 7

ExceptionMessage: "THIS RULE CHECK IS DEACTIVATED BY A LOCAL EXCEPTION REQUEST"
```

```sh
intercept config -a examples/exception/local_exception.yaml
```

Both files are merged and you can run the audit with the new exceptions in place

```sh
intercept audit -t examples/target/ -e dev
```

Redacted report:

<p align="center">
<img src="static/step04.png" style="border-radius:10px">
</p>

As you can notice rule 5 activated the exception but rule 7 just ignore it and returned a FATAL breach.

## 5. Enforcing **no exceptions** flag

By activating the **No Exceptions** flag (-x) all the exceptions will be ignored.

```
intercept audit -t examples/target/ -e prod -x
```

## 6. Policy File Explained

#### [policy/policy_rules.yaml](https://github.com/xfhg/intercept/tree/master/policy/policy_rules.yaml)

```yaml
Banner: |

  MULTI LINE TXT

ExitCritical: CRITICAL_ERROR_EXIT_TEXT
ExitWarning: WARNING_EXIT_TEXT
ExitClean: CLEAN_EXIT_TEXT

Rules:
  # SCAN TYPE

  - id: 1

    name: NAME_TEXT
    description: DESCRIPTION_TEXT
    solution: SOLUTION_TEXT
    error: ERROR_TEXT
    type: scan
    fatal: BOOL
    environment: TXT
    enforcement: BOOL
    patterns:
      - regex_1
      - regex_2
      - regex_3

  # COLLECT TYPE

  - name: NAME_TEXT
    description: DESCRIPTION_TEXT
    type: collect
    patterns:
      - regex_4
      - regex_5
```

#### [policy/policy_exceptions.yaml](https://github.com/xfhg/intercept/tree/master/policy/policy_exceptions.yaml)

```yaml
Exceptions:
  - RULE_ID
  - RULE_ID

ExceptionMessage: TXT_MESSAGE
```

<br>

# Policy Enforcement Levels

Enforcement levels are a first class concept in allowing pass/fail behavior to be associated separately from the policy logic. This enables any policy to be a warning, allow exceptions, or be absolutely mandatory. These levels can be coupled to environments, different uses of the same policy can have different enforcement levels per environment.

You can set three enforcement levels:

- **Advisory**: The policy is allowed to fail. However, a warning will be shown to the user or logged.

```
  - fatal: false
  - enforcement: false
  - environment : (all | optional)
```

- **Soft Mandatory**: The policy must pass unless an exception is specified. The purpose of this level is to provide a level of privilege separation for a behavior. Additionally, the exception provides non-repudiation since at least the primary actor was explicitly overriding a failed policy.

```
  - fatal: true
  - enforcement: false
  - environment : (all | optional)
```

- **Hard Mandatory**: The policy must pass no matter what. The only way to override a hard mandatory policy is to explicitly remove the policy. It should be used in situations where an exception is not possible.

```
  - fatal: true
  - enforcement: true
  - environment : (all | optional)
```

<br>

# Used in production

**INTERCEPT** was created to lint thousands of infra PRs and deployments a day with minor human intervention, the first MVP been running for more than a year saving countless hours of human debug/review time. Keep in mind INTERCEPT is not and does not pretend to be a security tool.
It's easy to circumvent a regex pattern once you know it, but the main objective of this tool is to pro-actively help the developers fix their code and assist with style/rule suggestions to keep the codebase clean and avoid trivial support requests from the uneducated crowd.

## Inspired by

- [ripgrep](https://github.com/BurntSushi/ripgrep)
- [Hashicorp Sentinel](https://www.hashicorp.com/sentinel/)
- [Open Policy Agent](https://www.openpolicyagent.org/)

## Standing on the shoulders of giants

### Why [ripgrep](https://github.com/BurntSushi/ripgrep) ? Why is it fast?

- It is built on top of Rust's regex engine. Rust's regex engine uses finite automata, SIMD and aggressive literal optimizations to make searching very fast. (PCRE2 support)
  Rust's regex library maintains performance with full Unicode support by building UTF-8 decoding directly into its deterministic finite automaton engine.

- It supports searching with either memory maps or by searching incrementally with an intermediate buffer. The former is better for single files and the latter is better for large directories. ripgrep chooses the best searching strategy for you automatically.

- Applies ignore patterns in .gitignore files using a RegexSet. That means a single file path can be matched against multiple glob patterns simultaneously.

- It uses a lock-free parallel recursive directory iterator

### Benchmark ripgrep

| Tool                  | Command                                                 | Line count | Time       |
| --------------------- | ------------------------------------------------------- | ---------- | ---------- |
| ripgrep (Unicode)     | `rg -n -w '[A-Z]+_SUSPEND'`                             | 450        | **0.106s** |
| git grep              | `LC_ALL=C git grep -E -n -w '[A-Z]+_SUSPEND'`           | 450        | 0.553s     |
| The Silver Searcher   | `ag -w '[A-Z]+_SUSPEND'`                                | 450        | 0.589s     |
| git grep (Unicode)    | `LC_ALL=en_US.UTF-8 git grep -E -n -w '[A-Z]+_SUSPEND'` | 450        | 2.266s     |
| sift                  | `sift --git -n -w '[A-Z]+_SUSPEND'`                     | 450        | 3.505s     |
| ack                   | `ack -w '[A-Z]+_SUSPEND'`                               | 1878       | 6.823s     |
| The Platinum Searcher | `pt -w -e '[A-Z]+_SUSPEND'`                             | 450        | 14.208s    |

<br>

<details><summary><b>Tools Reference</b></summary>

- [ripgrep](https://github.com/BurntSushi/ripgrep)
- [git grep](https://www.kernel.org/pub/software/scm/git/docs/git-grep.html)
- [The Silver Searcher](https://github.com/ggreer/the_silver_searcher)
- [git grep (Unicode)](https://www.kernel.org/pub/software/scm/git/docs/git-grep.html)
- [sift](https://github.com/svent/sift)
- [ack](https://github.com/beyondgrep/ack2)
- [The Platinum Searcher](https://github.com/monochromegane/the_platinum_searcher)

</details>

<br>

---

<br>

## Tests

#### Test Suite runs with [venom](https://github.com/ovh/venom)

```sh
venom run tests/suite.yml
```

## Vulnerabilities

#### Scanned with [Sonatype Nancy](https://github.com/sonatype-nexus-community/nancy)

```sh
make nancy
```

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ Summary                       ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━┫
┃ Audited Dependencies    ┃ 62  ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━╋━━━━━┫
┃ Vulnerable Dependencies ┃ 0   ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━┻━━━━━┛
```

from Sonatype OSS Index


## Latest Changes:

- complimentary JSON output

## TODO

- [ ] Complete the test suite

- [ ] POST results (in JSON or YAML) to a configurable webhook or notification stream

<br>

# PLAYGROUND

Full examples and functionality ready to mess around with :

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/xfhg/intercept)

```
make setup-dev
make out-linux
cd release
./interceptl config -a ../examples/policy/simple.yaml
./interceptl audit -t ../examples/target
```

---
