<p align="center">

<img src="static/intercept.png" width="250">

</p>

# INTERCEPT

Stupidly easy to use, small footprint **Policy as Code** subsecond command-line scanner that leverages the power of the fastest multi-line search tool to scan your codebase. It can be used as a linter, guard rail control or simple data collector and inspector. Consider it a weaponized ripgrep. Works on Mac, Linux and Windows

[![Latest release](https://img.shields.io/badge/release-MVP%20ONE-blue)](https://github.com/xfhg/intercept/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/xfhg/intercept)](https://goreportcard.com/report/github.com/xfhg/intercept)

## How it works

- intercept binary
- policies yaml file
- (included) latest [ripgrep](https://github.com/BurntSushi/ripgrep) binary
- (optional) latest [shellcheck](https://github.com/koalaman/shellcheck) binary
- (optional) exceptions yaml file

**Intercept** merges environment flags, policies yaml, exceptions yaml to generate a global config.
Uses **ripgrep** to scan a target path for policy breaches recursively against your code and generates a human readable detailed output of the findings.

<br>

#### Example Output

<p align="center">
<img src="static/output.png" >
</p>

<br>

### Use cases

- Simple and powerful free drop-in alternative for [Hashicorp Sentinel](https://www.hashicorp.com/sentinel/) if you are more comfortable writing and maintaining regular expressions than using a **new custom policy language**.

- Do you find [Open Policy Agent](https://www.openpolicyagent.org/) **rego** files too much sugar for your pipeline?

- Captures the patterns from [git-secrets](https://github.com/awslabs/git-secrets) and [trufflehog](https://github.com/dxa4481/truffleHog) and can prevent sensitive information to run through your pipeline. ([trufflehog regex](https://github.com/dxa4481/truffleHog/blob/dev/scripts/searchOrg.py))

- Identifies policy breach (files and line numbers), reports solutions/suggestions to its findings making it a great tool to ease onboarding developer teams to your unified deployment pipeline.

- Can enforce style-guides, coding-standards, best practices and also report on suboptimal configurations.

- Can collect patterns or high entropy data and output it in multiple formats.

- Anything you can crunch on a regular expression can be actioned on.

### Latest [Release](https://github.com/xfhg/intercept/releases) :

```sh

  # Standard package (intercept + ripgrep) for individual platforms
    -- intercept-rg-linux-*.zip
    -- intercept-rg-macos-*.zip
    -- intercept-rg-win-*.zip

  # Clean package (intercept only) for individual platforms
    -- raw-intercept-linux-*.zip
    -- raw-intercept-macos-*.zip
    -- raw-intercept-win-*.zip

  # Full package (intercept + ripgrep) for all platforms
    -- x-intercept-*.zip

  # Package needed to fully use the Makefile
    -- intercept-buildpack-*.zip

  # Package of the latest compatible release of ripgrep (doesn't include intercept)
    -- i-ripgrep-linux-*.zip
    -- i-ripgrep-macos-*.zip
    -- i-ripgrep-win-*.zip

```

<br>
Download the standard package for your platform to get started

<br>

## Step by Step

<details>
<summary><b>CLICK HERE FOR DETAILED INSTRUCTIONS</b></summary>

<br>

Let's take a real-world example and verify how the dev teams are using our compliant terraform modules

On the folder [examples/](https://github.com/xfhg/intercept/tree/master/examples) we will scan the imaginary infra repo that contains terraform code at [examples/target/](https://github.com/xfhg/intercept/tree/master/examples/target)

## 1. Write a global policy file ([examples/policy/simple.yaml](https://github.com/xfhg/intercept/tree/master/examples/policy/simple.yaml))

We intend to :

- **scan** if private keys are present on their infra code (rule 1)

  - we want this policy to be fatal (**fatal:true**) and accept no exceptions (**enforcement:true**)
  - by omission of environment, this policy is fatal regardless of where it is scanning

- **scan** if modules are being sourced from its compliant source and not locally or from git (rule 5)

  - we want this policy to be fatal (**fatal:true**) only when the environment is PROD (**environment:prod**)
  - this policy can accept local exceptions (**enforcement:false**)

- **collect** instances of tf resources outside of the module usage

<br>

**Note :** _full text on the file, some text redacted from the snippets below for clarity._

<br>

```yaml
Banner: |

  | Use this banner to post guidelines and documentation about this pipeline step

  Free multi-line text

Rules:
  - name: Private key committed in code
    id: 1
    description:
    solution:
    error:
    type: scan
    enforcement: true
    environment:
    fatal: true
    patterns:
      - \s*(-----BEGIN PRIVATE KEY-----)
      - \s*(-----BEGIN RSA PRIVATE KEY-----)
      - \s*(-----BEGIN DSA PRIVATE KEY-----)
      - \s*(-----BEGIN EC PRIVATE KEY-----)
      - \s*(-----BEGIN OPENSSH PRIVATE KEY-----)
      - \s*(-----BEGIN PGP PRIVATE KEY BLOCK-----)

  - name: Compliant module source
    id: 5
    description:
    error:
    type: scan
    solution:
    environment: prod
    fatal: true
    enforcement: false
    patterns:
      - source\s*.*\.git"
      - \s+source\s*=\s*"((?!https\:).)

  - name: Collect sparse TF resources outside of modules.
    description:
    type: collect
    patterns:
      - (resource)\s*"(.*)"
```

You should also customize the Exit Error messages on the same file.

```yaml
ExitCritical: "Critical irregularities found in your code"
ExitWarning: "Irregularities found in your code"
ExitClean: "Clean report"
```

Where Exit Critical represents exit code of 1 when a fatal policy is breached.

## 2. Add the config file to intercept

```sh
intercept config -a policy/simple.yaml
```

intercept will always create a config.yaml from the imported configuration files, at the moment it does not support merging of the same class of items

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

## 3. Run the scan against target/ directory

This is the simplest call of audit:

```sh
intercept audit -t target/
```

<p align="center">
<img src="static/step01.png" style="border-radius:10px">
</p>

Exiting with just a warning...

Adding **prod** as environment variable:

```
intercept audit -t target/ -e prod
```

<p align="center">
<img src="static/step02.png" style="border-radius:10px">
</p>

Notice the fatal exception and the exit code 1

## 5. Add more policies ([examples/policy/complex.yaml](https://github.com/xfhg/intercept/tree/master/examples/policy/complex.yaml))

Looks great so far... let's validate that networking resources are not being hardcoded and also intercept any module deployment with suboptimal configuration parameters.

- **scan** if any SUBNET or VPC ids are being hardcoded instead of captured via data lookups (rule 001)

  - we want this policy to be fatal (**fatal:true**) immediately on DEV environment (**environment:dev**)
  - accept no exceptions (**enforcement:true**)

- **scan** if modules are being setup with suboptimal configuration parameters. (rule 005)

  - we just want this policy to be a notice and recommendation

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
    - (subnets\s*=\s*\[\s*"\$\{v)
    - (subnets\s*=\s*\[\s*"[s])
    - (vpc_zone_identifier\s*=\s*\[\s*"\$\{v)
    - (vpc_zone_identifier\s*=\s*\[\s*"[v])
    - (vpc_zone_identifier\s*=\s*=\s*"\$\{v)
    - (vpc_id\s*=\s*"\s*[v])
    - (vpc_id\s*=\s*"\s*\$\{v)
    - (vpc_security_group_ids\s*=\s*\[\s*"\$\{v)
    - (vpc_security_group_ids\s*=\s*\[\s*"[sg])
    - (security_groups\s*=\s*\[\s*"\$\{v)
    - (security_groups\s*=\s*\[\s*"[sg])
    - ("subnet-)
    - ("sg-)
    - ("vpc-)

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
intercept config -a policy/complex.yaml
```

Let's pretend to run the audit on DEV environment and check the differences on the report :

```
intercept audit -t target/ -e DEV
```

Redacted report:

<p align="center">
<img src="static/step03.png" style="border-radius:10px">
</p>

## 6. Add local exceptions ([examples/exception/local_exception.yaml](https://github.com/xfhg/intercept/tree/master/examples/exception/local_exception.yaml))

Use case : If you parse the config file from a global location and need local (per repo) exceptions you can add a local yaml file and merge it to the main config.

We will try to have an exception on policy 5 (accepts exceptions) and policy 7 (doesn't accept exceptions)

```yaml
RulesDeactivated:
  - 5
  - 7

ExceptionMessage: "THIS RULE CHECK IS DEACTIVATED BY A LOCAL EXCEPTION REQUEST"
```

```sh
intercept config -a exception/local_exception.yaml
```

Both files are merged and you can run the audit with the new exceptions in place

```sh
intercept audit -t target/ -e dev
```

Redacted report:

<p align="center">
<img src="static/step04.png" style="border-radius:10px">
</p>

As you can notice rule 5 activated the exception but rule 7 just ignore it and returned a FATAL breach.

## 7. Enforcing no exceptions flag

By activating the No Exceptions flag (-x) all the exceptions will be ignored.

```
intercept audit -t target/ -e prod -x
```

## 8. Policy File Explained

#### [policy/policy_rules.yaml](https://github.com/xfhg/intercept/tree/master/policy/policy_rules.yaml)

```yaml
Banner: |

  MULTI LINE TXT

ExitCritical: CRITICAL_ERROR_EXIT_TEXT
ExitWarning: WARNING_EXIT_TEXT
ExitClean: CLEAN_EXIT_TEXT

Rules:
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

  - name: NAME_TEXT
    description: DESCRIPTION_TEXT

    type: collect

    patterns:
      - regex_4
      - regex_5
```

#### [policy/policy_exceptions.yaml](https://github.com/xfhg/intercept/tree/master/policy/policy_exceptions.yaml)

```yaml
RulesDeactivated:
  - RULE_ID
  - RULE_ID

ExceptionMessage: TXT_MESSAGE
```

</details>
<br>

## Used in production

INTERCEPT was created to lint thousands of infra deployments a day with minor human intervention, the first MVP been running for a year already with no reported flaws. Keep in mind INTERCEPT is not and does not pretend to be a security tool.
It's easy to circumvent a regex pattern once you know it, but the main objective of this tool is to pro-actively help the developers fix their code and assist with style suggestions to keep the codebase clean and avoid trivial support tickets for the uneducated crowd.

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

- It uses a lock-free parallel recursive directory iterator, courtesy of **crossbeam** and **ignore**.

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

```
Audited dependencies:41,Vulnerable:0
```

from Sonatype OSS Index

## TODO

- [ ] Complete the test suite

- [ ] Add [shellcheck](https://github.com/koalaman/shellcheck) to give warnings and suggestions for bash/sh shell scripts (optional, not core feature)

- [ ] Add system check and download of auxiliary tools (ripgrep and shellcheck)

- [ ] Configurable output types for data collection and overall report
