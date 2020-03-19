<p align="center">

<img src="static/intercept.png" width="250">

</p>

# INTERCEPT

Stupidly easy to use, small footprint **Policy as Code** subsecond command-line scanner that leverages the power of the fastest multi-line search tool to scan your codebase. It can be used as a linter, guard rail control or simple data collector and inspector. Consider it a weaponized ripgrep. Works on Mac, Linux and Windows

[![Latest release](https://img.shields.io/badge/release-mvp%201-blue)](https://github.com/xfhg/intercept/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/xfhg/intercept)](https://goreportcard.com/report/github.com/xfhg/intercept)

## How it works

- intercept binary
- policies yaml file
- (included) [ripgrep](https://github.com/BurntSushi/ripgrep) binary
- (optional) exceptions yaml file

**Intercept** merges environment flags, policies yaml, exceptions yaml to generate a global config.
Uses **ripgrep** to scan a target path for policy breaches recursively against your code and generates a humand readable detailed output of the findings.

<br>

<details><summary><b>Example output </b></summary>

```sh



```

</details>
<br>

### Use cases

- Simple and powerful free drop in alternative for [Hashicorp Sentinel](https://www.hashicorp.com/sentinel/) if you are more comfortable writing and maintaining regular expressions than using a **new custom policy language**.

- Do you find [Open Policy Agent](https://www.openpolicyagent.org/) **rego** files too much sugar for your pipeline ?

- Captures the patterns from [git-secrets](https://github.com/awslabs/git-secrets) and [trufflehog](https://github.com/dxa4481/truffleHog) and can prevent sensitive information to run through your pipeline.

- Identifies policy breach (files and line numbers), reports solutions/suggestions to its findings making it a great tool to ease onboarding developer teams to your unified deployment pipeline.

- Can enforce style-guides, coding-standards, best practices and also report on suboptimal configurations.

- Can collect patterns or high entropy data and output it in multiple formats.

- Anything you can crunch on a regular expression can be actioned on.

### Latest [Releases](https://github.com/xfhg/intercept/releases) :

[![Latest release](https://img.shields.io/badge/release-mvp%201-blue)](https://github.com/xfhg/intercept/releases)

```sh

  # Standard package (intercept + ripgrep) for individual platforms
    -- intercept-rg-linux-*.zip
    -- intercept-rg-macos-*.zip
    -- intercept-rg-win-*.zip

  # Clean package (intercept only) for individual platforms
    -- intercept-linux-*.zip
    -- intercept-macos-*.zip
    -- intercept-win-*.zip

  # Full package (intercept + ripgrep) for all platforms
    -- intercept-x-*.zip

  # Package needed to full use the Makefile
    -- intercept-buildpack-*.zip

  # Package of the latest compatible release of ripgrep (doesn't include intercept)
    -- i-ripgrep-linux-*.zip
    -- i-ripgrep-macos-*.zip
    -- i-ripgrep-win-*.zip

```

<br>

## Instructions

Let's take a real world example and police how the dev teams are using our compliant terraform modules

On the folder examples/ we will scan the imaginary infra repo that contains terraform code example.

### Simple Example

#### 1. Write the global policy file (examples/policy/simple.yaml)

```yaml
Banner: |

  | This is a simple policy file with just two main rules with 6 patterns to scan and 1 to collect
  | Use this banner to post guidelines and ocumentation about this pipeline step

  Free text

Rules:
  - name: Private key committed in code
    id: 001
    description: Private key inside code version control
    solution: Remove it, rewrite git history and use Vault or AWS Sectets Manager to store your private keys
    error: This violation immediatly blocks your code deployment
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

  - name: Module not sourced from artifactory
    id: 005
    description: In production environment modules should be sourced from artifactory and not locally nor from git
    error: This violation will block your code promotion to non-development environments
    type: scan
    solution: "\n\tSource your modules from their latest version on artifactory \n\tMore info at XXX"
    environment: prod
    fatal: true
    enforcement:
    patterns:
      - source\s*.*\.git"
      - \s+source\s*=\s*"((?!https\:).)

  - name: Collect TF resources outside of modules.
    description: Since we only authorize TF modules to be deployed we collect TF Resources to be validated
    type: collect
    patterns:
      - (resource)\s*"(.*)"
```

#### 2. Add the config file to intercept

```sh
intercept config -a policy/simple.yaml
```

#### 3. Run the scan against target/

```sh
intercept audit -t target
```

#### 4. Review the output

#### 5. Add more rules , configure environments and enforcement of fatal policies

#### 6. Add local exceptions

#### 7. Playing with flags and config

<br>

---

<details><summary><b>TODO</b></summary>

- [ ] Complete this README
- [ ] Tests obviously
- [ ] Configurable output types for main report
- [ ] Configurable output types for data collection
  </details>

---

## Used in production

INTERCEPT was created to lint more than 8000 infra deployments a day with minor human intervention, first mvp been running for a year already with no reported flaws. Keep in mind INTERCEPT is not and does not pretend to be a security tool.

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

- It uses a lock-free parallel recursive directory iterator, courtesy of crossbeam and ignore libraries.

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
