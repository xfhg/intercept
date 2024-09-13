# SCAN Policies

SCAN-type policies are designed to identify and flag patterns or a collection of known patterns within your target codebase that should ideally be absent. These undesirable patterns may include, but are not limited to, exposed API keys, secrets, overly permissive CIDR ranges in security groups, improperly enabled configuration parameters, or even instances of code, script, or proxy definition misuse that could pose risks to your environment.

This proactive approach allows developers and security professionals to systematically weed out configurations or code snippets that contradict best practices or security guidelines, thereby enhancing the overall security posture and compliance of the codebase.

With an extensive library of over 1,500 predefined patterns available for selection, our tool offers a comprehensive means to safeguard your codebase against common pitfalls and security vulnerabilities. To illustrate, consider the following simplified example:


## Examples

```yaml{3,22-28}
Policies:
  - id: "SCAN-001 Private Keys"
    type: "scan"
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
      name: "Detect private keys"
      description: "Scan for potential private key leaks in the codebase"
      msg_solution: "Remove the private key and use secure key management practices."
      msg_error: "Private key detected in the codebase."
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

::: tip
Group similar patterns per policy to improve report readability
:::

