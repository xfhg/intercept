<p align="center">

<img src="static/intercept.png" width="250">

</p>

# INTERCEPT v1 _ PRE-RELEASE

Intercept is a security tool designed to provide Static Application Security Testing (SAST) capabilities to software development teams. The tool aims to help developers identify and address security vulnerabilities in their code early in the software development life cycle, reducing the risk of security breaches and ensuring compliance with industry regulations. Intercept leverages a range of security scanning techniques to analyze code, including pattern matching, code analysis, and vulnerability scanning. The tool is designed to be easy to use, with a simple sub-second command-line interface and customizable configuration options. With Intercept, developers can integrate security testing into their development workflows and make security a seamless part of their software development process.

<br>


![GitHub release (latest by date)](https://img.shields.io/github/v/release/xfhg/intercept)
![GitHub Release Date](https://img.shields.io/github/release-date/xfhg/intercept)
![GitHub last commit](https://img.shields.io/github/last-commit/xfhg/intercept)
![GitHub commits since latest release (by date)](https://img.shields.io/github/commits-since/xfhg/intercept/latest)

![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/xfhg/intercept)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/xfhg/intercept)
![GitHub issues](https://img.shields.io/github/issues-raw/xfhg/intercept)
![GitHub pull requests](https://img.shields.io/github/issues-pr-raw/xfhg/intercept)


## Features


- Pattern matching: Intercept uses a pattern matching technique to scan code for known vulnerabilities, such as SQL injection and cross-site scripting (XSS), reducing the time and effort required to identify and fix these common issues.
- Customizable rules: Intercept allows users to customize the security rules used to scan their code, making it possible to tailor the scanning process to the specific requirements of their application or organization.
- Integration with CI/CD: Intercept can be integrated into continuous integration and continuous deployment (CI/CD) pipelines, allowing security testing to be performed automatically as part of the development process.
- Detailed reporting: Intercept provides detailed reports on vulnerabilities and security issues, including severity ratings and remediation advice, making it easy for developers to prioritize and address security concerns.
- Support for multiple programming languages: Intercept supports scanning for vulnerabilities in multiple programming languages, including Java, Python, Ruby, and Go, making it a versatile tool for security testing across a range of applications and environments.

## Policy as Code

Policy as code is an approach to defining and enforcing policies within an organization using code. Instead of writing policies in documents or spreadsheets, policy as code involves writing policies as code using a programming language. This code is then integrated into the organization's infrastructure, software, or workflow to enforce the policies automatically.

Main benefits:

- Consistency: Policies are enforced consistently across all systems and applications, reducing the risk of errors and vulnerabilities.
- Automation: Policies can be enforced automatically, reducing the need for manual intervention and improving efficiency.
- Transparency: Policies can be reviewed and audited more easily, providing greater transparency into how policies are being enforced.
- Flexibility: Policies can be updated and changed more easily, allowing organizations to adapt to changing requirements and regulations.

Policy as code can be used to enforce a wide range of policies, including security policies, compliance policies, and operational policies. It is often used in conjunction with infrastructure as code and other DevOps practices to provide a more automated and streamlined approach to managing IT operations.

## Secret Scanning

Intercept offers an extensive library of policies consisting of over a thousand regular expressions that can be used to detect sensitive data leakage and enforce security best practices in software development. This vast collection of pre-defined policies makes it easy for developers to get started with secret scanning and quickly identify potential issues in their code. The policies cover a range of security concerns, such as hard-coded passwords, API keys, and other secrets, and are continuously updated to keep up with the latest security threats and best practices. With the ability to customize policies or add new ones, developers can ensure that their applications are protected against known and emerging threats, reducing the risk of sensitive data leakage and improving the overall security posture of their organization.


## Work in progress

We are excited to announce that Intercept is now available in a pre-release version! Please note that this version is still being developed and may contain small bugs or other backward compatibility issues. However, we are committed to delivering a top-quality security testing tool and are working hard to complete a complete overhaul of the documentation, integration, and feature set. In the coming weeks, we will be adding new features and updating the tool with the latest security policies to ensure that your applications are protected against known and emerging threats. We appreciate your patience and support as we work to deliver the best possible security testing tool.

## Playground

Mess around with it :

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/xfhg/intercept)



