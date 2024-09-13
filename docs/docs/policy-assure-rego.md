# ASSURE REGO Policies

ASSURE-type policies serve as proactive compliance tools within your codebase or configuration settings, distinctly contrasting with the reactive nature of SCAN-type policies. 

Instead of searching for known issues or vulnerabilities, ASSURE policies help in affirming the presence of specific, desirable patterns within your target codebase. These patterns are pivotal in ensuring that your codebase adheres to compliance standards or configuration requirements.

This approach shifts the focus from merely identifying and rectifying problems to actively validating and ensuring the desired state of system configurations, thereby fostering a more secure and compliant infrastructure.




## REGO

Leveraging Rego, the policy language of the Open Policy Agent (OPA), for defining ASSURE-type policy patterns, furnishes organizations with an advanced toolkit for crafting and enforcing compliance and security policies across their services and infrastructure. Rego, with its unique design tailored for expressing policies in complex and dynamic environments, brings a myriad of features and benefits for managing compliance requirements:

### Expressiveness and Flexibility

Rego's high level of expressiveness allows for the articulation of intricate and nuanced policies that can evaluate and enforce compliance across diverse aspects of the infrastructure and codebase. This flexibility is crucial for defining ASSURE-type policies, where the presence of specific patterns or configurations needs to be validated against a broad spectrum of criteria.

### Declarative Approach

The declarative nature of Rego enables the straightforward specification of "what" should be enforced without prescribing "how" to enforce it. This approach simplifies the development of policies by focusing on the desired state or outcome, rather than the procedural steps to achieve it. As a result, policies are more maintainable and easier to understand and audit.

### Context-Aware Decision Making

Rego policies can seamlessly integrate with external data sources, allowing policies to be context-aware and make decisions based on the most current data. For ASSURE-type policies, this means the ability to validate configurations against dynamic conditions or external benchmarks, ensuring compliance is not only based on static rules but is also responsive to evolving contexts.


### Native Integration with Cloud-Native Ecosystems

Rego's integration with the cloud-native ecosystem, including Kubernetes, Terraform, and various cloud service providers, enables the direct enforcement of ASSURE-type policies within the deployment pipeline. This integration facilitates real-time compliance checks, preventing non-compliant configurations from being deployed.

### Continuous Compliance and Auditability

The ability of Rego to integrate into CI/CD pipelines and runtime environments supports continuous compliance monitoring and enforcement. Changes to the codebase or configurations can be automatically evaluated against ASSURE-type policies, ensuring ongoing compliance. Additionally, Rego policies can be used to generate detailed audit logs, providing evidence of compliance for internal audits or regulatory purposes.

### Fine-Grained Control with Minimal Performance Overhead

Rego's efficient evaluation engine ensures that compliance checks can be performed with minimal performance overhead, critical for high-availability systems. This efficiency, coupled with the language's support for fine-grained control over policy enforcement, means organizations can implement comprehensive compliance checks without compromising on system performance.


## Examples

```yaml{3-4,18-21}
Policies:
  - id: "REGO-001"
    type: "rego"
    filepattern: 'input_\w+\.json$'
    enforcement:
      - environment: "all"
        fatal: "false"
        exceptions: "development"
        confidence: "high"
    metadata:
      name: "REGO Schema Compliance"
      description: "Enforce schema compliance on JSON configuration file"
      tags:
        - "config"
        - "ini"
        - "schema"
      score: "8"
    _rego:
      policy_file: policies/rego/simple.rego
      policy_data: policies/rego/simple_policy_data.json
      policy_query: data.policies.allow
```

::: tip
check the /playground/policies/rego for example
:::

### policy_file
```rego
package policies

default allow = false

allow {
    input.user == "alice"
    input.action == "read"
}

allow {
    input.user == "bob"
    input.action == "write"
}
```
### policy_data

```json
{
    "user_roles": {
      "alice": ["reader", "commenter"],
      "bob": ["writer", "editor"],
      "charlie": ["viewer"]
    },
    "action_permissions": {
      "read": ["reader", "writer", "editor"],
      "write": ["writer", "editor"],
      "comment": ["commenter", "writer", "editor"],
      "view": ["viewer", "reader", "commenter", "writer", "editor"]
    }
  }
  
```