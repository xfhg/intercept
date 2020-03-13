[![Go Report Card](https://goreportcard.com/badge/github.com/xfhg/intercept)](https://goreportcard.com/report/github.com/xfhg/intercept)

<p align="center">

<img src="static/intercept.png" width="250">

</p>

# INTERCEPT

Easy to configure, small footprint Policy as Code command-line scanner that leverages the power of the fastest multi-line search tool to scan your codebase. It can be used as a linter, guard rail control or simple data collector and inspector. Consider it a weaponized ripgrep.

## How it works 

Consumes policy files with single or grouped rules ( regex patterns ) that will be scanned recursively against code (text) in a target path and can be customized/filtered to environments, exceptions, and criticality of the findings. Generates detailed human-readable reports.

## Use cases

- simple and powerful free alternative for Hashicorp Sentinel if you are more comfortable writing and maintaining regular expressions than a new custom policy language.

- adds the same functionality as git-secrets and can prevent sensitive information to run through your pipeline.

- identifies the policy breach (files and line numbers) and can report solutions/suggestions to its findings making it a great tool to ease onboarding developer teams to your unified deployment pipeline.

- can enforce style-guides, coding-standards, best practices and also report on suboptimal configurations.

- can collect patterns or high entropy data and output it in multiple formats.

- anything you can crunch on a regular expression can be actioned on.

```

```

### Why ripgrep? Why is it fast?

- It is built on top of Rust's regex engine. Rust's regex engine uses finite automata, SIMD and aggressive literal optimizations to make searching very fast. (PCRE2 support)
Rust's regex library maintains performance with full Unicode support by building UTF-8 decoding directly into its deterministic finite automaton engine.

- It supports searching with either memory maps or by searching incrementally with an intermediate buffer. The former is better for single files and the latter is better for large directories. ripgrep chooses the best searching strategy for you automatically.

- Applies ignore patterns in .gitignore files using a RegexSet. That means a single file path can be matched against multiple glob patterns simultaneously.

- It uses a lock-free parallel recursive directory iterator, courtesy of crossbeam and ignore libraries.

```

```

### Benchmark

| Tool | Command | Line count | Time |
| ---- | ------- | ---------- | ---- |
| ripgrep (Unicode) | `rg -n -w '[A-Z]+_SUSPEND'` | 450 | **0.106s** |
| [git grep](https://www.kernel.org/pub/software/scm/git/docs/git-grep.html) | `LC_ALL=C git grep -E -n -w '[A-Z]+_SUSPEND'` | 450 | 0.553s |
| [The Silver Searcher](https://github.com/ggreer/the_silver_searcher) | `ag -w '[A-Z]+_SUSPEND'` | 450 | 0.589s |
| [git grep (Unicode)](https://www.kernel.org/pub/software/scm/git/docs/git-grep.html) | `LC_ALL=en_US.UTF-8 git grep -E -n -w '[A-Z]+_SUSPEND'` | 450 | 2.266s |
| [sift](https://github.com/svent/sift) | `sift --git -n -w '[A-Z]+_SUSPEND'` | 450 | 3.505s |
| [ack](https://github.com/beyondgrep/ack2) | `ack -w '[A-Z]+_SUSPEND'` | 1878 | 6.823s |
| [The Platinum Searcher](https://github.com/monochromegane/the_platinum_searcher) | `pt -w -e '[A-Z]+_SUSPEND'` | 450 | 14.208s |

---

## Simple Example

```

```

## Complex Example

```

```

## Full Feature Example

```

```

## TODO

- full test suite
