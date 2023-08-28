<p align="center">

<img src="static/interceptv1.png" width="275">

</p>

# INTERCEPT v1.3.x

**intercept** is a DevSecOps cli multi-tool designed to provide Static Application Security Testing (SAST) capabilities to software development teams. The tool aims to help developers identify and address security vulnerabilities in their code and API endpoints early in the software development life cycle, reducing the risk of security breaches and ensuring compliance with industry regulations. intercept leverages a range of security scanning techniques to analyze code, including pattern matching, code analysis, and vulnerability scanning. It is seamless to integrate, with a simple sub-second command-line interface and granular customizable configuration options. With intercept, developers can speed up security and integration testing into their development workflows and make security a critical yet seamless part of their software development process.

<br>


![GitHub release (latest by date)](https://img.shields.io/github/v/release/xfhg/intercept)
![GitHub Release Date](https://img.shields.io/github/release-date/xfhg/intercept)
![GitHub last commit](https://img.shields.io/github/last-commit/xfhg/intercept)

![GitHub commits since latest release (by date)](https://img.shields.io/github/commits-since/xfhg/intercept/latest)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/xfhg/intercept)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/xfhg/intercept)

[![CodeQL](https://github.com/xfhg/intercept/actions/workflows/codeql.yml/badge.svg)](https://github.com/xfhg/intercept/actions/workflows/codeql.yml)
![GitHub pull requests](https://img.shields.io/github/issues-pr-raw/xfhg/intercept)
[![intercept Release](https://github.com/xfhg/intercept/actions/workflows/release.yml/badge.svg)](https://github.com/xfhg/intercept/actions/workflows/release.yml)

<br>
<br>

## Features


- **Pattern matching:** intercept uses regex pattern matching technique to scan code for known vulnerabilities and customised patterns, reducing the time and effort required to identify and fix these common issues. [Library with more than 1500 patterns](https://github.com/xfhg/intercept/tree/master/policy/unstable.yaml) 
- **Customizable rules:** intercept allows users to customize all security rules used to scan their code, making it possible to tailor the scanning process to the specific requirements of their application or organization.
- **API endpoint checks:** intercept supports gathering API data and applying the same pattern matching policies throughout it's response, complete with full telemetry.
- **Integration with CI/CD:** intercept can easily be integrated into continuous integration and continuous deployment (CI/CD) pipelines, allowing security testing to be performed automatically as part of the development process.
- **Detailed reporting:** intercept provides detailed reports on vulnerabilities and security issues, _fully compliant SARIF output_, including severity ratings and remediation advice, making it easy for developers to prioritize and address security concerns early on.
- **Support for any programming language:** intercept supports scanning through any programming languages or file types,  making it a versatile tool for security testing across a range of applications and environments.
- **CUE Lang compatible ruleset:** intercept CUE Lang policies reduce boilerplate in large-scale configurations.Validate text-based data files or programmatic data such as incoming RPCs and validate backwards compatibility.
- **No daemons, low footprint, self-updatable binary**
- **Ultra flexible fine-grained regex policies**
- **No custom policy language, reduced complexity**
- New **CUE Lang schemas** normalized and simplified representations of constraints
- **Weaponised** ripgrep on steroids (CUE Lang + API checks)
- **Open source, free as in beer**

<br>

## Policy as Code

Policy as code is an approach to defining and enforcing policies within an organization using code. Instead of writing policies in documents or spreadsheets, policy as code involves writing policies as code using a programming language. This code is then integrated into the organization's infrastructure, software, or workflow to enforce the policies automatically.

Main benefits:

- **Consistency:** Policies are enforced consistently across all systems and applications, reducing the risk of errors and vulnerabilities.
- **Automation:** Policies can be enforced automatically, reducing the need for manual intervention and improving efficiency.
- **Transparency:** Policies can be reviewed and audited more easily, providing greater transparency into how policies are being enforced.
- **Flexibility:** Policies can be updated and changed more easily, allowing organizations to adapt to changing requirements and regulations.

**Policy as code** can be used to enforce a wide range of policies, including security policies, compliance policies, and operational policies. It is often used in conjunction with infrastructure as code and other DevOps practices to provide a more automated and streamlined approach to managing IT operations.

## Secret Scanning

**intercept** offers an extensive library of policies consisting of over a thousand regular expressions that can be used to detect sensitive data leakage and enforce security best practices in software development. This vast collection of pre-defined policies makes it easy for developers to get started with secret scanning and quickly identify potential issues in their code. The policies cover a range of security concerns, such as hard-coded passwords, API keys, and other secrets, and are continuously updated to keep up with the latest security threats and best practices. With the ability to customize policies or add new ones, developers can ensure that their applications are protected against known and emerging threats, reducing the risk of sensitive data leakage and improving the overall security posture of their organization.

#### [Pre made patterns available](https://github.com/xfhg/intercept/tree/master/policy/unstable.yaml) 

<br>
<br>
<br>

# Quick Start

1. Grab the latest [RELEASE](https://github.com/xfhg/intercept/releases) of intercept bundle for your platform 

```
core-intercept-rg-x86_64-darwin.zip
core-intercept-rg-x86_64-linux.zip
core-intercept-rg-x86_64-windows.zip
```

2. Make sure you have the latest setup

```
intercept system --update
```

3. Load some [EXAMPLE](https://github.com/xfhg/intercept/tree/master/examples) policies and target files

```
start with the minimal.yaml
```

4. Configure intercept

```
intercept config -r
intercept config -a examples/minimal.yaml
```

5. Audit the target folder

```
intercept audit -t examples/target
```

6. Check the different output flavours

```
stdout human readable console report
individual json rule output with detailed matches
all findings compiled into intercept.output.json
fully compliant SARIF output into intercept.sarif.json
all SHA256 of the scanned files into intercept.scannedSHA256.json
```

7. Tune the scan with extra flags like ENVIRONMENT or TAGS filter

```
intercept audit -t examples/target -e "development" -i "AWS"
```

## Policy File Structure

These are 5 types of policies available :

- **scan** : where we enforce breaking rules on matched patterns
- **collect** : where we just collect matched patterns
- **assure** : where we expect matched patterns and detect if missing
- **api** : apply the assure rules into API endpoint data
- **yml** : scan assure rules with CUE Lang schemas for YAML

Easy to read and compose the rules file have this minimal required structure:
```
Banner: |

  | Example SCAN, ASSURE, COLLECT, API and YML RULES

Rules:
  - name: Private key committed in code
    id: 100
    description: Private key committed to code version control
    error: This violation immediately blocks your code deployment
    tags: KEY
    type: scan
    fatal: true
    enforcement: true
    environment: all
    confidence: high
    patterns:
      - \s*(-----BEGIN PRIVATE KEY-----)
      - \s*(-----BEGIN RSA PRIVATE KEY-----)
      - \s*(-----BEGIN DSA PRIVATE KEY-----)
      - \s*(-----BEGIN EC PRIVATE KEY-----)
      - \s*(-----BEGIN OPENSSH PRIVATE KEY-----)
      - \s*(-----BEGIN PGP PRIVATE KEY BLOCK-----)

  - name: Collect sparse TF resources outside of modules.
    id: 200
    description: The following resources were detected outside of compliant module usage
    type: collect
    tags: AWS,AZURE
    patterns:
      - (resource)\s*"(.*)"

  - name: ASSURE SSL
    id: 301
    description: Assure ssl_cyphers only contains GANSO_SSL
    error: Misconfiguration or omission is fatal
    tags: KEY
    type: assure
    fatal: true
    enforcement: true
    environment: all
    confidence: high
    patterns:
      - ssl_cyphers\s*=\s*"GANSO_SSL"

  - name: Check API config endpoint
    id: 405
    description: Ensure anonymous access is turned OFF
    error: Misconfiguration or omission
    tags: KEY
    type: api
    api_endpoint: https://localhost:3003/simple
    api_insecure: true
    api_request: POST
    api_body: <xml><rpc><show><config></config></show></rpc><xml>
    api_auth: basic
    api_auth_basic: RULE105
    api_auth_token:
    fatal: false
    enforcement: true
    environment: all
    confidence: high
    patterns:
      - \s*ANONYMOUS_ACCESS\s*:\s*OFF\s*

  - name: YML ASSURE Ingress enabled
    id: 506
    description: Assure ingress enabled
    error: Misconfiguration or omission is fatal
    tags: KEY
    type: yml
    fatal: true
    enforcement: true
    environment: all
    confidence: high
    yml_filepattern: "^development-\\d+\\.yml$"
    yml_structure: |
      ingress: {enabled: true}
      ...

ExitCritical: "Critical irregularities found in your code"
ExitWarning: "Irregularities found in your code"
ExitClean: "Clean report"

```
[Policy Schema](https://github.com/xfhg/intercept/tree/master/policy/schema.json) 

<br>
<br>

## Extra Configuration & Flags
all flags under the same instruction can be combined

<br>

- SHA256 Hash for configuration file 
```
intercept config -a examples/yourpolicies.yaml -k 201c8fe265808374f3354768410401216632b9f2f68f9b15c85403da75327396
```
- Download of configuration file
```
intercept config -a https://xxx.com/artifact/policy.yaml -k 201c8fe26580(...)
```
- Enviroment enforcement (check policy enforcement levels)
```
intercept audit -t examples/target -e "development"
```
- Rule Tag filter
```
intercept audit -t examples/target -i "AWS,OWASP"
```
- Run only the API type checks (but with tags)
```
intercept api -i "AWS,OWASP"
```
- TURBO SILENT mode
```
intercept audit -t examples/target -s true
```
- Run only the YML type checks
```
intercept yml -t examples/target 
```
- Disable pipeline break
```
intercept audit -t examples/target -b false
```
- No exceptions 
```
intercept audit -t examples/target -x
```
- Ignoring files and folders
```
use .ignore file
```
<br>
<br>
<br>


# Policy Enforcement Levels

Enforcement levels are a first class concept in allowing pass/fail behavior to be associated separately from the policy logic. This enables any policy to be a warning, allow exceptions, or be absolutely mandatory. These levels can be coupled to environments, different uses of the same policy can have different enforcement levels per environment.

You can set three enforcement levels:

- **Advisory**: The policy is allowed to fail. However, a warning will be shown to the user or logged.

```
  - fatal: false
  - enforcement: false
  - environment : (all | optional)
  - confidence : low | high
```

- **Soft Mandatory**: The policy must pass unless an exception is specified. The purpose of this level is to provide a level of privilege separation for a behavior. Additionally, the exception provides non-repudiation since at least the primary actor was explicitly overriding a failed policy.

```
  - fatal: true
  - enforcement: false
  - environment : (all | optional)
  - confidence : low | high
```

- **Hard Mandatory**: The policy must pass no matter what. The only way to override a hard mandatory policy is to explicitly remove the policy. It should be used in situations where an exception is not possible.

```
  - fatal: true
  - enforcement: true
  - environment : (all | optional)
  - confidence : high
```

<br>


# Sandbox Playground

Build & mess around with it :

```
make
```

<p align="center">

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/xfhg/intercept)

<br>
<br>
<br>



## Standing on the shoulders of giants - [ripgrep](https://github.com/BurntSushi/ripgrep) + [cue](https://github.com/cue-lang/cue)

- It is built on top of Rust's regex engine. Rust's regex engine uses finite automata, SIMD and aggressive literal optimizations to make searching very fast. (PCRE2 support)
- Rust's regex library maintains performance with full Unicode support by building UTF-8 decoding directly into its deterministic finite automaton engine.
It supports searching with either memory maps or by searching incrementally with an intermediate buffer. The former is better for single files and the latter is better for large directories. ripgrep chooses the best searching strategy for you automatically.
- Applies your ignore patterns in .gitignore files using a RegexSet. That means a single file path can be matched against multiple glob patterns simultaneously.
- It uses a lock-free parallel recursive directory iterator, courtesy of crossbeam and ignore.
- **same engine used on vscode search**
- **CUE** merges the notion of schema and data. The same CUE definition can simultaneously be used for validating data and act as a template to reduce boilerplate. Schema definition is enriched with fine-grained value definitions and default values. At the same time, data can be simplified by removing values implied by such detailed definitions. The merging of these two concepts enables many tasks to be handled in a principled way.
- Constraints provide a simple and well-defined, yet powerful, alternative to inheritance, a common source of complexity with configuration languages.


<br>

| Tool                  | Command                                                 | Line count | Time       |
| --------------------- | ------------------------------------------------------- | ---------- | ---------- |
| **INTERCEPT (ripgrep)**     | `rg -n -w '[A-Z]+_SUSPEND'`                             | 452        | **0.106s** |
| git grep              | `LC_ALL=C git grep -E -n -w '[A-Z]+_SUSPEND'`           | 452        | 0.553s     |
| The Silver Searcher   | `ag -w '[A-Z]+_SUSPEND'`                                | 452        | 0.589s     |
| git grep (Unicode)    | `LC_ALL=en_US.UTF-8 git grep -E -n -w '[A-Z]+_SUSPEND'` | 452        | 2.266s     |
| sift                  | `sift --git -n -w '[A-Z]+_SUSPEND'`                     | 452        | 3.505s     |
| ack                   | `ack -w '[A-Z]+_SUSPEND'`                               | 452       | 6.823s     |
| The Platinum Searcher | `pt -w -e '[A-Z]+_SUSPEND'`                             | 452        | 14.208s    |



Timings were collected on a system with an Intel
i7-6900K 3.2 GHz.
A straight-up comparison between ripgrep, ugrep and GNU grep on a single large file cached in memory (**~13GB**, [`OpenSubtitles.raw.en.gz`](http://opus.nlpl.eu/download.php?f=OpenSubtitles/v2018/mono/OpenSubtitles.raw.en.gz)):

| Tool | Command | Line count | Time |
| ---- | ------- | ---------- | ---- |
| **ripgrep** | `rg -w 'Sherlock [A-Z]\w+'` | 7882 | **2.769s** |
| [ugrep](https://github.com/Genivia/ugrep) | `ugrep -w 'Sherlock [A-Z]\w+'` | 7882 | 6.802s |
| [GNU grep](https://www.gnu.org/software/grep/) | `LC_ALL=en_US.UTF-8 egrep -w 'Sherlock [A-Z]\w+'` | 7882 | 9.027s |

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
<br>
<br>
<br>


## Patterns optimized with

<p>

<img src="static/openai.svg" width="275">

</p>
<br>
<br>

## Licensing & Compliance

<br>
<br>


[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fxfhg%2Fintercept.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fxfhg%2Fintercept?ref=badge_large)
