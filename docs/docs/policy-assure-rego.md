# ASSURE-REGO Type Policies

ASSURE-type policies serve as proactive compliance tools within your codebase or configuration settings, distinctly contrasting with the reactive nature of SCAN-type policies. 

Instead of searching for known issues or vulnerabilities, ASSURE policies help in affirming the presence of specific, desirable patterns within your target codebase. These patterns are pivotal in ensuring that your codebase adheres to compliance standards or configuration requirements.

This approach shifts the focus from merely identifying and rectifying problems to actively validating and ensuring the desired state of system configurations, thereby fostering a more secure and compliant infrastructure.


::: tip
check the /examples/rego for details
:::


## REGO

Leveraging Rego, the policy language of the Open Policy Agent (OPA), for defining ASSURE-type policy patterns, furnishes organizations with an advanced toolkit for crafting and enforcing compliance and security policies across their services and infrastructure. Rego, with its unique design tailored for expressing policies in complex and dynamic environments, brings a myriad of features and benefits for managing compliance requirements:

- Expressiveness and Flexibility

Rego's high level of expressiveness allows for the articulation of intricate and nuanced policies that can evaluate and enforce compliance across diverse aspects of the infrastructure and codebase. This flexibility is crucial for defining ASSURE-type policies, where the presence of specific patterns or configurations needs to be validated against a broad spectrum of criteria.

- Declarative Approach

The declarative nature of Rego enables the straightforward specification of "what" should be enforced without prescribing "how" to enforce it. This approach simplifies the development of policies by focusing on the desired state or outcome, rather than the procedural steps to achieve it. As a result, policies are more maintainable and easier to understand and audit.

- Context-Aware Decision Making

Rego policies can seamlessly integrate with external data sources, allowing policies to be context-aware and make decisions based on the most current data. For ASSURE-type policies, this means the ability to validate configurations against dynamic conditions or external benchmarks, ensuring compliance is not only based on static rules but is also responsive to evolving contexts.


- Native Integration with Cloud-Native Ecosystems

Rego's integration with the cloud-native ecosystem, including Kubernetes, Terraform, and various cloud service providers, enables the direct enforcement of ASSURE-type policies within the deployment pipeline. This integration facilitates real-time compliance checks, preventing non-compliant configurations from being deployed.

- Continuous Compliance and Auditability

The ability of Rego to integrate into CI/CD pipelines and runtime environments supports continuous compliance monitoring and enforcement. Changes to the codebase or configurations can be automatically evaluated against ASSURE-type policies, ensuring ongoing compliance. Additionally, Rego policies can be used to generate detailed audit logs, providing evidence of compliance for internal audits or regulatory purposes.

- Fine-Grained Control with Minimal Performance Overhead

Rego's efficient evaluation engine ensures that compliance checks can be performed with minimal performance overhead, critical for high-availability systems. This efficiency, coupled with the language's support for fine-grained control over policy enforcement, means organizations can implement comprehensive compliance checks without compromising on system performance.

## Example

### The setup

```sh{2,4}
intercept config -r 
intercept config -a /app/examples/policy/rego.yaml

intercept rego -t /app/examples/target/rego 

cat intercept.rego.sarif.json
```
::: info
All rule types can be filtered by a combination of TAGS, ENVIRONMENT name and their own ENFORCEMENT levels. Make sure to explore it.
:::


### The Policy

```yaml{7,12-16}

  - name: REGO COMPLEX With DATA
    id: 901
    description: Rego Validation with external data
    error: This violation immediately blocks your code deployment
    tags: REGO
    type: rego
    fatal: true
    enforcement: true
    environment: all
    confidence: high
    rego_filepattern: '^input_test.json$'
    rego_policy_file: /app/examples/data/complex.rego
    rego_policy_data: /app/examples/data/policy_data.json
    rego_policy_query: |
      data.app.abac.result

```



### SARIF Output 

```json{8,10,24}
{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    ... REDACTED
      "results": [
        {
          "ruleId": "intercept.cc.rego.policy.901: REGO COMPLEX WITH DATA",
          "ruleIndex": 0,
          "level": "note",
          "message": {
            "text": "Rego Validation with external data"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "/app/examples/target/rego/input_test.json"
                },
                "region": {
                  "startLine": 0,
                  "endLine": 0,
                  "snippet": {
                    "text": "COMPLIANT"
                  }
                }
              }
            }
          ]
    ... REDACTED
```

### Console Output

