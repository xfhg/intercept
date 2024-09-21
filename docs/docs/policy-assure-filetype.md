# ASSURE FILETYPE Policies

ASSURE-type policies serve as proactive compliance tools within your codebase or configuration settings, distinctly contrasting with the reactive nature of SCAN-type policies. 

Instead of searching for known issues or vulnerabilities, ASSURE policies help in affirming the presence of specific, desirable patterns within your target codebase. These patterns are pivotal in ensuring that your codebase adheres to compliance standards or configuration requirements.

This approach shifts the focus from merely identifying and rectifying problems to actively validating and ensuring the desired state of system configurations, thereby fostering a more secure and compliant infrastructure.



## CUE Lang

Defining the patterns for ASSURE-type policies using CUE Language (CUE Lang) schemas introduces a new level of precision, flexibility, and efficiency in policy specification. CUE Lang, a constraint-based configuration language, offers several distinct advantages for defining these patterns:

### Enhanced Precision and Clarity

CUE Lang's schema-based approach allows for the detailed specification of patterns with clear, unambiguous definitions. This precision ensures that policies are applied consistently, reducing the risk of misinterpretation or errors in policy enforcement. By explicitly outlining the structure and constraints of desired configurations, CUE Lang schemas make it easier to define complex compliance requirements in a way that is both human-readable and machine-enforceable.

### Reusability and Modularity

CUE Lang promotes reusability and modularity through its schema definitions. Patterns defined for one project can be easily adapted or extended for use in others, without the need for duplication of effort. This modularity not only speeds up the policy creation process but also ensures consistency across different projects and teams within an organization.

### Scalability and Evolution

As organizational needs evolve, so too can the ASSURE policies defined using CUE Lang. The language's flexible schema definitions enable policies to be updated or expanded with minimal effort, ensuring that compliance standards can keep pace with changing requirements, technologies, and industry best practices.

### Enhanced Error Detection and Correction

CUE Lang's constraint-based approach not only helps in defining what is correct but also aids in identifying and correcting deviations from the defined policies. When applied to ASSURE-type policies, this means that any discrepancies or omissions in the target codebase or configurations can be precisely pinpointed and addressed, enhancing the overall security and compliance posture.



## Examples

::: tip
for the **type** key you can select between **yml**,**json**,**ini**,**toml** 
:::

```yaml{3,4,20-23}
Policies:
  - id: "YAML-001"
    type: "yml"
    filepattern: "\\.ya?ml$"
    enforcement:
      - environment: "all"
        fatal: "false"
        exceptions: "development"
    metadata:
      name: "YAML Schema Compliance"
      description: "Enforce schema compliance on YAML configuration files"
      msg_solution: "Ensure all required fields are present and comply with the schema."
      msg_error: "YAML file does not comply with the required schema."
      tags:
        - "config"
        - "yaml"
        - "schema"
      confidence: "high"
      score: "8"
    _schema:
      strict: false
      patch: false
      structure: |
          service: {
            name:     string
            replicas: string | int
            ports: [...{
              port:       int
              targetPort: int
            }]
          }
          environment: [...{
            name:  string
            value: string | int
          }]
          resources: {
            limits: {
              cpu:    string
              memory: string
            }
            requests: {
              cpu:    "600m"
              memory: string
            }
          }

  # Same ASSURE TYPE but for JSON

  - id: "JSON-002"
    type: "json"
    filepattern: "development-2048.json$"
    enforcement:
      - environment: "all"
        fatal: "false"
        exceptions: "development"
    metadata:
      name: "JSON Schema Compliance"
      description: "Enforce schema compliance on JSON configuration files"
      tags:
        - "config"
        - "json"
        - "schema"
      confidence: "high"
      score: "8"
    _schema:
      strict: false
      patch: false
      structure: |
        {
          autoscaling: {enabled: true}
        }

```
::: tip
**strict** key implies that the file should meet the structure defined in the pattern. **patch** will trigger the creation of a patch if your defined structure provides the expected values (and not only the types) for the target files
:::

