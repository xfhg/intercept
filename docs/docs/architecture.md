
# Architecture

<br><br>

INTERCEPT provides an industry-standard, policy-based approach to security and compliance. It offers comprehensive capabilities to scan, analyze, and patch various types of files, systems, services, hardenings, setups, and endpoints. Intercept can operate in ad-hoc, real-time, or event-triggered modes.

---

<p align="center">
  <img alt="Intercept Arch" src="/arch.png">
</p>

---

To accomplish a thorough scan of your codebase or system state, the INTERCEPT Policy Engine allows you to define a rich multitude of contextualized policy types that will fit most, if not all, of your compliance and security requirements.


<br><br>


::: tip
Discover how INTERCEPT can transform your security operations and provide unparalleled visibility into your organization's compliance status. Check the SANDBOX for a 10 step approach.
:::


<br><br><br>

# 🔋 Batteries Included

## 2 Operational Modes

### INTERCEPT AUDIT

The AUDIT mode in INTERCEPT is a comprehensive scanning and analysis process that evaluates your entire codebase, infrastructure, or system state against predefined security and compliance policies. This mode operates as follows:

1. Policy Loading: INTERCEPT loads all defined policies from the policy file.

2. Target Scanning: It scans the specified target directory or system, collecting relevant information for policy evaluation.

3. Parallel Processing: INTERCEPT processes all applicable policies in parallel, optimizing performance and reducing overall audit time.

4. Comprehensive Evaluation: Each policy (SCAN, ASSURE, RUNTIME, REGO, API) is evaluated against the collected data.

5. Result Aggregation: Results from all policy evaluations are aggregated into a detailed SARIF report.

6. Output Generation: INTERCEPT generates detailed output in various formats (e.g., JSON, SARIF) based on configuration.

7. Webhook Integration: If configured, audit results can be sent to specified endpoints for further processing or notification.

The AUDIT mode provides a point-in-time assessment of your system's compliance status, making it ideal for periodic checks, pre-deployment verifications, or on-demand security audits. It offers a holistic view of your system's security posture and compliance adherence.

### INTERCEPT OBSERVE

The OBSERVE mode in INTERCEPT is a continuous monitoring and real-time analysis process that provides ongoing surveillance of your system's security and compliance status. This mode operates as follows:

1. Policy Selection: INTERCEPT loads and applies policies marked for continuous observation (either by schedule or path monitoring).

2. Continuous Monitoring: It constantly watches specified targets (files, directories, APIs, or system states) for changes.

3. Real-time Evaluation: When changes are detected, INTERCEPT immediately evaluates the affected areas against relevant policies.

4. Instant Alerting: If any non-compliance is detected, INTERCEPT can trigger immediate alerts or actions (through webhooks).

5. Incremental Processing: Only changed components are re-evaluated, ensuring efficient use of resources.

6. Live Reporting: INTERCEPT provides real-time updates on the system's compliance status.

The OBSERVE mode offers continuous, real-time protection and compliance monitoring, making it ideal for production environments, critical systems, or any scenario where immediate detection of and response to compliance violations is crucial. It provides an always-on, vigilant approach to maintaining your system's security and compliance posture.

---

#### Both modes can be used together. Improved integration with CI/CD to ensure compliance at every stage of development and deployment and extend your compliance assurance into runtime context.




## Policy Types

### SCAN Policies
Scan policies enable thorough examination of non-binary files for regex patterns. These policies are crucial for identifying potential security risks such as leaked or hardcoded API keys, SSL certificates, passwords, and authentication tokens. A compliant state for a scan policy is achieved when the defined patterns are not found within the specified target path.
### ASSURE Policies
Functioning inversely to scan policies, assurance policies enforce the presence of defined patterns. They utilize regex, CUE Lang schemas, or a combination of both. These policies are ideal for validating configuration files, log streams, and audit logs against expected patterns or structures. Compliance is achieved when the target matches the specified patterns, schemas, or values.
#### Configuration File Type Policies (JSON, YAML, TOML, INI)
For monitored files with specific target types,  can generate patches to bring non-compliant files back into compliance.
#### Endpoint API Policies
API policies apply assurance policy principles to API endpoints, ensuring they meet defined standards and expectations.
### RUNTIME Policies
Utilizing a YAML-based serverspec toolkit, runtime policies validate server configurations and real-time system states. They are essential for monitoring services and configuration states, enabling immediate response to changes and drifts in compliance.
### REGO Policies
Leveraging the Open Policy Agent (OPA) engine, Rego policies assess compliance in complex scenarios. They excel in contextualized compliance checks where dynamic input data is necessary for accurate compliance status calculation.

## Multiplatform Single Binary

Low footprint, works everywhere.

## Integration Webhooks

Send your Compliance Reports immediatly to the right stakeholders, multiple levels of detail are provided out of the box.


## Platform compatibility

| Platform | Release | Audit Feature Set | Runtime Feature Set |
|----------|-------------| ---- |---- |
| Linux (x86_64) | intercept-linux-amd64 | Full | Full |
| Linux (ARMv7) | intercept-linux-arm |  Full | Full |
| Linux (ARM64) | intercept-linux-arm64 |  Full | Full |
| macOS (Intel) | intercept-darwin-amd64 | Full |Limited |
| macOS (Apple Silicon) | intercept-darwin-arm64 | Full |Limited |
| Windows (x86_64) | intercept-windows-amd64.exe | Most |Limited |