```sh{8,9,10}
│ 
├ REGO Rule # 901
│ Rule name :  REGO COMPLEX With DATA
│ Rule description :  Rego Validation with external data
│ Impacted Env :  all
│ Confidence :  high
│ Tags :  REGO
│ REGO Input File Pattern :  ^input_test.json$
│ REGO Policy :  /app/examples/data/complex.rego
│ REGO Query :  data.app.abac.result

│ 
│ Scanning..
│ File : /app/examples/target/rego/input_test.json
│ Hash : 5684dbfd47b5f9b6d3e2013285bba8b77576f6ab98925655d0c336f35c0235f9
│
│ 
│ REGO Tracer : 

Enter data.app.abac.result = _
| Eval data.app.abac.result = _
| Unify data.app.abac.result = _
| Index data.app.abac.result (matched 1 rule, early exit)
| Enter data.app.abac.result
| | Eval __local6__ = data.app.abac.controlstat
| | Unify __local6__ = data.app.abac.controlstat
| | Index data.app.abac.controlstat (matched 1 rule, early exit)
| | Enter data.app.abac.controlstat
| | | Eval data.app.abac.allow
| | | Unify data.app.abac.allow = _
| | | Index data.app.abac.allow (matched 4 rules, early exit)
| | | Enter data.app.abac.allow
| | | | Eval data.app.abac.user_is_owner
| | | | Unify data.app.abac.user_is_owner = _
| | | | Index data.app.abac.user_is_owner (matched 1 rule, early exit)
| | | | Enter data.app.abac.user_is_owner
| | | | | Eval __local0__ = input.user
| | | | | Unify __local0__ = input.user
| | | | | Unify "bob" = __local0__
| | | | | Eval data.user_attributes[__local0__].title = "owner"
| | | | | Unify data.user_attributes[__local0__].title = "owner"
| | | | | Unify "owner" = "employee"
| | | | | Fail data.user_attributes[__local0__].title = "owner"
| | | | | Redo __local0__ = input.user
| | | | Fail data.app.abac.user_is_owner
| | | Enter data.app.abac.allow
| | | | Eval data.app.abac.user_is_employee
| | | | Unify data.app.abac.user_is_employee = _
| | | | Index data.app.abac.user_is_employee (matched 1 rule, early exit)
| | | | Enter data.app.abac.user_is_employee
| | | | | Eval __local1__ = input.user
| | | | | Unify __local1__ = input.user
| | | | | Unify "bob" = __local1__
| | | | | Eval data.user_attributes[__local1__].title = "employee"
| | | | | Unify data.user_attributes[__local1__].title = "employee"
| | | | | Unify "employee" = "employee"
| | | | | Exit data.app.abac.user_is_employee early
| | | | Unify true = _
| | | | Eval data.app.abac.action_is_read
| | | | Unify data.app.abac.action_is_read = _
| | | | Index data.app.abac.action_is_read (matched 1 rule, early exit)
| | | | Enter data.app.abac.action_is_read
| | | | | Eval input.action = "read"
| | | | | Unify input.action = "read"
| | | | | Unify "read" = "read"
| | | | | Exit data.app.abac.action_is_read early
| | | | Unify true = _
| | | | Exit data.app.abac.allow early
| | | Unify true = _
| | | Exit data.app.abac.controlstat early
| | Unify 5 = __local6__
| | Eval gt(__local6__, 3)
| | Exit data.app.abac.result early
| Unify true = _
| Exit data.app.abac.result = _
Redo data.app.abac.result = _
| Redo data.app.abac.result = _
| Redo data.app.abac.result
| | Redo gt(__local6__, 3)
| | Redo __local6__ = data.app.abac.controlstat
| | | Redo data.app.abac.allow
| | | | Redo data.app.abac.action_is_read
| | | | | Redo input.action = "read"
| | | | Redo data.app.abac.user_is_employee
| | | | | Redo data.user_attributes[__local1__].title = "employee"
| | | | | Redo __local1__ = input.user

│ 
│ Query 0 : data.app.abac.result 
│ Result 0 : true
│ 
│ 
│ Compliant
│ 
├── ─
│
│

```


# Run it

```sh
docker pull ghcr.io/xfhg/intercept:latest

docker run -v --rm -w $PWD -v $PWD:$PWD -e TERM=xterm-256color ghcr.io/xfhg/intercept intercept config -a examples/policy/rego.yaml

docker run -v --rm -w $PWD -v $PWD:$PWD -e TERM=xterm-256color ghcr.io/xfhg/intercept intercept rego -t examples/target/rego
```