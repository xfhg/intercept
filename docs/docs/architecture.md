
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

Send your Compliance Reports immediatly to the right stakeholders


## Platform compatibility

| Platform | Release | Audit Feature Set | Runtime Feature Set |
|----------|-------------| ---- |---- |
| Linux (x86_64) | intercept-linux-amd64 | Full | Full |
| Linux (ARMv7) | intercept-linux-arm |  Full | Full |
| Linux (ARM64) | intercept-linux-arm64 |  Full | Full |
| macOS (Intel) | intercept-darwin-amd64 | Full |Limited |
| macOS (Apple Silicon) | intercept-darwin-arm64 | Full |Limited |
| Windows (x86_64) | intercept-windows-amd64.exe | Limited |Limited |